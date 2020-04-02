package pipeline

import (
	"context"

	"github.com/d5/tengo/v2"
)

// Expression - Describes a expression which evaluates to a bool
type Expression struct {
	Script string
}

// Evaluate - Evaluates the expression to a bool
func (e *Expression) Evaluate(variables map[string]tengo.Object) (string, error) {
	// TODO: handle errors
	src := `out := ` + e.Script
	script := tengo.NewScript([]byte(src))

	// setup variables
	for k, v := range variables {
		_ = script.Add(k, v)
	}

	// compile script
	compiled, err := script.RunContext(context.Background())
	if err != nil {
		return "", err
	}

	// evaluate output
	out := compiled.Get("out")
	return out.String(), nil
}
