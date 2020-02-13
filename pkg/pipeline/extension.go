package pipeline

import (
	log "github.com/inconshreveable/log15"

	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
)

// Extension -
type Extension interface {
	Name() string
	Command() string
	Execute(pipelog log.Logger, env environment.Environment, cmd []string) error
}
