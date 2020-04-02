package main

import (
	"flag"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/inconshreveable/log15"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swagger generated docs
	_ "gitlab.strictlypaste.xyz/ko1n/dips/api/manager"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/messages"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/rest/manager"
)

// @title dips
// @version 0.1
// @description dips manager api

// @BasePath /

// generate swagger docs
//go:generate swag init -g manager.go --parseDependency --output ../../api/manager

// generate crud wrappers
//go:generate go run ../../internal/persistence/database/crud/generate_crud.go -type=model.Pipeline -output  ../../internal/persistence/database/crud/pipeline.go
//go:generate go run ../../internal/persistence/database/crud/generate_crud.go -type=model.Job -output  ../../internal/persistence/database/crud/job.go

type config struct {
	MongoDB  database.MongoDBConfig  `json:"mongodb" toml:"mongodb"`
	InfluxDB database.InfluxDBConfig `json:"influxdb" toml:"influxdb"`
	AMQP     amqp.Config             `json:"amqp" toml:"amqp"`
}

func main() {
	srvlog := log.New("cmd", "manager")

	// parse command line
	configPtr := flag.String("config", "config.toml", "config file")
	flag.Parse()

	// parse config
	var conf config
	if _, err := toml.DecodeFile(*configPtr, &conf); err != nil {
		srvlog.Crit("Config file could not be parsed", "error", err)
		return
	}

	// setup database
	mongodb, err := database.MongoDBConnect(conf.MongoDB)
	if err != nil {
		srvlog.Crit("Could not connect to mongodb instances", "error", err)
		return
	}

	// setup messages
	// setup logging
	influxdb, err := database.InfluxDBConnect(conf.InfluxDB)
	if err != nil {
		srvlog.Crit("Could not connect to influxdb instance", "error", err)
		return
	}

	messageHandler := messages.CreateMessageHandler(influxdb, conf.InfluxDB.Database)
	/*
		msg.Message("test", "Hello World 1!")
		msg.Message("test", "Hello World 2!")
		msg.Message("test", "Hello World 3!")

		msg.GetAll("test")
	*/

	// setup web service
	r := gin.Default()

	// add basic cors handling
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // TODO: secure me
	r.Use(cors.New(config))

	// setup manager api
	err = manager.CreateManagerAPI(manager.ManagerAPIConfig{
		Gin:            r,
		AMQP:           conf.AMQP,
		MongoDB:        mongodb,
		MessageHandler: messageHandler, // TODO: let managerAPI create its own db connections
	})
	if err != nil {
		srvlog.Crit("Unable to create Manager API", "error", err)
		return
	}

	// add swagger documentation on local dev builds
	mode := os.Getenv("GIN_MODE")
	if mode != "release" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		url := ginSwagger.URL("http://localhost:" + port + "/swagger/doc.json")
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
		srvlog.Info("Swagger setup at: http://localhost:" + port + "/swagger/index.html")
	}

	r.Run()
}
