package model

import "github.com/zebresel-com/mongodm"

// TODO: each job should contain all stages as a sublist when being inserted/parsed
// TODO: each stage should track progress individually + total job progress (tasknum / totaltasks)

// TODO: Name + initial variables of job should be tracked

// TODO: cross reference pipeline from job...
// TODO: would it be better to copy a pipeline here so if we change the pipeline this job wont be affected?

// Job - Database struct describing a pipeline job
type Job struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Name                 string    `json:"name" bson:"Name"`
	Pipeline             *Pipeline `json:"pipeline" bson:"pipeline"`
}

/*
// JobStage - Database struct describing a job stage
type JobStage struct {
	Name     string          `json:"name" bson:"name" required:"true"`
	Progress uint            `json:"progress" bson:"progress"`
	Tasks    []*JobStageTask `json:"tasks" bson:"tasks"`
}

// JobStageTask - Database struct describing a stage task
type JobStageTask struct {
	ID       uint     `json:"id" bson:"id" required:"true"`
	Name     string   `json:"name" bson:"name" required:"true"`
	Progress uint     `json:"progress" bson:"progress"`
	StdOut   []string `json:"stdout" bson:"stdout"`
	StdErr   []string `json:"stderr" bson:"stderr"`
}
*/
