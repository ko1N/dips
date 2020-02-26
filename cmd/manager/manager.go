package main

import (
	"flag"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/inconshreveable/log15"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zebresel-com/mongodm"

	// swagger generated docs
	_ "gitlab.strictlypaste.xyz/ko1n/dips/api/manager"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/rest"
)

// @title DIPS
// @version 0.1
// @description DIPS Manager API

// @BasePath /

// generate swagger docs
//go:generate swag init -g manager.go --parseDependency --output ../../api/manager

// Job - Database struct describing a pipeline job
type Job struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Pipeline             string `json:"pipeline"  bson:"pipeline" required:"true"`
}

// some test code
func createJob(db *mongodm.Connection, pipeline string) {
	jobModel := db.Model("Job")

	job := &Job{}
	jobModel.New(job)

	job.Pipeline = pipeline

	job.Save()
}

func main() {
	srvlog := log.New("cmd", "manager")

	// parse command line
	dbLangPtr := flag.String("dblang", "", "selected language")
	dbHostPtr := flag.String("dbhost", "localhost", "mongodb host")
	dbName := flag.String("dbname", "dips", "database name")
	dbUser := flag.String("dbuser", "dips", "database username")
	dbPass := flag.String("dbpass", "dips", "database password")

	// setup database
	db, err := persistence.Connect(*dbLangPtr, []string{*dbHostPtr}, *dbName, *dbUser, *dbPass)
	if err != nil {
		srvlog.Crit("Database connection could not be established", "error", err)
		return
	}

	// registration
	db.Register(&Job{}, "jobs")

	persistence.CreateCrudWrapper(db, &Job{})

	// test
	createJob(db, "test123")

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
