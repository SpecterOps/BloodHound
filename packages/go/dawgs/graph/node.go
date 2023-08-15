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
	"encoding/json"
	"github.com/RoaringBitmap/roaring"
	"math"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

const (
	UnregisteredNodeID ID = math.MaxUint32
)

func PrepareNode(properties *Properties, kinds ...Kind) *Node {
	return NewNode(UnregisteredNodeID, properties, kinds...)
}

func NewNode(id ID, properties *Properties, kinds ...Kind) *Node {
	return &Node{
		ID:         id,
		Kinds:      kinds,
		Properties: properties,
	}
}

type serializableNode struct {
	ID           ID          `json:"id"`
	Kinds        []string    `json:"kinds"`
	AddedKinds   []string    `json:"added_kinds"`
	DeletedKinds []string    `json:"deleted_kinds"`
	Properties   *Properties `json:"properties"`
}

type Node struct {
	ID           ID          `json:"id"`
	Kinds        Kinds       `json:"kinds"`
	AddedKinds   Kinds       `json:"added_kinds"`
	DeletedKinds Kinds       `json:"deleted_kinds"`
	Properties   *Properties `json:"properties"`
}

func (s *Node) Merge(other *Node) {
	s.Kinds = s.Kinds.Add(other.Kinds...)

	for _, otherKind := range other.AddedKinds {
		s.DeletedKinds = s.DeletedKinds.Remove(otherKind)
	}

	for _, otherKind := range other.DeletedKinds {
		s.Kinds = s.Kinds.Remove(otherKind)
		s.AddedKinds = s.AddedKinds.Remove(otherKind)
	}

	s.AddedKinds = s.AddedKinds.Add(other.AddedKinds...)
	s.DeletedKinds = s.DeletedKinds.Add(other.DeletedKinds...)

	s.Properties.Merge(other.Properties)
}

func (s *Node) SizeOf() size.Size {
	nodeSize := size.Of(s) + s.Kinds.SizeOf()

	if s.Properties != nil {
		nodeSize += s.Properties.SizeOf()
	}

	return nodeSize
}

func (s *Node) AddKinds(kinds ...Kind) {
	for _, kind := range kinds {
		s.Kinds = s.Kinds.Add(kind)
		s.AddedKinds = s.AddedKinds.Add(kind)
		s.DeletedKinds = s.DeletedKinds.Remove(kind)
	}
}

func (s *Node) DeleteKinds(kinds ...Kind) {
	for _, kind := range kinds {
		s.Kinds = s.Kinds.Remove(kind)
		s.AddedKinds = s.AddedKinds.Remove(kind)
		s.DeletedKinds = s.DeletedKinds.Add(kind)
	}
}

func (s *Node) MarshalJSON() ([]byte, error) {
	var (
		jsonNode = serializableNode{
			ID:           s.ID,
			Kinds:        s.Kinds.Strings(),
			AddedKinds:   s.AddedKinds.Strings(),
			DeletedKinds: s.DeletedKinds.Strings(),
			Properties:   s.Properties,
		}
	)

	return json.Marshal(jsonNode)
}

// NodeSet is a mapped index of Node instances and their ID fields.
type NodeSet map[ID]*Node

// Pick returns a single Node instance from this set. Repeated calls to this function are not guaranteed to return
// the same Node instance.
func (s NodeSet) Pick() *Node {
	for _, value := range s {
		return value
	}

	return nil
}

// ContainingNodeKinds returns a new NodeSet containing only Node instances that contain any one of the given Kind instances.
func (s NodeSet) ContainingNodeKinds(kinds ...Kind) NodeSet {
	newNodeSet := NodeSet{}
	for _, node := range s {
		if node.Kinds.ContainsOneOf(kinds...) {
			newNodeSet.Add(node)
		}
	}

	return newNodeSet
}

// Remove removes a Node from this set by its database ID.
func (s NodeSet) Remove(id ID) {
	delete(s, id)
}

// Get returns a Node from this set by its database ID.
func (s NodeSet) Get(id ID) *Node {
	return s[id]
}

// Len returns the number of unique Node instances in this set.
func (s NodeSet) Len() int {
	return len(s)
}

// Copy returns a shallow copy of this set.
func (s NodeSet) Copy() NodeSet {
	newSet := make(NodeSet, len(s))

	for k, v := range s {
		newSet[k] = v
	}

	return newSet
}

// KindSet returns a NodeKindSet constructed from the Node instances in this set.
func (s NodeSet) KindSet() NodeKindSet {
	nodeKindSet := NodeKindSet{}

	for _, node := range s {
		nodeKindSet.Add(node)
	}

	return nodeKindSet
}

// Contains returns true if the ID of the given Node is stored within this NodeSet.
func (s NodeSet) Contains(node *Node) bool {
	return s.ContainsID(node.ID)
}

// ContainsID returns true if the Node represented by the given ID is stored within this NodeSet.
func (s NodeSet) ContainsID(id ID) bool {
	_, seen := s[id]
	return seen
}

// Add adds a given Node to the NodeSet.
func (s NodeSet) Add(nodes ...*Node) {
	for _, node := range nodes {
		s[node.ID] = node
	}
}

