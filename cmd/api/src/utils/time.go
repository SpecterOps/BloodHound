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
	"time"
)

// ReadRFC3339 attempts to parse a RFC3339 string into a time.Time struct.
//
// If the RFC3339 string is empty this function returns the given default value. If the string is not formatted as a
// valid RFC3339 datetime then this function will return the resulting error.
func ReadRFC3339(valueStr string, defaultValue time.Time) (time.Time, error) {
	if valueStr == "" {
		return defaultValue, nil
	}

	return time.Parse(time.RFC3339Nano, valueStr)
}
