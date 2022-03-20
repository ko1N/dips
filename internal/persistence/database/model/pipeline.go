package model

import (
	"github.com/ko1N/dips/pkg/pipeline"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: properties struct
// TODO: task struct should contain all necessary infos + unique ids

// Pipeline - Database struct describing a pipeline
type Pipeline struct {
	Id       *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Revision uint                `json:"revision" bson:"revision"`
	Name     string              `json:"name" bson:"name" required:"true"`
	Script   string              `json:"script" bson:"script"`
	Pipeline *pipeline.Pipeline  `json:"pipeline" bson:"pipeline"`
}
