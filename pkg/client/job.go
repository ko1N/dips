package client

import (
	"encoding/json"

	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

type Job struct {
	jobQueue  (chan amqp.Message)
	job       *model.Job
	variables map[string]interface{}
}

type JobRequest struct {
	Job       *model.Job             `json:"job"`
	Variables map[string]interface{} `json:"variables"`
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

func (j *Job) Name(name string) *Job {
	if j.job == nil {
		j.job = &model.Job{}
	}
	j.job.Name = name
	return j
}

func (j *Job) Pipeline(script []byte) *Job {
	if j.job.Pipeline == nil {
		j.job.Pipeline = &model.Pipeline{}
	}
	j.job.Pipeline.Script = script
	return j
}

// Parameters - Sets the input parameters of the job
func (j *Job) Variables(variables map[string]interface{}) *Job {
	j.variables = variables
	return j
}

// Dispatches the job (and never blocks)
func (j *Job) Dispatch() {
	jobRequest := JobRequest{
		Job:       j.job,
		Variables: j.variables,
	}
	jobRequest.Job.Id = bson.NewObjectId()

	request, err := json.Marshal(&jobRequest)
	if err != nil {
		panic("Invalid job request: " + err.Error())
	}

	j.jobQueue <- amqp.Message{
		Payload: string(request),
	}
}

type JobWorker struct {
	client      *Client
	concurrency int
	jobQueue    (chan amqp.Message)
	handler     func(*JobContext) error
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

func (w *JobWorker) Concurrency(threads int) *JobWorker {
	w.concurrency = threads
	return w
}

func (w *JobWorker) Handler(handler func(*JobContext) error) *JobWorker {
	w.handler = handler
	return w
}

// Run - Starts a new goroutine for this worker
func (w *JobWorker) Run() {
	// TODO: graceful shutdown
	concurrency := w.concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for request := range w.jobQueue {
				var jobRequest JobRequest
				err := json.Unmarshal([]byte(request.Payload), &jobRequest)
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
}
