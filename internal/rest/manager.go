package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/crud"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/model"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// TODO: this should be self-contained and not have a global state!

// database
var jobs crud.JobWrapper

// amqp channels
var sendPipelineExecute chan string
var recvPipelineStatus chan string

// SuccessResponse - reponse for a successful operation
type SuccessResponse struct {
	Status string `json:"status"`
}

// FailureResponse - response for a failed operation
type FailureResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	Data   interface{} `json:"status"`
}

// ExecuteJob - executes a job
// @Summary executes a job
// @Description This method will execute the job sent via the post body
// @ID execute-job
// @Accept plain
// @Produce json
// @Param pipeline body string true "Pipeline Script"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/job/execute [post]
func ExecuteJob(c *gin.Context) {
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
	job := model.Job{
		Pipeline: string(body),
		Progress: 0,
	}
	for _, stage := range pi.Stages {
		js := model.JobStage{
			Name:     stage.Name,
			Progress: 0,
		}
		for _, task := range stage.Tasks {
			js.Tasks = append(js.Tasks, &model.JobStageTask{
				Name:     task.Name,
				Progress: 0,
			})
		}
		job.Stages = append(job.Stages, &js)
	}
	jobs.CreateJob(&job)

	// use job id

	// send pipeline to worker
	sendPipelineExecute <- string(body)

	c.JSON(http.StatusOK, SuccessResponse{
		Status: "pipeline created",
	})
}

// JobListResponse - response with a list of jobs
type JobListResponse struct {
	Jobs []*model.Job `json:"jobs"`
}

// JobList - lists all jobs
// @Summary lists all jobs
// @Description This method will return a list of all jobs
// @ID job-list
// @Produce json
// @Success 200 {object} JobListResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/job/list [get]
func JobList(c *gin.Context) {
	// TODO: pagination

	jobList, err := jobs.FindJobs()
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
// @ID pipeline-info
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} JobInfoResponse
// @Failure 400 {object} FailureResponse
// @Router /manager/jobs/info/{:id} [get]
func JobInfo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "invalid pipeline id",
			Error:  "pipeline id is \"\"",
		})
		return
	}

	job, err := jobs.FindJobByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailureResponse{
			Status: "unable to find pipeline with id " + id,
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, JobInfoResponse{
		Job: job,
	})
}

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(r *gin.Engine, db *mongodm.Connection, mq string) error {
	// setup database
	jobs = crud.CreateJobWrapper(db)

	// setup amqp
	client := amqp.Create(mq)
	sendPipelineExecute = client.RegisterProducer("pipeline_execute")
	recvPipelineStatus = client.RegisterConsumer("pipeline_status")
	client.Start()

	// setup rest routes
	r.POST("/manager/job/execute", ExecuteJob)
	r.GET("/manager/job/list", JobList) // TODO: /list/running, /finished, etc
	r.GET("/manager/job/info/:id", JobInfo)

	return nil
}
