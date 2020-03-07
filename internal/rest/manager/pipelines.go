package manager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
	"gopkg.in/mgo.v2/bson"
)

// list pipelines register pipelines, update pipelines, unregister pipelines

// PipelineCreateResponse - reponse for a successful pipeline creation
type PipelineCreateResponse struct {
	Status   string         `json:"status"`
	Pipeline model.Pipeline `json:"pipeline"`
}

// PipelineCreate - creates a pipeline
// @Summary creates a pipeline
// @Description This method will create the pipeline sent via the post body
// @ID create-pipeline
// @Tags pipelines
// @Accept plain
// @Produce json
// @Param pipeline body string true "Pipeline Script"
// @Success 200 {object} PipelineCreateResponse
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
	pl := model.Pipeline{
		Script:   string(body),
		Name:     pi.Name,
		Pipeline: &pi,
	}

	err = pipelines.CreatePipeline(&pl)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to create database entry for pipeline",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PipelineCreateResponse{
		Status:   "pipeline created",
		Pipeline: pl,
	})
}

// PipelineListResponse - response with a list of pipelines
type PipelineListResponse struct {
	Pipelines []*model.Pipeline `json:"pipelines"`
}

// PipelineList - lists all registered pipelines
// @Summary lists all registered pipelines
// @Description This method will return a list of all registered pipelines
// @ID pipeline-list
// @Tags pipelines
// @Produce json
// @Success 200 {object} PipelineListResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/job/list [get]
func PipelineList(c *gin.Context) {
	// TODO: pagination
	pipelineList := []*model.Pipeline{}
	err := pipelines.FindPipelinesQuery().
		Select(bson.M{"name": true}).
		Exec(&pipelineList)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find any pipelines",
			Error:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, PipelineListResponse{
		Pipelines: pipelineList,
	})
}
