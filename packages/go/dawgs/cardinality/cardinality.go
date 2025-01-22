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

package cardinality

type ProviderConstructor[T uint32 | uint64] func() Provider[T]
type SimplexConstructor[T uint32 | uint64] func() Simplex[T]
type DuplexConstructor[T uint32 | uint64] func() Duplex[T]

// Provider describes the most basic functionality of a cardinality provider algorithm: adding elements to the provider
// and producing the cardinality of those elements.
type Provider[T uint32 | uint64] interface {
	Add(value ...T)
	Or(other Provider[T])
	Clear()
	Cardinality() uint64
	CloneProvider() Provider[T]
}

func CloneProvider[T uint32 | uint64](provider Provider[T]) Provider[T] {
	switch typedProvider := provider.(type) {
	case Simplex[T]:
		return typedProvider.Clone()

	case Duplex[T]:
		return typedProvider.Clone()

	default:
		// TODO: Review this default case
		return provider
	}
}

// Simplex is a one-way cardinality provider that does not allow a user to retrieve encoded values back out of the
// provider. This interface is suitable for algorithms such as HyperLogLog which utilizes a hash function to merge
// identifiers into the cardinality provider.
type Simplex[T uint32 | uint64] interface {
	Provider[T]

	Clone() Simplex[T]
}

// Iterator allows enumeration of a duplex cardinality provider without requiring the allocation of the provider's set.
type Iterator[T uint32 | uint64] interface {
	HasNext() bool
	Next() T
}

// Duplex is a two-way cardinality provider that allows a user to retrieve encoded values back out of the provider. This
// interface is suitable for algorithms that behave similar to bitvectors.
type Duplex[T uint32 | uint64] interface {
	Provider[T]

	Xor(other Provider[T])
	And(other Provider[T])
	AndNot(other Provider[T])
	Remove(value T)
	Slice() []T
	Contains(value T) bool
	Each(delegate func(value T) bool)
	CheckedAdd(value T) bool
	Clone() Duplex[T]
}
