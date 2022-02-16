package client

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

const defaultTaskTimeout = 3 * time.Minute

// Task - A task instance to be dispatched to a worker
type Task struct {
	client       *Client
	service      string
	id           string
	timeout      time.Duration
	taskRequests (chan amqp.Message)
	taskResults  (chan amqp.Message)
	job          *model.Job
	name         string
	params       map[string]string
}

// TaskRequest - Request to start a task
type TaskRequest struct {
	TaskID  string            `json:"id" bson:"id"`
	Timeout time.Duration     `json:"timeout" bson:"timeout"`
	Job     *model.Job        `json:"job" bson:"job"`
	Name    string            `json:"name" bson:"name"`
	Params  map[string]string `json:"params" bson:"params"`
}

// A task that has been dispatched to a worker that awaits a response
type DispatchedTask struct {
	task *Task
}

type TaskResult struct {
	Error  *string                `json:"error" bson:"error"`
	Output map[string]interface{} `json:"output" bson:"output"`
}

// NewTask - Creates a new task to be dispatched to a worker
func (client *Client) NewTask(service string) *Task {
	// TODO: sanitize name
	taskId := bson.NewObjectId().Hex()
	return &Task{
		client:       client,
		service:      service,
		id:           taskId,
		timeout:      defaultTaskTimeout,
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
func (t *Task) Parameters(params map[string]string) *Task {
	t.params = params
	return t
}

// Dispatches the task (and never blocks)
func (t *Task) Dispatch() *DispatchedTask {
	taskRequest := TaskRequest{
		TaskID:  t.id,
		Timeout: t.timeout,
		Job:     t.job,
		Name:    t.name,
		Params:  t.params,
	}

	request, err := json.Marshal(&taskRequest)
	if err != nil {
		panic("Invalid task request: " + err.Error())
	}

	t.taskRequests <- amqp.Message{
		Expiration: strconv.Itoa(int(t.timeout.Milliseconds())),
		Payload:    string(request),
	}

	return &DispatchedTask{
		task: t,
	}
}

func (t *DispatchedTask) Await() (*TaskResult, error) {
	// release channel after this function returns
	defer t.Close()

	now := time.Now()
	for {
		select {
		case result := <-t.task.taskResults:
			var tr TaskResult
			err := json.Unmarshal([]byte(result.Payload), &tr)
			if err != nil {

			}
			if tr.Error != nil {
				return nil, errors.New(*tr.Error)
			}
			return &tr, nil

		default:
			if now.Add(t.task.timeout).Before(time.Now()) {
				return nil, errors.New("Timeout reached while executing task")
			}
			time.Sleep(1 * time.Millisecond)
			break
		}
	}
}

func (t *DispatchedTask) Close() {
	t.task.client.amqp.CloseResponseConsumer("dips.worker.task."+t.task.service+".result", t.task.id)
}

// TaskWorker - A worker service instance
type TaskWorker struct {
	client       *Client
	concurrency  int
	taskRequests (chan amqp.Message)
	taskResults  (chan amqp.Message)
	handler      func(*TaskContext) (map[string]interface{}, error)
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

func (worker *TaskWorker) Concurrency(threads int) *TaskWorker {
	worker.concurrency = threads
	return worker
}

// Handler - Sets the handler for this worker
func (worker *TaskWorker) Handler(handler func(*TaskContext) (map[string]interface{}, error)) *TaskWorker {
	worker.handler = handler
	return worker
}

// Run - Starts a new goroutine for this worker
func (worker *TaskWorker) Run() {
	// TODO: graceful shutdown
	concurrency := worker.concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for request := range worker.taskRequests {
				var taskRequest TaskRequest
				err := json.Unmarshal([]byte(request.Payload), &taskRequest)
				if err != nil {
					panic("Invalid task request: " + err.Error())
				}
				if worker.handler != nil {
					result, err := worker.handler(&TaskContext{
						Client:  worker.client,
						Worker:  worker,
						Request: &taskRequest,
					})

					response := TaskResult{
						Output: result,
					}
					if err != nil {
						e := err.Error()
						response.Error = &e
					}

					payload, err := json.Marshal(response)
					if err != nil {
						panic("Unable to marshal task result: " + err.Error())
					}

					// send response for task
					worker.taskResults <- amqp.Message{
						Expiration:    strconv.Itoa(int(taskRequest.Timeout.Milliseconds())),
						CorrelationId: taskRequest.TaskID,
						Payload:       string(payload),
					}
				} else {
					panic("handler not registered")
				}
			}
		}()
	}
}
