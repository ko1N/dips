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
	Execute([]string, func(string), func(string)) (ExecutionResult, error)
	Close()
}
