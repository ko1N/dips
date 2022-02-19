package manager

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/messages"
	"github.com/ko1N/dips/pkg/client"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: this should be self-contained and not have a global state!
const mongoTimeout = 5 * time.Second
const colPipeline = "pipeline"
const colJobs = "jobs"

// SuccessResponse - reponse for a successful operation
type SuccessResponse struct {
	Status string `json:"status"`
}

// FailureResponse - response for a failed operation
type FailureResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ManagerAPI struct {
	gin            *gin.Engine
	dipscl         *client.Client
	mongo          *mongo.Database
	messageHandler messages.MessageHandler
}

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(
	gin *gin.Engine,
	dipscl *client.Client,
	mongo *mongo.Database,
	messageHandler messages.MessageHandler,
) (*ManagerAPI, error) {
	api := &ManagerAPI{
		gin,
		dipscl,
		mongo,
		messageHandler,
	}

	// register event handlers
	api.dipscl.NewEventHandler().
		HandleMessage(api.handleMessage).
		HandleStatus(api.handleStatus).
		Run()

	// setup rest routes
	r := api.gin

	r.POST("/manager/pipeline/", api.PipelineCreate)
	r.GET("/manager/pipeline/all", api.PipelineList)
	r.GET("/manager/pipeline/details/:pipeline_id", api.PipelineDetails)
	r.PATCH("/manager/pipeline/:pipeline_id", api.PipelineUpdate)
	r.DELETE("/manager/pipeline/:pipeline_id", api.PipelineDelete)

	r.POST("/manager/pipeline/execute/:pipeline_id", api.PipelineExecute)

	r.GET("/manager/job/all", api.JobList) // TODO: add more queries, running, finished, etc
	r.GET("/manager/job/details/:job_id", api.JobDetails)

	return api, nil
}
