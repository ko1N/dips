package workflow

import (
	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/environment"
)

// Engine - engine instance
type Engine struct {
	Extensions []Extension
}

// TODO: load/register modules, instantiate executions, run workflows, etc
// TODO: channels between workflows to execute 'notify' ?

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

// ExecuteWorkflow - executed the given workflow on the engine
func (e *Engine) ExecuteWorkflow(env environment.Environment, wf Workflow) error {
	// create a channel for communication
	// + logging for this workflow, then exec it

	// TODO: read stages and execute them here,
	// TODO: properly parse the workflow interfaces into structures in the CreateWorkflow() func

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
						ext.Execute(env, cmd.Arguments)
					}
				}
			}
		}
	}

	return nil
}
