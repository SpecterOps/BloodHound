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

// Map is a generic construct that holds key-value pairs.
// Keys must be `comparable`, however, there is no constraint on the type of stored values.
type Map[T comparable, U any] map[T]U

// Set add or updates an entry in the Map with the specified key-value pair.
func (s Map[T, U]) Set(t T, u U) Map[T, U] {
	s[t] = u
	return s
}

// Delete removes the specified entry from the Map.
func (s Map[T, U]) Delete(t T) Map[T, U] {
	delete(s, t)
	return s
}

// Has returns true if a key exists within the Map
func (s Map[T, U]) Has(t T) bool {
	_, ok := s[t]
	return ok
}

// Clone returns a shallow copy of the Map.
func (s Map[T, U]) Clone() Map[T, U] {
	out := make(Map[T, U], len(s))
	for key, value := range s {
		out[key] = value
	}
	return out
}
