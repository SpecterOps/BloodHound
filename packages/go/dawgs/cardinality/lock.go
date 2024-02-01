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

import (
	"sync"
)

type threadSafeDuplex[T uint32 | uint64] struct {
	provider Duplex[T]
	lock     *sync.Mutex
}

func ThreadSafeDuplex[T uint32 | uint64](provider Duplex[T]) Duplex[T] {
	return threadSafeDuplex[T]{
		provider: provider,
		lock:     &sync.Mutex{},
	}
}

func (s threadSafeDuplex[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Clear()
}

func (s threadSafeDuplex[T]) Add(values ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Add(values...)
}

func (s threadSafeDuplex[T]) Remove(value T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Remove(value)
}

func (s threadSafeDuplex[T]) Xor(other Provider[T]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Xor(other)
}

func (s threadSafeDuplex[T]) And(other Provider[T]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.And(other)
}

func (s threadSafeDuplex[T]) Or(other Provider[T]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Or(other)
}

func (s threadSafeDuplex[T]) Cardinality() uint64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.provider.Cardinality()
}

func (s threadSafeDuplex[T]) Slice() []T {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.provider.Slice()
}

func (s threadSafeDuplex[T]) Contains(value T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.provider.Contains(value)
}

func (s threadSafeDuplex[T]) Each(delegate func(value T) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Each(delegate)
}

func (s threadSafeDuplex[T]) CheckedAdd(value T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.provider.CheckedAdd(value)
}

func (s threadSafeDuplex[T]) Clone() Duplex[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	return ThreadSafeDuplex(s.provider.Clone())
}

type threadSafeSimplex[T uint32 | uint64] struct {
	provider Simplex[T]
	lock     *sync.Mutex
}

func ThreadSafeSimplex[T uint32 | uint64](provider Simplex[T]) Simplex[T] {
	return threadSafeSimplex[T]{
		provider: provider,
		lock:     &sync.Mutex{},
	}
}

func (s threadSafeSimplex[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Clear()
}

func (s threadSafeSimplex[T]) Add(values ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Add(values...)
}

func (s threadSafeSimplex[T]) Or(other Provider[T]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.provider.Or(other)
}

func (s threadSafeSimplex[T]) Cardinality() uint64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.provider.Cardinality()
}

func (s threadSafeSimplex[T]) Clone() Simplex[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	return ThreadSafeSimplex(s.provider.Clone())
}
