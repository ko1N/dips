package environments

import "github.com/d5/tengo/v2"

// ExecutionResult - represents a execution
type ExecutionResult struct {
	ExitCode int
	StdOut   string
	StdErr   string
}

// ToScriptObject - converts the ExecutionResult to a script object
func (r *ExecutionResult) ToScriptObject() tengo.Object {
	failed := tengo.TrueValue
	if r.ExitCode == 0 {
		failed = tengo.FalseValue
	}

	return &tengo.Map{
		Value: map[string]tengo.Object{
			"failed": failed,
			"rc":     &tengo.Int{Value: int64(r.ExitCode)},
			"stderr": &tengo.String{Value: r.StdErr},
			// stderr_lines
			"stdout": &tengo.String{Value: r.StdOut},
			// stdout_lines
		},
	}
}

// Environment -
type Environment interface {
	Name() string
	Execute(cmd []string, stdout func(string), stderr func(string)) (ExecutionResult, error)

	CopyTo(from string, to string) error
	CopyFrom(from string, to string) error

	Close()
}
