package model

import (
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/pipeline"
)

// TODO: properties struct
// TODO: task struct should contain all necessary infos + unique ids

// Pipeline - Database struct describing a pipeline
type Pipeline struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Script               string             `json:"script" bson:"script" required:"true"`
	Revision             uint               `json:"revision" bson:"revision"`
	Name                 string             `json:"name" bson:"name" required:"true"`
	Pipeline             *pipeline.Pipeline `json:"pipeline" bson:"pipeline"`
	// TODO: properties
}
