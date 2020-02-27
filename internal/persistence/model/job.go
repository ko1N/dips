package model

import "github.com/zebresel-com/mongodm"

// Job - Database struct describing a pipeline job
type Job struct {
	mongodm.DocumentBase `json:",inline" bson:",inline"`
	Pipeline             string `json:"pipeline" bson:"pipeline" required:"true"`
}
