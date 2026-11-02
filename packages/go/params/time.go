// Copyright 2026 Specter Ops, Inc.
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

// This file provides a storage-agnostic model for parsing and validating sort query parameters.
// It intentionally carries no knowledge of how the resulting order is ultimately applied (SQL, graph,
// etc.) so that it can be reused across the API without coupling to any particular persistence layer.

package params

import (
	"net/url"
	"time"
)

const (
	// DefaultTimeRangeDuration is the default time range for range query parameters: 30 days prior to now
	DefaultTimeRangeDuration = -30 * 24 * time.Hour
)

func DefaultTimeRange() (time.Time, time.Time) {
	now := time.Now()

	return now, now.Add(DefaultTimeRangeDuration)
}

func ParseTimeQueryParameter(params url.Values, key string, defaultValue time.Time) (time.Time, error) {
	if param := params.Get(key); param != "" {
		return time.Parse(time.RFC3339Nano, param)
	}

	return defaultValue, nil
}
