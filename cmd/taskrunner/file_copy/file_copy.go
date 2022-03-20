package main

import (
	"fmt"
	"io"
	"io/ioutil"

	log "github.com/inconshreveable/log15"
	"gopkg.in/yaml.v2"

	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/ko1N/dips/pkg/execution/tracking"
	"github.com/ko1N/dips/pkg/taskstorage"
)

type Config struct {
	Dips DipsConfig `yaml:"dips"`
}

type DipsConfig struct {
	Host string `yaml:"host"`
}

func readConfig(filename string) (*Config, error) {
	fallback := Config{
		Dips: DipsConfig{
			Host: "rabbitmq:rabbitmq@localhost",
		},
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return &fallback, nil
	}

	var conf Config
	err = yaml.Unmarshal([]byte(contents), &conf)
	if err != nil {
		return &fallback, nil
	}
	return &conf, nil
}

func main() {
	conf, err := readConfig("config.yml")
	if err != nil {
		panic(err)
	}

	cl, err := dipscl.NewClient(conf.Dips.Host)
	if err != nil {
		panic(err)
	}

	cl.
		NewTaskWorker("file_copy").
		// TODO: task timeout??
		Concurrency(100).
		//Environment("shell").
		Filesystem("disk").
		Handler(fileCopyHandler).
		Run()

	fmt.Println("file_copy worker started")

	signal := make(chan struct{})
	<-signal
}

func fileCopyHandler(task *dipscl.TaskContext) (map[string]interface{}, error) {
	tracker := tracking.CreateTaskTracker(
		log.New("cmd", "file_copy"),
		task.Client,
		task.Request.Job.Id.Hex(),
		task.Request.TaskID)

	source := task.Request.Params["source"]
	if source == "" {
		return nil, fmt.Errorf("`source` variable must not be empty")
	}

	target := task.Request.Params["target"]
	if target == "" {
		return nil, fmt.Errorf("`target` variable must not be empty")
	}

	sourceUrl, err := taskstorage.ParseFileUrl(source)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
	}

	targetUrl, err := taskstorage.ParseFileUrl(target)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url in `target` variable: %s", err.Error())
	}

	// source store
	tracker.Info("connecting to source storage at", "source", source)
	sourceStore, err := taskstorage.ConnectStorage(sourceUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source storage: %s", err.Error())
	}
	defer sourceStore.Close()

	// target store
	tracker.Info("connecting to target storage at", "target", target)
	targetStore, err := taskstorage.ConnectStorage(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target storage: %s", err.Error())
	}
	defer targetStore.Close()

	reader, err := sourceStore.GetFileReader(sourceUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get reader for source storage: %s", err.Error())
	}
	defer reader.Close()

	writer, err := targetStore.GetFileWriter(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get writer for target storage: %s", err.Error())
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy between storages: %s", err.Error())
	}

	tracker.Info("file copy successful")
	return map[string]interface{}{
		"target": target,
	}, nil
}
