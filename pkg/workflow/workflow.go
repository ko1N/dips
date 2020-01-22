package workflow

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"gopkg.in/yaml.v2"
)

// Workflow -
type Workflow struct {
	Workflow interface{}
	//
	Variables map[interface{}]interface{}
}

// CreateFromBytes - loads a new workflow instance from a byte array
func CreateFromBytes(data []byte) (Workflow, error) {
	// TODO: multifile workflows
	if !strings.HasPrefix(string(data), "---\n") {
		return Workflow{}, errors.New("not a valid workflow yaml")
	}

	var workflow interface{}
	err := yaml.Unmarshal(data, &workflow)
	if err != nil {
		return Workflow{}, err
	}

	return Workflow{
		Workflow:  workflow,
		Variables: make(map[interface{}]interface{}),
	}, nil
}

// Run - Executes a workflow
func (w *Workflow) Run() error {
	for _, stage := range w.Workflow.([]interface{}) {
		fmt.Println("executing stage")
		err := w.runStage(stage.(map[interface{}]interface{}))
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: parseStage() function which returns a struct
func (w *Workflow) runStage(stage map[interface{}]interface{}) error {
	// has variables ?
	if vars, ok := stage["vars"]; ok {
		fmt.Printf("vars:\n%v\n\n", vars)
		for k, v := range vars.(map[interface{}]interface{}) {
			w.Variables[k] = v
		}
	}

	// has tasks ?
	if tasks, ok := stage["tasks"]; ok {
		//fmt.Printf("tasks:\n%v\n\n", tasks)
		for _, task := range tasks.([]interface{}) {
			//fmt.Printf("task:\n%v\n\n", task)
			w.runTask(task.(map[interface{}]interface{}))
		}
	} else {
		return errors.New("no tasks found in stage")
	}
	return nil
}

// TODO: parseTask() function which returns a struct
func (w *Workflow) runTask(task map[interface{}]interface{}) {
	if name, ok := task["name"]; ok {
		n, err := w.parseString(name.(string))
		if err != nil {
			// TODO: proper error handling
		}
		fmt.Println("executing task \"" + n + "\"")
	}

	// find a registered plugin (e.g. shell)
	if shell, ok := task["shell"]; ok {
		s, err := w.parseString(shell.(string))
		if err != nil {
			// TODO: proper error handling
		}
		fmt.Println("shell: " + s)
	}
}

func (w *Workflow) parseString(str string) (string, error) {
	// preprocess template
	str = strings.ReplaceAll(str, "{{", "{{index .Variables \"")
	str = strings.ReplaceAll(str, "}}", "\"}}")
	tpl, err := template.New("cmd").Parse(str)
	if err != nil {
		return "", err
	}

	var res bytes.Buffer
	err = tpl.Execute(&res, w)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}
