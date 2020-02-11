package main

import (
	"io/ioutil"
	"log"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/modules"
)

func main() {
	data, err := ioutil.ReadFile("../../test/test.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// create a global engine object for pipeline execution
	engine := pipeline.CreateEngine()
	engine.
		RegisterExtension(&modules.Shell{}).
		RegisterExtension(&modules.FFMpeg{})

	// create a new environment for this pipeline
	//var env environment.Environment = &environment.NativeEnvironment{}
	env, err := environment.MakeDockerEnvironment("alpine:latest")
	defer env.Close()

	pipeline, err := pipeline.CreateFromBytes(data)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// execute pipeline on engine
	err = engine.ExecutePipeline(&env, pipeline)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// ...
}
