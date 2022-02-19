package main

import (
	"flag"
	"io/ioutil"

	"github.com/ko1N/dips/internal/amqp"
	"github.com/ko1N/dips/pkg/client"

	"github.com/BurntSushi/toml"
	log "github.com/inconshreveable/log15"
)

type config struct {
	AMQP amqp.Config `json:"amqp" toml:"amqp"`
}

func main() {
	pipelinePtr := flag.String("pipeline", "", "the pipeline to execute")
	flag.Parse()

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

	// setup dips client
	cl, err := client.NewClient(conf.AMQP.Host)
	if err != nil {
		panic(err)
	}

	// execute pipeline
	contents, err := ioutil.ReadFile(*pipelinePtr)
	if err != nil {
		srvlog.Crit("unable to open pipeline script file", "error", err)
	}
	for i := 1; i < 1000; i++ {
		go func() {
			cl.NewJob().
				Name("test").
				Pipeline(contents).
				Variables(map[string]interface{}{
					"filename": "minio://minio:miniominio@172.17.0.1:9000/test/test.mp4",
				}).
				Dispatch()
		}()
	}

	signal := make(chan struct{})
	<-signal
}
