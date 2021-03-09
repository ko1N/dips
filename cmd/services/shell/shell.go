package main

import (
	"fmt"
	"time"

	"github.com/ko1N/dips/pkg/client"
)

func main() {
	cl, err := client.NewClient("rabbitmq:rabbitmq@localhost")
	if err != nil {
		panic(err)
	}

	cl.
		NewWorker("shell").
		Handler(shellHandler).
		Run()

	cl.
		NewTask("shell").
		Name("test task").
		Parameters(map[string]interface{}{"param": "variant"}).
		Dispatch()

	fmt.Println("workers started")
	for {
		time.Sleep(1 * time.Second)
	}
}

func shellHandler(job *client.TaskContext) error {
	fmt.Printf("handling job %s: %s\n", job.TaskRequest.Name, job.TaskRequest.Params)
	return nil
}