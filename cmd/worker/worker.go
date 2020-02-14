package main

import (
	"io/ioutil"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/modules"

	log "github.com/inconshreveable/log15"
)

func main() {
	// create global logger for this instance
	srvlog := log.New("cmd", "worker")

	data, err := ioutil.ReadFile("../../test/test.yml")
	if err != nil {
		srvlog.Crit("unable to open test.yml")
		return
	}

	// create a global engine object for pipeline execution
	engine := pipeline.CreateEngine()
	engine.
		RegisterExtension(&modules.Shell{}).
		RegisterExtension(&modules.FFMpeg{})

	// create logging instance for this pipeline
	pipelog := srvlog.New("pipeline", "test.yml") // TODO: generate ID
	pipelog.Info("pipeline created")

	// create a new environment for this pipeline
	//var env environment.Environment = &environment.NativeEnvironment{}
	env, err := environment.CreateDockerEnvironment(pipelog, "alpine:latest")
	if err != nil {
		pipelog.Info("unable to start docker environment", err)
		return
	}
	defer env.Close()

	pipeline, err := pipeline.CreateFromBytes(data)
	if err != nil {
		pipelog.Info("unable to parse pipeline file", err)
		return
	}

	// execute pipeline on engine
	err = engine.ExecutePipeline(pipelog, &env, pipeline)
	if err != nil {
		pipelog.Info("unable to execute pipeline", err)
		return
	}

	// ...
}
