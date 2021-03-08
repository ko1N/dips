package client

import (
	"encoding/json"

	"github.com/ko1N/dips/internal/persistence/database/model"
)

// Worker - A worker service instance
type Worker struct {
	workerQueue (chan string)
	handler     func(*TaskContext) error
}

// TaskRequest - Request to start a task
type TaskRequest struct {
	Job  *model.Job `json:"job"`
	Name string     `json:"name"`
	// TODO: task id ?
	Params map[string]interface{} `json:"params"`
}

// TaskContext - The TaskContext that is being sent to the task handler
type TaskContext struct {
	//Client      *Client
	Worker      *Worker
	TaskRequest *TaskRequest

	// TODO: input variables
	// TODO: logging function
	// TODO: status function
}

// NewWorker - Creates a new worker service with the given name
func (client *Client) NewWorker(name string) *Worker {
	// TODO: sanitize name
	workerQueue := client.amqp.RegisterConsumer("dips.worker.task." + name)
	return &Worker{
		workerQueue: workerQueue,
	}
}

// Handler - Sets the handler for this worker
func (worker *Worker) Handler(handler func(*TaskContext) error) *Worker {
	worker.handler = handler
	return worker
}

// Run - Starts a new goroutine for this worker
func (worker *Worker) Run() {
	// TODO: graceful shutdown
	go func() {
		for request := range worker.workerQueue {
			var taskRequest TaskRequest
			err := json.Unmarshal([]byte(request), &taskRequest)
			if err != nil {
				panic("Invalid job request: " + err.Error())
			}
			if worker.handler != nil {
				worker.handler(&TaskContext{
					Worker:      worker,
					TaskRequest: &taskRequest,
				})
			} else {
				// TODO: handle case?
			}
		}
	}()
}
