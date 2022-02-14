package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

// Task - A task instance to be dispatched to a worker
type Task struct {
	id           string
	timeout      time.Duration
	taskRequests (chan amqp.Message)
	taskResults  (chan amqp.Message)
	job          *model.Job
	name         string
	params       map[string]interface{}
}

// TaskRequest - Request to start a task
type TaskRequest struct {
	TaskID string                 `json:"id" bson:"id"`
	Job    *model.Job             `json:"job" bson:"job"`
	Name   string                 `json:"name" bson:"name"`
	Params map[string]interface{} `json:"params" bson:"params"`
}

// A task that has been dispatched to a worker that awaits a response
type DispatchedTask struct {
	task *Task
}

// NewTask - Creates a new task to be dispatched to a worker
func (client *Client) NewTask(service string) *Task {
	// TODO: sanitize name
	taskId := bson.NewObjectId().Hex()
	return &Task{
		id:           taskId,
		timeout:      30 * time.Minute,
		taskRequests: client.amqp.RegisterProducer("dips.worker.task." + service + ".request"),
		taskResults:  client.amqp.RegisterResponseConsumer("dips.worker.task."+service+".result", taskId),
	}
}

// Sets the timeout of the task
func (t *Task) Timeout(timeout time.Duration) *Task {
	t.timeout = timeout
	return t
}

// Job - Sets the job the task belongs to
func (t *Task) Job(job *model.Job) *Task {
	t.job = job
	return t
}

// Name - Sets the name of the task
func (t *Task) Name(name string) *Task {
	t.name = name
	return t
}

// Parameters - Sets the input parameters of the task
func (t *Task) Parameters(params map[string]interface{}) *Task {
	t.params = params
	return t
}

// Dispatches the task (and never blocks)
func (t *Task) Dispatch() *DispatchedTask {
	taskRequest := TaskRequest{
		TaskID: t.id,
		Job:    t.job,
		Name:   t.name,
		Params: t.params,
	}

	request, err := json.Marshal(&taskRequest)
	if err != nil {
		panic("Invalid task request: " + err.Error())
	}

	t.taskRequests <- amqp.Message{
		Expiration: strconv.Itoa(int(t.timeout.Milliseconds())), // TODO: check if this is correct?
		Payload:    string(request),
	}

	return &DispatchedTask{
		task: t,
	}
}

func (t *DispatchedTask) Await() error {
	now := time.Now()
outer:
	for {
		select {
		case result := <-t.task.taskResults:
			fmt.Println(result)
			break outer

		default:
			if now.Add(t.task.timeout).Before(time.Now()) {
				return errors.New("Timeout reached while executing task")
			}
			time.Sleep(1 * time.Millisecond)
			break
		}
	}

	return nil
}

// TaskWorker - A worker service instance
type TaskWorker struct {
	client       *Client
	taskRequests (chan amqp.Message)
	taskResults  (chan amqp.Message)
	handler      func(*TaskContext) error
}

// TaskContext - The TaskContext that is being sent to the task handler
type TaskContext struct {
	Client  *Client
	Worker  *TaskWorker
	Request *TaskRequest

	// TODO: input variables
	// TODO: logging function
	// TODO: status function
}

// NewWorker - Creates a new worker service with the given name
func (client *Client) NewTaskWorker(service string) *TaskWorker {
	// TODO: sanitize name
	return &TaskWorker{
		client:       client,
		taskRequests: client.amqp.RegisterConsumer("dips.worker.task." + service + ".request"),
		taskResults:  client.amqp.RegisterProducer("dips.worker.task." + service + ".result"),
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
		for request := range worker.taskRequests {
			var taskRequest TaskRequest
			err := json.Unmarshal([]byte(request.Payload), &taskRequest)
			if err != nil {
				panic("Invalid task request: " + err.Error())
			}
			if worker.handler != nil {
				worker.handler(&TaskContext{
					Client:  worker.client,
					Worker:  worker,
					Request: &taskRequest,
				})
			} else {
				// TODO: handle case?
			}

			// send response for task
			worker.taskResults <- amqp.Message{
				CorrelationId: taskRequest.TaskID,
				Payload:       "TODO",
			}
		}
	}()
}
