package main

import (
	"fmt"
	"strings"

	"github.com/ko1N/dips/pkg/dipscl"
)

func main() {
	cl, err := dipscl.NewClient("rabbitmq:rabbitmq@localhost")
	if err != nil {
		panic(err)
	}

	cl.
		NewTaskWorker("shell").
		// TODO: task timeout??
		Concurrency(100).
		//Environment("shell").
		Filesystem("disk").
		Handler(shellHandler).
		Run()

	fmt.Println("shell worker started")

	signal := make(chan struct{})
	<-signal
}

func shellHandler(task *dipscl.TaskContext) (map[string]interface{}, error) {
	fmt.Printf("handling 'shell' task %s: %s\n", task.Request.Name, task.Request.Params)

	executable := task.Request.Params[""]
	cmdline := strings.Split(executable, " ")

	res, err := task.Environment.Execute(
		cmdline[0], append(cmdline[1:], []string{}...),
		func(outmsg string) {
			//ctx.Tracker.Info(outmsg, "stream", "stdout")
			fmt.Printf("stdout: %s\n", outmsg)
		},
		func(errmsg string) {
			//ctx.Tracker.Info(errmsg, "stream", "stderr")
			fmt.Printf("stderr: %s\n", errmsg)
		})
	if err != nil {
		//ctx.Tracker.Crit("unable to execute video2x", "error", err)
		return nil, err
	}

	return map[string]interface{}{
		"rc":     res.ExitCode,
		"stdout": res.StdOut,
		"stderr": res.StdErr,
	}, nil
}
