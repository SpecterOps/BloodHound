// Copyright 2024 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

const (
	LogLevelVarName           = "SB_LOG_LEVEL"
	VersionVarName            = "SB_VERSION"
	PostgresConnectionVarName = "SB_PG_CONNECTION"
	YarnCmdVarName            = "SB_YARN_CMD"
)

// Environment is a string map representation of env vars
type Environment map[string]string

// NewEnvironment pulls os.Environ and converts to an Environment
func NewEnvironment() (Environment, error) {
	var (
		envVars = os.Environ()
		envMap  = make(Environment, len(envVars))
	)

	for _, env := range os.Environ() {
		envTuple := strings.SplitN(env, "=", 2)
		envMap[envTuple[0]] = envTuple[1]
	}

	// If yarn isn't available, it's catastrophic
	err := envMap.SetExecIfEmpty(YarnCmdVarName, []string{"yarn", "yarnpkg"})
	if err != nil {
		return nil, err
	}

	return envMap, nil
}

// SetIfEmpty sets a value only if the key currently has no value
func (s Environment) SetIfEmpty(key string, value string) {
	if _, ok := s[key]; !ok {
		s[key] = value
	}
}

// SetExecIfEmpty sets an system executable from an array of options if empty
func (s Environment) SetExecIfEmpty(key string, execCmds []string) error {
	if _, ok := s[key]; !ok {
		for _, execCmd := range execCmds {
			_, err := exec.LookPath(execCmd)
			if err == nil {
				s[key] = execCmd
				return nil
			}
		}
		options := strings.Join(execCmds, ", ")
		return fmt.Errorf("unable to locate any of %s executable(s) in path for undefined env %s", options, key)
	}
	return nil
}

// Overrides an environment variable with a new value
func (s Environment) Override(key string, value string) {
	slog.Info(fmt.Sprintf("Overriding environment variable %s with %s", key, value))
	s[key] = value
}

// Slice converts the Environment to a slice of strings in the form `KEY=VALUE` to send to external libraries
func (s Environment) Slice() []string {
	var envSlice = make([]string, 0, len(s))
	for key, val := range s {
		envSlice = append(envSlice, strings.Join([]string{key, val}, "="))
	}

	return envSlice
}
