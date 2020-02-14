package modules

import (
	"strconv"
	"strings"

	log "github.com/inconshreveable/log15"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
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
func (e *Shell) Execute(pipelog log.Logger, env environment.Environment, cmds []string) error {
	for _, cmd := range cmds {
		pipelog.Info("executing command `" + cmd + "`")
		result, err := env.Execute(append([]string{}, "/bin/sh", "-c", cmd), nil, nil)
		if err != nil {
			return err
		}

		if result.ExitCode == 0 {
			if result.StdOut != "" {
				pipelog.Info("command result: `" + strings.TrimSuffix(result.StdOut, "\n") + "`")
			} else if result.StdErr != "" {
				pipelog.Info("command result: `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
			}
		} else {
			if result.StdErr != "" {
				pipelog.Info("command failed with code " + strconv.Itoa(result.ExitCode) + ": `" + strings.TrimSuffix(result.StdErr, "\n") + "`")
			} else {
				pipelog.Info("command failed with code " + strconv.Itoa(result.ExitCode))
			}
		}
	}

	return nil
}
