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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Int32 is an nullable int32.
// It does not consider zero values to be null.
// It will decode to null, not zero, if null.
type Int32 struct {
	sql.NullInt32
}

// NewInt creates a new Int
func NewInt32(i int32, valid bool) Int32 {
	return Int32{
		NullInt32: sql.NullInt32{
			Int32: i,
			Valid: valid,
		},
	}
}

// Int64From creates a new Int that will always be valid.
func Int32From(i int32) Int32 {
	return NewInt32(i, true)
}

// Int64FromPtr creates a new Int that be null if i is nil.
func Int32FromPtr(i *int32) Int32 {
	if i == nil {
		return NewInt32(0, false)
	}
	return NewInt32(*i, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (s Int32) ValueOrZero() int32 {
	if !s.Valid {
		return 0
	}
	return s.Int32
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number, string, and null input.
// 0 will not be considered a null Int.
func (s *Int32) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(jsonNullLiteral)) {
		s.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &s.Int32); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need int or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseInt(str, 10, 32)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to int: %w", err)
			}
			s.Int32 = int32(n)
			s.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Int if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (s *Int32) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		s.Valid = false
		return nil
	}

	if intValue, err := strconv.ParseInt(string(text), 10, 32); err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	} else {
		s.Int32 = int32(intValue)
	}

	s.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Int is null.
func (s Int32) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(int64(s.Int32), 10)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Int is null.
func (s Int32) MarshalText() ([]byte, error) {
	if !s.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatInt(int64(s.Int32), 10)), nil
}

// SetValid changes this Int's value and also sets it to be non-null.
func (s *Int32) SetValid(n int32) {
	s.Int32 = n
	s.Valid = true
}

// Ptr returns a pointer to this Int's value, or a nil pointer if this Int is null.
func (s Int32) Ptr() *int32 {
	if !s.Valid {
		return nil
	}
	return &s.Int32
}

// IsZero returns true for invalid Ints, for future omitempty support (Go 1.4?)
// A non-null Int with a 0 value will not be considered zero.
func (s Int32) IsZero() bool {
	return !s.Valid
}

// Equal returns true if both ints have the same value or are both null.
func (s Int32) Equal(other Int32) bool {
	return s.Valid == other.Valid && (!s.Valid || s.Int32 == other.Int32)
}
