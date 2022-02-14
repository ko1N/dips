package client

import (
	"encoding/json"

	"github.com/ko1N/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

type Job struct {
	jobQueue (chan string)
	job      *model.Job
	params   map[string]interface{}
}

type JobRequest struct {
	Job    *model.Job             `json:"job"`
	Params map[string]interface{} `json:"params"`
}

func (c *Client) NewJob() *Job {
	return &Job{
		jobQueue: c.amqp.RegisterProducer("dips.worker.job"),
	}
}

// Job - Sets the job
func (j *Job) Job(job *model.Job) *Job {
	j.job = job
	return j
}

// Parameters - Sets the input parameters of the job
func (j *Job) Parameters(params map[string]interface{}) *Job {
	j.params = params
	return j
}

// Dispatch - Dispatches the task to the message queue
func (j *Job) Dispatch() {
	jobRequest := JobRequest{
		Job: j.job,
		// TODO: params
	}
	jobRequest.Job.Id = bson.NewObjectId()

	request, err := json.Marshal(&jobRequest)
	if err != nil {
		panic("Invalid job request: " + err.Error())
	}

	j.jobQueue <- string(request)
}

type JobWorker struct {
	client   *Client
	jobQueue (chan string)
	handler  func(*JobContext) error
}

type JobContext struct {
	Client  *Client
	Worker  *JobWorker
	Request *JobRequest
}

func (c *Client) NewJobWorker() *JobWorker {
	return &JobWorker{
		client:   c,
		jobQueue: c.amqp.RegisterConsumer("dips.worker.job"),
	}
}

func (w *JobWorker) Handler(handler func(*JobContext) error) *JobWorker {
	w.handler = handler
	return w
}

// Run - Starts a new goroutine for this worker
func (w *JobWorker) Run() {
	// TODO: graceful shutdown
	go func() {
		for request := range w.jobQueue {
			var jobRequest JobRequest
			err := json.Unmarshal([]byte(request), &jobRequest)
			if err != nil {
				panic("Invalid job request: " + err.Error())
			}
			if w.handler != nil {
				w.handler(&JobContext{
					Client:  w.client,
					Worker:  w,
					Request: &jobRequest,
				})
			} else {
				// TODO: handle case?
			}
		}
	}()
}
