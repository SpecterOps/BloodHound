// Copyright 2023 Specter Ops, Inc.
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

package utils

import (
	"fmt"
	"strings"

	"github.com/gobeam/stringy"
)

var MapErrorsToStrings func(collection []error, iteratee func(value error) string) []string

func init() {
	Bind(&MapErrorsToStrings, GenericMap)
}

type Errors []error

func (s Errors) AsStringSlice() []string {
	return MapErrorsToStrings(s, func(value error) string {
		return value.Error()
	})
}

func (s Errors) Error() string {
	return strings.Join(s.AsStringSlice(), "; ")
}

// SwitchCase is a helper function to change string case.
// This was originally charted out to be used for formatting error responses
func SwitchCase(input string, format string) (string, error) {
	stringyInput := stringy.New(input)
	switch format {
	case "lower":
		return stringyInput.ToLower(), nil
	case "upper":
		return stringyInput.ToUpper(), nil
	case "camel":
		return stringyInput.CamelCase(), nil
	case "snake":
		return stringyInput.SnakeCase().ToLower(), nil
	case "kebab":
		return stringyInput.KebabCase().ToLower(), nil
	default:
		return "", fmt.Errorf("invalid return format specified")
	}
}
