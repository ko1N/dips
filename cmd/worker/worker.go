package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"time"

	"github.com/d5/tengo/v2"
	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"github.com/ko1N/dips/pkg/client"
	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/tracking"

	"github.com/BurntSushi/toml"
	log "github.com/inconshreveable/log15"
)

type config struct {
	AMQP amqp.Config `json:"amqp" toml:"amqp"`
}

func main() {
	pipelinePtr := flag.String("pipeline", "", "the pipeline to execute")
	flag.Parse()

	// create global logger for this instance
	srvlog := log.New("cmd", "worker")

	// parse command line
	configPtr := flag.String("config", "config.toml", "config file")
	flag.Parse()

	// parse config
	var conf config
	if _, err := toml.DecodeFile(*configPtr, &conf); err != nil {
		srvlog.Crit("Config file could not be parsed", "error", err)
		return
	}

	// setup amqp
	cl, err := client.NewClient(conf.AMQP.Host)
	if err != nil {
		panic(err)
	}

	// TODO: configure concurrency
	cl.NewJobWorker().
		Handler(handleJob).
		Run()

		// TEST
	// parse pipeline
	content, err := ioutil.ReadFile(*pipelinePtr)
	if err != nil {
		srvlog.Crit("unable to open pipeline script file", "error", err)
	}
	cl.NewJob().
		Job(&model.Job{
			Name: "test",
			Pipeline: &model.Pipeline{
				Script: content,
			},
		}).
		Dispatch()

	for {
		time.Sleep(1 * time.Second)
	}
}

// TODO: send status updates containing log messages
// TODO: send status updates containing raw cmd exec log
func handleJob(job *client.JobContext) error {
	// create logging instance for this pipeline
	tracker := tracking.CreateJobTracker(&tracking.JobTrackerConfig{
		Logger: log.New("cmd", "worker"), // TODO: dep injection
		Client: job.Client,
		JobID:  job.Request.Job.Id.Hex(),
	})

	pi, err := pipeline.CreateFromBytes(job.Request.Job.Pipeline.Script)
	if err != nil {
		tracker.Logger().Crit("unable to create pipeline from bytes", "error", err)
		return errors.New("Unable to create pipeline from bytes")
	}

	// execute pipeline on engine
	exec := pipeline.NewExecutionContext(
		job.Request.Job.Id.Hex(),
		pi,
		tracker).
		TaskHandler(func(task *pipeline.Task) error {
			job.Client.NewTask(task.Service).
				Name(task.Name).
				//Parameters(task.). // TODO: parameters
				Dispatch()
			return nil
		})

	// setup parameters
	for paramName, paramValue := range job.Request.Params {
		exec.Variables[paramName] = &tengo.String{Value: paramValue.(string)}
	}

	// run execution
	err = exec.Run()
	if err != nil {
		tracker.Logger().Crit("unable to execute pipeline", "error", err)
	}

	return nil
}
