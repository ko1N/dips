package environment

// Environment -
type Environment interface {
	Execute([]string) (string, error)
}
