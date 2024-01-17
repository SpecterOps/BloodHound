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

package v2

import "time"

// Constants and functions relating to pagination

const (
	// DefaultTimeRangeDuration is the default time range for range query parameters: 30 days prior to now
	DefaultTimeRangeDuration = -30 * 24 * time.Hour
)

func DefaultTimeRange() (time.Time, time.Time) {
	now := time.Now()

	return now, now.Add(DefaultTimeRangeDuration)
}
