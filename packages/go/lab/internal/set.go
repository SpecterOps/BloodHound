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

package internal

// Set is a generic collection construct that holds unique values.
// Values must be `comparable` and only occur once in a collection.
type Set[T comparable] Map[T, struct{}]

// Add inserts a value into the Set if there isn't already a pre-existing entry.
func (s Set[T]) Add(t T) Set[T] {
	s[t] = struct{}{}
	return s
}

// Delete removes a value from the Set if it exists.
func (s Set[T]) Delete(t T) Set[T] {
	delete(s, t)
	return s
}

// Has returns true if a value exists within the Set
func (s Set[T]) Has(t T) bool {
	_, ok := s[t]
	return ok
}

// Clone returns a shallow copy of the Set
func (s Set[T]) Clone() Set[T] {
	out := make(Set[T], len(s))
	for item := range s {
		out.Add(item)
	}
	return out
}

func (s Set[T]) AsSlice() []T {
	out := make([]T, len(s))
	for value := range s {
		out = append(out, value)
	}
	return out
}
