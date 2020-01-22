package main

import (
	"io/ioutil"
	"log"

	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/workflow"
	"gitlab.strictlypaste.xyz/ko1n/transcode/pkg/workflow/modules"
)

func main() {
	data, err := ioutil.ReadFile("../../test/test.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	/*
		execution, err := workflow.Load(data)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		err = execution.Run()
		if err != nil {
			log.Fatalf("error: %v", err)
		}*/

	engine := workflow.CreateEngine()
	engine.
		RegisterExtension(modules.Shell{}).
		RegisterExtension(modules.FFMpeg{})

	workflow, err := workflow.CreateFromBytes(data)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// execute workflow on engine
	err = engine.ExecuteWorkflow(workflow)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// ...
}
