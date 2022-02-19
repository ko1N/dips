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

// Task -
type Task struct {
	Name         string            `json:"name" bson:"name"`
	Service      string            `json:"service" bson:"service"`
	Input        map[string]string `json:"input" bson:"input"`
	IgnoreErrors bool              `json:"ignore_errors" bson:"ignore_errors"`
	Register     string            `json:"register" bson:"register"`
	Notify       []string          `json:"notify" bson:"notify"` // NotifyRef
	When         Expression        `json:"when" bson:"when"`
}

// Stage -
type Stage struct {
	Name  string `json:"name" bson:"name"`
	Tasks []Task `json:"tasks" bson:"tasks"`
	//Variables   []Variable `json:"variables" bson:"variables"`
}

// Pipeline -
type Pipeline struct {
	Name       string      `json:"name" bson:"name"`
	Parameters []Parameter `json:"parameters" bson:"parameters"`
	Stages     []Stage     `json:"stages" bson:"stages"`
}

// CreateFromBytes - loads a new pipeline instance from a byte array
func CreateFromBytes(data string) (*Pipeline, error) {
	// TODO: multifile pipelines
	if !strings.HasPrefix(string(data), "---\n") {
		return nil, errors.New("Not a valid pipeline script. should start with `---`")
	}

	var script interface{}
	err := yaml.Unmarshal([]byte(data), &script)
	if err != nil {
		return nil, err
	}

	if s, ok := script.(map[interface{}]interface{}); ok {
		return parsePipeline(s)
	}

	return nil, errors.New("Not a valid pipeline script. Script should start with `name:` or `stages:`")
}

func parsePipeline(script map[interface{}]interface{}) (*Pipeline, error) {
	result := Pipeline{}

	if name, ok := script["name"]; ok {
		result.Name = name.(string)
	}

	var err error
	result.Parameters, err = parseParameters(script)
	if err != nil {
		return nil, err
	}

	if stages, ok := script["stages"]; ok {
		for _, s := range stages.([]interface{}) {
			stage, err := parseStage(s.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}
			result.Stages = append(result.Stages, stage)
		}
	}

	return &result, nil
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
			Name: stage.(string),
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
			result = append(result, *_task)
		}
	}
	return result, nil
}

func parseTask(script map[interface{}]interface{}) (*Task, error) {
	result := &Task{}

	for key, value := range script {
		switch key.(string) {
		case "name":
			result.Name = value.(string)
			break

		case "service":
			result.Service = value.(string)
			break

		case "input":
			// TODO: LIST
			result.Input = make(map[string]string)
			if val, ok := value.(map[interface{}]interface{}); ok {
				for key, val := range val {
					result.Input[key.(string)] = val.(string)
				}
			} else if val, ok := value.(string); ok {
				result.Input[""] = val
			}
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
						return nil, errors.New("Invalid syntax when parsing \"notify\". Should be a string or a list of strings")
					}
				}
			} else {
				return nil, errors.New("Invalid syntax when parsing \"notify\"")
			}
			break

		case "when":
			result.When = Expression{
				Script: value.(string),
			}
			break

		default:
			break
		}
	}

	// TODO: sanitize task
	if result.Service == "" {
		return nil, errors.New("task requires a service")
	}

	return result, nil
}
