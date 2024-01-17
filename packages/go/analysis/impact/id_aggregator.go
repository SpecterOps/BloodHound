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

package impact

import (
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"sync"
)

type PathAggregator interface {
	Cardinality(targets ...uint32) cardinality.Provider[uint32]
	Contains(target uint32) bool
	AddPath(path *graph.IDSegment)
	AddShortcut(path *graph.IDSegment)
}

type ThreadSafeAggregator struct {
	aggregator PathAggregator
	lock       *sync.RWMutex
}

func (s ThreadSafeAggregator) Contains(target uint32) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.aggregator.Contains(target)
}

func (s ThreadSafeAggregator) Cardinality(targets ...uint32) cardinality.Provider[uint32] {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.aggregator.Cardinality(targets...)
}

func (s ThreadSafeAggregator) AddPath(path *graph.IDSegment) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.aggregator.AddPath(path)
}

func (s ThreadSafeAggregator) AddShortcut(path *graph.IDSegment) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.aggregator.AddShortcut(path)
}

func NewThreadSafeAggregator(aggregator PathAggregator) PathAggregator {
	return &ThreadSafeAggregator{
		aggregator: aggregator,
		lock:       &sync.RWMutex{},
	}
}

// IDA is a cardinality aggregator for paths and shortcut paths.
//
// When encoding shortcut paths the aggregator will track node dependencies for nodes that otherwise would have missing
// cardinality entries. Dependencies are organized as an adjacency list for each node. These adjacency lists combine to
// make a dependency graph of cardinalities that can be traversed.
//
// Once all paths are encoded, shortcut or otherwise, into the aggregator, users may then resolve the full cardinality
// of nodes by calling the cardinality functions of the aggregator. Resolution is accomplished using a recursive
// depth-first strategy.
type IDA struct {
	resolved               cardinality.Duplex[uint32]
	cardinalities          *graph.IndexedSlice[uint32, cardinality.Provider[uint32]]
	dependencies           map[uint32]cardinality.Duplex[uint32]
	newCardinalityProvider cardinality.ProviderConstructor[uint32]
}

func (s IDA) Contains(target uint32) bool {
	return s.cardinalities.Has(target)
}

func NewIDA(newCardinalityProvider cardinality.ProviderConstructor[uint32]) IDA {
	return IDA{
		cardinalities:          graph.NewIndexedSlice[uint32, cardinality.Provider[uint32]](),
		dependencies:           map[uint32]cardinality.Duplex[uint32]{},
		resolved:               cardinality.NewBitmap32(),
		newCardinalityProvider: newCardinalityProvider,
	}
}

// pushDependency adds a new dependency for the given target.
func (s IDA) pushDependency(target, dependency uint32) {
	if dependencies, hasDependencies := s.dependencies[target]; hasDependencies {
		dependencies.Add(dependency)
	} else {
		newDependencies := cardinality.NewBitmap32()
		newDependencies.Add(dependency)

		s.dependencies[target] = newDependencies
	}
}

// popDependencies will take the simplex cardinality provider reference for the given target, remove it from the
// containing map in the aggregator and then return it
func (s IDA) popDependencies(targetUint32ID uint32) []uint32 {
	dependencies, hasDependencies := s.dependencies[targetUint32ID]
	delete(s.dependencies, targetUint32ID)

	if hasDependencies {
		return dependencies.Slice()
	}

	return nil
}

func (s IDA) membership(targetUint32ID uint32) cardinality.Provider[uint32] {
	return s.cardinalities.GetOr(targetUint32ID, s.newCardinalityProvider)
}

// idaRes is a cursor type that tracks the resolution of a node's pathMembers
type idaRes struct {
	// target is the uint32 ID of the node being resolved
	target uint32

	// pathMembers stores the cardinality of the target's path membership
	pathMembers cardinality.Provider[uint32]

	// completions are cardinality providers that will have this resolution's pathMembers merged into them
	completions []cardinality.Provider[uint32]

	// dependencies contains a slice of uint32 node IDs that this resolution depends on
	dependencies []uint32
}

