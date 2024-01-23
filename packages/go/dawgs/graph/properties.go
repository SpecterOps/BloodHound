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

package graph

import (
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

func formatPropertyTypeError(expectedTypeName string, value any) error {
	return fmt.Errorf("expected property type %s but received %T", expectedTypeName, value)
}

func int64SliceToIDSlice(int64Slice []int64) []ID {
	idSlice := make([]ID, len(int64Slice))

	for idx := 0; idx < len(int64Slice); idx++ {
		idSlice[idx] = ID(int64Slice[idx])
	}

	return idSlice
}

func interfaceSliceToStringSlice(value any) ([]string, error) {
	if typedSlice, typeOk := value.([]any); !typeOk {
		return nil, formatPropertyTypeError("[]any", value)
	} else {
		stringSlice := make([]string, len(typedSlice))

		for idx := 0; idx < len(typedSlice); idx++ {
			switch typedValue := typedSlice[idx].(type) {
			case string:
				stringSlice[idx] = typedValue
			default:
				return nil, formatPropertyTypeError("string", typedSlice[idx])
			}
		}

		return stringSlice, nil
	}
}

func interfaceSliceToInt64Slice(value any) ([]int64, error) {
	if typedSlice, typeOK := value.([]any); !typeOK {
		return nil, formatPropertyTypeError("[]any", value)
	} else {
		int64Slice := make([]int64, len(typedSlice))

		for idx := 0; idx < len(typedSlice); idx++ {
			switch typedValue := typedSlice[idx].(type) {
			case int64:
				int64Slice[idx] = typedValue

			case float64:
				int64Slice[idx] = int64(typedValue)

			default:
				return nil, formatPropertyTypeError("int64", typedSlice[idx])
			}
		}

		return int64Slice, nil
	}
}

// safePropertyValue is a generic implementation of the PropertyValue interface.
type safePropertyValue struct {
	key          string
	value        any
	defaultValue any
}

func (s safePropertyValue) getValue() (any, error) {
	if s.IsNil() {
		if s.defaultValue != nil {
			return s.defaultValue, nil
		}

		return nil, fmt.Errorf("property %s: %w", s.key, ErrPropertyNotFound)
	} else {
		return s.value, nil
	}
}

func (s safePropertyValue) IsNil() bool {
	return s.value == nil
}

func (s safePropertyValue) Any() any {
	return s.value
}

func (s safePropertyValue) Int64Slice() ([]int64, error) {
	if rawValue, err := s.getValue(); err != nil {
		return nil, err
	} else if slice, err := interfaceSliceToInt64Slice(rawValue); err != nil {
		return nil, err
	} else {
		return slice, nil
	}
}

func (s safePropertyValue) IDSlice() ([]ID, error) {
	if int64Slice, err := s.Int64Slice(); err != nil {
		return nil, err
	} else {
		return int64SliceToIDSlice(int64Slice), nil
	}
}

func (s safePropertyValue) StringSlice() ([]string, error) {
	if rawValue, err := s.getValue(); err != nil {
		return nil, err
	} else {
		return interfaceSliceToStringSlice(rawValue)
	}
}

func (s safePropertyValue) Bool() (bool, error) {
	if rawValue, err := s.getValue(); err != nil {
		return false, err
	} else if typedValue, typeOK := rawValue.(bool); !typeOK {
		err := formatPropertyTypeError("bool", rawValue)

		return false, err
	} else {
		return typedValue, nil
	}
}

func (s safePropertyValue) Int() (int, error) {
	if rawValue, err := s.getValue(); err != nil {
		return 0, err
	} else {
		return AsNumeric[int](rawValue)
	}
}

func (s safePropertyValue) Int64() (int64, error) {
	if rawValue, err := s.getValue(); err != nil {
		return 0, err
	} else {
		return AsNumeric[int64](rawValue)
	}
}

func (s safePropertyValue) Uint64() (uint64, error) {
	if rawValue, err := s.getValue(); err != nil {
		return 0, err
	} else {
		return AsNumeric[uint64](rawValue)
	}
}

func (s safePropertyValue) Float64() (float64, error) {
	if rawValue, err := s.getValue(); err != nil {
		return 0, err
	} else {
		return AsNumeric[float64](rawValue)
	}
}

func (s safePropertyValue) String() (string, error) {
	if rawValue, err := s.getValue(); err != nil {
		return "", err
	} else if typedValue, typeOK := rawValue.(string); !typeOK {
		err := formatPropertyTypeError("string", rawValue)

		return "", err
	} else {
		return typedValue, nil
	}
}

func (s safePropertyValue) Time() (time.Time, error) {
	if rawValue, err := s.getValue(); err != nil {
		return time.Time{}, err
	} else {
		switch typed := rawValue.(type) {
		case string:
			if parsedTime, err := time.Parse(time.RFC3339Nano, typed); err != nil {
				return time.Time{}, err
			} else {
				return parsedTime, nil
			}

		case dbtype.Time:
			return typed.Time(), nil

		case dbtype.LocalTime:
			return typed.Time(), nil

		case dbtype.Date:
			return typed.Time(), nil

		case dbtype.LocalDateTime:
			return typed.Time(), nil

		case float64:
			return time.Unix(int64(typed), 0), nil

		case int64:
			return time.Unix(typed, 0), nil

		case time.Time:
			return typed, nil

		default:
			err := fmt.Errorf("invalid time type %T - expected either 'string' or 'time.Time'", rawValue)

			return time.Time{}, err
		}
	}
}

// NewPropertyResult takes a bare any type and returns a generic type negotiation wrapper that adheres to the
// PropertyValue interface.
func NewPropertyResult(key string, value any) PropertyValue {
	return &safePropertyValue{
		key:   key,
		value: value,
	}
}

// Properties is a map type that satisfies the Properties interface.
type Properties struct {
	Map      map[string]any      `json:"map"`
	Deleted  map[string]struct{} `json:"deleted"`
	Modified map[string]struct{} `json:"modified"`
}

func (s *Properties) Merge(other *Properties) {
	for otherKey, otherValue := range other.Map {
		s.Map[otherKey] = otherValue
	}

	for otherModifiedKey := range other.Modified {
		s.Modified[otherModifiedKey] = struct{}{}

		delete(s.Deleted, otherModifiedKey)
	}

	for otherDeletedKey := range other.Deleted {
		s.Deleted[otherDeletedKey] = struct{}{}

		delete(s.Map, otherDeletedKey)
		delete(s.Modified, otherDeletedKey)
	}
}

func (s *Properties) MapOrEmpty() map[string]any {
	if s == nil || s.Map == nil {
		return map[string]any{}
	}

	return s.Map
}

func (s *Properties) SizeOf() size.Size {
	instanceSize := size.Of(*s)

	if s.Map != nil {
		instanceSize += size.OfMapValues(s.Map)
	}

	if s.Deleted != nil {
		instanceSize += size.OfMapValues(s.Deleted)
	}

	if s.Modified != nil {
		instanceSize += size.OfMapValues(s.Modified)
	}

	return instanceSize
}

func (s *Properties) Len() int {
	if s.Map == nil {
		return 0
	}

	return len(s.Map)
}

func (s *Properties) ModifiedProperties() map[string]any {
	if s.Modified == nil {
		return map[string]any{}
	}

	properties := make(map[string]any, len(s.Modified))

	for key := range s.Modified {
		properties[key] = s.Map[key]
	}

	return properties
}

func (s *Properties) DeletedProperties() []string {
	if s.Deleted == nil {
		return nil
	}

	deletedPropertyKeys := make([]string, 0, len(s.Deleted))

	for key := range s.Deleted {
		deletedPropertyKeys = append(deletedPropertyKeys, key)
	}

	return deletedPropertyKeys
}

// Get fetches a value from the Properties by key and returns a tuple containing the value and a boolean informing
// the caller if the value was found. If the value was not found the value portion of the return tuple is nil.
func (s *Properties) Get(key string) PropertyValue {
	if s.Map == nil {
		return NewPropertyResult(key, nil)
	}

	return NewPropertyResult(key, s.Map[key])
}

// Set sets a value within the PropertyMap.
func (s *Properties) Set(key string, value any) *Properties {
	if s.Map == nil {
		s.Map = map[string]any{
			key: value,
		}
	} else {
		s.Map[key] = value
	}

	// If we set a property we track it as modified and remove it from the deleted properties index
	if s.Modified == nil {
		s.Modified = map[string]struct{}{
			key: {},
		}
	} else {
		s.Modified[key] = struct{}{}
	}

	if s.Deleted != nil {
		delete(s.Deleted, key)
	}

	return s
}

func (s *Properties) SetAll(other map[string]any) *Properties {
	for key, value := range other {
		s.Set(key, value)
	}

	return s
}

func (s *Properties) Clone() *Properties {
	newProperties := &Properties{
		Map:      map[string]any{},
		Deleted:  map[string]struct{}{},
		Modified: map[string]struct{}{},
	}

	for key, value := range s.Map {
		newProperties.Map[key] = value
	}

	for key := range s.Modified {
		newProperties.Modified[key] = struct{}{}
	}

	for key := range s.Deleted {
		newProperties.Deleted[key] = struct{}{}
	}

	return newProperties
}

// Exists returns true if a value exists for the given key, false otherwise.
func (s *Properties) Exists(key string) bool {
	if s.Map == nil {
		return false
	}

	_, found := s.Map[key]
	return found
}

// GetOrDefault fetches a value from the Properties by key. If the key is not present in the Properties this
// function returns the given default value instead.
func (s *Properties) GetOrDefault(key string, defaultValue any) PropertyValue {
	value := defaultValue

	if s.Map != nil {
		if mapValue, found := s.Map[key]; found && mapValue != nil {
			value = mapValue
		}
	}

	return NewPropertyResult(key, value)
}

func (s *Properties) Delete(key string) *Properties {
	if s.Map != nil {
		delete(s.Map, key)
	}

	if s.Deleted == nil {
		s.Deleted = map[string]struct{}{
			key: {},
		}
	} else {
		s.Deleted[key] = struct{}{}
	}

	if s.Modified != nil {
		delete(s.Modified, key)
	}

	return s
}

// TODO: This function does not correctly communicate that it is lazily instantiated
func NewProperties() *Properties {
	return &Properties{}
}

func NewPropertiesRed() *Properties {
	return &Properties{
		Map:      map[string]any{},
		Modified: make(map[string]struct{}),
		Deleted:  make(map[string]struct{}),
	}
}

type PropertyMap map[String]any

func symbolMapToStringMap(props map[String]any) map[string]any {
	store := make(map[string]any, len(props))

	for k, v := range props {
		store[k.String()] = v
	}

	return store
}

func AsProperties[T PropertyMap | map[String]any | map[string]any](rawStore T) *Properties {
	var store map[string]any

	switch typedStore := any(rawStore).(type) {
	case PropertyMap:
		store = symbolMapToStringMap(typedStore)

	case map[String]any:
		store = symbolMapToStringMap(typedStore)

	case map[string]any:
		store = typedStore
	}

	return &Properties{
		Map:      store,
		Modified: make(map[string]struct{}),
		Deleted:  make(map[string]struct{}),
	}
}
