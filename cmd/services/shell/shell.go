package main

import (
	"fmt"
	"time"

	"github.com/ko1N/dips/internal/persistence/database/model"
	"github.com/ko1N/dips/pkg/client"
)

func main() {
	cl, err := client.NewClient("rabbitmq:rabbitmq@localhost")
	if err != nil {
		panic(err)
	}

	cl.NewJobWorker().
		Handler(jobHandler).
		Run()

	cl.NewJob().
		Job(&model.Job{
			Name: "test",
		}).
		Dispatch()

	/*
		cl.
			NewTaskWorker("shell").
			Handler(shellHandler).
			Run()

		cl.
			NewTask("shell").
			Name("test task").
			Parameters(map[string]interface{}{"param": "variant"}).
			Dispatch()
	*/

	fmt.Println("workers started")
	for {
		time.Sleep(1 * time.Second)
	}
}

func jobHandler(job *client.JobContext) error {
	fmt.Printf("handling job %+v\n", job.Request.Job)
	return nil
}

func shellHandler(task *client.TaskContext) error {
	fmt.Printf("handling job %s: %s\n", task.Request.Name, task.Request.Params)
	return nil
}
