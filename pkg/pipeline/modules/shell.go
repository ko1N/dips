package modules

import (
	"strconv"
	"strings"

	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/environments"
)

// pipeline module for shell cmds

// Shell -
type Shell struct{}

// Name -
func (e *Shell) Name() string {
	return "Shell"
}

// Command -
func (e *Shell) Command() string {
	return "shell"
}

// StartPipeline -
func (e *Shell) StartPipeline(ctx *pipeline.ExecutionContext) error {
	return nil
}

// FinishPipeline -
func (e *Shell) FinishPipeline(ctx *pipeline.ExecutionContext) error {
	return nil
}

// Execute - Executes the set of commands as shell commands in the environment
func (e *Shell) Execute(ctx *pipeline.ExecutionContext, cmd string) (environments.ExecutionResult, error) {
	ctx.Tracker.Logger().Info("executing command `" + cmd + "`")
	result, err := ctx.Environment.Execute(
		append([]string{}, "/bin/sh", "-c", cmd),
		func(outmsg string) {
			ctx.Tracker.StdOut(outmsg)
		},
		func(errmsg string) {
			ctx.Tracker.StdErr(errmsg)
		})
	if err != nil {
		return environments.ExecutionResult{}, err
	}

	// TODO: move this into engine.go
	if result.ExitCode == 0 {
		if result.StdOut != "" {
			ctx.Tracker.Logger().Info("command result: `" + strings.TrimSuffix(result.StdOut, "\n") + "`")
		} else if result.StdErr != "" {
			ctx.Tracker.Logger().Info("command result: `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
		}
	} else {
		if result.StdErr != "" {
			ctx.Tracker.Logger().Warn("command failed with code " + strconv.Itoa(result.ExitCode) + ": `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
		} else {
			ctx.Tracker.Logger().Warn("command failed with code " + strconv.Itoa(result.ExitCode))
		}
	}

	return result, nil
}
