package manager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"github.com/ko1N/dips/pkg/pipeline"
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
// @ID pipeline-create
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

	// write pipeline to database
	pl := model.Pipeline{
		Revision: 0,
		Name:     pi.Name,
		Script:   body,
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
// @Router /manager/pipeline/all [get]
func PipelineList(c *gin.Context) {
	// TODO: pagination
	pipelineList := []*model.Pipeline{}
	err := pipelines.FindPipelinesQuery(bson.M{"deleted": false}).
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

// PipelineDetailsResponse - response with pipeline details
type PipelineDetailsResponse struct {
	Pipeline *model.Pipeline `json:"pipeline"`
}

// PipelineDetails - find a pipeline by it's id and shows all fields
// @Summary find a pipeline by it's id and shows all fields
// @Description This method will return a single pipeline by it's id or an error.
// @ID pipeline-details
// @Tags pipelines
// @Produce json
// @Param pipeline_id path string true "Pipeline ID"
// @Success 200 {object} PipelineDetailsResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/details/{pipeline_id} [get]
func PipelineDetails(c *gin.Context) {
	id := c.Param("pipeline_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id must not be empty",
		})
		return
	}

	pipe, err := pipelines.FindPipelineByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PipelineDetailsResponse{
		Pipeline: pipe,
	})
}

// PipelineUpdate - updates the pipeline with the given id
// @Summary updates the pipeline with the given id
// @Description This method will update the given pipeline from a provided script
// @ID pipeline-update
// @Tags pipelines
// @Accept plain
// @Produce json
// @Param pipeline_id path string true "Pipeline ID"
// @Param pipeline body string true "Pipeline Script"
// @Success 200 {object} PipelineDetailsResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/{pipeline_id} [patch]
func PipelineUpdate(c *gin.Context) {
	// read pipeline from db
	id := c.Param("pipeline_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id must not be empty",
		})
		return
	}

	pipe, err := pipelines.FindPipelineByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  err.Error(),
		})
		return
	}

	// parse body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to read post body",
			Error:  err.Error(),
		})
		return
	}

	// validate body
	pi, err := pipeline.CreateFromBytes(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to parse pipeline",
			Error:  err.Error(),
		})
		return
	}

	if string(pipe.Script) != string(body) {
		// update pipeline script
		pipe.Revision = pipe.Revision + 1
		pipe.Name = pi.Name
		pipe.Script = body

		err = pipe.Save()
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "unable to update database entry for pipeline",
				Error:  err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, PipelineDetailsResponse{
		Pipeline: pipe,
	})
}

// PipelineDelete - deletes the pipeline with the given id
// @Summary deletes the pipeline with the given id
// @Description This method will delete the given pipeline
// @ID pipeline-delete
// @Tags pipelines
// @Produce json
// @Param pipeline_id path string true "Pipeline ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/pipeline/{pipeline_id} [delete]
func PipelineDelete(c *gin.Context) {
	// read pipeline from db
	id := c.Param("pipeline_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id must not be empty",
		})
		return
	}

	pipe, err := pipelines.FindPipelineByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  err.Error(),
		})
		return
	}

	err = pipe.Delete()
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to delete database entry for pipeline",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Status: "pipeline `" + id + "` deleted",
	})
}
