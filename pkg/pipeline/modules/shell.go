package modules

import (
	"strconv"
	"strings"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
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
func (e *Shell) StartPipeline(ctx pipeline.ExecutionContext) error {
	return nil
}

// FinishPipeline -
func (e *Shell) FinishPipeline(ctx pipeline.ExecutionContext) error {
	return nil
}

// Execute - Executes the set of commands as shell commands in the environment
func (e *Shell) Execute(ctx pipeline.ExecutionContext, cmds []string) error {
	for _, cmd := range cmds {
		ctx.Tracker.Logger().Info("executing command `" + cmd + "`")
		result, err := ctx.Environment.Execute(
			append([]string{}, "/bin/sh", "-c", cmd),
			func(outmsg string) {
				ctx.Tracker.TrackStdOut(outmsg)
			},
			func(errmsg string) {
				ctx.Tracker.TrackStdErr(errmsg)
			})
		if err != nil {
			return err
		}

		if result.ExitCode == 0 {
			if result.StdOut != "" {
				ctx.Tracker.Logger().Info("command result: `" + strings.TrimSuffix(result.StdOut, "\n") + "`")
			} else if result.StdErr != "" {
				ctx.Tracker.Logger().Info("command result: `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
			}
		} else {
			if result.StdErr != "" {
				ctx.Tracker.Logger().Info("command failed with code " + strconv.Itoa(result.ExitCode) + ": `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
			} else {
				ctx.Tracker.Logger().Info("command failed with code " + strconv.Itoa(result.ExitCode))
			}
		}
	}

	return nil
}
