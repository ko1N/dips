package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// amqp channels
var sendPipelineExecute chan string
var recvPipelineStatus chan string

// ExecutePipeline - executes a pipeline
// @Summary executes a pipeline
// @Description This method will execute the pipeline sent via the post body
// @ID execute-pipeline
// @Accept plain
// @Produce json
// @Success 200
// @Router /pipeline/execute [post]
func ExecutePipeline(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "unable to read post body",
			"error":  err.Error(),
		})
		return
	}

	// pre-validate body
	_, err = pipeline.CreateFromBytes(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "unable to parse pipeline",
			"error":  err.Error(),
		})
		return
	}

	// send pipeline to worker
	sendPipelineExecute <- string(body)

	c.JSON(http.StatusOK, gin.H{
		"status": "pipeline created",
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
