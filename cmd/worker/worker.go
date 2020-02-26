package main

import (
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/modules"

	"github.com/google/uuid"
	log "github.com/inconshreveable/log15"
)

// amqp channels
var recvPipelineExecute chan string
var sendPipelineStatus chan string

// TODO: send status updates containing log messages
// TODO: send status updates containing raw cmd exec log
func executePipeline(srvlog log.Logger, engine *pipeline.Engine, payload string) {
	// create logging instance for this pipeline
	id := uuid.New().String()
	pipelog := srvlog.New("pipeline", id)
	pipelog.Info("pipeline `" + id + "` created")

	// parse pipeline script
	pipeline, err := pipeline.CreateFromBytes([]byte(payload))
	if err != nil {
		pipelog.Info("unable to parse pipeline file", err)
		return
	}

	// execute pipeline on engine
	err = engine.ExecutePipeline(pipelog, pipeline)
	if err != nil {
		pipelog.Info("unable to execute pipeline", err)
		return
	}
}

func main() {
	// create global logger for this instance
	srvlog := log.New("cmd", "worker")

	// create a global engine object for pipeline execution
	engine := pipeline.CreateEngine()
	engine.
		RegisterExtension(&modules.Shell{}).
		RegisterExtension(&modules.WGet{}).
		RegisterExtension(&modules.FFMpeg{})

	// setup amqp
	client := amqp.Create("rabbitmq:rabbitmq@localhost")
	recvPipelineExecute = client.RegisterConsumer("pipeline_execute")
	sendPipelineStatus = client.RegisterProducer("pipeline_status")
	client.Start()

	for payload := range recvPipelineExecute {
		go executePipeline(srvlog, &engine, payload)
	}
}
