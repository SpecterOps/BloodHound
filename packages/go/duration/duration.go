// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// duration provides a shared duration type that covers missing generic functionality
// from the builtin time.duration type.
package duration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
)

const (
	// Date components require at least one unit in Y/M/D order.
	// Months are recognized here so the parser can report its unsupported-month error.
	iso8601DateComponentPattern = `(?:\d+Y(?:\d+M)?(?:\d+D)?|\d+M(?:\d+D)?|\d+D)`
	// Time components require at least one unit in H/M/S order.
	iso8601TimeComponentPattern = `(?:\d+H(?:\d+M)?(?:\d+S)?|\d+M(?:\d+S)?|\d+S)`
)

// Needed to ensure a complete ISO duration; the parser otherwise accepts -P7D and P7Dfoo.
// ^P and $ anchor the input; | separates weeks, date units with optional time, and time only.
// \d+ matches one or more digits; W means weeks and T starts the time section.
// (?:...) groups without capturing; ? makes the preceding group optional.
var iso8601DurationPattern = regexp.MustCompile(
	`^P(?:\d+W|` + iso8601DateComponentPattern + `(?:T` + iso8601TimeComponentPattern + `)?|T` + iso8601TimeComponentPattern + `)$`,
)

// ISODuration wraps time.Duration for json unmarshalling from ISO duration values.
// Note durations from the ISO duration format are inexact.
// The builtin type conversions only guarantee support for days, hours, and minutes and
// do not support negative durations.
type ISODuration time.Duration

// UnmarshalJSON validates a complete ISO duration string before converting it to time.Duration.
func (s *ISODuration) UnmarshalJSON(b []byte) error {
	var durationString string
	if err := json.Unmarshal(b, &durationString); err != nil {
		return err
	}

	if !iso8601DurationPattern.MatchString(durationString) {
		return fmt.Errorf("invalid ISO 8601 duration: %w", iso8601.ErrBadFormat)
	}

	// Parse the ISO 8601 duration string (e.g., "PT1H30M")
	parsed, err := iso8601.FromString(durationString)
	if err != nil {
		return fmt.Errorf("invalid ISO 8601 duration: %w", err)
	}

	// Convert the parsed struct into a standard time.Duration
	*s = ISODuration(parsed.ToDuration())
	return nil
}

func (s ISODuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ToISO8601String())
}

// ToISO8601String only supports days, hours and minutes.
// since ISO durations are inexact to begin with, we trucate to a minute.
// Does not support negative values
func (s *ISODuration) ToISO8601String() string {
	d := time.Duration(*s)
	if d == 0 {
		return "PT0S"
	}

	if d < 0 {
		d = -d
	}

	result := "P"

	days := int(d / (24 * time.Hour))
	if days > 0 {
		result += fmt.Sprintf("%dD", days)
	}

	hours := int(d/(time.Hour)) % 24
	minutes := int(d/(time.Minute)) % 60

	if hours > 0 || minutes > 0 {
		result += "T"

		if hours > 0 {
			result += fmt.Sprintf("%dH", hours)
		}
		if minutes > 0 {
			result += fmt.Sprintf("%dM", minutes)
		}
	}

	// Fallback if everything truncated to 0 but duration is non-zero
	if days == 0 && hours == 0 && minutes == 0 {
		return "PT0S"
	}

	return result
}
