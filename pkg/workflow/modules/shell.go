package modules

import (
	"fmt"

	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/environment"
)

// workflow module for shell cmds

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
func (e *Shell) Execute(env environment.Environment, cmds []string) error {
	for _, cmd := range cmds {
		buf, err := env.Execute(append([]string{}, "/bin/sh", "-c", cmd))
		if err != nil {
			return err
		}
		fmt.Println("result: " + buf)
	}
	return nil
}