func (s NodeSet) AddIfNotExists(node *Node) bool {
	if _, exists := s[node.ID]; exists {
		return false
	}

	s[node.ID] = node
	return true
}

// AddSet merges all Nodes from the given NodeSet into this NodeSet.
func (s NodeSet) AddSet(set NodeSet) {
	for k, v := range set {
		s[k] = v
	}
}

// Slice returns a slice of the Node instances stored in this NodeSet.
func (s NodeSet) Slice() []*Node {
	slice := make([]*Node, 0, len(s))

	for _, v := range s {
		slice = append(slice, v)
	}

	return slice
}

// IDs returns a slice of database IDs for all nodes in the set.
func (s NodeSet) IDs() []ID {
	idList := make([]ID, 0, len(s))

	for _, node := range s {
		idList = append(idList, node.ID)
	}

	return idList
}

// IDBitmap returns a new roaring64.Bitmap instance containing all Node ID values in this NodeSet.
func (s NodeSet) IDBitmap() *roaring.Bitmap {
	bitmap := roaring.New()

	for id := range s {
		bitmap.Add(id.Uint32())
	}

	return bitmap
}

func (s *NodeSet) UnmarshalJSON(input []byte) error {
	var (
		tmpMap map[ID]serializableNode
	)

	if err := json.Unmarshal(input, &tmpMap); err != nil {
		return err
	}

	nodeSet := make(NodeSet, len(tmpMap))
	for key, value := range tmpMap {
		nodeSet[key] = &Node{
			ID:           value.ID,
			Kinds:        StringsToKinds(value.Kinds),
			AddedKinds:   StringsToKinds(value.AddedKinds),
			DeletedKinds: StringsToKinds(value.DeletedKinds),
			Properties:   value.Properties,
		}
	}

	*s = nodeSet

	return nil
}

type ThreadSafeNodeSet struct {
	nodeSet *NodeSet
	rwLock  *sync.RWMutex
}

func NewThreadSafeNodeSet(nodeSet NodeSet) *ThreadSafeNodeSet {
	return &ThreadSafeNodeSet{
		nodeSet: &nodeSet,
		rwLock:  &sync.RWMutex{},
	}
}

// Pick returns a single Node instance from this set. Repeated calls to this function are not guaranteed to return
// the same Node instance.
func (s ThreadSafeNodeSet) Pick() *Node {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.Pick()
}

// ContainingNodeKinds returns a new ThreadSafeNodeSet containing only Node instances that contain any one of the given Kind instances.
func (s ThreadSafeNodeSet) ContainingNodeKinds(kinds ...Kind) NodeSet {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.ContainingNodeKinds(kinds...)
}

// Remove removes a Node from this set by its database ID.
func (s *ThreadSafeNodeSet) Remove(id ID) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.nodeSet.Remove(id)
}

// Get returns a Node from this set by its database ID.
func (s ThreadSafeNodeSet) Get(id ID) *Node {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.Get(id)
}

// Len returns the number of unique Node instances in this set.
func (s ThreadSafeNodeSet) Len() int {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.Len()
}

// Copy returns a shallow copy of this set.
func (s ThreadSafeNodeSet) Copy() ThreadSafeNodeSet {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return *NewThreadSafeNodeSet(s.nodeSet.Copy())
}

// KindSet returns a NodeKindSet constructed from the Node instances in this set.
func (s ThreadSafeNodeSet) KindSet() NodeKindSet {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.KindSet()
}

// Contains returns true if the ID of the given Node is stored within this NodeSet.
func (s ThreadSafeNodeSet) Contains(node *Node) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.Contains(node)
}

// ContainsID returns true if the Node represented by the given ID is stored within this NodeSet.
func (s ThreadSafeNodeSet) ContainsID(id ID) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.ContainsID(id)
}

// Add adds a given Node to the NodeSet.
func (s *ThreadSafeNodeSet) Add(nodes ...*Node) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.nodeSet.Add(nodes...)
}

func (s *ThreadSafeNodeSet) AddIfNotExists(node *Node) bool {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	return s.nodeSet.AddIfNotExists(node)
}

// AddSet merges all Nodes from the given NodeSet into this NodeSet.
func (s *ThreadSafeNodeSet) AddSet(set NodeSet) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.nodeSet.AddSet(set)
}

// Slice returns a slice of the Node instances stored in this NodeSet.
func (s ThreadSafeNodeSet) Slice() []*Node {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.Slice()
}

// IDs returns a slice of database IDs for all nodes in the set.
func (s ThreadSafeNodeSet) IDs() []ID {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.IDs()
}

// IDBitmap returns a new roaring64.Bitmap instance containing all Node ID values in this NodeSet.
func (s ThreadSafeNodeSet) IDBitmap() *roaring.Bitmap {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.nodeSet.IDBitmap()
}

func Uint32SliceToIDs(raw []uint32) []ID {
	ids := make([]ID, len(raw))

	for idx, rawID := range raw {
		ids[idx] = ID(rawID)
	}

	return ids
}

