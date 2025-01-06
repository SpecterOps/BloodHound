// Copyright 2024 Specter Ops, Inc.
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

package translate

type IndexedSlice[K comparable, V any] struct {
	index  map[K]int
	values []V
}

func NewIndexedSlice[K comparable, V any]() *IndexedSlice[K, V] {
	return &IndexedSlice[K, V]{
		index: make(map[K]int),
	}
}

func (s *IndexedSlice[K, V]) Keys() []K {
	keys := make([]K, 0, len(s.index))

	for key := range s.index {
		keys = append(keys, key)
	}

	return keys
}

func (s *IndexedSlice[K, V]) Values() []V {
	return s.values
}

func (s *IndexedSlice[K, V]) Merge(other *IndexedSlice[K, V]) {
	for key, idx := range other.index {
		s.Put(key, other.values[idx])
	}
}

// Len returns the number of values stored.
func (s *IndexedSlice[K, V]) Len() int {
	return len(s.values)
}

func (s *IndexedSlice[K, V]) Get(key K) (V, bool) {
	if valueIdx, hasValue := s.index[key]; hasValue {
		return s.values[valueIdx], true
	}

	var empty V
	return empty, false
}

func (s *IndexedSlice[K, V]) Has(key K) bool {
	_, hasValue := s.index[key]
	return hasValue
}

func (s *IndexedSlice[K, V]) GetOr(key K, defaultConstructor func() V) V {
	if valueIdx, hasValue := s.index[key]; hasValue {
		return s.values[valueIdx]
	}

	defaultValue := defaultConstructor()

	s.Put(key, defaultValue)
	return defaultValue
}

// CheckedGet returns a tuple containing the value and a boolean representing if a value was found for the
// given key.
func (s *IndexedSlice[K, V]) CheckedGet(key K) (V, bool) {
	if valueIdx, hasValue := s.index[key]; hasValue {
		return s.values[valueIdx], true
	}

	var empty V
	return empty, false
}

// GetAll returns all found values for a given slice of keys. Any keys that do not have stored values
// in this IndexedSlice are returned as the second value of the tuple return for this function.
func (s *IndexedSlice[K, V]) GetAll(keys []K) ([]V, []K) {
	var (
		values      = make([]V, 0, len(keys))
		missingKeys = make([]K, 0, len(keys))
	)

	for _, key := range keys {
		if valueIdx, hasValue := s.index[key]; hasValue {
			values = append(values, s.values[valueIdx])
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	return values, missingKeys
}

// GetAllIndexed returns all found values for a given slice of keys. Any keys that do not have stored values
// in this IndexedSlice are returned as the second value of the tuple return for this function.
func (s *IndexedSlice[K, V]) GetAllIndexed(keys []K) (*IndexedSlice[K, V], []K) {
	var (
		values      = NewIndexedSlice[K, V]()
		missingKeys = make([]K, 0, len(keys))
	)

	for _, key := range keys {
		if valueIdx, hasValue := s.index[key]; hasValue {
			values.Put(key, s.values[valueIdx])
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	return values, missingKeys
}

// Put inserts the given value with the given key.
func (s *IndexedSlice[K, V]) Put(key K, value V) {
	if valueIdx, hasValue := s.index[key]; hasValue {
		s.values[valueIdx] = value
	} else {
		s.values = append(s.values, value)
		s.index[key] = len(s.values) - 1
	}
}

// WithPut inserts the given value with the given key and then returns the indexed slice pointer.
func (s *IndexedSlice[K, V]) WithPut(key K, value V) *IndexedSlice[K, V] {
	s.Put(key, value)

	// This is fluent to allow for additions during construction
	return s
}

func (s *IndexedSlice[K, V]) Each(delegate func(key K, value V) bool) {
	for id, idx := range s.index {
		if !delegate(id, s.values[idx]) {
			break
		}
	}
}

func (s *IndexedSlice[K, V]) Copy() *IndexedSlice[K, V] {
	sCopy := &IndexedSlice[K, V]{
		index:  s.index,
		values: make([]V, len(s.values)),
	}

	copy(sCopy.values, s.values)
	return sCopy
}
