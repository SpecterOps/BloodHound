package environment

import (
	"os"
	"strings"
)

// Environment is a string map representation of env vars
type Environment map[string]string

// NewEnvironment pulls os.Environ and converts to an Environment
func NewEnvironment() Environment {
	var (
		envVars = os.Environ()
		envMap  = make(Environment, len(envVars))
	)

	for _, env := range os.Environ() {
		envTuple := strings.SplitN(env, "=", 2)
		envMap[envTuple[0]] = envTuple[1]
	}

	return envMap
}

// SetIfEmpty sets a value only if the key currently has no value
func (s Environment) SetIfEmpty(key string, value string) {
	if _, ok := s[key]; !ok {
		s[key] = value
	}
}

// Slice converts the Environment to a slice of strings in the form `KEY=VALUE` to send to external libraries
func (s Environment) Slice() []string {
	var envSlice = make([]string, 0, len(s))
	for key, val := range s {
		envSlice = append(envSlice, strings.Join([]string{key, val}, "="))
	}

	return envSlice
}
