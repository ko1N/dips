package main

import (
	"encoding/json"
	"flag"

	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/storage"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/rest/manager"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/modules"

	"github.com/BurntSushi/toml"
	log "github.com/inconshreveable/log15"
)

type config struct {
	AMQP    amqp.Config          `json:"amqp" toml:"amqp"`
	Storage *storage.MinIOConfig `json:"storage" toml:"storage"`
}

// amqp channels
var recvJobExecute chan string
var sendJobStatus chan string

// TODO: send status updates containing log messages
// TODO: send status updates containing raw cmd exec log
func executePipeline(srvlog log.Logger, engine *pipeline.Engine, payload string) {
	msg := manager.ExecuteJobMessage{}
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		srvlog.Crit("unable to unmarshal payload", "error", err)
		return
	}

	// create logging instance for this pipeline
	tracker := pipeline.CreateJobTracker(srvlog, sendJobStatus, msg.ID)

	// parse pipeline script
	pipe, err := pipeline.CreateFromBytes([]byte(msg.Pipeline))
	if err != nil {
		tracker.Logger().Crit("unable to parse pipeline file", "error", err)
		return
	}

	// execute pipeline on engine
	exec := engine.CreateExecution(msg.ID, pipe, tracker)
	err = exec.Run()
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

	// create a global engine object for pipeline execution
	engine := pipeline.CreateEngine()
	engine.
		RegisterExtension(&modules.Shell{}).
		RegisterExtension(&modules.WGet{}).
		RegisterExtension(&modules.FFMpeg{})

	// setup storage
	if conf.Storage != nil {
		store, err := storage.ConnectMinIO(*conf.Storage)
		if err != nil {
			srvlog.Crit("Could not connect to minio storage server", "error", err)
			return
		}

		engine.RegisterExtension(&modules.Storage{Storage: &store})
	}

	// setup amqp
	client := amqp.Create(conf.AMQP)
	recvJobExecute = client.RegisterConsumer("pipeline_execute")
	sendJobStatus = client.RegisterProducer("pipeline_status")
	client.Start()

	for payload := range recvJobExecute {
		go executePipeline(srvlog, &engine, payload)
	}
}
