package main

import (
	"flag"
	"io/ioutil"

	log "github.com/inconshreveable/log15"

	"github.com/ko1N/dips/pkg/pipeline"
	"github.com/ko1N/dips/pkg/pipeline/tracking"
)

func main() {
	pipelinePtr := flag.String("pipeline", "", "the pipeline to execute")
	flag.Parse()

	// setup engine
	srvlog := log.New("cmd", "worker")

	// create logging instance for this pipeline
	tracker := tracking.CreateJobTracker(tracking.JobTrackerConfig{
		Logger: srvlog,
		Client: nil,
		JobID:  "manual",
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
	exec := pipeline.NewExecutionContext(
		"manual",
		pi,
		tracker)
	err = exec.Run()
	if err != nil {
		srvlog.Crit("unable to execute pipeline", "error", err)
		return
	}
}
