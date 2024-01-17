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
)

// Aggregator is a cardinality aggregator for paths and shortcut paths.
//
// When encoding shortcut paths the aggregator will track node dependencies for nodes that otherwise would have missing
// cardinality entries. Dependencies are organized as an adjacency list for each node. These adjacency lists combine to
// make a dependency graph of cardinalities that can be traversed.
//
// Once all paths are encoded, shortcut or otherwise, into the aggregator, users may then resolve the full cardinality
// of nodes by calling the cardinality functions of the aggregator. Resolution is accomplished using a recursive
// depth-first strategy.
type Aggregator struct {
	resolved               cardinality.Duplex[uint32]
	cardinalities          *graph.IndexedSlice[uint32, cardinality.Provider[uint32]]
	dependencies           map[uint32]cardinality.Duplex[uint32]
	newCardinalityProvider cardinality.ProviderConstructor[uint32]
}

func NewAggregator(newCardinalityProvider cardinality.ProviderConstructor[uint32]) Aggregator {
	return Aggregator{
		cardinalities:          graph.NewIndexedSlice[uint32, cardinality.Provider[uint32]](),
		dependencies:           map[uint32]cardinality.Duplex[uint32]{},
		resolved:               cardinality.NewBitmap32(),
		newCardinalityProvider: newCardinalityProvider,
	}
}

// pushDependency adds a new dependency for the given target.
func (s Aggregator) pushDependency(target, dependency uint32) {
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
func (s Aggregator) popDependencies(targetUint32ID uint32) []uint32 {
	dependencies, hasDependencies := s.dependencies[targetUint32ID]
	delete(s.dependencies, targetUint32ID)

	if hasDependencies {
		return dependencies.Slice()
	}

	return nil
}

func (s Aggregator) getImpact(targetUint32ID uint32) cardinality.Provider[uint32] {
	return s.cardinalities.GetOr(targetUint32ID, s.newCardinalityProvider)
}

// resolution is a cursor type that tracks the resolution of a node's impact
type resolution struct {
	// target is the uint32 ID of the node being resolved
	target uint32

	// impact stores the cardinality of the target's impact
	impact cardinality.Provider[uint32]

	// completions are cardinality providers that will have this resolution's impact merged into them
	completions []cardinality.Provider[uint32]

	// dependencies contains a slice of uint32 node IDs that this resolution depends on
	dependencies []uint32
}

// resolve takes the target uint32 ID of a node and calculates the cardinality of nodes that have a path that traverse
// it
func (s Aggregator) resolve(targetUint32ID uint32) cardinality.Provider[uint32] {
	var (
		targetImpact = s.getImpact(targetUint32ID)
		resolutions  = map[uint32]*resolution{
			targetUint32ID: {
				target:       targetUint32ID,
				impact:       targetImpact,
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
				next.impact.Or(s.cardinalities.Get(nextDependency))
			} else if inProgressResolution, hasResolution := resolutions[nextDependency]; hasResolution {
				// If this dependency is in the process of being resolved; track this node (var next) as a completion
				// to or with the in progress resolutions pathMembers once fully resolved
				inProgressResolution.completions = append(inProgressResolution.completions, next.impact)
			} else {
				// For each dependency not already resolved or in-progress is descended into as a new resolution
				stack = append(stack, nextDependency)
				resolutions[nextDependency] = &resolution{
					target:       nextDependency,
					impact:       s.getImpact(nextDependency),
					completions:  []cardinality.Provider[uint32]{next.impact},
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
			nextCompletion.Or(nextResolution.impact)
		}
	}

	// Second resolution pass for completion dependencies that were not fully resolved on the first pass
	for _, nextResolution := range resolutions {
		for _, nextCompletion := range nextResolution.completions {
			nextCompletion.Or(nextResolution.impact)
		}

		s.resolved.Add(nextResolution.target)
	}

	return targetImpact
}

func (s Aggregator) Cardinality(targets ...uint32) cardinality.Provider[uint32] {
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

func (s Aggregator) AddPath(path *graph.PathSegment, impactKinds graph.Kinds) {
	var impactingNodes []uint32

	if path.Node.Kinds.ContainsOneOf(impactKinds...) {
		impactingNodes = append(impactingNodes, path.Node.ID.Uint32())
	}

	for cursor := path.Trunk; cursor != nil; cursor = cursor.Trunk {
		// Only pull the pathMembers from the map if we have nodes that should be counted for this cursor
		if len(impactingNodes) > 0 {
			s.getImpact(cursor.Node.ID.Uint32()).Add(impactingNodes...)
		}

		// Only roll up cardinalities for nodes that belong to the set of impacting kinds
		if cursor.Node.Kinds.ContainsOneOf(impactKinds...) {
			impactingNodes = append(impactingNodes, cursor.Node.ID.Uint32())
		}
	}
}

func (s Aggregator) AddShortcut(path *graph.PathSegment, impactKinds graph.Kinds) {
	var (
		terminalUint32ID = path.Node.ID.Uint32()
		impactingNodes   []uint32
	)

	// Only add the terminal to the impacting nodes if it's a type that imparts impact - this does not remove the
	// shortcut from dependency tracking of upstream impacted nodes
	if path.Node.Kinds.ContainsOneOf(impactKinds...) {
		impactingNodes = append(impactingNodes, terminalUint32ID)
	}

	for cursor := path.Trunk; cursor != nil; cursor = cursor.Trunk {
		cursorNodeUint32ID := cursor.Node.ID.Uint32()

		// Add the terminal shortcut as a dependency to each ascending node
		s.pushDependency(cursorNodeUint32ID, terminalUint32ID)

		// Only pull the pathMembers from the map if we have nodes that should be counted for this cursor
		if len(impactingNodes) > 0 {
			s.getImpact(cursorNodeUint32ID).Add(impactingNodes...)
		}

		// Only roll up cardinalities for nodes that belong to the set of impacting kinds
		if cursor.Node.Kinds.ContainsOneOf(impactKinds...) {
			impactingNodes = append(impactingNodes, cursor.Node.ID.Uint32())
		}
	}
}

func (s Aggregator) Resolved() cardinality.Duplex[uint32] {
	return s.resolved
}
