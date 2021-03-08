package main

import (
	"encoding/json"
	"flag"

	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/internal/rest/manager"
	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/tracking"

	"github.com/BurntSushi/toml"
	"github.com/d5/tengo/v2"
	log "github.com/inconshreveable/log15"
)

type config struct {
	AMQP amqp.Config `json:"amqp" toml:"amqp"`
}

// amqp channels
var recvPipelineExecute chan string
var sendJobStatus chan string
var sendJobMessage chan string

// TODO: send status updates containing log messages
// TODO: send status updates containing raw cmd exec log
func executePipeline(srvlog log.Logger, payload string) {
	msg := manager.ExecutePipelineMessage{}
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		srvlog.Crit("unable to unmarshal payload", "error", err)
		return
	}

	// create logging instance for this pipeline
	tracker := tracking.CreateJobTracker(tracking.JobTrackerConfig{
		Logger:          srvlog,
		ProgressChannel: sendJobStatus,
		MessageChannel:  sendJobMessage,
		JobID:           msg.Job.Id.Hex(),
	})

	// execute pipeline on engine
	exec := pipeline.NewExecutionContext(
		msg.Job.Id.Hex(),
		msg.Job.Pipeline.Pipeline,
		tracker)

	// setup parameters
	for _, param := range msg.Job.Parameters {
		exec.Variables[param.Name] = &tengo.String{Value: param.Value}
	}

	// run execution
	err := exec.Run()
	if err != nil {
		tracker.Logger().Crit("unable to execute pipeline", "error", err)
		return
	}
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
	client := amqp.NewAMQP(conf.AMQP)
	recvPipelineExecute = client.RegisterConsumer("pipeline_execute")
	sendJobStatus = client.RegisterProducer("job_status")
	sendJobMessage = client.RegisterProducer("job_message")

	for payload := range recvPipelineExecute {
		go executePipeline(srvlog, payload)
	}
}
