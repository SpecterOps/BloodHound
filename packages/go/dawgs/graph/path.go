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
	"strings"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type Path struct {
	Nodes []*Node
	Edges []*Relationship
}

func AllocatePath(pathDepth int) Path {
	return Path{
		Nodes: make([]*Node, pathDepth+1),
		Edges: make([]*Relationship, pathDepth),
	}
}

func (s Path) Walk(delegate func(start, end *Node, relationship *Relationship) bool) {
	for idx := 1; idx < len(s.Nodes); idx++ {
		if shouldContinue := delegate(s.Nodes[idx-1], s.Nodes[idx], s.Edges[idx-1]); !shouldContinue {
			break
		}
	}
}

func (s Path) WalkReverse(delegate func(start, end *Node, relationship *Relationship) bool) {
	for idx := len(s.Nodes) - 2; idx >= 0; idx-- {
		if shouldContinue := delegate(s.Nodes[idx], s.Nodes[idx+1], s.Edges[idx]); !shouldContinue {
			break
		}
	}
}

func (s Path) Root() *Node {
	return s.Nodes[0]
}

func (s Path) ContainsNode(id ID) bool {
	for _, node := range s.Nodes {
		if node.ID == id {
			return true
		}
	}

	return false
}

func (s Path) Terminal() *Node {
	return s.Nodes[len(s.Nodes)-1]
}

type IDPath struct {
	Nodes []ID
	Edges []ID
}

func AllocateIDPath(pathDepth int) IDPath {
	return IDPath{
		Nodes: make([]ID, pathDepth+1),
		Edges: make([]ID, pathDepth),
	}
}

func (s IDPath) Walk(delegate func(start, end, relationship ID) bool) {
	for idx := 1; idx < len(s.Nodes); idx++ {
		if shouldContinue := delegate(s.Nodes[idx-1], s.Nodes[idx], s.Edges[idx-1]); !shouldContinue {
			break
		}
	}
}

func (s IDPath) WalkReverse(delegate func(start, end, relationship ID) bool) {
	for idx := len(s.Nodes) - 2; idx >= 0; idx-- {
		if shouldContinue := delegate(s.Nodes[idx], s.Nodes[idx+1], s.Edges[idx]); !shouldContinue {
			break
		}
	}
}

func (s IDPath) Root() ID {
	return s.Nodes[0]
}

func (s IDPath) Terminal() ID {
	return s.Nodes[len(s.Nodes)-1]
}

type IDSegment struct {
	Node     ID
	Edge     ID
	Trunk    *IDSegment
	Branches []*IDSegment
	size     size.Size
}

func NewRootIDSegment(root ID) *IDSegment {
	newSegment := &IDSegment{
		Node: root,
	}

	newSegment.updateSize()
	return newSegment
}

func (s *IDSegment) sizeOfBranches() size.Size {
	var sizeOfBranches size.Size

	for _, branch := range s.Branches {
		sizeOfBranches += branch.SizeOf()
	}

	return sizeOfBranches
}

// TODO: Warning, this is a data race condition that should be covered by atomics
func (s *IDSegment) updateSize() {
	var (
		previousSize = s.size
		newSize      = size.Of(*s) + size.OfSlice(s.Branches) + s.sizeOfBranches()
	)

	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		cursor.size -= previousSize
		cursor.size += newSize
	}
}

func (s *IDSegment) SizeOf() size.Size {
	sizeCopy := s.size
	return sizeCopy
}

func (s *IDSegment) IsRoot() bool {
	return s.Edge == 0 && s.Trunk == nil
}

func (s *IDSegment) Descend(endID, edgeID ID) *IDSegment {
	var (
		branchSliceCapacity = cap(s.Branches)
		nextEdge            = &IDSegment{
			Node:  endID,
			Edge:  edgeID,
			Trunk: s,
		}
	)

	// Track this edge on the list of branches
	s.Branches = append(s.Branches, nextEdge)

	// Did we allocate more slice space, if so update our size
	if branchSliceCapacity != cap(s.Branches) {
		s.updateSize()
	}

	// Update sizing information up the tree for the new segment and then return it
	nextEdge.updateSize()
	return nextEdge
}

