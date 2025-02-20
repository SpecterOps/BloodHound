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

package graph

import (
	"sync"

	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type KindBitmaps map[string]cardinality.Duplex[uint64]

func (s KindBitmaps) Get(kinds ...Kind) cardinality.Duplex[uint64] {
	bitmap := cardinality.NewBitmap64()

	if len(kinds) == 0 {
		for _, kindBitmap := range s {
			bitmap.Or(kindBitmap)
		}
	} else {
		for _, kind := range kinds {
			if kindBitmap, hasKind := s[kind.String()]; hasKind {
				bitmap.Or(kindBitmap)
			}
		}
	}

	return bitmap
}

func (s KindBitmaps) Count(kinds ...Kind) uint64 {
	return s.Get(kinds...).Cardinality()
}

func (s KindBitmaps) Or(bitmaps KindBitmaps) {
	for kindStr, leftBitmap := range bitmaps {
		if rightBitmap, hasRightBitmap := s[kindStr]; !hasRightBitmap {
			newRightBitmap := cardinality.NewBitmap64()
			newRightBitmap.Or(leftBitmap)

			s[kindStr] = newRightBitmap
		} else {
			rightBitmap.Or(leftBitmap)
		}
	}
}

func (s KindBitmaps) AddSets(nodeSets ...NodeSet) {
	for _, nodeSet := range nodeSets {
		for _, node := range nodeSet {
			s.AddIDToKinds(node.ID, node.Kinds)
		}
	}
}

func (s KindBitmaps) OrAll() cardinality.Duplex[uint64] {
	all := cardinality.NewBitmap64()

	for _, bitmap := range s {
		all.Or(bitmap)
	}

	return all
}

func (s KindBitmaps) Contains(node *Node) bool {
	for _, bitmap := range s {
		if bitmap.Contains(node.ID.Uint64()) {
			return true
		}
	}

	return false
}

func (s KindBitmaps) AddDuplexToKind(ids cardinality.Duplex[uint64], kind Kind) {
	kindStr := kind.String()

	if bitmap, hasBitmap := s[kindStr]; !hasBitmap {
		newBitmap := cardinality.NewBitmap64()
		newBitmap.Or(ids)

		s[kindStr] = newBitmap
	} else {
		bitmap.Or(ids)
	}
}

func (s KindBitmaps) AddIDToKind(id ID, kind Kind) {
	var (
		nodeID  = id.Uint64()
		kindStr = kind.String()
	)

	if bitmap, hasBitmap := s[kindStr]; !hasBitmap {
		newBitmap := cardinality.NewBitmap64()
		newBitmap.Add(nodeID)

		s[kindStr] = newBitmap
	} else {
		bitmap.Add(nodeID)
	}
}

func (s KindBitmaps) AddIDToKinds(id ID, kinds Kinds) {
	nodeID := id.Uint64()

	for _, kind := range kinds {
		kindStr := kind.String()

		if bitmap, hasBitmap := s[kindStr]; !hasBitmap {
			newBitmap := cardinality.NewBitmap64()
			newBitmap.Add(nodeID)

			s[kindStr] = newBitmap
		} else {
			bitmap.Add(nodeID)
		}
	}
}

func (s KindBitmaps) AddNodes(nodes ...*Node) {
	for _, node := range nodes {
		s.AddIDToKinds(node.ID, node.Kinds)
	}
}

type ThreadSafeKindBitmap struct {
	bitmaps KindBitmaps
	rwLock  *sync.RWMutex
}

func NewThreadSafeKindBitmap() *ThreadSafeKindBitmap {
	return &ThreadSafeKindBitmap{
		bitmaps: map[string]cardinality.Duplex[uint64]{},
		rwLock:  &sync.RWMutex{},
	}
}

func (s ThreadSafeKindBitmap) Count(kinds ...Kind) uint64 {
	return s.Get(kinds...).Cardinality()
}

func (s ThreadSafeKindBitmap) Get(kinds ...Kind) cardinality.Duplex[uint64] {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	bitmap := cardinality.NewBitmap64()

	if len(kinds) == 0 {
		for _, kindBitmap := range s.bitmaps {
			bitmap.Or(kindBitmap)
		}
	} else {
		for _, kind := range kinds {
			if kindBitmap, hasKind := s.bitmaps[kind.String()]; hasKind {
				bitmap.Or(kindBitmap)
			}
		}
	}

	return bitmap
}

func (s ThreadSafeKindBitmap) Or(kind Kind, other cardinality.Duplex[uint64]) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if kindBitmap, hasKind := s.bitmaps[kind.String()]; hasKind {
		kindBitmap.Or(other)
	} else {
		s.bitmaps[kind.String()] = other
	}
}

func (s ThreadSafeKindBitmap) Cardinality(kinds ...Kind) uint64 {
	return s.Get(kinds...).Cardinality()
}

func (s ThreadSafeKindBitmap) Clone() *ThreadSafeKindBitmap {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	clone := NewThreadSafeKindBitmap()
	for kind, kindBitmap := range s.bitmaps {
		clone.bitmaps[kind] = kindBitmap.Clone()
	}

	return clone
}

func (s ThreadSafeKindBitmap) Contains(kind Kind, value uint64) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	if kindBitmap, hasKind := s.bitmaps[kind.String()]; hasKind {
		return kindBitmap.Contains(value)
	}

	return false
}

func (s ThreadSafeKindBitmap) Add(kind Kind, value uint64) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if kindBitmap, hasKind := s.bitmaps[kind.String()]; hasKind {
		kindBitmap.Add(value)
	} else {
		kindBitmap = cardinality.NewBitmap64()
		kindBitmap.Add(value)

		s.bitmaps[kind.String()] = kindBitmap
	}
}

