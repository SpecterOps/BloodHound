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

package models

// Optional is a simple generic optional type.
//
// See: https://en.wikipedia.org/wiki/Option_type
type Optional[T any] struct {
	Value T
	Set   bool
}

func ValueOptional[T any](value T) Optional[T] {
	return Optional[T]{
		Value: value,
		Set:   true,
	}
}

func PointerOptional[T any](value *T) Optional[T] {
	if value == nil {
		return EmptyOptional[T]()
	}

	return Optional[T]{
		Value: *value,
		Set:   true,
	}
}

func EmptyOptional[T any]() Optional[T] {
	var emptyT T

	return Optional[T]{
		Value: emptyT,
		Set:   false,
	}
}
