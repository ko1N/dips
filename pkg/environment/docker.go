package environment

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// implements a dockerized environment

// DockerEnvironment -
type DockerEnvironment struct {
	ctx context.Context
	cli *client.Client
	id  string
}

// MakeDockerEnvironment -
func MakeDockerEnvironment(image string) (DockerEnvironment, error) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(
		ctx,
		fmt.Sprintf("docker.io/library/%s", image),
		types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"/bin/sh"},
		Tty:   true,
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return DockerEnvironment{
		ctx: ctx,
		cli: cli,
		id:  resp.ID,
	}, nil
}

// Execute - executes the given cmd inside a docker container
func (e *DockerEnvironment) Execute(cmd []string) (string, error) {
	fmt.Printf("exec: '%s'\n", strings.Join(cmd, " "))

	cfg := types.ExecConfig{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	exec, err := e.cli.ContainerExecCreate(e.ctx, e.id, cfg)
	if err != nil {
		return "", err
	}
	resp, err := e.cli.ContainerExecAttach(
		e.ctx,
		exec.ID,
		cfg,
	)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	buf, _, _ := resp.Reader.ReadLine()
	return strings.TrimSpace(string(buf)), nil
}

// Close - shuts down the environment and the corresponding container
func (e *DockerEnvironment) Close() {
	timeout := 5 * time.Second
	e.cli.ContainerStop(e.ctx, e.id, &timeout)
}
