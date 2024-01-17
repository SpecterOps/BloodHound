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

package null

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Time is a nullable time.Time. It supports SQL and JSON serialization.
// It will marshal to null if null.
type Time struct {
	sql.NullTime
}

// Value implements the driver Valuer interface.
func (s Time) Value() (driver.Value, error) {
	if !s.Valid {
		return nil, nil
	}
	return s.Time, nil
}

// NewTime creates a new Time.
func NewTime(t time.Time, valid bool) Time {
	return Time{
		NullTime: sql.NullTime{
			Time:  t,
			Valid: valid,
		},
	}
}

// TimeFrom creates a new Time that will always be valid.
func TimeFrom(t time.Time) Time {
	return NewTime(t, true)
}

// TimeFromPtr creates a new Time that will be null if t is nil.
func TimeFromPtr(t *time.Time) Time {
	if t == nil {
		return NewTime(time.Time{}, false)
	}
	return NewTime(*t, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (s Time) ValueOrZero() time.Time {
	if !s.Valid {
		return time.Time{}
	}
	return s.Time
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this time is null.
func (s Time) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return s.Time.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input.
func (s *Time) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(jsonNullLiteral)) {
		s.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &s.Time); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.Valid = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It returns an empty string if invalid, otherwise time.Time's MarshalText.
func (s Time) MarshalText() ([]byte, error) {
	if !s.Valid {
		return []byte{}, nil
	}
	return s.Time.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It has backwards compatibility with v3 in that the string "null" is considered equivalent to an empty string
// and unmarshaling will succeed. This may be removed in a future version.
func (s *Time) UnmarshalText(text []byte) error {
	str := string(text)
	// allowing "null" is for backwards compatibility with v3
	if str == "" || str == "null" {
		s.Valid = false
		return nil
	}
	if err := s.Time.UnmarshalText(text); err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}
	s.Valid = true
	return nil
}

// SetValid changes this Time's value and sets it to be non-null.
func (s *Time) SetValid(v time.Time) {
	s.Time = v
	s.Valid = true
}

// Ptr returns a pointer to this Time's value, or a nil pointer if this Time is null.
func (s Time) Ptr() *time.Time {
	if !s.Valid {
		return nil
	}
	return &s.Time
}

// IsZero returns true for invalid Times, hopefully for future omitempty support.
// A non-null Time with a zero value will not be considered zero.
func (s Time) IsZero() bool {
	return !s.Valid
}

// Equal returns true if both Time objects encode the same time or are both null.
// Two times can be equal even if they are in different locations.
// For example, 6:00 +0200 CEST and 4:00 UTC are Equal.
func (s Time) Equal(other Time) bool {
	return s.Valid == other.Valid && (!s.Valid || s.Time.Equal(other.Time))
}

// ExactEqual returns true if both Time objects are equal or both null.
// ExactEqual returns false for times that are in different locations or
// have a different monotonic clock reading.
func (s Time) ExactEqual(other Time) bool {
	return s.Valid == other.Valid && (!s.Valid || s.Time == other.Time)
}
