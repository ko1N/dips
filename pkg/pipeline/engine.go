package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/d5/tengo/v2"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
)

// Engine - engine instance
type Engine struct {
	Extensions []Extension
}

// TODO: load/register modules, instantiate executions, run pipelines, etc
// TODO: channels between pipelines to execute 'notify' ?

// CreateEngine - creates a new engine instance
func CreateEngine() Engine {
	return Engine{
		Extensions: []Extension{},
	}
}

// RegisterExtension - registers the extension in this engine
func (e *Engine) RegisterExtension(ext Extension) *Engine {
	e.Extensions = append(e.Extensions, ext)
	return e
}

// ExecutionContext - context for a execution
type ExecutionContext struct {
	JobID       string
	Pipeline    Pipeline
	Tracker     JobTracker
	Environment environment.Environment
	Variables   map[string]tengo.Object
}

// ExecutePipeline - executed the given pipeline on the engine
func (e *Engine) ExecutePipeline(ctx ExecutionContext) error {
	ctx.Tracker.Logger().Info("------ Starting Pipeline: " + ctx.JobID)
	defer ctx.Tracker.Logger().Info("------ Finished Pipeline: " + ctx.JobID)

	// TODO: add CreateExecutionContext
	// TODO: move context into seperate file
	// TODO: decouple this function into ExecutionContext
	ctx.Variables = make(map[string]tengo.Object)

	// call startPipline extension hooks
	for _, ext := range e.Extensions {
		if err := ext.StartPipeline(ctx); err != nil {
			return err
		}
	}

	defer func() {
		// call finishPipline extension hooks (regardless of failure)
		for _, ext := range e.Extensions {
			ext.FinishPipeline(ctx)
		}
	}()

	var taskID uint
	for _, stage := range ctx.Pipeline.Stages {
		ctx.Tracker.Logger().Info("------ Performing Stage: " + stage.Name)

		// setup environment
		ctx.Tracker.Logger().Info("--- Creating environment: " + stage.Environment)
		env, err := e.createEnvironment(stage.Environment, ctx.Tracker)
		if err != nil {
			ctx.Tracker.Logger().Crit("unable to create environment `" + stage.Environment + "`")
			return err
		}
		ctx.Environment = env
		defer func() {
			env.Close()
			ctx.Environment = nil
		}()

		// execute tasks in pipeline
		for _, task := range stage.Tasks {
			ctx.Tracker.Logger().Info("--- Executing Task: " + task.Name)
			ctx.Tracker.TrackTask(taskID)

			// check "when" condition
			if task.When.Script != "" {
				res, err := task.When.Evaluate(ctx.Variables)
				if err != nil {
					ctx.Tracker.Logger().Crit("unable to compile expression", "error", err)
					return err
				}
				if res == false {
					ctx.Tracker.Logger().Info("`when` condition not met, skipping task")
					continue
				}
			}

			// TODO: new func + throw error if command was not found!
			for _, cmd := range task.Command {
				for _, ext := range e.Extensions {
					if ext.Command() == cmd.Name {
						for _, line := range cmd.Lines {
							result, err := ext.Execute(ctx, line)
							if err != nil {
								ctx.Tracker.Logger().Crit("task execution failed", "error", err)
								return err
							}

							if !task.IgnoreErrors && result.ExitCode != 0 {
								ctx.Tracker.Logger().Crit("aborting pipeline execution", "error", "task failed to exit properly (exitcode "+strconv.Itoa(result.ExitCode)+")")
								return nil
							}

							// convert result into tengo objects and store it
							if task.Register != "" {
								ctx.Variables[task.Register] = result.ToScriptObject()
							}
						}
					}
				}
			}

			// if this task doesnt support tracking we just increase it to 100%
			ctx.Tracker.TrackProgress(100)
			taskID++
		}
	}

	return nil
}

func (e *Engine) createEnvironment(env string, tracker JobTracker) (environment.Environment, error) {
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
