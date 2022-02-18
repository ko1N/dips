package manager

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"github.com/ko1N/dips/pkg/client"
)

// PipelineExecuteRequest - Request Body when executing a pipeline
type PipelineExecuteRequest struct {
	Name      string                 `json:"name"`
	Variables map[string]interface{} `json:"variables"`
}

// PipelineExecute - executes a pipeline
// @Summary executes a pipeline
// @Description This method will execute the pipeline with the given id
// @ID pipeline-execute
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipeline_id path string true "Pipeline ID"
// @Param execute_request body PipelineExecuteRequest true "Request Body"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/execute/{pipeline_id} [post]
func PipelineExecute(cl *client.Client) func(*gin.Context) {
	return func(c *gin.Context) {
		// try to find requested pipeline
		pipelineID := c.Param("pipeline_id")
		if pipelineID == "" {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "invalid pipeline_id",
				Error:  "pipeline_id must not be empty",
			})
			return
		}

		pipe, err := pipelines.FindPipelineByID(pipelineID)
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "unable to find pipeline with id `" + pipelineID + "`",
				Error:  err.Error(),
			})
			return
		}

		// try to parse body
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "unable to read body",
				Error:  err.Error(),
			})
			return
		}

		var request PipelineExecuteRequest
		err = json.Unmarshal(body, &request)
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "unable to parse request body",
				Error:  err.Error(),
			})
			return
		}

		// create the job in the database
		job := model.Job{
			Name:      request.Name,
			Variables: request.Variables,
			Pipeline:  pipe,
		}

		err = jobs.CreateJob(&job)
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "unable to create database entry for job",
				Error:  err.Error(),
			})
			return
		}

		// send pipeline to worker
		cl.NewJob().
			Job(&job).
			Dispatch()

		// return success
		c.JSON(http.StatusOK, SuccessResponse{
			Status: "pipeline job created",
		})
	}
}
