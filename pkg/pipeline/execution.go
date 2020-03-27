package pipeline

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/d5/tengo/v2"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/environments"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/tracking"
)

// ExecutionContext - context for a execution
type ExecutionContext struct {
	Engine      *Engine
	JobID       string
	Pipeline    *Pipeline
	Tracker     tracking.JobTracker
	Environment environments.Environment
	Variables   map[string]tengo.Object
}

// Run - runs the execution
func (e *ExecutionContext) Run() error {
	e.Tracker.Status("------ Starting Pipeline: " + e.JobID)
	defer e.Tracker.Status("------ Finished Pipeline: " + e.JobID)

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
		e.Tracker.Status("------ Performing Stage: " + stage.Name)

		// setup environment
		e.Tracker.Status("--- Creating environment: " + stage.Environment)
		env, err := e.createEnvironment(stage.Environment, e.Tracker)
		if err != nil {
			e.Tracker.Error("unable to create environment `"+stage.Environment+"`", nil)
			return err
		}
		e.Environment = env
		defer func() {
			env.Close()
			e.Environment = nil
		}()

		// execute tasks in pipeline
		for _, task := range stage.Tasks {
			e.Tracker.Status("--- Executing Task: " + task.Name)
			e.Tracker.TrackTask(taskID)

			// check "when" condition
			if task.When.Script != "" {
				res, err := task.When.Evaluate(e.Variables)
				if err != nil {
					e.Tracker.Error("unable to compile expression", err)
					return err
				}
				if res == false {
					e.Tracker.Status("`when` condition not met, skipping task")
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
								e.Tracker.Error("task execution failed", err)
								return err
							}

							if !task.IgnoreErrors && result.ExitCode != 0 {
								e.Tracker.Error("aborting pipeline execution", errors.New("task failed to exit properly (exitcode "+strconv.Itoa(result.ExitCode)+")"))
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

//
// TODO: remove Logger() dependency from environment but rather use the tracker!
//

func (e *ExecutionContext) createEnvironment(env string, tracker tracking.JobTracker) (environments.Environment, error) {
	split := strings.Split(env, "/")

	switch split[0] {
	case "native":
		env, err := environments.CreateNativeEnvironment(tracker)
		if err != nil {
			return nil, err
		}
		return &env, nil
	case "docker":
		image := "alpine:latest"
		if len(split) > 1 {
			image = split[1]
		}
		env, err := environments.CreateDockerEnvironment(tracker, image)
		if err != nil {
			return nil, err
		}
		return &env, nil
	}

	return nil, fmt.Errorf("environment `%s` not found", split[0])
}
