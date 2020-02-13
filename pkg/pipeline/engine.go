package pipeline

import (
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
func (e *Engine) ExecutePipeline(pipelog log.Logger, env environment.Environment, wf Pipeline) error {
	// create a channel for communication
	// + logging for this pipeline, then exec it

	// TODO: read stages and execute them here,
	// TODO: properly parse the pipeline interfaces into structures in the CreatePipeline() func

	//wf.Run()

	for _, stage := range wf.Stages {
		for _, task := range stage.Tasks {
			//fmt.Printf("executing:\n")
			//fmt.Printf("%v\n", task)

			// TODO: new func?
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
