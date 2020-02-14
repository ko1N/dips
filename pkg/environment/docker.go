package environment

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// implements a dockerized environment

// DockerEnvironment -
type DockerEnvironment struct {
	ctx context.Context
	cli *client.Client
	id  string
}

// CreateDockerEnvironment -
func CreateDockerEnvironment(pipelog log.Logger, image string) (DockerEnvironment, error) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		pipelog.Crit("unable to create docker environment", err)
		return DockerEnvironment{}, err
	}

	_, err = cli.ImagePull(
		ctx,
		fmt.Sprintf("docker.io/library/%s", image),
		types.ImagePullOptions{})
	if err != nil {
		pipelog.Crit("unable to pull docker image `"+image+"`", err)
		return DockerEnvironment{}, err
	}
	//io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"/bin/sh"},
		Tty:   true,
	}, nil, nil, "")
	if err != nil {
		pipelog.Crit("unable to create container with image `"+image+"`", err)
		return DockerEnvironment{}, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		pipelog.Crit("unable to start container with image `"+image+"`", err)
		return DockerEnvironment{}, err
	}

	return DockerEnvironment{
		ctx: ctx,
		cli: cli,
		id:  resp.ID,
	}, nil
}

// Name - returns the name of the docker environment
func (e *DockerEnvironment) Name() string {
	return "docker"
}

// Execute - executes the given cmd inside a docker container
func (e *DockerEnvironment) Execute(cmd []string, stdout func(string), stderr func(string)) (ExecutionResult, error) {
	//fmt.Printf("exec: '%s'\n", strings.Join(cmd, " "))

	cfg := types.ExecConfig{
		Detach:       false,
		Tty:          false,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		User:         "root", // TODO: make configurable
	}

	exec, err := e.cli.ContainerExecCreate(e.ctx, e.id, cfg)
	if err != nil {
		return ExecutionResult{}, err
	}
	stream, err := e.cli.ContainerExecAttach(
		e.ctx,
		exec.ID,
		types.ExecConfig{
			Tty: false,
		},
	)
	if err != nil {
		return ExecutionResult{}, err
	}
	defer stream.Close()

	// create stdout/stderr pipes
	outpr, outpw := io.Pipe()
	outsig := make(chan struct{})
	var outBuf bytes.Buffer

	errpr, errpw := io.Pipe()
	errsig := make(chan struct{})
	var errBuf bytes.Buffer

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

	// run blocking stdCopy
	_, err = stdcopy.StdCopy(outpw, errpw, stream.Reader)

	// close the pipes regardless of an error in stdCopy
	outpw.Close()
	errpw.Close()

	// synchronize with stdout/stderr tracking
	<-outsig
	<-errsig

	// check error from stdCopy
	if err != nil {
		return ExecutionResult{}, err
	}

	// wait for the task to finish and return
	for {
		inspect, err := e.cli.ContainerExecInspect(e.ctx, exec.ID)
		if err != nil {
			return ExecutionResult{}, err
		}

		if !inspect.Running {
			return ExecutionResult{
				ExitCode: inspect.ExitCode,
				StdOut:   outBuf.String(),
				StdErr:   errBuf.String(),
			}, nil
		}

		time.Sleep(time.Second)
	}
}

// Close - shuts down the environment and the corresponding container
func (e *DockerEnvironment) Close() {
	timeout := 5 * time.Second
	e.cli.ContainerStop(e.ctx, e.id, &timeout)
}
