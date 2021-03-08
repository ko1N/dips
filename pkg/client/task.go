package client

import (
	"encoding/json"

	"github.com/ko1N/dips/internal/persistence/database/model"
)

// Task - A task instance to be dispatched to a worker
type Task struct {
	taskQueue (chan string)
	job       *model.Job
	name      string
	params    map[string]interface{}
}

// NewTask - Creates a new task to be dispatched to a worker
func (client *Client) NewTask(service string) *Task {
	// TODO: sanitize name
	return &Task{
		taskQueue: client.amqp.RegisterProducer("dips.worker.task." + service),
	}
}

// Job - Sets the job the task belongs to
func (task *Task) Job(job *model.Job) *Task {
	task.job = job
	return task
}

// Name - Sets the name of the task
func (task *Task) Name(name string) *Task {
	task.name = name
	return task
}

// Parameters - Sets the input parameters of the task
func (task *Task) Parameters(params map[string]interface{}) *Task {
	task.params = params
	return task
}

// Dispatch - Dispatches the task to the message queue
func (task *Task) Dispatch() {
	taskRequest := TaskRequest{
		Job:    task.job,
		Name:   task.name,
		Params: task.params,
	}

	request, err := json.Marshal(&taskRequest)
	if err != nil {
		panic("Invalid task request: " + err.Error())
	}

	task.taskQueue <- string(request)
}