// resolve takes the target uint32 ID of a node and calculates the cardinality of nodes that have a path that traverse
// it
func (s IDA) resolve(targetUint32ID uint32) cardinality.Provider[uint32] {
	var (
		targetImpact = s.membership(targetUint32ID)
		resolutions  = map[uint32]*idaRes{
			targetUint32ID: {
				target:       targetUint32ID,
				pathMembers:  targetImpact,
				dependencies: s.popDependencies(targetUint32ID),
			},
		}
		stack = []uint32{targetUint32ID}
	)

	for len(stack) > 0 {
		// Pick up the next resolution
		next := resolutions[stack[len(stack)-1]]

		// Exhaust the resolution's dependencies
		if len(next.dependencies) > 0 {
			nextDependency := next.dependencies[len(next.dependencies)-1]
			next.dependencies = next.dependencies[:len(next.dependencies)-1]

			if s.resolved.Contains(nextDependency) {
				// If this dependency has already been resolved, fetch and or it with this resolution's pathMembers
				next.pathMembers.Or(s.cardinalities.Get(nextDependency))
			} else if inProgressResolution, hasResolution := resolutions[nextDependency]; hasResolution {
				// If this dependency is in the process of being resolved; track this node (var next) as a completion
				// to or with the in progress resolutions pathMembers once fully resolved
				inProgressResolution.completions = append(inProgressResolution.completions, next.pathMembers)
			} else {
				// For each dependency not already resolved or in-progress is descended into as a new resolution
				stack = append(stack, nextDependency)
				resolutions[nextDependency] = &idaRes{
					target:       nextDependency,
					pathMembers:  s.membership(nextDependency),
					completions:  []cardinality.Provider[uint32]{next.pathMembers},
					dependencies: s.popDependencies(nextDependency),
				}
			}
		} else {
			// Pop the resolution from our dependency unwind
			stack = stack[:len(stack)-1]
		}
	}

	// First resolution pass for completion dependencies
	for _, nextResolution := range resolutions {
		for _, nextCompletion := range nextResolution.completions {
			nextCompletion.Or(nextResolution.pathMembers)
		}
	}

	// Second resolution pass for completion dependencies that were not fully resolved on the first pass
	for _, nextResolution := range resolutions {
		for _, nextCompletion := range nextResolution.completions {
			nextCompletion.Or(nextResolution.pathMembers)
		}

		s.resolved.Add(nextResolution.target)
	}

	return targetImpact
}

func (s IDA) Cardinality(targets ...uint32) cardinality.Provider[uint32] {
	log.Debugf("Calculating pathMembers cardinality for %d targets", len(targets))
	defer log.Measure(log.LevelDebug, "Calculated pathMembers cardinality for %d targets", len(targets))()

	impact := s.newCardinalityProvider()

	for _, target := range targets {
		if s.resolved.Contains(target) {
			impact.Or(s.cardinalities.Get(target))
		} else {
			impact.Or(s.resolve(target))
		}
	}

	return impact
}

func (s IDA) AddPath(path *graph.IDSegment) {
	pathMembers := []uint32{path.Node.Uint32()}

	for cursor := path.Trunk; cursor != nil; cursor = cursor.Trunk {
		cursorNodeUint32ID := cursor.Node.Uint32()

		// Roll up cardinalities for nodes that belong to the path
		s.membership(cursorNodeUint32ID).Add(pathMembers...)
		pathMembers = append(pathMembers, cursor.Node.Uint32())
	}
}

func (s IDA) AddShortcut(path *graph.IDSegment) {
	var (
		terminalUint32ID = path.Node.Uint32()
		pathMembers      = []uint32{terminalUint32ID}
	)

	for cursor := path.Trunk; cursor != nil; cursor = cursor.Trunk {
		cursorNodeUint32ID := cursor.Node.Uint32()

		// The terminal node of this path was not fully traversed, so push it as a dependency of all ascending nodes
		// above it
		s.pushDependency(cursorNodeUint32ID, terminalUint32ID)

		// Roll up cardinalities for nodes that belong to the path
		s.membership(cursorNodeUint32ID).Add(pathMembers...)
		pathMembers = append(pathMembers, cursorNodeUint32ID)
	}
}

func (s IDA) Resolved() cardinality.Duplex[uint32] {
	return s.resolved
}
