package crud

import (
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/model"
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

// Create - creates a new document
func (c *JobWrapper) Create(document *model.Job) error {
	mdl := c.db.Model("Job")
	mdl.New(document)
	return document.Save()
}

func (c *JobWrapper) FindOne(query ...interface{}) (*model.Job, error) {
	mdl := c.db.Model("Job")
	value := &model.Job{}
	err := mdl.FindOne(query...).Exec(value)
	return value, err
}
