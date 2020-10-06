package environments

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/system"
	"github.com/ko1N/dips/pkg/pipeline/tracking"
)

// implements a dockerized environment

// DockerEnvironment -
type DockerEnvironment struct {
	ctx context.Context
	cli *client.Client
	id  string
}

// CreateDockerEnvironment -
func CreateDockerEnvironment(tracker tracking.JobTracker, image string) (DockerEnvironment, error) {
	tracker.Status("creating docker environment for image `" + image + "`")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		tracker.Error("unable to create docker environment", err)
		return DockerEnvironment{}, err
	}

	_, err = cli.ImagePull(
		ctx,
		fmt.Sprintf("docker.io/library/%s", image),
		types.ImagePullOptions{})
	if err != nil {
		tracker.Error("unable to pull docker image `"+image+"`. trying to use latest local image.", err)
		//return DockerEnvironment{}, err // TODO: handle docker login
	}
	//io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   []string{"/bin/sh"},
		Tty:   true,
	}, nil, nil, "")
	if err != nil {
		tracker.Error("unable to create container with image `"+image+"`", err)
		return DockerEnvironment{}, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		tracker.Error("unable to start container with image `"+image+"`", err)
		return DockerEnvironment{}, err
	}

	tracker.Status("created docker environment for image `" + image + "` with id `" + resp.ID + "`")
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
		//WorkingDir: ...
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

// see https://github.com/docker/cli/blob/2298e6a3fe24d3ac9276acdd35c24c06c8f58125/cli/command/utils.go#L159
// validateOutputPathFileMode validates the output paths of the `cp` command and serves as a
// helper to `ValidateOutputPath`
func validateOutputPathFileMode(fileMode os.FileMode) error {
	switch {
	case fileMode&os.ModeDevice != 0:
		return errors.New("got a device")
	case fileMode&os.ModeIrregular != 0:
		return errors.New("got an irregular file")
	}
	return nil
}

// CopyTo - copies a file to the docker container
func (e *DockerEnvironment) CopyTo(from string, to string) error {
	// see https://github.com/docker/cli/blob/master/cli/command/container/cp.go#L186

	// prepare source file info
	srcInfo, err := archive.CopyInfoSourcePath(from, false)
	if err != nil {
		return err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return err
	}
	defer srcArchive.Close()

	// prepare destination file info
	dstInfo := archive.CopyInfo{Path: to}
	dstStat, err := e.cli.ContainerStatPath(e.ctx, e.id, to)

	// if the destination is a symbolic link, we should evaluate it
	if err == nil && dstStat.Mode&os.ModeSymlink != 0 {
		linkTarget := dstStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// join with the parent directory
			dstParent, _ := archive.SplitPathDirEntry(to)
			linkTarget = filepath.Join(dstParent, linkTarget)
		}

		dstInfo.Path = linkTarget
		dstStat, err = e.cli.ContainerStatPath(e.ctx, e.id, linkTarget)
	}

	// validate the destination path
	if err := validateOutputPathFileMode(dstStat.Mode); err != nil {
		//return errors.Wrapf(err, "destination `%s` must be a directory or a regular file", to)
		return err
	}

	// prepare copy
	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return err
	}
	defer preparedArchive.Close()

	// copy the archive
	return e.cli.CopyToContainer(
		e.ctx,
		e.id,
		dstDir,
		preparedArchive,
		types.CopyToContainerOptions{
			AllowOverwriteDirWithFile: false,
			//CopyUIDGID:                false,
		})
}

// CopyFrom - copies a file from the docker container
func (e *DockerEnvironment) CopyFrom(from string, to string) error {
	// see https://github.com/docker/cli/blob/master/cli/command/container/cp.go#L119

	content, stat, err := e.cli.CopyFromContainer(e.ctx, e.id, from)
	if err != nil {
		return err
	}
	defer content.Close()

	srcInfo := archive.CopyInfo{
		Path:       from,
		Exists:     true,
		IsDir:      stat.Mode.IsDir(),
		RebaseName: "", // we ignore symlink following atm
	}

	preArchive := content
	if len(srcInfo.RebaseName) != 0 {
		_, srcBase := archive.SplitPathDirEntry(srcInfo.Path)
		preArchive = archive.RebaseArchiveEntries(content, srcBase, srcInfo.RebaseName)
	}
	return archive.CopyTo(preArchive, srcInfo, to)
}

// Close - shuts down the environment and the corresponding container
func (e *DockerEnvironment) Close() {
	timeout := 5 * time.Second
	e.cli.ContainerStop(e.ctx, e.id, &timeout)
}
