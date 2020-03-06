package model

import ()

// TODO: properties struct
// TODO: task struct should contain all necessary infos + unique ids

// Pipeline - Database struct describing a pipeline
type Pipeline struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Script               string `json:"script" bson:"script" required:"true"`
	// TODO: properties
	Script string           `json:"name" bson:"name" required:"true"`
	Stages []*PipelineStage `json:"stages" bson:"stages"`
}

// PipelineStage - Database struct describing a pipeline stage
type PipelineStage struct {
	Name  string          `json:"name" bson:"name" required:"true"`
	Tasks []*JobStageTask `json:"tasks" bson:"tasks"`
}

// PipelineTask - Database struct describing a pipeline stage task
type PipelineTask struct {
	ID   uint   `json:"id" bson:"id" required:"true"`
	Name string `json:"name" bson:"name" required:"true"`
}
