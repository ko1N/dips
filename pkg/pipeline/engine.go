package pipeline

import (
	"github.com/d5/tengo/v2"
	"github.com/ko1N/dips/pkg/pipeline/tracking"
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

// CreateExecution - creates a new execution context
func (e *Engine) CreateExecution(jobID string, pipeline *Pipeline, tracker tracking.JobTracker) ExecutionContext {
	return ExecutionContext{
		Engine:    e,
		JobID:     jobID,
		Pipeline:  pipeline,
		Tracker:   tracker,
		Variables: make(map[string]tengo.Object),
	}
}
