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
	"gitlab.strictlypaste.xyz/ko1n/dips/internal/persistence/model"
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

// Create - creates a new document
func (c *{{.TypeName}}Wrapper) Create(document *{{.Type}}) error {
	mdl := c.db.Model("{{.TypeName}}")
	mdl.New(document)
	return document.Save()
}

// FindOne - finds a single document based on the bson query
func (c *{{.TypeName}}Wrapper) FindOne(query ...interface{}) (*{{.Type}}, error) {
	mdl := c.db.Model("{{.TypeName}}")
	value := &{{.Type}}{}
	err := mdl.FindOne(query...).Exec(value)
	return value, err
}
`
