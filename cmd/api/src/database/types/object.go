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

package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"github.com/specterops/bloodhound/errors"
)

type JSONBObject struct {
	scannedBytes []byte
	Object       any
}

func NewJSONBObject(object any) (JSONBObject, error) {
	bytes, err := json.Marshal(object)
	if err != nil {
		return JSONBObject{}, fmt.Errorf("error marshaling scannedBytes for JSONBObject: %w", err)
	}

	return JSONBObject{
		scannedBytes: bytes,
		Object:       object,
	}, nil
}

// Scan parses the input value (expected to be JSON) to []byte and then attempts to unmarshal it into the receiver
func (s *JSONBObject) Scan(value any) error {
	if bytes, typeOK := value.([]byte); !typeOK {
		return fmt.Errorf("expected JSONB type of []byte but received %T", value)
	} else {
		s.scannedBytes = bytes
	}

	return nil
}

// Map maps the value of the JSON blob onto the given interface
func (s *JSONBObject) Map(target any) error {
	if len(s.scannedBytes) == 0 {
		if s.Object == nil {
			return errors.Error("JSONObject is nil")
		}

		if content, err := json.Marshal(s.Object); err != nil {
			return err
		} else {
			s.scannedBytes = content
		}
	}

	return json.Unmarshal(s.scannedBytes, target)
}

// The raw byte slice is stored in the object, so only return that part of the object and let
// the json lib take over from there. This prevents the value from being returned under the
// Object attribute.
func (s *JSONBObject) MarshalJSON() ([]byte, error) {
	return s.scannedBytes, nil
}

// Override default unmarshal behavior to save raw data in scannedBytes and to put the
// unmarshalled value directly into Object. This works around the lack of Object
// attribute in the displayed JSON.
func (s *JSONBObject) UnmarshalJSON(data []byte) error {
	var object interface{}

	if err := json.Unmarshal(data, &object); err != nil {
		return fmt.Errorf("error unmarshaling data for JSONBObject: %w", err)
	} else {
		s.Object = object
		s.scannedBytes = data

		return nil
	}
}

// Value returns the json-marshaled value of the receiver
func (s JSONBObject) Value() (driver.Value, error) {
	return json.Marshal(s.Object)
}

// GormDBDataType returns JSONB if postgres, otherwise panics due to lack of DB type support
func (s JSONBObject) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch dbDialect := db.Dialector.Name(); dbDialect {
	case "postgres":
		return "JSONB"

	default:
		panic(fmt.Sprintf("Unsupported database dialect for JSON datatype: %s", dbDialect))
	}
}

type JSONUntypedObject map[string]any

// Scan parses the input value (expected to be JSON) to []byte and then attempts to unmarshal it into the receiver
func (s *JSONUntypedObject) Scan(value any) error {
	if bytes, ok := value.([]byte); !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	} else {
		if err := json.Unmarshal(bytes, s); err != nil {
			return err
		}

		return nil
	}
}

func (s JSONUntypedObject) Map(value any) error {
	return nil
}

// Value returns the json-marshaled value of the receiver
func (s JSONUntypedObject) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// GormDBDataType returns JSONB if postgres, otherwise panics due to lack of DB type support
func (s JSONUntypedObject) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch dbDialect := db.Dialector.Name(); dbDialect {
	case "postgres":
		return "JSONB"

	default:
		panic(fmt.Sprintf("Unsupported database dialect for JSON datatype: %s", dbDialect))
	}
}
