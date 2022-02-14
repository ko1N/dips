package pipeline

import (
	"errors"
	"strconv"

	"github.com/d5/tengo/v2"
	"github.com/ko1N/dips/pkg/pipeline/tracking"
)

// ExecutionContext - context for a execution
type ExecutionContext struct {
	JobID       string
	Pipeline    *Pipeline
	Tracker     tracking.JobTracker
	Variables   map[string]tengo.Object
	taskHandler func(*Task) (*ExecutionResult, error)
}

type ExecutionResult struct {
	Error  *string                `json:"error" bson:"error"`
	Output map[string]interface{} `json:"output" bson:"output"`
}

func (r *ExecutionResult) ToScriptObject() tengo.Object {
	if r.Error == nil {
		return &tengo.Map{
			Value: map[string]tengo.Object{
				"failed": tengo.FalseValue,
				"error":  nil,
				// TODO: key value mapping
			}}
	} else {
		return &tengo.Map{
			Value: map[string]tengo.Object{
				"failed": tengo.TrueValue,
				"error":  &tengo.String{Value: *r.Error},
				// TODO: key value mapping
			}}
	}
}

func NewExecutionContext(jobID string, pipeline *Pipeline, tracker tracking.JobTracker) *ExecutionContext {
	return &ExecutionContext{
		JobID:     jobID,
		Pipeline:  pipeline,
		Tracker:   tracker,
		Variables: make(map[string]tengo.Object),
	}
}

// Handler - Sets the handler for this worker
func (e *ExecutionContext) TaskHandler(handler func(*Task) (*ExecutionResult, error)) *ExecutionContext {
	// TODO: multiple handlers
	e.taskHandler = handler
	return e
}

// Run - runs the execution
func (e *ExecutionContext) Run() error {
	e.Tracker.Status("------ Starting Pipeline: " + e.JobID)
	defer e.Tracker.Status("------ Finished Pipeline: " + e.JobID)

	// TODO: add CreateExecutionContext
	// TODO: move context into seperate file
	// TODO: decouple this function into ExecutionContext

	var taskID uint
	for _, stage := range e.Pipeline.Stages {
		e.Tracker.Status("------ Performing Stage: " + stage.Name)

		//expression := regexp.MustCompile(`{{.*?}}`)

		// execute tasks in pipeline
		for _, task := range stage.Tasks {
			e.Tracker.Status("--- Executing Task " + strconv.Itoa(int(taskID)) + ": " + task.Service + " (" + task.Name + ")")
			e.Tracker.TrackTask(taskID)

			// TODO: put this logic in seperate objects
			// check "when" condition
			if task.When.Script != "" {
				res, err := task.When.Evaluate(e.Variables)
				if err != nil {
					e.Tracker.Error("unable to compile expression", err)
					return err
				}
				if res != "true" {
					e.Tracker.Status("`when` condition not met, skipping task")
					continue
				}
			}

			// dispatch task
			if e.taskHandler != nil {
				result, err := (e.taskHandler)(&task)
				if err != nil {
					e.Tracker.Error("task execution failed", err)
					return err
				}

				// TODO: duistingish between service error and actual execution error
				if !task.IgnoreErrors && err != nil {
					e.Tracker.Error("aborting pipeline execution", errors.New("task failed to exit properly ("+err.Error()+")"))
					return nil
				}

				// convert result into tengo objects and store it
				if task.Register != "" {
					e.Variables[task.Register] = result.ToScriptObject()
				}
			}

			// TODO: new func + throw error if command was not found!
			/*
				for _, cmd := range task.Command {
					for _, ext := range e.Engine.Extensions {
						if ext.Command() == cmd.Name {
							for _, rawLine := range cmd.Lines {
								e.Tracker.Status("$ " + rawLine)

								// TODO: put this logic in seperate objects
								line := string(expression.ReplaceAllFunc([]byte(rawLine), func(m []byte) []byte {
									t := strings.TrimSpace(string(m[2 : len(m)-2]))
									v, err := (&Expression{Script: string(t)}).Evaluate(e.Variables)
									if err != nil {
										// TODO:
									}
									return []byte(v)
								}))

								e.Tracker.StdIn(line)

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
			*/

			// if this task doesnt support tracking we just increase it to 100%
			e.Tracker.TrackProgress(100)
			taskID++
		}
	}

	return nil
}
