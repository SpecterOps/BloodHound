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

// Foldl returns an accumulative value by applying the provided function to each
// element in the list and the accumulation starting with the first list element.
//
// Foldl is considered left-associative or that the parenthetical grouping of operations is evaluated from left to right.
// For example, `1 + 2 + 3 + 4 + 5` would be evaluated as `(((1 + 2) + 3) + 4) + 5`. Generally, left folds are used to
// transform lists in reverse order or when the folding function is commutative.
//
// Foldl is eagerly evaluated. Eager evaluation typically yields better performance but does not allow the
// folding function to terminate early.
func Foldl[Value, Accumulation any](initVal Accumulation, list []Value, reducers ...func(Accumulation, Value) Accumulation) Accumulation {
	var (
		acc = initVal
		l   = list
	)
	for len(l) != 0 {
		head := Head(l)
		for _, reduce := range reducers {
			acc = reduce(acc, head)
		}
		l = Tail(l)
	}
	return acc
}

// FoldlLazy returns an accumulative value by recursively applying the provided function to each
// element in the list and the accumulation starting with the first list element.
//
// Foldl is considered left-associative or that the parenthetical grouping of operations is evaluated from left to right.
// For example, `1 + 2 + 3 + 4 + 5` would be evaluated as `(((1 + 2) + 3) + 4) + 5`. Generally, left folds are used to
// transform lists in reverse order or when the folding function is commutative.
//
// FoldlLazy is lazily evaluated. Lazy evaluation allows the folding function to short-circuit the operation by returning
// a value that does not depend on the accumulating parameter. Depending on the folding function, lazy evaluation can
// impact time and space performance.
func FoldlLazy[Value, Accumulation any](initval Accumulation, list []Value, reducers ...func(Accumulation, Value) Accumulation) Accumulation {
	if len(list) == 0 {
		return initval
	}

	var (
		head = Head(list)
		tail = Tail(list)
		acc  = initval
	)
	for _, reduce := range reducers {
		acc = reduce(acc, head)
	}
	return FoldlLazy(acc, tail, reducers...)
}
