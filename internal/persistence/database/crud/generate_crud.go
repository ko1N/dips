// +build ignore

package main

import (
	"flag"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/joncalhoun/pipe"
)

type data struct {
	Type           string
	TypeName       string
	CollectionName string
	Output         string
}

func main() {
	var d data
	flag.StringVar(&d.Type, "type", "", "datatype that will be wrapped")
	flag.StringVar(&d.Output, "output", "", "generated file name")
	flag.Parse()

	s := strings.Split(d.Type, ".")
	d.TypeName = s[len(s)-1]
	d.CollectionName = strings.ToLower(d.TypeName) + "s"

	t := template.Must(template.New("queue").Parse(tpl))
	rc, wc, _ := pipe.Commands(
		exec.Command("gofmt"),
		exec.Command("goimports"),
	)
	t.Execute(wc, d)
	wc.Close()

	dst, _ := os.Create(d.Output)
	io.Copy(dst, rc)
}

var tpl = `
package crud

import (
	"github.com/zebresel-com/mongodm"
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/database/model"
	"gopkg.in/mgo.v2/bson"
)

// {{.TypeName}}Wrapper - Represents a crud wrapper and all required data
type {{.TypeName}}Wrapper struct {
	db *mongodm.Connection
}

// Create{{.TypeName}}Wrapper - Creates a new CRUD wrapper for the given types
func Create{{.TypeName}}Wrapper(db *mongodm.Connection) {{.TypeName}}Wrapper {
	db.Register(&{{.Type}}{}, "{{.CollectionName}}")
	return {{.TypeName}}Wrapper{
		db: db,
	}
}

// Create{{.TypeName}} - creates a new document
func (c *{{.TypeName}}Wrapper) Create{{.TypeName}}(document *{{.Type}}) error {
	mdl := c.db.Model("{{.TypeName}}")
	mdl.New(document)
	return document.Save()
}

// Find{{.TypeName}} - finds a single document based on the bson query
func (c *{{.TypeName}}Wrapper) Find{{.TypeName}}(query ...interface{}) (*{{.Type}}, error) {
	value := &{{.Type}}{}
	err := c.Find{{.TypeName}}Query(query...).Exec(value)
	return value, err
}

// Find{{.TypeName}}Query - returns a query to find a single document based on the bson query
func (c *{{.TypeName}}Wrapper) Find{{.TypeName}}Query(query ...interface{}) *mongodm.Query {
	mdl := c.db.Model("{{.TypeName}}")
	return mdl.FindOne(query...)
}

// Find{{.TypeName}}s - finds a list of documents based on the bson query
func (c *{{.TypeName}}Wrapper) Find{{.TypeName}}s(query ...interface{}) ([]*{{.Type}}, error) {
	value := []*{{.Type}}{}
	err := c.Find{{.TypeName}}sQuery(query...).Exec(&value)
	if _, ok := err.(*mongodm.NotFoundError); ok {
		return value, nil // not found will not result in an error but in an empty list
	}
	return value, err
}

// Find{{.TypeName}}sQuery - finds a list of documents based on the bson query
func (c *{{.TypeName}}Wrapper) Find{{.TypeName}}sQuery(query ...interface{}) *mongodm.Query {
	mdl := c.db.Model("{{.TypeName}}")
	return mdl.Find(query...)
}

// Find{{.TypeName}}ByID - finds a single document based on its hex id
func (c *{{.TypeName}}Wrapper) Find{{.TypeName}}ByID(id string) (*{{.Type}}, error) {
	value := &{{.Type}}{}
	err := c.Find{{.TypeName}}ByIDQuery(id).Exec(value)
	return value, err
}

// Find{{.TypeName}}ByIDQuery - finds a single document based on its hex id
func (c *{{.TypeName}}Wrapper) Find{{.TypeName}}ByIDQuery(id string) *mongodm.Query {
	mdl := c.db.Model("{{.TypeName}}")
	return mdl.FindId(bson.ObjectIdHex(id))
}
`
