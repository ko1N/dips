package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// TODO: this should be self-contained and not have a global state!

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

	// send pipeline to worker
	sendPipelineExecute <- string(body)

	c.JSON(http.StatusOK, SuccessResponse{
		Status: "pipeline created",
	})
}

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(r *gin.Engine, mq string) error {
	client := amqp.Create(mq)
	sendPipelineExecute = client.RegisterProducer("pipeline_execute")
	recvPipelineStatus = client.RegisterConsumer("pipeline_status")
	client.Start()

	r.POST("/manager/pipeline/execute", ExecutePipeline)

	return nil
}
