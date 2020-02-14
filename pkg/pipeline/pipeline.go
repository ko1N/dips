package pipeline

import (
	"errors"
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
	Name        string
	Environment string
	Tasks       []Task
	Variables   []Variable
}

// Pipeline -
type Pipeline struct {
	Stages []Stage
}

// CreateFromBytes - loads a new pipeline instance from a byte array
func CreateFromBytes(data []byte) (Pipeline, error) {
	// TODO: multifile pipelines
	if !strings.HasPrefix(string(data), "---\n") {
		return Pipeline{}, errors.New("not a valid pipeline yaml")
	}

	var script interface{}
	err := yaml.Unmarshal(data, &script)
	if err != nil {
		return Pipeline{}, err
	}

	return parsePipeline(script)
}

func parsePipeline(script interface{}) (Pipeline, error) {
	result := Pipeline{}
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
		//fmt.Println("Parsing stage" + stage.(string))
		result := Stage{
			Name:        stage.(string),
			Environment: "native",
		}

		if env, ok := script["environment"]; ok {
			result.Environment = env.(string)
		}

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
		//fmt.Printf("vars:\n%v\n\n", vars)
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
			//fmt.Printf("task:\n%v\n\n", task)
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
