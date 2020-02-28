package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// pipeline module for ffmpeg

// FFMpeg -
type FFMpeg struct{}

// Name -
func (e *FFMpeg) Name() string {
	return "FFMpeg"
}

// Command -
func (e *FFMpeg) Command() string {
	return "ffmpeg"
}

// Execute -
func (e *FFMpeg) Execute(env environment.Environment, cmds []string, tracker pipeline.JobTracker) error {
	for _, cmd := range cmds {
		tracker.Logger().Info("probing input files")
		file, duration, err := e.estimateDuration(env, cmd, tracker)
		if err != nil {
			tracker.Logger().Crit("unable to estimate file duration")
			return err
		}
		tracker.Logger().Info(fmt.Sprintf("input file `%s` is %f seconds long", file, duration))

		// run ffmpeg and track progress
		tracker.Logger().Info("executing ffmpeg `" + cmd + "`")
		result, err := env.Execute(
			append([]string{}, "/bin/sh", "-c", "ffmpeg -v warning -progress /dev/stdout "+cmd),
			func(outmsg string) {
				tracker.TrackStdOut(outmsg)
				s := strings.Split(outmsg, "=")
				if len(s) == 2 && s[0] == "out_time_us" {
					time, err := strconv.Atoi(s[1])
					if err == nil {
						progress := float64(time) / (duration * 1000.0 * 1000.0)
						tracker.TrackProgress(uint(progress * 100.0))
					}
				}
			},
			func(errmsg string) {
				tracker.TrackStdErr(errmsg)
			})
		if err != nil {
			tracker.Logger().Crit("execution of ffmpeg failed")
			return err
		}

		if result.ExitCode == 0 {
			tracker.Logger().Info("ffmpeg transcode finished")
		} else {
			// TODO: handle error
			return errors.New("unable to transcode video")
		}
	}

	return nil
}

func (e *FFMpeg) estimateDuration(env environment.Environment, cmd string, tracker pipeline.JobTracker) (string, float64, error) {
	// parse argument list and figure out the input file(s)
	var opts struct {
		Input string `short:"i" long:"input"`
		// TODO: figure out shorted flag, -t, etc
	}
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(strings.Split(cmd, " "))
	if err != nil {
		tracker.Logger().Crit("unable to parse input command line `" + cmd + "`")
		return "", 0, err
	}

	// probe inputs
	probeResult, err := env.Execute(
		append([]string{}, "/bin/sh", "-c", "ffprobe -v quiet -print_format json -show_format -show_streams -i "+opts.Input), nil, nil)
	if err != nil {
		tracker.Logger().Crit("unable to execute ffprobe")
		return "", 0, err
	}

	var probe interface{}
	err = json.Unmarshal([]byte(probeResult.StdOut), &probe)
	if err != nil {
		tracker.Logger().Crit("unable to unmarshal ffprobe result")
		return "", 0, err
	}

	format, ok := probe.(map[string]interface{})["format"]
	if !ok {
		tracker.Logger().Crit("could not locate `format`in ffprobe result")
		return "", 0, errors.New("unable to parse ffprobe result")
	}

	durationStr, ok := format.(map[string]interface{})["duration"].(string)
	if !ok {
		tracker.Logger().Crit("could not locate `dration`in ffprobe result")
		return "", 0, errors.New("unable to parse ffprobe result")
	}

	duration, err := strconv.ParseFloat(durationStr, 32)
	if err != nil {
		tracker.Logger().Crit("could not parse duration `" + durationStr + "` as number in ffprobe result")
		return "", 0, errors.New("unable to parse ffprobe result")
	}

	return opts.Input, duration, nil
}
