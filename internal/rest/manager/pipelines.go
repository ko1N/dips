package manager

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"github.com/ko1N/dips/pkg/pipeline"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func (a *ManagerAPI) PipelineCreate(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to read post body",
			Error:  err.Error(),
		})
		return
	}

	// pre-validate body
	pi, err := pipeline.CreateFromBytes(string(body))
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
		Script:   string(body),
		Pipeline: pi,
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	ires, err := a.mongo.
		Collection(colPipeline).
		InsertOne(ctx, &pl)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to create database entry for pipeline",
			Error:  err.Error(),
		})
		return
	}
	id := ires.InsertedID.(primitive.ObjectID)
	pl.Id = &id

	c.JSON(http.StatusOK, PipelineCreateResponse{
		Status:   "pipeline created",
		Pipeline: pl,
	})
}

// PipelineListResponse - response with a list of pipelines
type PipelineListResponse struct {
	Pipelines []model.Pipeline `json:"pipelines"`
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
func (a *ManagerAPI) PipelineList(c *gin.Context) {
	// TODO: pagination
	// TODO: filter name
	pipelineList := []model.Pipeline{}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := a.mongo.Collection(colPipeline).Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "error fetching pipelines",
			Error:  err.Error(),
		})
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var pipe model.Pipeline
		err := cur.Decode(&pipe)
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "error fetching pipelines2",
				Error:  err.Error(),
			})
			return
		}
		pipelineList = append(pipelineList, pipe)
	}
	if err := cur.Err(); err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "error fetching pipelines3",
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
func (a *ManagerAPI) PipelineDetails(c *gin.Context) {
	id := c.Param("pipeline_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id must not be empty",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	oid, _ := primitive.ObjectIDFromHex(id)
	fres := a.mongo.
		Collection(colPipeline).
		FindOne(ctx, bson.M{"_id": oid})
	if fres.Err() != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  fres.Err().Error(),
		})
		return
	}
	var pipeline model.Pipeline
	err := fres.Decode(&pipeline)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PipelineDetailsResponse{
		Pipeline: &pipeline,
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
func (a *ManagerAPI) PipelineUpdate(c *gin.Context) {
	// read pipeline from db
	id := c.Param("pipeline_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id must not be empty",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	oid, _ := primitive.ObjectIDFromHex(id)
	fres := a.mongo.
		Collection(colPipeline).
		FindOne(ctx, bson.M{"_id": oid})
	if fres.Err() != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  fres.Err().Error(),
		})
		return
	}
	var pipe model.Pipeline
	err := fres.Decode(&pipe)
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
	pi, err := pipeline.CreateFromBytes(string(body))
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
		pipe.Script = string(body)
		pipe.Pipeline = pi

		_, err := a.mongo.
			Collection(colPipeline).
			UpdateByID(ctx, oid, bson.M{"$set": &pipe})
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "unable to update database entry for pipeline",
				Error:  err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, PipelineDetailsResponse{
		Pipeline: &pipe,
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
func (a *ManagerAPI) PipelineDelete(c *gin.Context) {
	// read pipeline from db
	id := c.Param("pipeline_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id must not be empty",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	oid, _ := primitive.ObjectIDFromHex(id)
	fres := a.mongo.
		Collection(colPipeline).
		FindOne(ctx, bson.M{"_id": oid})
	if fres.Err() != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id `" + id + "`",
			Error:  fres.Err().Error(),
		})
		return
	}

	_, err := a.mongo.
		Collection(colPipeline).
		DeleteOne(ctx, bson.M{"_id": oid})
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
