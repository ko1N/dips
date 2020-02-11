package workflow

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// Variable -
type Variable struct {
	Name  string
	Value string
}

// Command -
type Command struct {
	Name      string
	Arguments []string
}

// Task -
type Task struct {
	Name     string
	Command  []Command
	Register string   // VariableRef
	Notify   []string // NotifyRef
}

// Stage -
type Stage struct {
	Tasks     []Task
	Variables []Variable
}

// Workflow -
type Workflow struct {
	Stages []Stage
}

// CreateFromBytes - loads a new workflow instance from a byte array
func CreateFromBytes(data []byte) (Workflow, error) {
	// TODO: multifile workflows
	if !strings.HasPrefix(string(data), "---\n") {
		return Workflow{}, errors.New("not a valid workflow yaml")
	}

	var script interface{}
	err := yaml.Unmarshal(data, &script)
	if err != nil {
		return Workflow{}, err
	}

	return parseWorkflow(script)
}

func parseWorkflow(script interface{}) (Workflow, error) {
	result := Workflow{}
	for _, s := range script.([]interface{}) {
		stage, err := parseStage(s.(map[interface{}]interface{}))
		if err != nil {
			return result, err
		}
		result.Stages = append(result.Stages, stage)
	}
	return result, nil
}

func parseStage(script map[interface{}]interface{}) (Stage, error) {
	if stage, ok := script["stage"]; ok {
		fmt.Println("Parsing stage" + stage.(string))
		result := Stage{}

		var err error
		result.Variables, err = parseVariables(script)
		if err != nil {
			return result, err
		}

		result.Tasks, err = parseTasks(script)
		if err != nil {
			return result, err
		}

		return result, nil
	}

	return Stage{}, errors.New("Malformed stage. Should start with \"- stage: [Name]\"")
}

func parseVariables(script map[interface{}]interface{}) ([]Variable, error) {
	var result []Variable
	if vars, ok := script["vars"]; ok {
		fmt.Printf("vars:\n%v\n\n", vars)
		for key, value := range vars.(map[interface{}]interface{}) {
			result = append(result, Variable{
				Name:  key.(string),
				Value: value.(string),
			})
		}
	}
	return result, nil
}

func parseTasks(script map[interface{}]interface{}) ([]Task, error) {
	var result []Task
	if tasks, ok := script["tasks"]; ok {
		//fmt.Printf("tasks:\n%v\n\n", tasks)
		for _, task := range tasks.([]interface{}) {
			fmt.Printf("task:\n%v\n\n", task)
			_task, err := parseTask(task.(map[interface{}]interface{}))
			if err != nil {
				return result, err
			}
			result = append(result, _task)
		}
	}
	return result, nil
}

func parseTask(script map[interface{}]interface{}) (Task, error) {
	result := Task{}

	for key, value := range script {
		switch key.(string) {
		case "name":
			result.Name = value.(string)
			break

		case "register":
			result.Register = value.(string)
			break

		case "notify":
			if val, ok := value.(string); ok {
				result.Notify = []string{val}
			} else if list, ok := value.([]interface{}); ok {
				for _, val := range list {
					if str, ok := val.(string); ok {
						result.Notify = append(result.Notify, str)
					} else {
						return result, errors.New("Invalid syntax when parsing \"notify\". Should be a string or a list of strings")
					}
				}
			} else {
				return result, errors.New("Invalid syntax when parsing \"notify\"")
			}
			break

		default:
			cmd, err := parseCommand(key.(string), value)
			if err != nil {
				return result, err
			}
			result.Command = append(result.Command, cmd)
			break
		}
	}

	return result, nil
}

func parseCommand(cmd string, args interface{}) (Command, error) {
	if val, ok := args.(string); ok {
		return Command{
			Name:      cmd,
			Arguments: []string{val},
		}, nil
	} else if list, ok := args.([]interface{}); ok {
		var arguments []string
		for _, val := range list {
			if str, ok := val.(string); ok {
				arguments = append(arguments, str)
			} else {
				return Command{}, errors.New("Invalid syntax when parsing \"" + cmd + "\". Should be a string or a list of strings")
			}
		}
		return Command{
			Name:      cmd,
			Arguments: arguments,
		}, nil
	} else {
		return Command{}, errors.New("Invalid syntax when parsing \"" + cmd + "\"")
	}
}

/*
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
*/
