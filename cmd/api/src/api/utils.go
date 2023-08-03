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

package api

import (
	"fmt"
	"strconv"
)

func ParseOptionalBool(value string, defaultValue bool) (bool, error) {
	if value != "" {
		if boolValue, err := strconv.ParseBool(value); err != nil {
			return defaultValue, fmt.Errorf("error converting value to boolean, defaulting to %t: %w", defaultValue, err)
		} else {
			return boolValue, nil
		}
	} else {
		return defaultValue, nil
	}
}

func FilterStructSlice[T any](slice []T, conditional func(t T) bool) []T {
	result := []T{}
	for _, v := range slice {
		if conditional(v) {
			result = append(result, v)
		}
	}
	return result
}
