package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/inconshreveable/log15"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swagger generated docs
	_ "gitlab.strictlypaste.xyz/ko1n/dips/cmd/manager/docs"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/rest"
)

// @title DIPS
// @version 0.1
// @description DIPS Manager API

// @BasePath /

// generate swagger docs
//go:generate swag init --parseDependency

func main() {
	srvlog := log.New("cmd", "manager")

	r := gin.Default()

	// add basic cors handling
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // TODO: secure me
	r.Use(cors.New(config))

	// setup manager api
	rest.CreateManagerAPI(r, "rabbitmq:rabbitmq@localhost")

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
