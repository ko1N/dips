package client

import (
	"encoding/json"

	"github.com/ko1N/dips/internal/persistence/database/model"
)

type Job struct {
	jobQueue (chan string)
	job      *model.Job
}

type JobRequest struct {
	Job *model.Job `json:"job"`
}

func (client *Client) NewJob() *Job {
	return &Job{
		jobQueue: client.amqp.RegisterProducer("dips.worker.job"),
	}
}

// Job - Sets the job the task belongs to
func (job *Job) Job(jobModel *model.Job) *Job {
	job.job = jobModel
	return job
}

// Dispatch - Dispatches the task to the message queue
func (job *Job) Dispatch() {
	jobRequest := JobRequest{
		Job: job.job,
	}

	request, err := json.Marshal(&jobRequest)
	if err != nil {
		panic("Invalid job request: " + err.Error())
	}

	job.jobQueue <- string(request)
}

type JobWorker struct {
	jobQueue (chan string)
	handler  func(*JobContext) error
}

type JobContext struct {
	Worker  *JobWorker
	Request *JobRequest
}

func (client *Client) NewJobWorker() *JobWorker {
	return &JobWorker{
		jobQueue: client.amqp.RegisterConsumer("dips.worker.job"),
	}
}

func (worker *JobWorker) Handler(handler func(*JobContext) error) *JobWorker {
	worker.handler = handler
	return worker
}

// Run - Starts a new goroutine for this worker
func (worker *JobWorker) Run() {
	// TODO: graceful shutdown
	go func() {
		for request := range worker.jobQueue {
			var jobRequest JobRequest
			err := json.Unmarshal([]byte(request), &jobRequest)
			if err != nil {
				panic("Invalid job request: " + err.Error())
			}
			if worker.handler != nil {
				worker.handler(&JobContext{
					Worker:  worker,
					Request: &jobRequest,
				})
			} else {
				// TODO: handle case?
			}
		}
	}()
}