func (s ThreadSafeKindBitmap) CheckedAdd(kind Kind, value uint64) bool {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if kindBitmap, hasKind := s.bitmaps[kind.String()]; hasKind {
		return kindBitmap.CheckedAdd(value)
	} else {
		kindBitmap = cardinality.NewBitmap64()
		kindBitmap.Add(value)

		s.bitmaps[kind.String()] = kindBitmap

		return true
	}
}

// IndexedSlice is a structure maps a comparable key to a value that implements size.Sizable.
type IndexedSlice[K comparable, V any] struct {
	index  map[K]int
	values []V
	size   size.Size
}

func NewIndexedSlice[K comparable, V any]() *IndexedSlice[K, V] {
	return &IndexedSlice[K, V]{
		index: make(map[K]int),
		size:  0,
	}
}

func (s *IndexedSlice[K, V]) Keys() []K {
	keys := make([]K, 0, len(s.index))

	for key := range s.index {
		keys = append(keys, key)
	}

	return keys
}

func (s *IndexedSlice[K, V]) Values() []V {
	return s.values
}

func (s *IndexedSlice[K, V]) Merge(other *IndexedSlice[K, V]) {
	for key, idx := range other.index {
		s.Put(key, other.values[idx])
	}
}

// Len returns the number of values stored.
func (s *IndexedSlice[K, V]) Len() int {
	return len(s.values)
}

// SizeOf returns the relative size of the IndexedSlice instance.
func (s *IndexedSlice[K, V]) SizeOf() size.Size {
	return s.size
}

func (s *IndexedSlice[K, V]) Get(key K) V {
	if valueIdx, hasValue := s.index[key]; hasValue {
		return s.values[valueIdx]
	}

	var empty V
	return empty
}

func (s *IndexedSlice[K, V]) Has(key K) bool {
	_, hasValue := s.index[key]
	return hasValue
}

func (s *IndexedSlice[K, V]) GetOr(key K, defaultConstructor func() V) V {
	if valueIdx, hasValue := s.index[key]; hasValue {
		return s.values[valueIdx]
	}

	defaultValue := defaultConstructor()

	s.Put(key, defaultValue)
	return defaultValue
}

// CheckedGet returns a tuple containing the value and a boolean representing if a value was found for the
// given key.
func (s *IndexedSlice[K, V]) CheckedGet(key K) (V, bool) {
	if valueIdx, hasValue := s.index[key]; hasValue {
		return s.values[valueIdx], true
	}

	var empty V
	return empty, false
}

// GetAll returns all found values for a given slice of keys. Any keys that do not have stored values
// in this IndexedSlice are returned as the second value of the tuple return for this function.
func (s *IndexedSlice[K, V]) GetAll(keys []K) ([]V, []K) {
	var (
		values      = make([]V, 0, len(keys))
		missingKeys = make([]K, 0, len(keys))
	)

	for _, key := range keys {
		if valueIdx, hasValue := s.index[key]; hasValue {
			values = append(values, s.values[valueIdx])
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	return values, missingKeys
}

// GetAllIndexed returns all found values for a given slice of keys. Any keys that do not have stored values
// in this IndexedSlice are returned as the second value of the tuple return for this function.
func (s *IndexedSlice[K, V]) GetAllIndexed(keys []K) (*IndexedSlice[K, V], []K) {
	var (
		values      = NewIndexedSlice[K, V]()
		missingKeys = make([]K, 0, len(keys))
	)

	for _, key := range keys {
		if valueIdx, hasValue := s.index[key]; hasValue {
			values.Put(key, s.values[valueIdx])
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	return values, missingKeys
}

func sizeOf(value any) size.Size {
	if sizeable, typeOK := value.(size.Sizable); typeOK {
		return sizeable.SizeOf()
	}

	return size.Of(value)
}

// Put inserts the given value with the given key.
func (s *IndexedSlice[K, V]) Put(key K, value V) {
	s.size += sizeOf(value)

	if valueIdx, hasValue := s.index[key]; hasValue {
		s.size -= sizeOf(s.values[valueIdx])
		s.values[valueIdx] = value
	} else {
		s.values = append(s.values, value)
		s.index[key] = len(s.values) - 1
	}
}

func (s *IndexedSlice[K, V]) Each(delegate func(key K, value V) bool) {
	for id, idx := range s.index {
		if !delegate(id, s.values[idx]) {
			break
		}
	}
}

// DuplexToGraphIDs takes a Duplex provider and returns a slice of graph IDs.
func DuplexToGraphIDs[T uint32 | uint64](provider cardinality.Duplex[T]) []ID {
	ids := make([]ID, 0, provider.Cardinality())

	provider.Each(func(value T) bool {
		ids = append(ids, ID(value))
		return true
	})

	return ids

}

// NodeSetToDuplex takes a graph NodeSet and returns a Duplex provider that contains all node IDs.
func NodeSetToDuplex(nodes NodeSet) cardinality.Duplex[uint64] {
	duplex := cardinality.NewBitmap64()

	for nodeID := range nodes {
		duplex.Add(nodeID.Uint64())
	}

	return duplex
}

// NodeSetToDuplex takes a graph NodeSet and returns a Duplex provider that contains all node IDs.
func NodeIDsToDuplex(nodeIDs []ID) cardinality.Duplex[uint64] {
	duplex := cardinality.NewBitmap64()

	for _, nodeID := range nodeIDs {
		duplex.Add(nodeID.Uint64())
	}

	return duplex
}
