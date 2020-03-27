package manager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/messages"
	"gopkg.in/mgo.v2/bson"
)

// JobListResponse - response with a list of jobs
type JobListResponse struct {
	Jobs []*model.Job `json:"jobs"`
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
func JobList(c *gin.Context) {
	// TODO: pagination
	jobList := []*model.Job{}
	err := jobs.FindJobsQuery(bson.M{"deleted": false}).
		Select(bson.M{"name": true, "progress": true}).
		Exec(&jobList)
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
func JobDetails(c *gin.Context) {
	id := c.Param("job_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid job id",
			Error:  "job id must not be empty",
		})
		return
	}

	job, err := jobs.FindJobByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find job with id " + id,
			Error:  err.Error(),
		})
		return
	}

	// fetch job messages
	messages := messageHandler.GetAll(id)

	c.JSON(http.StatusOK, JobDetailsResponse{
		Job:      job,
		Messages: messages,
	})
}
