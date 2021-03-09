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

// TaskRequest - Request to start a task
type TaskRequest struct {
	Job  *model.Job `json:"job"`
	Name string     `json:"name"`
	// TODO: task id ?
	Params map[string]interface{} `json:"params"`
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

// TaskWorker - A worker service instance
type TaskWorker struct {
	taskQueue (chan string)
	handler   func(*TaskContext) error
}

// TaskContext - The TaskContext that is being sent to the task handler
type TaskContext struct {
	//Client      *Client
	Worker  *TaskWorker
	Request *TaskRequest

	// TODO: input variables
	// TODO: logging function
	// TODO: status function
}

// NewWorker - Creates a new worker service with the given name
func (client *Client) NewTaskWorker(name string) *TaskWorker {
	// TODO: sanitize name
	return &TaskWorker{
		taskQueue: client.amqp.RegisterConsumer("dips.worker.task." + name),
	}
}

// Handler - Sets the handler for this worker
func (worker *TaskWorker) Handler(handler func(*TaskContext) error) *TaskWorker {
	worker.handler = handler
	return worker
}

// Run - Starts a new goroutine for this worker
func (worker *TaskWorker) Run() {
	// TODO: graceful shutdown
	go func() {
		for request := range worker.taskQueue {
			var taskRequest TaskRequest
			err := json.Unmarshal([]byte(request), &taskRequest)
			if err != nil {
				panic("Invalid task request: " + err.Error())
			}
			if worker.handler != nil {
				worker.handler(&TaskContext{
					Worker:  worker,
					Request: &taskRequest,
				})
			} else {
				// TODO: handle case?
			}
		}
	}()
}
