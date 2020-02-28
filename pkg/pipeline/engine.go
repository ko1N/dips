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

// ExecutePipeline - executed the given pipeline on the engine
func (e *Engine) ExecutePipeline(wf Pipeline, tracker JobTracker) error {
	// create a channel for communication
	// + logging for this pipeline, then exec it

	// TODO: read stages and execute them here,
	// TODO: properly parse the pipeline interfaces into structures in the CreatePipeline() func

	//wf.Run()

	var taskID uint
	for _, stage := range wf.Stages {
		tracker.Logger().Info("------ Performing Stage: " + stage.Name)

		// setup environment
		tracker.Logger().Info("--- Creating environment: " + stage.Environment)
		env, err := e.createEnvironment(stage.Environment, tracker)
		if err != nil {
			tracker.Logger().Crit("unable to create environment `" + stage.Environment + "`")
			return err
		}
		defer env.Close()

		// execute tasks in pipeline
		for _, task := range stage.Tasks {
			tracker.Logger().Info("--- Executing Task: " + task.Name)
			tracker.TrackTask(taskID)

			// TODO: new func + throw error if command was not found!
			for _, cmd := range task.Command {
				for _, ext := range e.Extensions {
					if ext.Command() == cmd.Name {
						//fmt.Println("executing cmd " + ext.Name())
						ext.Execute(env, cmd.Arguments, tracker)
					}
				}
			}

			// if this task doesnt support tracking we just increase it to 100%
			tracker.TrackProgress(100)
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
