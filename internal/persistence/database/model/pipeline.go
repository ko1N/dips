package model

import (
	"github.com/zebresel-com/mongodm"
)

// TODO: properties struct
// TODO: task struct should contain all necessary infos + unique ids

// Pipeline - Database struct describing a pipeline
type Pipeline struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Script               string `json:"script" bson:"script" required:"true"`
	Revision             uint   `json:"revision" bson:"revision"`
	Name                 string `json:"name" bson:"name" required:"true"`
	Pipeline             []byte `json:"pipeline" bson:"pipeline"`
	// TODO: properties
}
