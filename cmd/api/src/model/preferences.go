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

package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// PreferenceItem represents a single user preference with its value and an enterprise flag.
type PreferenceItem struct {
	Value      any  `json:"value"`
	Enterprise bool `json:"enterprise"`
}

// Preferences represents a map of preference names to their PreferenceItem values, stored as JSONB in the database.
type Preferences map[string]PreferenceItem

// Scan verifies that the input value is []byte (the Go representation of JSONB) and unmarshals it into the receiver.
// Uses a pointer receiver because it must modify the map. This matches the pattern in types.JSONUntypedObject.
// sql.Scanner implementation
func (s *Preferences) Scan(value any) error {
	var (
		bytes  []byte
		typeOK bool
	)
	if bytes, typeOK = value.([]byte); !typeOK {
		return fmt.Errorf("expected []byte representation of JSONB, received %T", value)
	}

	return json.Unmarshal(bytes, s)
}

// Value marshals the Preferences map into []byte for storage as JSONB.
// driver.Valuer implementation
func (s Preferences) Value() (driver.Value, error) {
	if len(s) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(s)
}

// GormDBDataType returns JSONB if postgres, otherwise panics due to lack of DB type support.
// GORM schema.DataTypeInterface implementation
func (s Preferences) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch dbDialect := db.Name(); dbDialect {
	case "postgres":
		return "JSONB"

	default:
		panic(fmt.Sprintf("Unsupported database dialect for JSON datatype: %s", dbDialect))
	}
}
