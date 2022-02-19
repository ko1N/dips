package model

import "gopkg.in/mgo.v2/bson"

// TODO: properties struct
// TODO: task struct should contain all necessary infos + unique ids

// Pipeline - Database struct describing a pipeline
type Pipeline struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Revision uint          `json:"revision" bson:"revision"`
	Name     string        `json:"name" bson:"name" required:"true"`
	Script   []byte        `json:"script" bson:"script"`
	// TODO: properties
}
