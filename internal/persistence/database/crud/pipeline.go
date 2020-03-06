package crud

import (
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

// PipelineWrapper - Represents a crud wrapper and all required data
type PipelineWrapper struct {
	db *mongodm.Connection
}

// CreatePipelineWrapper - Creates a new CRUD wrapper for the given types
func CreatePipelineWrapper(db *mongodm.Connection) PipelineWrapper {
	db.Register(&model.Pipeline{}, "pipelines")
	return PipelineWrapper{
		db: db,
	}
}

// CreatePipeline - creates a new document
func (c *PipelineWrapper) CreatePipeline(document *model.Pipeline) error {
	mdl := c.db.Model("Pipeline")
	mdl.New(document)
	return document.Save()
}

// FindPipeline - finds a single document based on the bson query
func (c *PipelineWrapper) FindPipeline(query ...interface{}) (*model.Pipeline, error) {
	value := &model.Pipeline{}
	err := c.FindPipelineQuery(query...).Exec(value)
	return value, err
}

// FindPipelineQuery - returns a query to find a single document based on the bson query
func (c *PipelineWrapper) FindPipelineQuery(query ...interface{}) *mongodm.Query {
	mdl := c.db.Model("Pipeline")
	return mdl.FindOne(query...)
}

// FindPipelines - finds a list of documents based on the bson query
func (c *PipelineWrapper) FindPipelines(query ...interface{}) ([]*model.Pipeline, error) {
	value := []*model.Pipeline{}
	err := c.FindPipelinesQuery(query...).Exec(&value)
	if _, ok := err.(*mongodm.NotFoundError); ok {
		return value, nil // not found will not result in an error but in an empty list
	}
	return value, err
}

// FindPipelinesQuery - finds a list of documents based on the bson query
func (c *PipelineWrapper) FindPipelinesQuery(query ...interface{}) *mongodm.Query {
	mdl := c.db.Model("Pipeline")
	return mdl.Find(query...)
}

// FindPipelineByID - finds a single document based on its hex id
func (c *PipelineWrapper) FindPipelineByID(id string) (*model.Pipeline, error) {
	value := &model.Pipeline{}
	err := c.FindPipelineByIDQuery(id).Exec(value)
	return value, err
}

// FindPipelineByIDQuery - finds a single document based on its hex id
func (c *PipelineWrapper) FindPipelineByIDQuery(id string) *mongodm.Query {
	mdl := c.db.Model("Pipeline")
	return mdl.FindId(bson.ObjectIdHex(id))
}
