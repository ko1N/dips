package pipeline

import "github.com/ko1N/dips/pkg/pipeline/environments"

// Extension -
type Extension interface {
	Name() string
	Command() string

	StartPipeline(ctx *ExecutionContext) error
	FinishPipeline(ctx *ExecutionContext) error

	Execute(ctx *ExecutionContext, cmd string) (environments.ExecutionResult, error)
}
