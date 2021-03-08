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
		Handler(shellHandler)

	cl.Run()

	fmt.Println("workers started")
	for {
		time.Sleep(1 * time.Second)
	}
}

func shellHandler(job *client.Job) error {
	fmt.Println("handling job")
	return nil
}
