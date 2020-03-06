package manager

import (
	"github.com/gin-gonic/gin"
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/amqp"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/crud"
)

// TODO: this should be self-contained and not have a global state!

// database
var pipelines crud.PipelineWrapper
var jobs crud.JobWrapper

// amqp channels
var sendPipelineExecute chan string
var recvPipelineStatus chan string

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(r *gin.Engine, db *mongodm.Connection, mq amqp.Config) error {
	// setup database
	pipelines = crud.CreatePipelineWrapper(db)
	jobs = crud.CreateJobWrapper(db)

	// setup amqp
	client := amqp.Create(mq)
	sendPipelineExecute = client.RegisterProducer("pipeline_execute")
	recvPipelineStatus = client.RegisterConsumer("pipeline_status")
	go recvJobStatus()
	client.Start()

	// setup rest routes
	r.POST("/manager/pipeline/", PipelineCreate)
	// TODO: PipelineList
	// TODO: PipelineGet ?
	// TODO: PipelineUpdate
	// TODO: PipelineDelete

	r.POST("/manager/job/execute", ExecuteJob)
	r.GET("/manager/job/list", JobList) // TODO: /list/running, /finished, etc
	r.GET("/manager/job/info/:id", JobInfo)

	return nil
}
