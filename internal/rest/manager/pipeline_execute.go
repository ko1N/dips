package manager

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
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
func (a *ManagerAPI) PipelineExecute(c *gin.Context) {
	// try to find requested pipeline
	pipelineID := c.Param("pipeline_id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline_id",
			Error:  "pipeline_id must not be empty",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	fres := a.mongo.
		Collection(colPipeline).
		FindOne(ctx, bson.D{{"_id", pipelineID}})
	if fres.Err() != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + pipelineID + "`",
			Error:  fres.Err().Error(),
		})
		return
	}
	var pipeline model.Pipeline
	err := fres.Decode(&pipeline)
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
		Pipeline:  &pipeline,
	}

	ires, err := a.mongo.
		Collection(colJobs).
		InsertOne(ctx, &job)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to create database entry for job",
			Error:  err.Error(),
		})
		return
	}

	// send pipeline to worker
	a.dipscl.NewJob().
		Id(ires.InsertedID).
		Job(&job).
		Dispatch()

	// return success
	c.JSON(http.StatusOK, SuccessResponse{
		Status: "pipeline job created",
	})
}
