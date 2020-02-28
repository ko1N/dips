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
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/rest"
)

// @title dips
// @version 0.1
// @description dips manager api

// @BasePath /

// generate swagger docs
//go:generate swag init -g manager.go --parseDependency --output ../../api/manager

// generate crud wrappers
//go:generate go run ../../internal/persistence/crud/generate_crud.go -type=model.Job -output  ../../internal/persistence/crud/job.go

type config struct {
	Database persistence.Config `json:"db" toml:"db"`
	AMQP     amqp.Config        `json:"amqp" toml:"amqp"`
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
	db, err := persistence.Connect(conf.Database)
	if err != nil {
		srvlog.Crit("Database connection could not be established", "error", err)
		return
	}

	// setup web service
	r := gin.Default()

	// add basic cors handling
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // TODO: secure me
	r.Use(cors.New(config))

	// setup manager api
	rest.CreateManagerAPI(r, db, conf.AMQP)

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
