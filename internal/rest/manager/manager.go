package manager

import (
	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/crud"
	"github.com/ko1N/dips/internal/persistence/messages"
	"github.com/ko1N/dips/pkg/client"
	"github.com/zebresel-com/mongodm"
)

// TODO: this should be self-contained and not have a global state!

// database
var pipelines crud.PipelineWrapper
var jobs crud.JobWrapper

var messageHandler messages.MessageHandler

// SuccessResponse - reponse for a successful operation
type SuccessResponse struct {
	Status string `json:"status"`
}

// FailureResponse - response for a failed operation
type FailureResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// ManagerAPIConfig - config required to run a manager
type ManagerAPIConfig struct {
	Gin            *gin.Engine
	Dips           *client.Client
	MongoDB        *mongodm.Connection
	MessageHandler messages.MessageHandler
}

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(conf ManagerAPIConfig) error {
	// setup crud wrappers
	pipelines = crud.CreatePipelineWrapper(conf.MongoDB)
	jobs = crud.CreateJobWrapper(conf.MongoDB)

	messageHandler = conf.MessageHandler

	conf.Dips.NewEventHandler().
		HandleMessage(handleMessage(conf.Dips)).
		HandleStatus(handleStatus(conf.Dips)).
		Run()

	// setup rest routes
	r := conf.Gin

	r.POST("/manager/pipeline/", PipelineCreate)
	r.GET("/manager/pipeline/all", PipelineList)
	r.GET("/manager/pipeline/details/:pipeline_id", PipelineDetails)
	r.PATCH("/manager/pipeline/:pipeline_id", PipelineUpdate)
	r.DELETE("/manager/pipeline/:pipeline_id", PipelineDelete)

	r.POST("/manager/pipeline/execute/:pipeline_id", PipelineExecute(conf.Dips))

	r.GET("/manager/job/all", JobList) // TODO: add more queries, running, finished, etc
	r.GET("/manager/job/details/:job_id", JobDetails)

	return nil
}
