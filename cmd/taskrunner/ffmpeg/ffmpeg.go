package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"

	log "github.com/inconshreveable/log15"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"

	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/ko1N/dips/pkg/execution/tracking"
	"github.com/ko1N/dips/pkg/taskstorage"
)

type Config struct {
	Dips   DipsConfig    `yaml:"dips"`
	FFmpeg *FFmpegConfig `yaml:"ffmpeg"`
}

type DipsConfig struct {
	Host string `yaml:"host"`
}

type FFmpegConfig struct {
	FFprobeExecutable string `yaml:"ffprobe"`
	FFmpegExecutable  string `yaml:"ffmpeg"`
}

func readConfig(filename string) (*Config, error) {
	fallback := Config{
		Dips: DipsConfig{
			Host: "rabbitmq:rabbitmq@localhost",
		},
		FFmpeg: &FFmpegConfig{
			FFprobeExecutable: "/usr/bin/ffprobe",
			FFmpegExecutable:  "/usr/bin/ffmpeg",
		},
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return &fallback, nil
	}

	var conf Config
	err = yaml.Unmarshal([]byte(contents), &conf)
	if err != nil {
		return &fallback, nil
	}
	return &conf, nil
}

func main() {
	conf, err := readConfig("config.yml")
	if err != nil {
		panic(err)
	}

	cl, err := dipscl.NewClient(conf.Dips.Host)
	if err != nil {
		panic(err)
	}

	cl.
		NewTaskWorker("ffprobe").
		// TODO: task timeout??
		Concurrency(10).
		//Environment("shell").
		Filesystem("disk").
		Handler(ffprobeHandler(conf.FFmpeg)).
		Run()

	cl.
		NewTaskWorker("ffmpeg").
		// TODO: task timeout??
		Concurrency(10).
		//Environment("shell").
		Filesystem("disk").
		Handler(ffmpegHandler(conf.FFmpeg)).
		Run()

	fmt.Println("ffmpeg worker started")

	signal := make(chan struct{})
	<-signal
}

func ffprobeHandler(conf *FFmpegConfig) func(*dipscl.TaskContext) (map[string]interface{}, error) {
	return func(task *dipscl.TaskContext) (map[string]interface{}, error) {
		tracker := tracking.CreateTaskTracker(
			log.New("cmd", "ffmpeg"),
			task.Client,
			task.Request.Job.Id.Hex(),
			task.Request.TaskID)

		// input video
		source := task.Request.Params["source"]
		if source == "" {
			return nil, fmt.Errorf("`source` variable must not be empty")
		}

		url, err := taskstorage.ParseFileUrl(source)
		if err != nil {
			return nil, fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
		}

		err = task.Filesystem.AddInput(url)
		if err != nil {
			return nil, fmt.Errorf("unable to add input file '%s': %s", url.URL.String(), err.Error())
		}

		// ffprobe
		probe, err := executeFFmpegProbe(task, conf, &tracker, url.FilePath)
		if err != nil {
			return nil, fmt.Errorf("ffprobe failed: %s", err.Error())
		}

		tracker.Info("ffmpeg-probe successful")
		return map[string]interface{}{
			"probe": probe,
		}, nil
	}
}

func executeFFmpegProbe(task *dipscl.TaskContext, conf *FFmpegConfig, tracker *tracking.JobTracker, filename string) (map[string]interface{}, error) {
	tracker.Info("probing input file: %s", filename)

	// probe inputs
	executable := "ffprobe"
	if conf != nil {
		executable = conf.FFprobeExecutable
	}
	cmdline := strings.Split(executable, " ")

	probeResult, err := task.Environment.Execute(
		cmdline[0], append(cmdline[1:], []string{"-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", "-i", filename}...),
		func(outmsg string) {
			// TODO: detect true ffmpeg errors
			tracker.StdOut(outmsg)
		},
		func(errmsg string) {
			// TODO: detect true ffmpeg errors
			tracker.StdErr(errmsg)
		})
	if err != nil {
		tracker.Crit("unable to execute ffprobe: %s", err)
		return nil, err
	}

	var probe interface{}
	err = json.Unmarshal([]byte(probeResult.StdOut), &probe)
	if err != nil {
		tracker.Crit("unable to unmarshal ffprobe result")
		return nil, err
	}

	probeMap := probe.(map[string]interface{})
	return probeMap, nil
}

type FFMpegArgs struct {
	Source string
	Target string
}

