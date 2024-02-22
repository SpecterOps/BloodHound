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

package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/specterops/bloodhound/src/serde"
)

var structTagRegex = regexp.MustCompile(`(\w+):"([^"]+)"`)
var ErrInvalidConfigurationPath = errors.New("unable to find a configuration element by path")

// taggedField represents a struct field by its index and a parsed representation of any tags associated with the
// struct field.
type taggedField struct {
	Field int
	Tag   tag
}

// tag is a parsed struct tag definition. It contains an encoding which represents the prefix of a golang struct tag.
// All other arguments in the struct tag are encoded as an array of values.
//
// For example:
//
//	type A struct {
//		  Test string `json:"test"`
//	}
//
// Would contain a tag struct defined as such:
//
//	tag {
//	   Prefix: "json",
//	   Arguments: []string{
//	      "test",
//	   },
//	}
type tag struct {
	Prefix    string
	Arguments []string
}

// Name returns the first argument of the tag argument list or "" if the tag contained no arguments.
func (s tag) Name() string {
	if len(s.Arguments) > 0 {
		return s.Arguments[0]
	}

	return ""
}

// parseTag attempts to parse the contents of a reflect.StructTag type and return the resulting values. This function
// returns false if the function is unable to parse the struct tag contents.
func parseTag(rawTag reflect.StructTag) (tag, bool) {
	var tag tag

	if capture := structTagRegex.FindStringSubmatch(string(rawTag)); capture != nil {
		tag.Prefix = capture[1]

		for _, tagArgument := range strings.Split(capture[2], ",") {
			tag.Arguments = append(tag.Arguments, strings.TrimSpace(tagArgument))
		}

		return tag, true
	}

	return tag, false
}

// parseTaggedFields takes a reflection type struct and returns an array of taggedField structs representing the
// values of all of the tags and their associated struct field index.
func parseTaggedFields(target reflect.Type) []taggedField {
	var taggedFields []taggedField

	for fieldIdx := 0; fieldIdx < target.NumField(); fieldIdx++ {
		field := target.Field(fieldIdx)

		if tag, ok := parseTag(field.Tag); ok {
			taggedFields = append(taggedFields, taggedField{
				Field: fieldIdx,
				Tag:   tag,
			})
		}
	}

	return taggedFields
}

func valueOf(target any) reflect.Value {
	if valueRef, isValueType := target.(reflect.Value); isValueType {
		return valueRef
	}

	return reflect.ValueOf(target)
}

// indirectOf takes an interface and inspects if it's a pointer type. If so, the function returns the reflection value
// type of the value the pointer references. If not, the reflection value of the interface is returned.
func indirectOf(target any) reflect.Value {
	if valueRef := valueOf(target); valueRef.Kind() == reflect.Ptr {
		return reflect.Indirect(valueRef)
	} else {
		return valueRef
	}
}

// setRawValue takes a target interface and a string value representation. This function only supports golang primitive
// types. The type of the target is switched on which selects the appropriate parsing function for the string value. If
// an error is encountered during parsing of the string value, the resulting parsing error is returned.
func setRawValue(targetAddr any, value string) error {
	switch casted := targetAddr.(type) {
	case *uint:
		if parsed, err := strconv.ParseUint(value, 10, 64); err != nil {
			return err
		} else {
			*casted = uint(parsed)
		}

	case *uint8:
		if parsed, err := strconv.ParseUint(value, 10, 8); err != nil {
			return err
		} else {
			*casted = uint8(parsed)
		}

	case *uint16:
		if parsed, err := strconv.ParseUint(value, 10, 64); err != nil {
			return err
		} else {
			*casted = uint16(parsed)
		}

	case *uint32:
		if parsed, err := strconv.ParseUint(value, 10, 64); err != nil {
			return err
		} else {
			*casted = uint32(parsed)
		}

	case *uint64:
		if parsed, err := strconv.ParseUint(value, 10, 64); err != nil {
			return err
		} else {
			*casted = parsed
		}

	case *int:
		if parsed, err := strconv.Atoi(value); err != nil {
			return err
		} else {
			*casted = parsed
		}

	case *int8:
		if parsed, err := strconv.ParseInt(value, 10, 8); err != nil {
			return err
		} else {
			*casted = int8(parsed)
		}

	case *int16:
		if parsed, err := strconv.ParseInt(value, 10, 16); err != nil {
			return err
		} else {
			*casted = int16(parsed)
		}

	case *int32:
		if parsed, err := strconv.ParseInt(value, 10, 32); err != nil {
			return err
		} else {
			*casted = int32(parsed)
		}

	case *int64:
		if parsed, err := strconv.ParseInt(value, 10, 64); err != nil {
			return err
		} else {
			*casted = parsed
		}

	case *float32:
		if parsed, err := strconv.ParseFloat(value, 32); err != nil {
			return err
		} else {
			*casted = float32(parsed)
		}

	case *float64:
		if parsed, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		} else {
			*casted = parsed
		}

	case *string:
		*casted = value

	case *bool:
		if parsed, err := strconv.ParseBool(value); err != nil {
			return err
		} else {
			*casted = parsed
		}

	case *time.Duration:
		if parsed, err := time.ParseDuration(value); err != nil {
			return err
		} else {
			*casted = parsed
		}

	case *serde.URL:
		if parsed, err := serde.ParseURL(value); err != nil {
			return err
		} else {
			*casted = parsed
		}

	default:
		return fmt.Errorf("unsupported type %T", targetAddr)
	}

	return nil
}

func SetValue(target any, path, value string) error {
	if path == "" {
		return fmt.Errorf("trying to set a value on the root configuration object is not allowed")
	}

	var (
		pathParts = strings.Split(path, environmentVariablePathSeparator)
		cursor    = indirectOf(target)
	)

	for idx := 0; idx < len(pathParts); idx++ {
		var (
			nextPathPart = pathParts[idx]
			cursorType   = cursor.Type()
			taggedFields = parseTaggedFields(cursorType)
		)

		found := false
		for _, taggedField := range taggedFields {
			taggedFieldName := taggedField.Tag.Name()

			if taggedFieldName == nextPathPart {
				cursor = cursor.Field(taggedField.Field)
				found = true
				break
			}

			lookahead := idx

			for lookahead < len(pathParts) {
				// Make sure to add one to lookahead, as we want to get the range starting at base and going 1 or more indexes beyond it
				remainingFullPath := strings.Join(pathParts[idx:lookahead+1], "_")

				if taggedFieldName == remainingFullPath {
					cursor = cursor.Field(taggedField.Field)
					found = true
					idx = lookahead
					break
				}

				lookahead++
			}

			if found {
				break
			}
		}

		if !found {
			return fmt.Errorf("%w: %s", ErrInvalidConfigurationPath, path)
		}
	}

	if !cursor.CanAddr() {
		return fmt.Errorf("type %s is not addressable from parent type %T", cursor.Type().Name(), target)
	}

	return setRawValue(cursor.Addr().Interface(), value)
}
