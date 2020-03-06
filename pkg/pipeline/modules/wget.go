package modules

import (
	"errors"
	"strconv"
	"strings"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// pipeline module for wget

// WGet -
type WGet struct{}

// Name -
func (e *WGet) Name() string {
	return "WGet"
}

// Command -
func (e *WGet) Command() string {
	return "wget"
}

// StartPipeline -
func (e *WGet) StartPipeline(ctx *pipeline.ExecutionContext) error {
	return nil
}

// FinishPipeline -
func (e *WGet) FinishPipeline(ctx *pipeline.ExecutionContext) error {
	return nil
}

// Execute -
func (e *WGet) Execute(ctx *pipeline.ExecutionContext, cmd string) (environment.ExecutionResult, error) {
	// run wget and track progress
	ctx.Tracker.Logger().Info("executing wget `" + cmd + "`")
	result, err := ctx.Environment.Execute(
		append([]string{}, "/bin/sh", "-c", "wget -q --show-progress "+cmd),
		func(outmsg string) {
			ctx.Tracker.TrackStdOut(outmsg)
		},
		func(errmsg string) {
			ctx.Tracker.TrackStdErr(errmsg)
			if strings.Contains(errmsg, "%") {
				split := strings.Split(errmsg, " ")
				for _, part := range split {
					if strings.LastIndexByte(part, '%') == len(part)-1 {
						progress, err := strconv.Atoi(strings.TrimSuffix(part, "%"))
						if err == nil {
							ctx.Tracker.TrackProgress(uint(progress))
						}
					}
				}
			}
		})
	if err != nil {
		ctx.Tracker.Logger().Crit("execution of wget failed")
		return environment.ExecutionResult{}, err
	}

	if result.ExitCode == 0 {
		ctx.Tracker.Logger().Info("wget finished")
	} else {
		// TODO: handle error
		return environment.ExecutionResult{}, errors.New("wget failed")
	}

	return environment.ExecutionResult{}, nil
}
