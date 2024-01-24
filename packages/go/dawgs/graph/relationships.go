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
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type Relationship struct {
	ID         ID
	StartID    ID
	EndID      ID
	Kind       Kind
	Properties *Properties
}

func (s *Relationship) Merge(other *Relationship) {
	s.Properties.Merge(other.Properties)
}

func (s *Relationship) SizeOf() size.Size {
	relSize := size.Of(s) + size.Of(s.Kind)

	if s.Properties != nil {
		relSize += s.Properties.SizeOf()
	}

	return relSize
}

func PrepareRelationship(properties *Properties, kind Kind) *Relationship {
	return &Relationship{
		Kind:       kind,
		Properties: properties,
	}
}

func NewRelationship(id, startID, endID ID, properties *Properties, kind Kind) *Relationship {
	return &Relationship{
		ID:         id,
		StartID:    startID,
		EndID:      endID,
		Kind:       kind,
		Properties: properties,
	}
}

// RelationshipSet is a mapped index of Relationship instances and their ID fields.
type RelationshipSet map[ID]*Relationship

// Len returns the number of unique Relationship instances in this set.
func (s RelationshipSet) Len() int {
	return len(s)
}

// Get returns a Relationship from this set by its database ID.
func (s RelationshipSet) Get(id ID) *Relationship {
	return s[id]
}

// Contains returns true if the ID of the given Relationship is stored within this RelationshipSet.
func (s RelationshipSet) Contains(relationship *Relationship) bool {
	return s.ContainsID(relationship.ID)
}

// ContainsID returns true if the Relationship represented by the given ID is stored within this RelationshipSet.
func (s RelationshipSet) ContainsID(id ID) bool {
	_, seen := s[id]
	return seen
}

// Add adds a given Relationship to the RelationshipSet.
func (s RelationshipSet) Add(relationships ...*Relationship) {
	for _, relationship := range relationships {
		s[relationship.ID] = relationship
	}
}

// Slice returns a slice of the Relationship instances stored in this RelationshipSet.
func (s RelationshipSet) Slice() []*Relationship {
	slice := make([]*Relationship, 0, len(s))

	for _, v := range s {
		slice = append(slice, v)
	}

	return slice
}

// AddSet merges all Relationships from the given RelationshipSet into this RelationshipSet.
func (s RelationshipSet) AddSet(other RelationshipSet) {
	for k, v := range other {
		s[k] = v
	}
}

// IDBitmap returns a new roaring64.Bitmap instance containing all Relationship ID values in this RelationshipSet.
func (s RelationshipSet) IDBitmap() *roaring64.Bitmap {
	bitmap := roaring64.New()

	for id := range s {
		bitmap.Add(uint64(id))
	}

	return bitmap
}

// NewRelationshipSet returns a new RelationshipSet from the given Relationship slice.
func NewRelationshipSet(relationships ...*Relationship) RelationshipSet {
	newSet := make(RelationshipSet, len(relationships))

	for _, relationship := range relationships {
		newSet[relationship.ID] = relationship
	}

	return newSet
}
