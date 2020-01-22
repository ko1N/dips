package workflow

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
func (e *Engine) ExecuteWorkflow(wf Workflow) error {
	// create a channel for communication
	// + logging for this workflow, then exec it

	// TODO: read stages and execute them here,
	// TODO: properly parse the workflow interfaces into structures in the CreateWorkflow() func

	wf.Run()
	return nil
}
