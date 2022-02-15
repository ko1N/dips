package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"github.com/ko1N/dips/pkg/client"
)

func main() {
	cl, err := client.NewClient("rabbitmq:rabbitmq@localhost")
	if err != nil {
		panic(err)
	}

	cl.
		NewTaskWorker("shell").
		// TODO: task timeout??
		Handler(shellHandler).
		Run()

	fmt.Println("shell worker started")

	signal := make(chan struct{})
	<-signal
}

func shellHandler(task *client.TaskContext) (map[string]interface{}, error) {
	fmt.Printf("handling 'shell' task %s: %s\n", task.Request.Name, task.Request.Params)

	//fmt.Printf("exec: `%s %s`\n", cmd, strings.Join(args, " "))
	exc := exec.Command("/bin/sh", "-c", task.Request.Params[""])

	// create stdout/stderr pipes
	stdoutpipe, err := exc.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrpipe, err := exc.StderrPipe()
	if err != nil {
		return nil, err
	}

	exc.Stdin = nil

	// start process
	err = exc.Start()
	if err != nil {
		return nil, err
	}

	// track stderr
	stderrsig := make(chan struct{})
	var errBuf bytes.Buffer
	go func() {
		reader := bufio.NewReader(stderrpipe)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			//if stderr != nil {
			//	stderr(strings.TrimSuffix(text, "\n"))
			//}
			errBuf.Write([]byte(text))
		}
		stderrsig <- struct{}{}
	}()

	// track stdout
	stdoutsig := make(chan struct{})
	var outBuf bytes.Buffer
	go func() {
		reader := bufio.NewReader(stdoutpipe)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			//if stdout != nil {
			//	stdout(strings.TrimSuffix(text, "\n"))
			//}
			outBuf.Write([]byte(text))
		}
		stdoutsig <- struct{}{}
	}()

	// wait for both pipes to be closed before calling wait
	<-stderrsig
	<-stdoutsig

	// wait for exc to finish
	err = exc.Wait()
	//if err != nil {
	//	return nil, err
	//}

	return map[string]interface{}{
		"rc":     exc.ProcessState.ExitCode(),
		"stdout": outBuf.String(),
		"stderr": errBuf.String(),
	}, nil
}
