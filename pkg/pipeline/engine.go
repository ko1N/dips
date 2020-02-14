package pipeline

import (
	"fmt"
	"strings"

	log "github.com/inconshreveable/log15"

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
func (e *Engine) ExecutePipeline(pipelog log.Logger, wf Pipeline) error {
	// create a channel for communication
	// + logging for this pipeline, then exec it

	// TODO: read stages and execute them here,
	// TODO: properly parse the pipeline interfaces into structures in the CreatePipeline() func

	//wf.Run()

	for _, stage := range wf.Stages {
		pipelog.Info("------ Performing Stage: " + stage.Name)

		// setup environment
		pipelog.Info("--- Creating environment: " + stage.Environment)
		env, err := e.createEnvironment(pipelog, stage.Environment)
		if err != nil {
			pipelog.Crit("unable to create environment `" + stage.Environment + "`")
			return err
		}
		defer env.Close()

		// execute tasks in pipeline
		for _, task := range stage.Tasks {
			pipelog.Info("--- Executing Task: " + task.Name)

			// TODO: new func + throw error if command was not found!
			for _, cmd := range task.Command {
				for _, ext := range e.Extensions {
					if ext.Command() == cmd.Name {
						//fmt.Println("executing cmd " + ext.Name())
						ext.Execute(pipelog, env, cmd.Arguments)
					}
				}
			}
		}
	}

	return nil
}

func (e *Engine) createEnvironment(pipelog log.Logger, env string) (environment.Environment, error) {
	split := strings.Split(env, "/")

	switch split[0] {
	case "native":
		env, err := environment.CreateNativeEnvironment(pipelog)
		if err != nil {
			return nil, err
		}
		return &env, nil
	case "docker":
		if len(split) == 1 {
			env, err := environment.CreateDockerEnvironment(pipelog, "alpine:latest")
			if err != nil {
				return nil, err
			}
			return &env, nil
		} else {
			env, err := environment.CreateDockerEnvironment(pipelog, split[1])
			if err != nil {
				return nil, err
			}
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment `%s` not found", split[0])
}
