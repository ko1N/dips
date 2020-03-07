package manager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
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
// @Router /manager/job/list [get]
func JobList(c *gin.Context) {
	// TODO: pagination
	jobList := []*model.Job{}
	err := jobs.FindJobsQuery().
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

// JobInfoResponse - response for a failed operation
type JobInfoResponse struct {
	Job *model.Job `json:"job"`
	// TODO: log?
}

// JobInfo - find a single job by it's id
// @Summary find a single job by it's id
// @Description This method will return a single job by it's id or an error.
// @ID job-info
// @Tags jobs
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} JobInfoResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/job/info/{job_id} [get]
func JobInfo(c *gin.Context) {
	id := c.Param("id")
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

	c.JSON(http.StatusOK, JobInfoResponse{
		Job: job,
	})
}
