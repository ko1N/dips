package main

import (
	"errors"
	"flag"
	"time"

	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/pkg/dipscl"
	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/tracking"

	"github.com/BurntSushi/toml"
	log "github.com/inconshreveable/log15"
)

type config struct {
	AMQP amqp.Config `json:"amqp" toml:"amqp"`
}

func main() {
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
	cl, err := dipscl.NewClient(conf.AMQP.Host)
	if err != nil {
		panic(err)
	}

	// TODO: configure concurrency, timeouts, etc
	cl.NewJobWorker().
		Concurrency(10).
		Handler(handleJob).
		Run()

	signal := make(chan struct{})
	<-signal
}

// TODO: send status updates containing log messages
// TODO: send status updates containing raw cmd exec log
func handleJob(job *dipscl.JobContext) error {
	// create logging instance for this pipeline
	tracker := tracking.CreateJobTracker(log.New("cmd", "worker"), job.Client, job.Request.Job.Id.Hex())

	pi, err := pipeline.CreateFromBytes(job.Request.Job.Pipeline.Script)
	if err != nil {
		tracker.Crit("unable to create pipeline from bytes", "error", err)
		return errors.New("Unable to create pipeline from bytes")
	}

	// execute pipeline on engine
	exec := pipeline.
		NewExecutionContext(job.Request.Job.Id.Hex(), pi, tracker).
		Variables(job.Request.Variables).
		TaskHandler(func(task *pipeline.Task, input map[string]string) (*pipeline.ExecutionResult, error) {
			retries := 3
			for {
				result, err := job.Client.
					NewTask(task.Service).
					Name(task.Name).
					Job(job.Request.Job).
					Timeout(10 * time.Second).
					Parameters(input).
					Dispatch().
					Await()
				if err != nil {
					if retries > 0 {
						retries--
						continue
					} else {
						return nil, err
					}
				}
				return &pipeline.ExecutionResult{
					Success: result.Error == nil,
					Error:   result.Error,
					Output:  result.Output,
				}, nil
			}
		})

	// run execution
	err = exec.Run()
	if err != nil {
		tracker.Crit("error while executing pipeline", "error", err)
	}

	return nil
}
