package model

import "github.com/zebresel-com/mongodm"

// TODO: each job should contain all stages as a sublist when being inserted/parsed
// TODO: each stage should track progress individually + total job progress (tasknum / totaltasks)

// TODO: Name + initial variables of job should be tracked

// Job - Database struct describing a pipeline job
type Job struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Pipeline             string      `json:"pipeline" bson:"pipeline" required:"true"`
	Progress             uint        `json:"progress" bson:"progress" required:"true"`
	Stages               []*JobStage `json:"stages" bson:"stages"`
}

// JobStage - Database struct describing a job stage
type JobStage struct {
	Name     string          `json:"name" bson:"name" required:"true"`
	Progress uint            `json:"progress" bson:"progress" required:"true"`
	Tasks    []*JobStageTask `json:"tasks" bson:"tasks"`
}

// JobStageTask - Database struct describing a stage task
type JobStageTask struct {
	Name     string `json:"name" bson:"name" required:"true"`
	Progress uint   `json:"progress" bson:"progress" required:"true"`
}
