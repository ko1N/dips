package modules

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/inconshreveable/log15"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
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

// Execute -
func (e *WGet) Execute(pipelog log.Logger, env environment.Environment, cmds []string) error {
	for _, cmd := range cmds {
		// run wget and track progress
		pipelog.Info("executing wget `" + cmd + "`")
		result, err := env.Execute(
			append([]string{}, "/bin/sh", "-c", "wget -q --show-progress "+cmd),
			nil,
			func(out string) {
				if strings.Contains(out, "%") {
					split := strings.Split(out, " ")
					for _, part := range split {
						if strings.LastIndexByte(part, '%') == len(part)-1 {
							progress, err := strconv.Atoi(strings.TrimSuffix(part, "%"))
							if err == nil {
								fmt.Printf("progress: %d\n", progress)
							}
						}
					}
				}
			})
		if err != nil {
			pipelog.Crit("execution of wget failed")
			return err
		}

		if result.ExitCode == 0 {
			pipelog.Info("wget finished")
		} else {
			// TODO: handle error
			return errors.New("wget failed")
		}
	}

	return nil
}