func (s *IDSegment) Depth() int {
	depth := 0

	for cursor := s; cursor.Trunk != nil; cursor = cursor.Trunk {
		depth++
	}

	return depth
}

func (s *IDSegment) WalkReverse(delegate func(nextSegment *IDSegment) bool) {
	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		if !delegate(cursor) {
			break
		}
	}
}

func (s *IDSegment) Detach() {
	if s.Trunk != nil {
		// If there's a trunk, remove this node from the trunk root's edges
		for idx, trunkBranch := range s.Trunk.Branches {
			if trunkBranch.Edge == s.Edge {
				s.Trunk.Branches = append(s.Trunk.Branches[:idx], s.Trunk.Branches[idx+1:]...)
				s.Trunk.updateSize()
			}
		}
	}
}

func (s *IDSegment) Path() IDPath {
	var (
		depthIdx = s.Depth()
		path     = AllocateIDPath(depthIdx)
	)

	s.WalkReverse(func(nextSegment *IDSegment) bool {
		path.Nodes[depthIdx] = nextSegment.Node

		if nextSegment.Trunk != nil {
			path.Edges[depthIdx-1] = nextSegment.Edge
		}

		depthIdx--
		return true
	})

	return path
}

func (s *IDSegment) IsCycle() bool {
	bitmap := roaring64.NewBitmap()

	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		if !bitmap.CheckedAdd(cursor.Node.Uint64()) {
			return true
		}
	}

	return false
}

type Tree struct {
	Root *PathSegment
}

func NewTree(root *Node) Tree {
	return Tree{
		Root: &PathSegment{
			Node: root,
		},
	}
}

func (s Tree) SizeOf() size.Size {
	return size.Of(s) + s.Root.size
}

type PathSegment struct {
	Node     *Node
	Trunk    *PathSegment
	Edge     *Relationship
	Branches []*PathSegment
	Tag      any
	size     size.Size
}

func sizeOfPathSegment(segment *PathSegment) size.Size {
	return size.Of(*segment) + segment.Node.SizeOf() + size.Of(segment.Branches)*size.Size(cap(segment.Branches))
}

func NewRootPathSegment(root *Node) *PathSegment {
	newSegment := &PathSegment{
		Node: root,
	}

	newSegment.size = sizeOfPathSegment(newSegment)
	return newSegment
}

func (s *PathSegment) GetTrunkSegment() *PathSegment {
	if s.Trunk != nil {
		return s.Trunk
	}

	return nil
}

func (s *PathSegment) SizeOf() size.Size {
	return s.size
}

func (s *PathSegment) IsCycle() bool {
	if s.Trunk != nil {
		var (
			terminal = s.Node
			cursor   = s.Trunk
		)

		for {
			if terminal.ID == cursor.Node.ID {
				return true
			}

			if cursor.Trunk != nil {
				cursor = cursor.Trunk
			} else {
				break
			}
		}
	}

	return false
}

func (s *PathSegment) Depth() int {
	depth := 0

	for cursor := s; cursor.Trunk != nil; cursor = cursor.Trunk {
		depth++
	}

	return depth
}

func (s *PathSegment) Path() Path {
	var (
		depthIdx = s.Depth()
		path     = AllocatePath(depthIdx)
	)

	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		path.Nodes[depthIdx] = cursor.Node

		if cursor.Trunk != nil {
			path.Edges[depthIdx-1] = cursor.Edge
		}

		depthIdx--
	}

	return path
}

func (s *PathSegment) Slice() []*PathSegment {
	var (
		containerIdx = s.Depth()
		container    = make([]*PathSegment, containerIdx)
	)

	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		containerIdx--
		container[containerIdx] = cursor
	}

	return container
}

func (s *PathSegment) WalkReverse(delegate func(nextSegment *PathSegment) bool) {
	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		if !delegate(cursor) {
			break
		}
	}
}

func (s *PathSegment) Search(delegate func(nextSegment *PathSegment) bool) *Node {
	for cursor := s; cursor != nil; cursor = cursor.Trunk {
		if delegate(cursor) {
			return cursor.Node
		}
	}

	return nil
}

