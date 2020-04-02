package model

import "github.com/zebresel-com/mongodm"

// TODO: each job should contain all stages as a sublist when being inserted/parsed
// TODO: each stage should track progress individually + total job progress (tasknum / totaltasks)

// TODO: Name + initial variables of job should be tracked

// TODO: cross reference pipeline from job...
// TODO: would it be better to copy a pipeline here so if we change the pipeline this job wont be affected?

// JobParameter - Database struct describing a pipeline job parameter
type JobParameter struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

// Job - Database struct describing a pipeline job
type Job struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Name                 string         `json:"name" bson:"name"`
	Parameters           []JobParameter `json:"parameters" bson:"parameters"`
	Pipeline             *Pipeline      `json:"pipeline" bson:"pipeline"`
}
