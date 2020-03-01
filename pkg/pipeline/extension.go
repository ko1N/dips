package pipeline

// Extension -
type Extension interface {
	Name() string
	Command() string

	StartPipeline(ctx ExecutionContext) error
	FinishPipeline(ctx ExecutionContext) error

	Execute(ctx ExecutionContext, cmd []string) error
}
