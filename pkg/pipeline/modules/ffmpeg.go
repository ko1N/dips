package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/environments"
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

// StartPipeline -
func (e *FFMpeg) StartPipeline(ctx *pipeline.ExecutionContext) error {
	return nil
}

// FinishPipeline -
func (e *FFMpeg) FinishPipeline(ctx *pipeline.ExecutionContext) error {
	return nil
}

// Execute -
func (e *FFMpeg) Execute(ctx *pipeline.ExecutionContext, cmd string) (environments.ExecutionResult, error) {
	ctx.Tracker.Logger().Info("probing input files")
	file, duration, err := e.estimateDuration(ctx, cmd)
	if err != nil {
		ctx.Tracker.Logger().Crit("unable to estimate file duration")
		return environments.ExecutionResult{}, err
	}
	ctx.Tracker.Logger().Info(fmt.Sprintf("input file `%s` is %f seconds long", file, duration))

	// run ffmpeg and track progress
	ctx.Tracker.Logger().Info("executing ffmpeg `" + cmd + "`")
	result, err := ctx.Environment.Execute(
		append([]string{}, "/bin/sh", "-c", "ffmpeg -v warning -progress /dev/stdout "+cmd),
		func(outmsg string) {
			ctx.Tracker.StdOut(outmsg)
			s := strings.Split(outmsg, "=")
			if len(s) == 2 && s[0] == "out_time_us" {
				time, err := strconv.Atoi(s[1])
				if err == nil {
					progress := float64(time) / (duration * 1000.0 * 1000.0)
					ctx.Tracker.TrackProgress(uint(progress * 100.0))
				}
			}
		},
		func(errmsg string) {
			ctx.Tracker.StdErr(errmsg)
		})
	if err != nil {
		ctx.Tracker.Logger().Crit("execution of ffmpeg failed")
		return environments.ExecutionResult{}, err
	}

	if result.ExitCode == 0 {
		ctx.Tracker.Logger().Info("ffmpeg transcode finished")
	} else {
		// TODO: handle error
		return environments.ExecutionResult{}, errors.New("unable to transcode video")
	}

	return environments.ExecutionResult{}, nil
}

func (e *FFMpeg) estimateDuration(ctx *pipeline.ExecutionContext, cmd string) (string, float64, error) {
	// parse argument list and figure out the input file(s)
	var opts struct {
		Input string `short:"i" long:"input"`
		// TODO: figure out shorted flag, -t, etc
	}
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(strings.Split(cmd, " "))
	if err != nil {
		ctx.Tracker.Logger().Crit("unable to parse input command line `" + cmd + "`")
		return "", 0, err
	}

	// probe inputs
	probeResult, err := ctx.Environment.Execute(
		append([]string{}, "/bin/sh", "-c", "ffprobe -v quiet -print_format json -show_format -show_streams -i "+opts.Input), nil, nil)
	if err != nil {
		ctx.Tracker.Logger().Crit("unable to execute ffprobe")
		return "", 0, err
	}

	var probe interface{}
	err = json.Unmarshal([]byte(probeResult.StdOut), &probe)
	if err != nil {
		ctx.Tracker.Logger().Crit("unable to unmarshal ffprobe result")
		return "", 0, err
	}

	format, ok := probe.(map[string]interface{})["format"]
	if !ok {
		ctx.Tracker.Logger().Crit("could not locate `format`in ffprobe result")
		return "", 0, errors.New("unable to parse ffprobe result")
	}

	durationStr, ok := format.(map[string]interface{})["duration"].(string)
	if !ok {
		ctx.Tracker.Logger().Crit("could not locate `dration`in ffprobe result")
		return "", 0, errors.New("unable to parse ffprobe result")
	}

	duration, err := strconv.ParseFloat(durationStr, 32)
	if err != nil {
		ctx.Tracker.Logger().Crit("could not parse duration `" + durationStr + "` as number in ffprobe result")
		return "", 0, errors.New("unable to parse ffprobe result")
	}

	return opts.Input, duration, nil
}
