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
		NewTaskWorker("shell").
		Handler(shellHandler).
		Run()

	fmt.Println("shell worker started")

	signal := make(chan struct{})
	<-signal
}

func shellHandler(task *client.TaskContext) error {
	fmt.Printf("handling 'shell' task %s: %s\n", task.Request.Name, task.Request.Params)
	time.Sleep(1 * time.Second)
	return nil
}
