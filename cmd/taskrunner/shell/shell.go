package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/inconshreveable/log15"
	"gopkg.in/yaml.v2"

	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/ko1N/dips/pkg/execution/tracking"
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
