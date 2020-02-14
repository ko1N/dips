package environment

// implements a native environment ("bare metal")

// NativeEnvironment -
type NativeEnvironment struct{}

// Execute -
func (e *NativeEnvironment) Execute(cmd []string, stdout func(string), stderr func(string)) (string, error) {
	//fmt.Printf("executing '%s'\n", strings.Join(cmd, " "))
	return "", nil
}
