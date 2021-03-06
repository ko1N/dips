package modules

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/environments"
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
func (e *WGet) Execute(ctx *pipeline.ExecutionContext, cmd string) (environments.ExecutionResult, error) {
	// run wget and track progress
	ctx.Tracker.Logger().Info("executing wget `" + cmd + "`")
	result, err := ctx.Environment.Execute(
		append([]string{}, "/bin/sh", "-c", "wget -q --show-progress "+cmd),
		func(outmsg string) {
			ctx.Tracker.StdOut(outmsg)
		},
		func(errmsg string) {
			ctx.Tracker.StdErr(errmsg)
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
		return environments.ExecutionResult{}, err
	}

	if result.ExitCode == 0 {
		ctx.Tracker.Logger().Info("wget finished")
	} else {
		// TODO: handle error
		return environments.ExecutionResult{}, errors.New("wget failed")
	}

	return environments.ExecutionResult{}, nil
}
