package environment

import (
	"os"
	"strings"
)

type Environment map[string]string

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

func (s Environment) Slice() []string {
	var envSlice = make([]string, 0, len(s))
	for key, val := range s {
		envSlice = append(envSlice, strings.Join([]string{key, val}, "="))
	}

	return envSlice
}
