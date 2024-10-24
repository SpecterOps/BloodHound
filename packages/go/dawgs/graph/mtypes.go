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
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"sync"
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