func (s *PathSegment) Detach() {
	var (
		sizeDetached = s.SizeOf()
	)

	if s.Trunk != nil {
		// If there's a trunk, remove this node from the trunk root's edges
		for idx, trunkRootBranch := range s.Trunk.Branches {
			if trunkRootBranch.Edge.ID == s.Edge.ID && trunkRootBranch.Edge.Kind.Is(s.Edge.Kind) {
				s.Trunk.Branches = append(s.Trunk.Branches[:idx], s.Trunk.Branches[idx+1:]...)
				break
			}
		}
	}

	// Update size of the path tree now that this segment has been detached
	for sizeCursor := s; sizeCursor != nil; sizeCursor = sizeCursor.Trunk {
		sizeCursor.size -= sizeDetached
	}
}

func (s *PathSegment) Descend(node *Node, relationship *Relationship) *PathSegment {
	var (
		nextSegment = &PathSegment{
			Node:  node,
			Trunk: s,
			Edge:  relationship,
		}
		sizeAdded         = sizeOfPathSegment(nextSegment)
		oldBranchCapacity = cap(s.Branches)
	)

	// Track the size of the segment
	nextSegment.size = sizeAdded

	// Track this edge on the list of branches
	s.Branches = append(s.Branches, nextSegment)

	// Update the size if we increased the capacity of the branches slice for this segment
	if newCapacity := cap(s.Branches); newCapacity != oldBranchCapacity {
		capacityAdded := newCapacity - oldBranchCapacity
		sizeAdded += size.Of(s.Branches) * size.Size(capacityAdded)
	}

	// Track size on the root segment of this path tree
	for sizeCursor := s; sizeCursor != nil; sizeCursor = sizeCursor.Trunk {
		sizeCursor.size += sizeAdded
	}

	return nextSegment
}

// FormatPathSegment outputs a cypher-formatted path from the given PathSegment pointer
func FormatPathSegment(segment *PathSegment) string {
	formatted := strings.Builder{}

	segment.WalkReverse(func(nextSegment *PathSegment) bool {
		formatted.WriteString("(")
		formatted.WriteString(nextSegment.Node.ID.String())
		formatted.WriteString(")")

		if nextSegment.Trunk != nil {
			formatted.WriteString("<-[")
			formatted.WriteString(nextSegment.Edge.Kind.String())
			formatted.WriteString("]-")
		}

		return true
	})

	return formatted.String()
}

// PathSet is a collection of graph traversals stored as Path instances.
type PathSet []Path

func NewPathSet(paths ...Path) PathSet {
	return paths
}

func (s PathSet) FilterByEdge(filter func(edge *Relationship) bool) PathSet {
	var matchingPaths PathSet

	for _, path := range s {
		include := true

		for _, edge := range path.Edges {
			if !filter(edge) {
				include = false
				break
			}
		}

		if include {
			matchingPaths = append(matchingPaths, path)
		}
	}

	return matchingPaths
}

func (s PathSet) IncludeByEdgeKinds(edgeKinds Kinds) PathSet {
	return s.FilterByEdge(func(edge *Relationship) bool {
		return edgeKinds.ContainsOneOf(edge.Kind)
	})
}

func (s PathSet) ExcludeByEdgeKinds(edgeKinds Kinds) PathSet {
	return s.FilterByEdge(func(edge *Relationship) bool {
		return !edgeKinds.ContainsOneOf(edge.Kind)
	})
}

func (s PathSet) Paths() []Path {
	return s
}

func (s PathSet) Len() int {
	return len(s)
}

func (s PathSet) Roots() NodeSet {
	nodes := NewNodeSet()

	for _, nextPath := range s {
		nodes.Add(nextPath.Root())
	}

	return nodes
}

func (s PathSet) Terminals() NodeSet {
	nodes := NewNodeSet()

	for _, nextPath := range s {
		nodes.Add(nextPath.Terminal())
	}

	return nodes
}

func (s PathSet) AllNodes() NodeSet {
	nodes := NewNodeSet()

	for _, nextPath := range s {
		nodes.Add(nextPath.Nodes...)
	}

	return nodes
}

func (s *PathSet) AddPath(path Path) {
	*s = append(*s, path)
}

func (s *PathSet) AddPathSet(pathSet PathSet) {
	*s = append(*s, pathSet.Paths()...)
}
