package environment

import (
	"fmt"
	"strings"
)

// implements a native environment ("bare metal")

// NativeEnvironment -
type NativeEnvironment struct{}

// Execute -
func (e *NativeEnvironment) Execute(cmd []string) (string, error) {
	fmt.Printf("executing '%s'\n", strings.Join(cmd, " "))
	return "", nil
}
