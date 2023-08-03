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

	"github.com/specterops/bloodhound/dawgs/graph"
)

type KindBitmaps map[string]Duplex[uint32]

func (s KindBitmaps) Get(kinds ...graph.Kind) Duplex[uint32] {
	intersection := NewBitmap32()

	for _, kind := range kinds {
		if bitmap, hasBitmap := s[kind.String()]; hasBitmap {
			intersection.Or(bitmap)
		}
	}

	return intersection
}

func (s KindBitmaps) Count(kind graph.Kind) uint64 {
	if bitmap, hasBitmap := s[kind.String()]; hasBitmap {
		return bitmap.Cardinality()
	}

	return 0
}

func (s KindBitmaps) CountAll(kinds ...graph.Kind) uint64 {
	allNodes := NewBitmap32()

	for _, kind := range kinds {
		if bitmap, hasBitmap := s[kind.String()]; hasBitmap {
			allNodes.Or(bitmap)
		}
	}

	return allNodes.Cardinality()
}

func (s KindBitmaps) Or(bitmaps KindBitmaps) {
	for kindStr, leftBitmap := range bitmaps {
		if rightBitmap, hasRightBitmap := s[kindStr]; !hasRightBitmap {
			newRightBitmap := NewBitmap32()
			newRightBitmap.Or(leftBitmap)

			s[kindStr] = newRightBitmap
		} else {
			rightBitmap.Or(leftBitmap)
		}
	}
}

func (s KindBitmaps) AddSets(nodeSets ...graph.NodeSet) {
	for _, nodeSet := range nodeSets {
		for _, node := range nodeSet {
			s.AddIDKindsPair(node.ID, node.Kinds)
		}
	}
}

func (s KindBitmaps) OrAll() Duplex[uint32] {
	all := NewBitmap32()

	for _, bitmap := range s {
		all.Or(bitmap)
	}

	return all
}

func (s KindBitmaps) Contains(node *graph.Node) bool {
	for _, bitmap := range s {
		if bitmap.Contains(node.ID.Uint32()) {
			return true
		}
	}

	return false
}

func (s KindBitmaps) AddIDKindsPair(id graph.ID, kinds graph.Kinds) {
	nodeID := id.Uint32()

	for _, nodeKind := range kinds {
		nodeKindStr := nodeKind.String()

		if bitmap, hasBitmap := s[nodeKindStr]; !hasBitmap {
			newBitmap := NewBitmap32()
			newBitmap.Add(nodeID)

			s[nodeKindStr] = newBitmap
		} else {
			bitmap.Add(nodeID)
		}
	}
}

func (s KindBitmaps) AddNodes(nodes ...*graph.Node) {
	for _, node := range nodes {
		s.AddIDKindsPair(node.ID, node.Kinds)
	}
}

type ThreadSafeKindBitmap struct {
	bitmaps map[graph.Kind]Duplex[uint32]
	rwLock  *sync.RWMutex
}

func NewThreadSafeKindBitmap() *ThreadSafeKindBitmap {
	return &ThreadSafeKindBitmap{
		bitmaps: map[graph.Kind]Duplex[uint32]{},
		rwLock:  &sync.RWMutex{},
	}
}

func (s ThreadSafeKindBitmap) Get(kinds ...graph.Kind) Duplex[uint32] {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	bitmap := NewBitmap32()

	if len(kinds) == 0 {
		for _, kindBitmap := range s.bitmaps {
			bitmap.Or(kindBitmap)
		}
	} else {
		for _, kind := range kinds {
			if kindBitmap, hasKind := s.bitmaps[kind]; hasKind {
				bitmap.Or(kindBitmap)
			}
		}
	}

	return bitmap
}

func (s ThreadSafeKindBitmap) Cardinality(kinds ...graph.Kind) uint64 {
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

func (s ThreadSafeKindBitmap) Contains(kind graph.Kind, value uint32) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	if kindBitmap, hasKind := s.bitmaps[kind]; hasKind {
		return kindBitmap.Contains(value)
	}

	return false
}

func (s ThreadSafeKindBitmap) Add(kind graph.Kind, value uint32) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if kindBitmap, hasKind := s.bitmaps[kind]; hasKind {
		kindBitmap.Add(value)
	} else {
		kindBitmap = NewBitmap32()
		kindBitmap.Add(value)

		s.bitmaps[kind] = kindBitmap
	}
}

func (s ThreadSafeKindBitmap) CheckedAdd(kind graph.Kind, value uint32) bool {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if kindBitmap, hasKind := s.bitmaps[kind]; hasKind {
		return kindBitmap.CheckedAdd(value)
	} else {
		kindBitmap = NewBitmap32()
		kindBitmap.Add(value)

		s.bitmaps[kind] = kindBitmap

		return true
	}
}
