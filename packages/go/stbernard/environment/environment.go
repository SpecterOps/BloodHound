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
	"os"
	"strings"

	"github.com/specterops/bloodhound/log"
)

const (
	LogLevelVarName = "SB_LOG_LEVEL"
	VersionVarName  = "SB_VERSION"
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

// Overrides an environment variable with a new value
func (s Environment) Override(key string, value string) {
	log.Infof("Overriding environment variable %s with %s", key, value)
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
