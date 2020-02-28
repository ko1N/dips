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

// Execute -
func (e *WGet) Execute(env environment.Environment, cmds []string, tracker pipeline.JobTracker) error {
	for _, cmd := range cmds {
		// run wget and track progress
		tracker.Logger().Info("executing wget `" + cmd + "`")
		result, err := env.Execute(
			append([]string{}, "/bin/sh", "-c", "wget -q --show-progress "+cmd),
			func(outmsg string) {
				tracker.TrackStdOut(outmsg)
			},
			func(errmsg string) {
				tracker.TrackStdErr(errmsg)
				if strings.Contains(errmsg, "%") {
					split := strings.Split(errmsg, " ")
					for _, part := range split {
						if strings.LastIndexByte(part, '%') == len(part)-1 {
							progress, err := strconv.Atoi(strings.TrimSuffix(part, "%"))
							if err == nil {
								tracker.TrackProgress(uint(progress))
							}
						}
					}
				}
			})
		if err != nil {
			tracker.Logger().Crit("execution of wget failed")
			return err
		}

		if result.ExitCode == 0 {
			tracker.Logger().Info("wget finished")
		} else {
			// TODO: handle error
			return errors.New("wget failed")
		}
	}

	return nil
}
