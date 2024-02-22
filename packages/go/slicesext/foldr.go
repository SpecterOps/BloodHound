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

package slicesext

// Foldr returns an accumulative value by recursively applying the provided function to each element in the list and the accumulation
// starting with the last list element.
//
// Foldr is the most commonly used fold and is considered right-associative or that the parenthetical grouping of operations is
// evaluated from right to left.
// For example: `1 + 2 + 3 + 4 + 5` would be evaluated as `1 + (2 + (3 + (4 + 5)))`
//
// Generally, right folds are used to transform lists into lists with related elements in the same order.
//
// Foldr is lazily evaluated. Lazy evaluation allows the folding function to terminate early by returning a value that does not depend
// on the accumulating parameter. This ability to short-circuit the operation is offset by a potential cost to time and space
// performance. However, a tail-recursive folding function can allow for efficient processing of very large or even infinite lists.
func Foldr[Value, Accumulation any](initval Accumulation, list []Value, reducers ...func(Accumulation, Value) Accumulation) Accumulation {
	if len(list) == 0 {
		return initval
	}

	var (
		head = Head(list)
		tail = Tail(list)
		acc  = Foldr(initval, tail, reducers...)
	)

	for _, reduce := range reducers {
		acc = reduce(acc, head)
	}

	return acc
}

// FoldrEager returns an accumulative value by applying the provided function to each element in the list and the accumulation
// starting with the last list element.
//
// FoldrEager is considered right-associative or that the parenthetical grouping of operations is evaluated from right to left.
// For example, `1 + 2 + 3 + 4 + 5` would be evaluated as `1 + (2 + (3 + (4 + 5)))`
//
// Generally, right folds are used to transform lists into lists with related elements in the same order.
//
// FoldrEager is eagerly evaluated. Transforming particularly large lists will typically have better time and space performance
// using eager evaluation but does not allow the operation to short-circuit.
func FoldrEager[Value, Accumulation any](initval Accumulation, list []Value, reducers ...func(Accumulation, Value) Accumulation) Accumulation {
	var (
		acc = initval
		l   = list
	)
	for len(l) != 0 {
		last := Last(l)
		for _, reduce := range reducers {
			acc = reduce(acc, last)
		}
		l = Init(l)
	}
	return acc
}
