package modules

import (
	"gitlab.strictlypaste.xyz/ko1n/dips/pkg/environment"
)

// pipeline module for ffmpeg

// FFMpeg -
type FFMpeg struct{}

// Execute -
func (e *FFMpeg) Name() string {
	return "FFMpeg"
}

// Execute -
func (e *FFMpeg) Command() string {
	return "ffmpeg"
}

// Execute -
func (e *FFMpeg) Execute(env environment.Environment, cmds []string) error {
	return nil
}
