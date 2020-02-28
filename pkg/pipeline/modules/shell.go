package modules

import (
	"strconv"
	"strings"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
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

// Execute - Executes the set of commands as shell commands in the environment
func (e *Shell) Execute(env environment.Environment, cmds []string, tracker pipeline.JobTracker) error {
	for _, cmd := range cmds {
		tracker.Logger().Info("executing command `" + cmd + "`")
		result, err := env.Execute(
			append([]string{}, "/bin/sh", "-c", cmd),
			func(outmsg string) {
				tracker.TrackStdOut(outmsg)
			},
			func(errmsg string) {
				tracker.TrackStdErr(errmsg)
			})
		if err != nil {
			return err
		}

		if result.ExitCode == 0 {
			if result.StdOut != "" {
				tracker.Logger().Info("command result: `" + strings.TrimSuffix(result.StdOut, "\n") + "`")
			} else if result.StdErr != "" {
				tracker.Logger().Info("command result: `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
			}
		} else {
			if result.StdErr != "" {
				tracker.Logger().Info("command failed with code " + strconv.Itoa(result.ExitCode) + ": `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
			} else {
				tracker.Logger().Info("command failed with code " + strconv.Itoa(result.ExitCode))
			}
		}
	}

	return nil
}
