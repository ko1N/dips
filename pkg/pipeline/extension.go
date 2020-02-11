package pipeline

import "gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"

// Extension -
type Extension interface {
	Name() string
	Command() string
	Execute(env environment.Environment, cmd []string) error
}
