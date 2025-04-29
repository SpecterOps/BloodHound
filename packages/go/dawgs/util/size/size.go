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

package size

import (
	"unsafe"
)

// Sizable is an interface describing a type that may have its runtime memory resident size
// estimated.
type Sizable interface {
	// SizeOf returns the size of this entity.
	SizeOf() Size
}

// Of returns the size of an object. When the input is a pointer
// to an object, the size of the pointer itself is returned,
// as opposed to the size of the memory referenced. This function
// cannot infer the data type of nil, so nil checks need to be
// performed before calling this function.
func Of[T any](templatedValue T) Size {
	totalSize := Size(unsafe.Sizeof(templatedValue))

	switch typedValue := any(templatedValue).(type) {
	case string:
		// Strings must have their inner allocation counted
		totalSize *= Size(len(typedValue))
	}

	return totalSize
}

// OfSlice will return the size of a type templated slice.
//
// The slice descriptor consists of three fields:
// * A pointer to the underlying array.
// * The length of the slice (the number of elements currently in use).
// * The capacity of the slice (the total number of elements in the underlying array, from the start of the slice).
//
// On a 64-bit architecture, each of these fields is 8 bytes, so the size of a slice is 24 bytes. On a 32-bit
// architecture, each field is 4 bytes, making the slice size 12 bytes. The unsafe.Sizeof() function will return
// this size of the descriptor, not the size of the underlying array.
//
// For correctly typed templates the capacity of the slice is multiplied by the size of the template.
func OfSlice[T comparable](sliceT []T) Size {
	totalSize := Of(sliceT)

	switch any(sliceT).(type) {
	case []string:
		// Strings are variable allocations and must be walked
		for _, strValue := range sliceT {
			totalSize += Of(strValue)
		}

	default:
		var template T
		totalSize += Of(template) * Size(cap(sliceT))
	}

	return totalSize
}

// OfAny attempts to return the size of a value passed as `any`. This is done via
// type negotiation to unwrap the `any` type in a best-effort attempt to estimate
// the value's size.
//
// It is important to note that this function comes with usage caveats:
// * Maps are not estimated by this function
// * Nested slices of type `[]any` are supported but have ill-defined runtime measurement costs
// * If the type of the value is known, this function is less performant that calling Of(...)
func OfAny(value any) Size {
	switch typedValue := value.(type) {
	case int:
		return Of(typedValue)
	case int8:
		return Of(typedValue)
	case int16:
		return Of(typedValue)
	case int32:
		return Of(typedValue)
	case int64:
		return Of(typedValue)
	case uint:
		return Of(typedValue)
	case uint8:
		return Of(typedValue)
	case uint16:
		return Of(typedValue)
	case uint32:
		return Of(typedValue)
	case uint64:
		return Of(typedValue)
	case float32:
		return Of(typedValue)
	case float64:
		return Of(typedValue)
	case string:
		return Of(typedValue)
	case bool:
		return Of(typedValue)
	case uintptr:
		return Of(typedValue)
	case complex64:
		return Of(typedValue)
	case complex128:
		return Of(typedValue)
	case []int:
		return OfSlice(typedValue)
	case []int8:
		return OfSlice(typedValue)
	case []int16:
		return OfSlice(typedValue)
	case []int32:
		return OfSlice(typedValue)
	case []int64:
		return OfSlice(typedValue)
	case []uint:
		return OfSlice(typedValue)
	case []uint8:
		return OfSlice(typedValue)
	case []uint16:
		return OfSlice(typedValue)
	case []uint32:
		return OfSlice(typedValue)
	case []uint64:
		return OfSlice(typedValue)
	case []float32:
		return OfSlice(typedValue)
	case []float64:
		return OfSlice(typedValue)
	case []string:
		return OfSlice(typedValue)
	case []bool:
		return OfSlice(typedValue)
	case []uintptr:
		return OfSlice(typedValue)
	case []complex64:
		return OfSlice(typedValue)
	case []complex128:
		return OfSlice(typedValue)
	case []any:
		return OfAnySlice(typedValue)
	case struct{}:
		return 0
	default:
		// Best effort which is the pointer size of the any-value
		return Size(unsafe.Sizeof(value))
	}
}

// OfAnySlice calculates the estimated size of a given slice of type `[]any`.
func OfAnySlice(anySlice []any) Size {
	sliceValueSize := Size(0)

	for _, anyValue := range anySlice {
		sliceValueSize += OfAny(anyValue)
	}

	return Of(anySlice) + Size(cap(anySlice)) + sliceValueSize
}
