package manager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// list pipelines register pipelines, update pipelines, unregister pipelines

// PipelineCreate - creates a pipeline
// @Summary creates a pipeline
// @Description This method will create the pipeline sent via the post body
// @ID create-pipeline
// @Tags pipeline
// @Accept plain
// @Produce json
// @Param pipeline body string true "Pipeline Script"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/ [post]
func PipelineCreate(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to read post body",
			Error:  err.Error(),
		})
		return
	}

	// pre-validate body
	pi, err := pipeline.CreateFromBytes(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to parse pipeline",
			Error:  err.Error(),
		})
		return
	}

	// TODO: create func for this...
	// write pipeline to database
	pipeline := model.Pipeline{
		Script: string(body),
	}
	var taskID uint
	for _, stage := range pi.Stages {
		js := model.PipelineStage{
			Name: stage.Name,
		}
		for _, task := range stage.Tasks {
			js.Tasks = append(js.Tasks, &model.PipelineTask{
				ID:   taskID,
				Name: task.Name,
			})
			taskID++
		}
		pipeline.Stages = append(pipeline.Stages, &js)
	}

	err = pipelines.CreatePipeline(&pipeline)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to create database entry for pipeline",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Status: "pipeline created",
	})
}
