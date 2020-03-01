package crud

import (
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

// JobWrapper - Represents a crud wrapper and all required data
type JobWrapper struct {
	db *mongodm.Connection
}

// CreateJobWrapper - Creates a new CRUD wrapper for the given types
func CreateJobWrapper(db *mongodm.Connection) JobWrapper {
	db.Register(&model.Job{}, "jobs")
	return JobWrapper{
		db: db,
	}
}

// CreateJob - creates a new document
func (c *JobWrapper) CreateJob(document *model.Job) error {
	mdl := c.db.Model("Job")
	mdl.New(document)
	return document.Save()
}

// FindJob - finds a single document based on the bson query
func (c *JobWrapper) FindJob(query ...interface{}) (*model.Job, error) {
	value := &model.Job{}
	err := c.FindJobQuery(query...).Exec(value)
	return value, err
}

// FindJobQuery - returns a query to find a single document based on the bson query
func (c *JobWrapper) FindJobQuery(query ...interface{}) *mongodm.Query {
	mdl := c.db.Model("Job")
	return mdl.FindOne(query...)
}

// FindJobs - finds a list of documents based on the bson query
func (c *JobWrapper) FindJobs(query ...interface{}) ([]*model.Job, error) {
	value := []*model.Job{}
	err := c.FindJobsQuery(query...).Exec(&value)
	if _, ok := err.(*mongodm.NotFoundError); ok {
		return value, nil // not found will not result in an error but in an empty list
	}
	return value, err
}

// FindJobsQuery - finds a list of documents based on the bson query
func (c *JobWrapper) FindJobsQuery(query ...interface{}) *mongodm.Query {
	mdl := c.db.Model("Job")
	return mdl.Find(query...)
}

// FindJobByID - finds a single document based on its hex id
func (c *JobWrapper) FindJobByID(id string) (*model.Job, error) {
	value := &model.Job{}
	err := c.FindJobByIDQuery(id).Exec(value)
	return value, err
}

// FindJobByIDQuery - finds a single document based on its hex id
func (c *JobWrapper) FindJobByIDQuery(id string) *mongodm.Query {
	mdl := c.db.Model("Job")
	return mdl.FindId(bson.ObjectIdHex(id))
}
