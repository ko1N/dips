package environment

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path"

	"github.com/google/uuid"
	log "github.com/inconshreveable/log15"
)

// implements a native environment ("bare metal")

// NativeEnvironment -
type NativeEnvironment struct {
	PWD string
}

// CreateNativeEnvironment -
func CreateNativeEnvironment(pipelog log.Logger) (NativeEnvironment, error) {
	tempFolder := path.Join(".", "temp", uuid.New().String())
	err := os.MkdirAll(tempFolder, os.ModePerm)
	if err != nil {
		return NativeEnvironment{}, err
	}

	return NativeEnvironment{
		PWD: tempFolder,
	}, nil
}

// Name - returns the name of the native environment
func (e *NativeEnvironment) Name() string {
	return "native"
}

// Execute -
func (e *NativeEnvironment) Execute(cmd []string, stdout func(string), stderr func(string)) (ExecutionResult, error) {
	//fmt.Printf("exec: '%s'\n", strings.Join(cmd, " "))

	exc := exec.Command(cmd[0], cmd[1:]...)
	exc.Dir = e.PWD

	// create stdout/stderr pipes
	outpr, err := exc.StdoutPipe()
	if err != nil {
		return ExecutionResult{}, err
	}
	defer outpr.Close()
	outsig := make(chan struct{})
	var outBuf bytes.Buffer

	errpr, err := exc.StderrPipe()
	if err != nil {
		return ExecutionResult{}, err
	}
	defer errpr.Close()
	errsig := make(chan struct{})
	var errBuf bytes.Buffer

	// start process
	err = exc.Start()
	if err != nil {
		return ExecutionResult{}, err
	}

	// track stdout
	go func() {
		reader := bufio.NewScanner(outpr)
		for reader.Scan() {
			if stdout != nil {
				stdout(reader.Text())
			}
			outBuf.Write([]byte(reader.Text()))
			outBuf.Write([]byte("\n"))
		}
		outsig <- struct{}{}
	}()

	// track stderr
	go func() {
		reader := bufio.NewScanner(errpr)
		for reader.Scan() {
			if stderr != nil {
				stderr(reader.Text())
			}
			errBuf.Write([]byte(reader.Text()))
			errBuf.Write([]byte("\n"))
		}
		errsig <- struct{}{}
	}()

	// wait for exc to finish
	err = exc.Wait()
	if err != nil {
		return ExecutionResult{}, err
	}

	// synchropnize with stdout/stderr
	<-outsig
	<-errsig

	return ExecutionResult{
		ExitCode: exc.ProcessState.ExitCode(),
		StdOut:   outBuf.String(),
		StdErr:   errBuf.String(),
	}, nil
}

// Close - shuts down the environment and removes the temp folder
func (e *NativeEnvironment) Close() {
	if e.PWD != "" {
		os.RemoveAll(e.PWD)
	}
}
