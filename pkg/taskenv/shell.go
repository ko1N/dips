package taskenv

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/ko1N/dips/pkg/taskfs"
)

// implements a native environment ("bare metal")

// NativeEnvironment -
type NativeEnvironment struct {
	fs taskfs.FileSystem
}

// CreateNativeEnvironment -
func CreateNativeEnvironment(fs taskfs.FileSystem) (*NativeEnvironment, error) {
	return &NativeEnvironment{
		fs: fs,
	}, nil
}

// Execute -
func (e *NativeEnvironment) Execute(cmd string, args []string, stdout func(string), stderr func(string)) (*ExecutionResult, error) {
	//fmt.Printf("exec: `%s %s`\n", cmd, strings.Join(args, " "))

	exc := exec.Command(cmd, args...)
	fullPath := e.fs.RootPath()
	exc.Dir = fullPath

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
			if stderr != nil {
				stderr(strings.TrimSuffix(text, "\n"))
			}
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
			if stdout != nil {
				stdout(strings.TrimSuffix(text, "\n"))
			}
			outBuf.Write([]byte(text))
		}
		stdoutsig <- struct{}{}
	}()

	// wait for both pipes to be closed before calling wait
	<-stderrsig
	<-stdoutsig

	// wait for exc to finish
	err = exc.Wait()
	if err != nil {
		return nil, err
	}

	return &ExecutionResult{
		ExitCode: exc.ProcessState.ExitCode(),
		StdOut:   outBuf.String(),
		StdErr:   errBuf.String(),
	}, nil
}

// Close - shuts down the environment (no-op)
func (e *NativeEnvironment) Close() error {
	return nil
}
