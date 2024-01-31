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

// Package slicesext extends the standard library slices package with additional slice utilities
package slicesext

import "slices"

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

// MapWithErr applies a mapping/transformation function over each element in a given slice and returns a new slice of the transformed values with type T
// An error from the mapping function will be returned directly to the caller
// Note: MapWithErr is primarily designed for applying type conversion functions, where you want a slice of type U from a slice of type T
func MapWithErr[T, U any](s []T, f func(T) (U, error)) ([]U, error) {
	conv := make([]U, 0, len(s))

	for _, e := range s {
		if t, err := f(e); err != nil {
			return nil, err
		} else {
			conv = append(conv, t)
		}
	}

	return conv, nil
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

// Copyright 2021 The Go Authors. All rights reserved.
// Concat returns a new slice concatenating the passed in slices.
// This was ripped from go1.22 source and should be replaced with the stdlib implementation when we move to 1.22
// Original source: https://github.com/golang/go/blob/5c0d0929d3a6378c710376b55a49abd55b31a805/src/slices/slices.go#L502
func Concat[S ~[]E, E any](s ...S) S {
	size := 0
	for _, slice := range s {
		size += len(slice)
		if size < 0 {
			panic("len out of range")
		}
	}
	newslice := slices.Grow[S](nil, size)
	for _, s := range s {
		newslice = append(newslice, s...)
	}
	return newslice
}
