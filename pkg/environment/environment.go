package environment

// ExecutionResult - represents a execution
type ExecutionResult struct {
	ExitCode int
	StdOut   string
	StdErr   string
}

// Environment -
type Environment interface {
	Name() string
	Execute(cmd []string, stdout func(string), stderr func(string)) (ExecutionResult, error)

	CopyTo(from string, to string) error
	CopyFrom(from string, to string) error

	Close()
}
