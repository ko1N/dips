package main

import (
	"fmt"
	"strings"

	log "github.com/inconshreveable/log15"

	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/ko1N/dips/pkg/pipeline/tracking"
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
	tracker := tracking.CreateTaskTracker(
		log.New("cmd", "shell"),
		task.Client,
		task.Request.Job.Id.Hex(),
		task.Request.TaskID)

	executable := task.Request.Params["cmd"]
	cmdline := strings.Split(executable, " ")

	res, err := task.Environment.Execute(
		cmdline[0], append(cmdline[1:], []string{}...),
		func(outmsg string) {
			tracker.StdOut(outmsg)
		},
		func(errmsg string) {
			tracker.StdErr(errmsg)
		})
	if err != nil {
		tracker.Crit("unable to execute shell command: %s", err)
		return nil, err
	}

	return map[string]interface{}{
		"rc":     res.ExitCode,
		"stdout": res.StdOut,
		"stderr": res.StdErr,
	}, nil
}
