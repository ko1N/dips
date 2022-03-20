package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopkg.in/yaml.v2"

	// swagger generated docs
	_ "github.com/ko1N/dips/api/manager"
	"github.com/ko1N/dips/internal/persistence/database"
	"github.com/ko1N/dips/internal/persistence/messages"
	"github.com/ko1N/dips/internal/rest/manager"
	"github.com/ko1N/dips/pkg/dipscl"
)

// @title dips
// @version 0.1
// @description dips manager api

// @BasePath /

// generate swagger docs
//go:generate swag init -g manager.go --parseDependency --output ../../api/manager

type Config struct {
	Dips     DipsConfig              `yaml:"dips"`
	MongoDB  database.MongoDBConfig  `yaml:"mongodb"`
	InfluxDB database.InfluxDBConfig `yaml:"influxdb"`
}

type DipsConfig struct {
	Host string `yaml:"host"`
}

func readConfig(filename string) (*Config, error) {
	fallback := Config{
		Dips: DipsConfig{
			Host: "rabbitmq:rabbitmq@localhost",
		},
		MongoDB: database.MongoDBConfig{
			Hosts:         []string{"mongodb://localhost:27017"},
			AuthMechanism: "SCRAM-SHA-256",
			AuthSource:    "dips",
			Username:      "dips",
			Password:      "dips",
			Database:      "dips",
		},
		InfluxDB: database.InfluxDBConfig{
			Host:     "http://localhost:8086",
			Database: "dips",
			Username: "dips",
			Password: "dips",
		},
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return &fallback, nil
	}

	var conf Config
	err = yaml.Unmarshal([]byte(contents), &conf)
	if err != nil {
		return &fallback, nil
	}
	return &conf, nil
}

func main() {
	conf, err := readConfig("config.yml")
	if err != nil {
		panic(err)
	}

	// setup dips client
	cl, err := dipscl.NewClient(conf.Dips.Host)
	if err != nil {
		panic(err)
	}

	// setup database
	mongo, err := database.MongoDBConnect(&conf.MongoDB)
	if err != nil {
		panic(err)
	}

	// setup messages
	// setup logging
	influxdb, err := database.InfluxDBConnect(&conf.InfluxDB)
	if err != nil {
		panic(err)
	}

	// setup web service
	r := gin.Default()

	// add basic cors handling
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // TODO: secure me
	r.Use(cors.New(config))

	// setup manager api
	messageHandler := messages.CreateMessageHandler(influxdb, conf.InfluxDB.Database)
	_, err = manager.CreateManagerAPI(r, cl, mongo, messageHandler)
	if err != nil {
		panic(err)
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
		fmt.Println("Swagger setup at: http://localhost:" + port + "/swagger/index.html")
	}

	r.Run()
}
