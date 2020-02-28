package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/crud"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/model"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// TODO: this should be self-contained and not have a global state!

// database
var jobs crud.JobWrapper

// amqp channels
var sendPipelineExecute chan string
var recvPipelineStatus chan string

// SuccessResponse - reponse for a successful operation
type SuccessResponse struct {
	Status string `json:"status"`
}

// FailureResponse - response for a failed operation
type FailureResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// ExecutePipeline - executes a pipeline
// @Summary executes a pipeline
// @Description This method will execute the pipeline sent via the post body
// @ID execute-pipeline
// @Accept plain
// @Produce json
// @Param pipeline body string true "Pipeline Script"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/execute [post]
func ExecutePipeline(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to read post body",
			Error:  err.Error(),
		})
		return
	}

	// pre-validate body
	_, err = pipeline.CreateFromBytes(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to parse pipeline",
			Error:  err.Error(),
		})
		return
	}

	// write pipeline to database
	job := model.Job{
		Pipeline: string(body),
	}
	jobs.CreateJob(&job)

	// use job id

	// send pipeline to worker
	sendPipelineExecute <- string(body)

	c.JSON(http.StatusOK, SuccessResponse{
		Status: "pipeline created",
	})
}

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(r *gin.Engine, db *mongodm.Connection, mq string) error {
	// setup database
	jobs = crud.CreateJobWrapper(db)

	// setup amqp
	client := amqp.Create(mq)
	sendPipelineExecute = client.RegisterProducer("pipeline_execute")
	recvPipelineStatus = client.RegisterConsumer("pipeline_status")
	client.Start()

	// setup rest routes
	r.POST("/manager/pipeline/execute", ExecutePipeline)
	// /manager/pipeline/list
	// /manager/pipeline/info/{id}

	return nil
}