func IDsToUint32Slice(ids []ID) []uint32 {
	rawIDs := make([]uint32, len(ids))

	for idx, id := range ids {
		rawIDs[idx] = id.Uint32()
	}

	return rawIDs
}

// NewNodeSet returns a new NodeSet from the given Node slice.
func NewNodeSet(nodes ...*Node) NodeSet {
	newSet := NodeSet{}

	for _, node := range nodes {
		newSet[node.ID] = node
	}

	return newSet
}

func MergeNodeSets(sets ...NodeSet) NodeSet {
	newSet := NodeSet{}

	for _, set := range sets {
		newSet.AddSet(set)
	}

	return newSet
}

func EmptyNodeSet() NodeSet {
	return NodeSet{}
}

type NodeKindSet map[string]NodeSet

func NewNodeKindSet(nodeSets ...NodeSet) NodeKindSet {
	newKindSet := NodeKindSet{}
	newKindSet.AddSets(nodeSets...)

	return newKindSet
}

// GetCombined returns a NodeSet of all nodes contained in this set that match the given kinds.
func (s NodeKindSet) GetCombined(kinds ...Kind) NodeSet {
	combinedSet := NodeSet{}

	for _, kind := range kinds {
		if set, hasSet := s[kind.String()]; hasSet {
			combinedSet.AddSet(set)
		}
	}

	return combinedSet
}

// EachNode iterates through each node contained within this set.
func (s NodeKindSet) EachNode(delegate func(node *Node) error) error {
	visitedIDs := roaring64.New()

	for _, set := range s {
		for _, node := range set {
			if nextID := node.ID.Uint64(); !visitedIDs.Contains(nextID) {
				visitedIDs.Add(nextID)

				if err := delegate(node); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetNode fetches a Node from this set by its database ID.
func (s NodeKindSet) GetNode(id ID) *Node {
	for _, set := range s {
		if node := set.Get(id); node != nil {
			return node
		}
	}

	return nil
}

// AllNodeIDs returns all node IDs contained with in this set.
func (s NodeKindSet) AllNodeIDs() []ID {
	allIDs := roaring64.New()

	for _, set := range s {
		for _, node := range set {
			allIDs.Add(node.ID.Uint64())
		}
	}

	var (
		uint64Array = allIDs.ToArray()
		returnArray = make([]ID, len(uint64Array))
	)

	for idx, value := range uint64Array {
		returnArray[idx] = ID(value)
	}

	return returnArray
}

// AllNodes  returns all nodes present in this set as a NodeSet.
func (s NodeKindSet) AllNodes() NodeSet {
	var allSets = NodeSet{}

	for _, set := range s {
		allSets.AddSet(set)
	}

	return allSets
}

// CountAll returns the count of all unique nodes in the set.
func (s NodeKindSet) CountAll() int64 {
	var bitmap = roaring64.New()

	for _, set := range s {
		for _, node := range set {
			bitmap.Add(node.ID.Uint64())
		}
	}

	return int64(bitmap.GetCardinality())
}

// Copy returns a shallow copy of this set.
func (s NodeKindSet) Copy() NodeKindSet {
	newKindSet := NodeKindSet{}
	newKindSet.AddKindSet(s)

	return newKindSet
}

// RemoveNode removes a Node from this set by its database ID.
func (s NodeKindSet) RemoveNode(id ID) {
	for _, nodeSet := range s {
		nodeSet.Remove(id)
	}
}

// Count returns the count unique nodes for each given kind, summed.
func (s NodeKindSet) Count(kinds ...Kind) int64 {
	var bitmap = roaring64.New()

	for _, kind := range kinds {
		if set, hasKind := s[kind.String()]; hasKind {
			for _, node := range set {
				bitmap.Add(node.ID.Uint64())
			}
		}
	}

	return int64(bitmap.GetCardinality())
}

// Get returns the NodeSet for a given Kind. If there is no NodeSet for the given Kind then an empty NodeSet is returned.
func (s NodeKindSet) Get(kind Kind) NodeSet {
	if set, found := s[kind.String()]; found {
		return set
	}

	return EmptyNodeSet()
}

func (s NodeKindSet) addNode(node *Node) {
	for _, nodeKind := range node.Kinds {
		if existingNodeSet, found := s[nodeKind.String()]; !found {
			newNodeSet := NodeSet{
				node.ID: node,
			}

			s[nodeKind.String()] = newNodeSet
		} else {
			existingNodeSet.Add(node)
		}
	}
}

// Add adds the given list of Node types to this NodeKindSet.
func (s NodeKindSet) Add(nodes ...*Node) {
	for _, node := range nodes {
		s.addNode(node)
	}
}

// AddSets adds the given NodeSet instances to this NodeKindSet.
func (s NodeKindSet) AddSets(nodeSets ...NodeSet) {
	for _, nodeSet := range nodeSets {
		for _, node := range nodeSet {
			s.addNode(node)
		}
	}
}

// AddKindSet adds the given NodeKindSet to this NodeKindSet.
func (s NodeKindSet) AddKindSet(set NodeKindSet) {
	for _, nodeSet := range set {
		s.AddSets(nodeSet)
	}
}
