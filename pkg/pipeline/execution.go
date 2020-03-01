package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/d5/tengo/v2"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
)

// ExecutionContext - context for a execution
type ExecutionContext struct {
	Engine      *Engine
	JobID       string
	Pipeline    Pipeline                // TODO ptr?
	Tracker     JobTracker              // TODO ptr?
	Environment environment.Environment // TODO ptr?
	Variables   map[string]tengo.Object
}

// Run - runs the execution
func (e *ExecutionContext) Run() error {
	e.Tracker.Logger().Info("------ Starting Pipeline: " + e.JobID)
	defer e.Tracker.Logger().Info("------ Finished Pipeline: " + e.JobID)

	// TODO: add CreateExecutionContext
	// TODO: move context into seperate file
	// TODO: decouple this function into ExecutionContext
	e.Variables = make(map[string]tengo.Object)

	// call startPipline extension hooks
	for _, ext := range e.Engine.Extensions {
		if err := ext.StartPipeline(e); err != nil {
			return err
		}
	}

	defer func() {
		// call finishPipline extension hooks (regardless of failure)
		for _, ext := range e.Engine.Extensions {
			ext.FinishPipeline(e)
		}
	}()

	var taskID uint
	for _, stage := range e.Pipeline.Stages {
		e.Tracker.Logger().Info("------ Performing Stage: " + stage.Name)

		// setup environment
		e.Tracker.Logger().Info("--- Creating environment: " + stage.Environment)
		env, err := e.createEnvironment(stage.Environment, e.Tracker)
		if err != nil {
			e.Tracker.Logger().Crit("unable to create environment `" + stage.Environment + "`")
			return err
		}
		e.Environment = env
		defer func() {
			env.Close()
			e.Environment = nil
		}()

		// execute tasks in pipeline
		for _, task := range stage.Tasks {
			e.Tracker.Logger().Info("--- Executing Task: " + task.Name)
			e.Tracker.TrackTask(taskID)

			// check "when" condition
			if task.When.Script != "" {
				res, err := task.When.Evaluate(e.Variables)
				if err != nil {
					e.Tracker.Logger().Crit("unable to compile expression", "error", err)
					return err
				}
				if res == false {
					e.Tracker.Logger().Info("`when` condition not met, skipping task")
					continue
				}
			}

			// TODO: new func + throw error if command was not found!
			for _, cmd := range task.Command {
				for _, ext := range e.Engine.Extensions {
					if ext.Command() == cmd.Name {
						for _, line := range cmd.Lines {
							result, err := ext.Execute(e, line)
							if err != nil {
								e.Tracker.Logger().Crit("task execution failed", "error", err)
								return err
							}

							if !task.IgnoreErrors && result.ExitCode != 0 {
								e.Tracker.Logger().Crit("aborting pipeline execution", "error", "task failed to exit properly (exitcode "+strconv.Itoa(result.ExitCode)+")")
								return nil
							}

							// convert result into tengo objects and store it
							if task.Register != "" {
								e.Variables[task.Register] = result.ToScriptObject()
							}
						}
					}
				}
			}

			// if this task doesnt support tracking we just increase it to 100%
			e.Tracker.TrackProgress(100)
			taskID++
		}
	}

	return nil
}

func (e *ExecutionContext) createEnvironment(env string, tracker JobTracker) (environment.Environment, error) {
	split := strings.Split(env, "/")

	switch split[0] {
	case "native":
		env, err := environment.CreateNativeEnvironment(tracker.Logger())
		if err != nil {
			return nil, err
		}
		return &env, nil
	case "docker":
		image := "alpine:latest"
		if len(split) > 1 {
			image = split[1]
		}
		env, err := environment.CreateDockerEnvironment(tracker.Logger(), image)
		if err != nil {
			return nil, err
		}
		return &env, nil
	}

	return nil, fmt.Errorf("environment `%s` not found", split[0])
}
