package environment

import (
	"bytes"
	"context"
	"fmt"
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

// Execute - executes the given cmd inside a docker container
func (e *DockerEnvironment) Execute(cmd []string) (ExecutionResult, error) {
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

	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&stdOutBuf, &stdErrBuf, stream.Reader)
	if err != nil {
		return ExecutionResult{}, err
	}

	for {
		inspect, err := e.cli.ContainerExecInspect(e.ctx, exec.ID)
		if err != nil {
			return ExecutionResult{}, err
		}

		if !inspect.Running {
			return ExecutionResult{
				ExitCode: inspect.ExitCode,
				StdOut:   stdOutBuf.String(),
				StdErr:   stdErrBuf.String(),
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