func ffmpegHandler(conf *FFmpegConfig) func(*dipscl.TaskContext) (map[string]interface{}, error) {
	return func(task *dipscl.TaskContext) (map[string]interface{}, error) {
		tracker := tracking.CreateTaskTracker(
			log.New("cmd", "ffmpeg"),
			task.Client,
			task.Request.Job.Id.Hex(),
			task.Request.TaskID)

		// inputs + outputs
		source := task.Request.Params["source"]
		if source == "" {
			return nil, fmt.Errorf("`source` variable must not be empty")
		}

		target := task.Request.Params["target"]
		if target == "" {
			return nil, fmt.Errorf("`target` variable must not be empty")
		}

		sourceUrl, err := taskstorage.ParseFileUrl(source)
		if err != nil {
			return nil, fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
		}

		targetUrl, err := taskstorage.ParseFileUrl(target)
		if err != nil {
			return nil, fmt.Errorf("unable to parse url in `target` variable: %s", err.Error())
		}

		// add input + output
		err = task.Filesystem.AddInput(sourceUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to add input file '%s': %s", sourceUrl.URL.String(), err.Error())
		}

		err = task.Filesystem.AddOutput(targetUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to add output file '%s': %s", targetUrl.URL.String(), err.Error())
		}

		// ffmpeg
		argopts := FFMpegArgs{
			Source: sourceUrl.FilePath,
			Target: targetUrl.FilePath,
		}
		argsstr := strings.Replace(strings.Replace(task.Request.Params["args"], "[", "{{.", -1), "]", "}}", -1)
		argtpl, err := template.New("args").Parse(argsstr)
		if err != nil {
			return nil, fmt.Errorf("invalid ffmpeg args: %s", err)
		}

		var args bytes.Buffer
		err = argtpl.Execute(&args, argopts)
		if err != nil {
			return nil, fmt.Errorf("malformed ffmpeg args")
		}

		// ffmpeg
		err = executeFFmpegTranscode(task, conf, &tracker, args.String())
		if err != nil {
			return nil, fmt.Errorf("ffmpeg failed: %s", err.Error())
		}

		tracker.Info("ffmpeg-transcode successful")
		return map[string]interface{}{
			"target": target,
		}, nil
	}
}

func executeFFmpegTranscode(task *dipscl.TaskContext, conf *FFmpegConfig, tracker *tracking.JobTracker, cmd string) error {
	tracker.Info("probing input files")
	duration, err := estimateDuration(task, conf, tracker, cmd)
	if err != nil {
		tracker.Crit("unable to estimate file duration")
		return err
	}

	// run ffmpeg and track progress
	// due to the nature of sending a custom command line
	// to the sub-process we want to run it in a seperate subshell
	// so commands are being executed properly
	tracker.Info("executing ffmpeg: %s", cmd)
	executable := "ffmpeg"
	if conf != nil {
		executable = conf.FFmpegExecutable
	}
	cmdline := strings.Split("/bin/sh -c", " ")

	result, err := task.Environment.Execute(
		cmdline[0], append(cmdline[1:], []string{executable + " -v warning -progress /dev/stdout " + cmd}...),
		func(outmsg string) {
			tracker.StdOut(outmsg)

			s := strings.Split(outmsg, "=")
			if len(s) == 2 && s[0] == "out_time_us" {
				time, err := strconv.Atoi(s[1])
				if err == nil {
					progress := float64(time) / (duration * 1000.0 * 1000.0)
					tracker.Progress(uint(progress * 100.0))
				}
			}
		},
		func(errmsg string) {
			tracker.StdErr(errmsg)
		})
	if err != nil {
		tracker.Crit("execution of ffmpeg failed")
		return err
	}

	if result.ExitCode == 0 {
		tracker.Progress(100)
	} else {
		// TODO: handle error
		return errors.New("unable to transcode video")
	}

	return nil
}

func estimateDuration(task *dipscl.TaskContext, conf *FFmpegConfig, tracker *tracking.JobTracker, cmd string) (float64, error) {
	// parse argument list and figure out the input file(s)
	var opts struct {
		Input string `short:"i" long:"input"`
		// TODO: handle shorted flag, -t, etc
	}
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(strings.Split(cmd, " "))
	if err != nil {
		tracker.Crit("unable to parse input command line `%s`", cmd)
		return 0, err
	}

	// probe inputs
	probe, err := executeFFmpegProbe(task, conf, tracker, opts.Input)
	if err != nil {
		tracker.Crit("unable to probe result", "error", err)
		return 0, err
	}

	format, ok := probe["format"]
	if !ok {
		tracker.Crit("could not locate `format`in ffprobe result")
		return 0, errors.New("unable to parse ffprobe result")
	}

	durationStr, ok := format.(map[string]interface{})["duration"].(string)
	if !ok {
		tracker.Crit("could not locate `dration`in ffprobe result")
		return 0, errors.New("unable to parse ffprobe result")
	}

	duration, err := strconv.ParseFloat(durationStr, 32)
	if err != nil {
		tracker.Crit("could not parse duration `" + durationStr + "` as number in ffprobe result")
		return 0, errors.New("unable to parse ffprobe result")
	}

	tracker.Info("input file length: %f", duration)
	return duration, nil
}
