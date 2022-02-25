package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/ko1N/dips/pkg/taskstorage"
)

// TODO: clientconfig

type FFmpegConfig struct {
	FFprobeExecutable string `yaml:"ffprobe"`
	FFmpegExecutable  string `yaml:"ffmpeg"`
}

/*
func ReadConfig(filename string) (*Config, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal([]byte(contents), &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
*/

func main() {
	cl, err := dipscl.NewClient("rabbitmq:rabbitmq@localhost")
	if err != nil {
		panic(err)
	}

	ffmpegConf := FFmpegConfig{
		FFprobeExecutable: "/usr/bin/ffprobe",
		FFmpegExecutable:  "/usr/bin/ffmpeg",
	}

	cl.
		NewTaskWorker("ffprobe").
		// TODO: task timeout??
		Concurrency(10).
		//Environment("shell").
		Filesystem("disk").
		Handler(ffprobeHandler(&ffmpegConf)).
		Run()

		/*
			cl.
				NewTaskWorker("ffmpeg").
				// TODO: task timeout??
				Concurrency(10).
				//Environment("shell").
				Filesystem("disk").
				Handler(ffmpegHandler).
				Run()
		*/

	fmt.Println("ffmpeg worker started")

	signal := make(chan struct{})
	<-signal
}

func ffprobeHandler(conf *FFmpegConfig) func(*dipscl.TaskContext) (map[string]interface{}, error) {
	return func(task *dipscl.TaskContext) (map[string]interface{}, error) {
		/*
			tracker := tracking.CreateJobTracker(&tracking.JobTrackerConfig{
				Logger: log.New("cmd", "worker"), // TODO: dep injection
				Client: job.Client,
				JobId:  job.Request.Job.Id.Hex(),
			})
		*/
		// TODO: CreateTaskTracker
		fmt.Printf("handling 'ffprobe' task %s: %s\n", task.Request.Name, task.Request.Params)

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
		probe, err := executeFFmpegProbe(task, conf, url.FilePath)
		if err != nil {
			return nil, fmt.Errorf("ffprobe failed: %s", err.Error())
		}

		//ctx.Tracker.Info("ffmpeg-probe successful")
		return map[string]interface{}{
			"probe": probe,
		}, nil
	}
}

func executeFFmpegProbe(task *dipscl.TaskContext, conf *FFmpegConfig, filename string) (map[string]interface{}, error) {
	//ctx.Tracker.Info("probing input file", "filename", filename)

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
			//ctx.Tracker.Info(outmsg, "stream", "stdout")
		},
		func(errmsg string) {
			// TODO: detect true ffmpeg errors
			//ctx.Tracker.Info(errmsg, "stream", "stderr")
		})
	if err != nil {
		//ctx.Tracker.Crit("unable to execute ffprobe", "error", err)
		return nil, err
	}

	var probe interface{}
	err = json.Unmarshal([]byte(probeResult.StdOut), &probe)
	if err != nil {
		//ctx.Tracker.Crit("unable to unmarshal ffprobe result")
		return nil, err
	}

	probeMap := probe.(map[string]interface{})
	return probeMap, nil
}
