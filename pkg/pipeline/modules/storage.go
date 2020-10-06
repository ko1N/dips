package modules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ko1N/dips/internal/persistence/storage"
	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/environments"

	"github.com/google/uuid"
)

// pipeline module for storage

// Storage -
type Storage struct {
	Storage storage.Storage
}

// Name -
func (e *Storage) Name() string {
	return "Storage"
}

// Command -
func (e *Storage) Command() string {
	return "storage"
}

// StartPipeline -
func (e *Storage) StartPipeline(ctx *pipeline.ExecutionContext) error {
	ctx.Tracker.Status("creating storage bucket `" + ctx.JobID + "`")
	return e.Storage.CreateBucket(ctx.JobID)
}

// FinishPipeline -
func (e *Storage) FinishPipeline(ctx *pipeline.ExecutionContext) error {
	ctx.Tracker.Status("deleting storage bucket `" + ctx.JobID + "`")
	return e.Storage.DeleteBucket(ctx.JobID)
}

// Execute - Executes the given storage command
func (e *Storage) Execute(ctx *pipeline.ExecutionContext, cmd string) (environments.ExecutionResult, error) {
	cmdSplit := strings.Split(cmd, " ")
	switch cmdSplit[0] {
	case "ls":
		if err := e.listFiles(ctx, cmdSplit[1:]); err != nil {
			ctx.Tracker.Error("invalid storage command: `"+cmd+"`", err)
		}
		break
	case "get":
		if err := e.getFile(ctx, cmdSplit[1:]); err != nil {
			ctx.Tracker.Error("invalid storage command: `"+cmd+"`", err)
		}
		break
	case "put":
		if err := e.putFile(ctx, cmdSplit[1:]); err != nil {
			ctx.Tracker.Error("invalid storage command: `"+cmd+"`", err)
		}
		break
	default:
		ctx.Tracker.Error("invalid storage command: `"+cmd+"`. usage: ls/get/put", nil)
		break
	}

	return environments.ExecutionResult{}, nil
}

// TODO: get this back into the execution context / variables?
func (e *Storage) listFiles(ctx *pipeline.ExecutionContext, args []string) error {
	if len(args) > 0 {
		return errors.New("ls command does not have any parameters")
	}

	//tracker.Logger().Info("ls valid`" + cmd + "`")
	files, err := e.Storage.List(ctx.JobID)
	if err != nil {
		return err
	}

	for _, file := range files {
		fmt.Printf("file: %+v\n", file)
	}

	return nil
}

func (e *Storage) getFile(ctx *pipeline.ExecutionContext, args []string) error {
	// TODO: handle multiple files (wildcards?)
	// TODO: handle get [filename] [outfilename]
	if len(args) != 1 {
		return errors.New("get command does not have enough parameters. usage: get [filename]")
	}

	tempFileName := uuid.New().String()
	tempFile := "/tmp/" + ctx.JobID + "/" + tempFileName
	err := e.Storage.GetFile(ctx.JobID, filepath.Base(args[0]), tempFile)
	if err != nil {
		ctx.Tracker.Error("unable to copy file `"+tempFile+"` from storage", err)
		return err
	}

	// TODO: get proper PWD and track it in pipeline
	// copy file to env
	err = ctx.Environment.CopyTo(tempFile, args[0]) // TODO: get pwd?
	if err != nil {
		ctx.Tracker.Error("unable to copy file from temporary folder to environment `"+args[0]+"`", err)
		return err
	}

	return os.Remove(tempFile)
}

func (e *Storage) putFile(ctx *pipeline.ExecutionContext, args []string) error {
	// TODO: handle multiple files (wildcards?)
	// TODO: handle put [filename] [outfilename]
	if len(args) != 1 {
		return errors.New("put command does not have enough parameters. usage: put [filename]")
	}

	// TODO: get proper PWD and track it in pipeline
	// extract file from env
	tempFolder := "/tmp/" + ctx.JobID
	err := os.MkdirAll(tempFolder, os.ModePerm)
	if err != nil {
		ctx.Tracker.Error("unable to create temporary folder `"+tempFolder+"`", err)
		return err
	}

	tempFileName := uuid.New().String()
	tempFile := tempFolder + "/" + tempFileName
	err = ctx.Environment.CopyFrom(args[0], tempFile)
	if err != nil {
		ctx.Tracker.Error("unable to copy file from environment to temporary folder `"+tempFolder+"`", err)
		return err
	}

	err = e.Storage.PutFile(ctx.JobID, tempFile, filepath.Base(args[0]))
	if err != nil {
		ctx.Tracker.Error("unable to copy temporary file `"+args[0]+"` to storage", err)
		return err
	}

	return nil
}
