package main

import (
	"io/ioutil"
	"log"

	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/environment"
	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/workflow"
	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/workflow/modules"
)

func main() {
	data, err := ioutil.ReadFile("../../test/test.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// create a global engine object for workflow execution
	engine := workflow.CreateEngine()
	engine.
		RegisterExtension(&modules.Shell{}).
		RegisterExtension(&modules.FFMpeg{})

	// create a new environment for this workflow
	//var env environment.Environment = &environment.NativeEnvironment{}
	env, err := environment.MakeDockerEnvironment("alpine:latest")
	defer env.Close()

	workflow, err := workflow.CreateFromBytes(data)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// execute workflow on engine
	err = engine.ExecuteWorkflow(&env, workflow)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// ...
}
