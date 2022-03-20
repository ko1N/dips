package manager

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// PipelineExecuteRequest - Request Body when executing a pipeline
type PipelineExecuteRequest struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// PipelineExecuteResponse - Response when a pipeline was started
type PipelineExecuteResponse struct {
	Job *model.Job `json:"job"`
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
// @Success 200 {object} PipelineExecuteResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/execute/{pipeline_id} [post]
func (a *ManagerAPI) PipelineExecute(c *gin.Context) {
	// try to find requested pipeline
	pipelineId := c.Param("pipeline_id")
	if pipelineId == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline_id",
			Error:  "pipeline_id must not be empty",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	oid, _ := primitive.ObjectIDFromHex(pipelineId)
	fres := a.mongo.
		Collection(colPipeline).
		FindOne(ctx, bson.M{"_id": oid})
	if fres.Err() != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + pipelineId + "`",
			Error:  fres.Err().Error(),
		})
		return
	}
	var pipeline model.Pipeline
	err := fres.Decode(&pipeline)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + pipelineId + "`",
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
	// input parameters are mapped to variables here
	job := model.Job{
		Name:      request.Name,
		Variables: request.Parameters,
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
	id := ires.InsertedID.(primitive.ObjectID)
	job.Id = &id

	// send pipeline to worker
	a.dipscl.NewJob().
		Job(&job).
		Dispatch()

	// return success
	c.JSON(http.StatusOK, PipelineExecuteResponse{
		Job: &job,
	})
}
