package manager

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ko1N/dips/internal/persistence/database/model"
	"github.com/ko1N/dips/internal/persistence/messages"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// JobListResponse - response with a list of jobs
type JobListResponse struct {
	Jobs []model.Job `json:"jobs"`
}

// JobList - lists all jobs
// @Summary lists all jobs
// @Description This method will return a list of all jobs
// @ID job-list
// @Tags jobs
// @Produce json
// @Success 200 {object} JobListResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/job/all [get]
func (a *ManagerAPI) JobList(c *gin.Context) {
	// TODO: pagination
	// TODO: filter just name
	jobList := []model.Job{}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	cur, err := a.mongo.Collection(colJobs).Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "error fetching jobs",
			Error:  err.Error(),
		})
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var job model.Job
		err := cur.Decode(&job)
		if err != nil {
			c.JSON(http.StatusBadRequest, FailureResponse{
				Status: "error fetching jobs",
				Error:  err.Error(),
			})
			return
		}
		jobList = append(jobList, job)
	}
	if err := cur.Err(); err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "error fetching jobs",
			Error:  err.Error(),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipelines",
			Error:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, JobListResponse{
		Jobs: jobList,
	})
}

// JobDetailsResponse - response with job details
type JobDetailsResponse struct {
	Job      *model.Job         `json:"job"`
	Messages []messages.Message `json:"messages"`
}

// JobDetails - find a single job by it's id and shows all fields
// @Summary find a single job by it's id and shows all fields
// @Description This method will return a single job by it's id or an error.
// @ID job-details
// @Tags jobs
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} JobDetailsResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/job/details/{job_id} [get]
func (a *ManagerAPI) JobDetails(c *gin.Context) {
	id := c.Param("job_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid job id",
			Error:  "job id must not be empty",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()
	oid, _ := primitive.ObjectIDFromHex(id)
	fres := a.mongo.
		Collection(colJobs).
		FindOne(ctx, bson.M{"_id": oid})
	if fres.Err() != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find job with id " + id,
			Error:  fres.Err().Error(),
		})
		return
	}
	var job model.Job
	err := fres.Decode(&job)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find job with id " + id,
			Error:  err.Error(),
		})
		return
	}

	// fetch job messages
	messages := a.messageHandler.GetAll(id)

	c.JSON(http.StatusOK, JobDetailsResponse{
		Job:      &job,
		Messages: messages,
	})
}
