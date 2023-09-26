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

package slices

// Filter applies a predicate function over each element in a given slice and returns a new slice containing only the elements in which the predicate returns true
func Filter[T any](slice []T, fn func(T) bool) []T {
	var out []T
	for _, t := range slice {
		if fn(t) {
			out = append(out, t)
		}
	}
	return out
}

// Map applies a mapping/transformation function over each element in a given slice and returns a new slice of the transformed values or nil
func Map[T, U any](slice []T, fn func(T) U) []U {
	var out []U
	for _, t := range slice {
		out = append(out, fn(t))
	}
	return out
}

// FlatMap applies a mapping/transformation function over each element in a given slice, concatenates the results and returns a new flattened slice of the transformed values or nil
func FlatMap[T, U any](slice []T, fn func(T) []U) []U {
	var out []U
	for _, t := range slice {
		out = append(out, fn(t)...)
	}
	return out
}

// Unique returns a new slice that is free of any duplicate values in the original
func Unique[T comparable](slice []T) []T {
	var (
		exists    = struct{}{}
		existsMap = map[T]struct{}{}
		out       = make([]T, 0)
	)

	for _, value := range slice {
		if _, ok := existsMap[value]; !ok {
			existsMap[value] = exists
			out = append(out, value)
		}
	}

	return out
}

// UniqueBy returns a new slice that is free of any duplicate values in the original
func UniqueBy[T any, U comparable](slice []T, fn func(T) U) []T {
	var (
		exists    = struct{}{}
		existsMap = map[U]struct{}{}
		out       = make([]T, 0)
	)

	for _, value := range slice {
		key := fn(value)
		if _, ok := existsMap[key]; !ok {
			existsMap[key] = exists
			out = append(out, value)
		}
	}

	return out
}

// Contains returns true if a slice contains an element that is equal to the given value
func Contains[T comparable](slice []T, value T) bool {
	for _, sliceValue := range slice {
		if sliceValue == value {
			return true
		}
	}
	return false
}

func Head[T any](list []T) T {
	return list[0]
}

func Tail[T any](list []T) []T {
	return list[1:]
}

func Last[T any](list []T) T {
	return list[len(list)-1]
}

func Init[T any](list []T) []T {
	return list[:len(list)-1]
}

// Reverse reverses the order of a slice by doing an in place reverse
// This will reorder the provided slice in place, and uses zero allocations
func Reverse[T any](list []T) []T {
	for low, high := 0, len(list)-1; low < high; low, high = low+1, high-1 {
		list[low], list[high] = list[high], list[low]
	}

	return list
}
