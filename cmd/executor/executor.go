package main

import (
	"flag"
	"io/ioutil"

	log "github.com/inconshreveable/log15"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/modules"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline/tracking"
)

func main() {
	pipelinePtr := flag.String("pipeline", "", "the pipeline to execute")
	flag.Parse()

	// setup engine
	srvlog := log.New("cmd", "worker")

	// TODO: should the logger be inside engine?
	engine := pipeline.CreateEngine()
	engine.
		RegisterExtension(&modules.Shell{}).
		RegisterExtension(&modules.WGet{}).
		RegisterExtension(&modules.FFMpeg{})

	// create logging instance for this pipeline
	tracker := tracking.CreateJobTracker(tracking.JobTrackerConfig{
		Logger:          srvlog,
		ProgressChannel: nil,
		MessageChannel:  nil,
		JobID:           "manual",
	})

	// parse pipeline
	content, err := ioutil.ReadFile(*pipelinePtr)
	if err != nil {
		srvlog.Crit("unable to open pipeline script file", "error", err)
	}

	pi, err := pipeline.CreateFromBytes(content)
	if err != nil {
		srvlog.Crit("unable to create pipeline from bytes", "error", err)
	}

	// execute pipeline on engine
	exec := engine.CreateExecution(
		"manual",
		&pi,
		tracker)
	err = exec.Run()
	if err != nil {
		srvlog.Crit("unable to execute pipeline", "error", err)
		return
	}
}
