package pipeline

import (
	"fmt"
	"strings"

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
}

// ExecutePipeline - executed the given pipeline on the engine
func (e *Engine) ExecutePipeline(ctx ExecutionContext) error {
	// create a channel for communication
	// + logging for this pipeline, then exec it

	// TODO: read stages and execute them here,
	// TODO: properly parse the pipeline interfaces into structures in the CreatePipeline() func

	//wf.Run()

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

			// TODO: new func + throw error if command was not found!
			for _, cmd := range task.Command {
				for _, ext := range e.Extensions {
					if ext.Command() == cmd.Name {
						//fmt.Println("executing cmd " + ext.Name())
						ext.Execute(ctx, cmd.Arguments)
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
