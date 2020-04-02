package pipeline

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v2"
)

// Variable -
type Variable struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

// Parameter -
type Parameter struct {
	Name string `json:"name" bson:"name"`
}

// Command -
type Command struct {
	Name  string   `json:"name" bson:"name"`
	Lines []string `json:"lines" bson:"lines"`
}

// Task -
type Task struct {
	Name         string     `json:"name" bson:"name"`
	Command      []Command  `json:"command" bson:"command"`
	IgnoreErrors bool       `json:"ignore_errors" bson:"ignore_errors"`
	Register     string     `json:"register" bson:"register"` // VariableRef
	Notify       []string   `json:"notify" bson:"notify"`     // NotifyRef
	When         Expression `json:"when" bson:"when"`
}

// Stage -
type Stage struct {
	Name        string `json:"name" bson:"name"`
	Environment string `json:"environment" bson:"environment"`
	Tasks       []Task `json:"tasks" bson:"tasks"`
	//Variables   []Variable `json:"variables" bson:"variables"`
}

// Pipeline -
type Pipeline struct {
	Name       string      `json:"name" bson:"name"`
	Parameters []Parameter `json:"parameters" bson:"parameters"`
	Stages     []Stage     `json:"stages" bson:"stages"`
}

// CreateFromBytes - loads a new pipeline instance from a byte array
func CreateFromBytes(data []byte) (Pipeline, error) {
	// TODO: multifile pipelines
	if !strings.HasPrefix(string(data), "---\n") {
		return Pipeline{}, errors.New("Not a valid pipeline script. should start with `---`")
	}

	var script interface{}
	err := yaml.Unmarshal(data, &script)
	if err != nil {
		return Pipeline{}, err
	}

	if s, ok := script.(map[interface{}]interface{}); ok {
		return parsePipeline(s)
	}

	return Pipeline{}, errors.New("Not a valid pipeline script. Script should start with `name:` or `stages:`")
}

func parsePipeline(script map[interface{}]interface{}) (Pipeline, error) {
	result := Pipeline{}

	if name, ok := script["name"]; ok {
		result.Name = name.(string)
	}

	var err error
	result.Parameters, err = parseParameters(script)
	if err != nil {
		return result, err
	}

	if stages, ok := script["stages"]; ok {
		for _, s := range stages.([]interface{}) {
			stage, err := parseStage(s.(map[interface{}]interface{}))
			if err != nil {
				return result, err
			}
			result.Stages = append(result.Stages, stage)
		}
	}

	return result, nil
}

func parseParameters(script map[interface{}]interface{}) ([]Parameter, error) {
	var result []Parameter
	if params, ok := script["parameters"]; ok {
		for _, value := range params.([]interface{}) {
			result = append(result, Parameter{
				Name: value.(string),
			})
		}
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
		result.Tasks, err = parseTasks(script)
		if err != nil {
			return result, err
		}

		return result, nil
	}

	return Stage{}, errors.New("Malformed stage. Should start with \"- stage: [Name]\"")
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

		case "ignore_errors":
			result.IgnoreErrors = strings.ToLower(value.(string)) == "true"
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

		case "when":
			result.When = Expression{
				Script: value.(string),
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
			Name:  cmd,
			Lines: []string{val},
		}, nil
	} else if list, ok := args.([]interface{}); ok {
		var lines []string
		for _, val := range list {
			if str, ok := val.(string); ok {
				lines = append(lines, str)
			} else {
				return Command{}, errors.New("Invalid syntax when parsing \"" + cmd + "\". Should be a string or a list of strings")
			}
		}
		return Command{
			Name:  cmd,
			Lines: lines,
		}, nil
	} else {
		return Command{}, errors.New("Invalid syntax when parsing \"" + cmd + "\"")
	}
}
