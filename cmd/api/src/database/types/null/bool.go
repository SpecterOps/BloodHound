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
)

// Bool is a nullable bool.
// It does not consider false values to be null.
// It will decode to null, not false, if null.
type Bool struct {
	sql.NullBool
}

// NewBool creates a new Bool
func NewBool(b bool, valid bool) Bool {
	return Bool{
		NullBool: sql.NullBool{
			Bool:  b,
			Valid: valid,
		},
	}
}

// BoolFrom creates a new Bool that will always be valid.
func BoolFrom(b bool) Bool {
	return NewBool(b, true)
}

// BoolFromPtr creates a new Bool that will be null if f is nil.
func BoolFromPtr(b *bool) Bool {
	if b == nil {
		return NewBool(false, false)
	}
	return NewBool(*b, true)
}

// ValueOrZero returns the inner value if valid, otherwise false.
func (s Bool) ValueOrZero() bool {
	return s.Valid && s.Bool
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Bool.
func (s *Bool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(jsonNullLiteral)) {
		s.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &s.Bool); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (s *Bool) UnmarshalText(text []byte) error {
	str := string(text)
	switch str {
	case "", "null":
		s.Valid = false
		return nil
	case "true":
		s.Bool = true
	case "false":
		s.Bool = false
	default:
		return errors.New("null: invalid input for UnmarshalText:" + str)
	}
	s.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Bool is null.
func (s Bool) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	if !s.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Bool is null.
func (s Bool) MarshalText() ([]byte, error) {
	if !s.Valid {
		return []byte{}, nil
	}
	if !s.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// SetValid changes this Bool's value and also sets it to be non-null.
func (s *Bool) SetValid(v bool) {
	s.Bool = v
	s.Valid = true
}

// Ptr returns a pointer to this Bool's value, or a nil pointer if this Bool is null.
func (s Bool) Ptr() *bool {
	if !s.Valid {
		return nil
	}
	return &s.Bool
}

// IsZero returns true for invalid Bools, for future omitempty support (Go 1.4?)
// A non-null Bool with a 0 value will not be considered zero.
func (s Bool) IsZero() bool {
	return !s.Valid
}

// Equal returns true if both booleans have the same value or are both null.
func (s Bool) Equal(other Bool) bool {
	return s.Valid == other.Valid && (!s.Valid || s.Bool == other.Bool)
}
