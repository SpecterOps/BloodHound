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

package ops

import (
	"errors"
	"fmt"

	"github.com/RoaringBitmap/roaring/roaring64"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

type LimitSkipTracker struct {
	Limit int
	seen  int
	Skip  int
}

func (s *LimitSkipTracker) AtLimit() bool {
	return s.Limit > 0 && s.seen >= s.Limit
}

func (s *LimitSkipTracker) ShouldCollect() bool {
	// If a skip was set then we tick down until s.Skip is no longer greater than 0
	if s.Skip > 0 {
		s.Skip--
		return false
	}

	// If a limit was set, and we're not yet at it, increment the seen count and then return true
	if !s.AtLimit() {
		s.seen++
		return true
	}

	// Otherwise, we are at our limit and should no longer collect
	return false
}

var (
	ErrHaltTraversal        = errors.New("halt traversal")
	ErrTraversalMemoryLimit = errors.New("traversal required more memory than allowed")
)

// PathFilter is invoked on completed paths identified during a graph traversal. It may return a boolean value
// representing if the path was consumed. If consumed, a rendered path is then tracked for traversal plan limit
// specifications.
type PathFilter func(ctx *TraversalContext, segment *graph.PathSegment) bool

// PathVisitor is invoked on completed paths identified during a graph traversal. It may return an error value in the
// case where a fatal error condition has been encountered, rendering further traversal moot.
type PathVisitor func(ctx *TraversalContext, segment *graph.PathSegment) error

// DepthExceptionHandler is invoked on paths that exceed depth traversal plan depth limits.
type DepthExceptionHandler func(ctx *TraversalContext, segment *graph.PathSegment)

type SegmentFilter func(ctx *TraversalContext, segment *graph.PathSegment) bool

type NodeFilter func(node *graph.Node) bool

type TraversalPlan struct {
	Root                  *graph.Node
	Direction             graph.Direction
	BranchQuery           graph.CriteriaProvider
	DepthExceptionHandler DepthExceptionHandler
	DescentFilter         SegmentFilter
	PathFilter            PathFilter
	Skip                  int
	Limit                 int
}

func nextTraversal(tx graph.Transaction, segment *graph.PathSegment, direction graph.Direction, branchFilter graph.CriteriaProvider, requireOrder bool) ([]*graph.PathSegment, error) {
	var (
		branches           []*graph.PathSegment
		nextTraversalQuery = tx.Relationships().Filterf(func() graph.Criteria {
			var filters []graph.Criteria

			if branchFilter != nil {
				filters = append(filters, branchFilter())
			}

			switch direction {
			case graph.DirectionOutbound:
				filters = append(filters, query.InIDs(query.StartID(), segment.Node.ID))

			case graph.DirectionInbound:
				filters = append(filters, query.InIDs(query.EndID(), segment.Node.ID))
			}

			return query.And(
				filters...,
			)
		})
	)

	if requireOrder {
		nextTraversalQuery.OrderBy(query.Order(query.Relationship(), query.Ascending()))
	}

	switch direction {
	case graph.DirectionOutbound:
		return branches, ForEachEndNode(nextTraversalQuery, func(relationship *graph.Relationship, node *graph.Node) error {
			branches = append(branches, segment.Descend(node, relationship))
			return nil
		})

	case graph.DirectionInbound:
		return branches, ForEachStartNode(nextTraversalQuery, func(relationship *graph.Relationship, node *graph.Node) error {
			branches = append(branches, segment.Descend(node, relationship))
			return nil
		})

	default:
		return nil, fmt.Errorf("invalid direction %d", direction)
	}
}

type TraversalContext struct {
	LimitSkipTracker LimitSkipTracker
}

func AcyclicTraversal(tx graph.Transaction, plan TraversalPlan, pathVisitors ...PathVisitor) error {
	var (
		descentFilter = plan.DescentFilter
		visitedBitmap = roaring64.New()
	)

	plan.DescentFilter = func(ctx *TraversalContext, segment *graph.PathSegment) bool {
		if terminalID := segment.Node.ID.Uint64(); visitedBitmap.Contains(terminalID) {
			return false
		} else {
			visitedBitmap.Add(terminalID)
		}

		return descentFilter == nil || descentFilter(ctx, segment)
	}

	return Traversal(tx, plan, pathVisitors...)
}

func Traversal(tx graph.Transaction, plan TraversalPlan, pathVisitors ...PathVisitor) error {
	var (
		pathVisitor           PathVisitor
		requireTraversalOrder = plan.Limit > 0 || plan.Skip > 0
		rootSegment           = graph.NewRootPathSegment(plan.Root)
		stack                 = []*graph.PathSegment{rootSegment}
		ctx                   *TraversalContext
	)

	ctx = &TraversalContext{
		LimitSkipTracker: LimitSkipTracker{
			Limit: plan.Limit,
			Skip:  plan.Skip,
		},
	}

	if pvLen := len(pathVisitors); pvLen > 1 {
		return fmt.Errorf("specifying more than 1 path visitor is not supported")
	} else if pvLen == 1 {
		pathVisitor = pathVisitors[0]
	}

	for len(stack) > 0 {
		next := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if pathTreeSize := rootSegment.SizeOf(); pathTreeSize > tx.TraversalMemoryLimit() {
			return fmt.Errorf("%w - Limit: %.2f MB - Memory In-Use: %.2f MB", ErrTraversalMemoryLimit, tx.TraversalMemoryLimit().Mebibytes(), pathTreeSize.Mebibytes())
		}

		if descendents, err := nextTraversal(tx, next, plan.Direction, plan.BranchQuery, requireTraversalOrder); err != nil {
			// If the error value is the halt traversal sentinel then don't relay any error upstream
			if err == ErrHaltTraversal {
				break
			}

			return err
		} else {
			stackLengthBeforeDescent := len(stack)

			if plan.DescentFilter != nil {
				// If there's a descent filter specified we need to unwind all possible descent candidates and test
				// them. To avoid the annoying additional memory pressure that comes from the `range` keyword we iterate
				// in-place.
				for idx := 0; idx < len(descendents); idx++ {
					if nextDescendent := descendents[idx]; plan.DescentFilter(ctx, nextDescendent) {
						stack = append(stack, nextDescendent)
					}
				}
			} else {
				// No filter means descend into all potential paths
				stack = append(stack, descendents...)
			}

			// If this node does not have descendents then it's a path terminal
			if pathVisitor != nil && stackLengthBeforeDescent == len(stack) && next.Depth() > 0 && (plan.PathFilter == nil || plan.PathFilter(ctx, next)) {
				if err := pathVisitor(ctx, next); err != nil {
					return err
				}
			}
		}

		// Break if we're at our limit
		if ctx.LimitSkipTracker.AtLimit() {
			break
		}
	}

	return nil
}

// TraverseIntermediaryPaths NodeFilter is used to select candidate nodes for adding to the results
func TraverseIntermediaryPaths(tx graph.Transaction, plan TraversalPlan, nodeFilter NodeFilter) (graph.PathSet, error) {
	var (
		paths         = graph.NewPathSet()
		descentFilter = plan.DescentFilter
	)

	plan.DescentFilter = func(ctx *TraversalContext, segment *graph.PathSegment) bool {
		if descentFilter != nil && !descentFilter(ctx, segment) {
			return false
		}

		if nodeFilter(segment.Node) && ctx.LimitSkipTracker.ShouldCollect() {
			paths.AddPath(segment.Path())
		}

		return true
	}

	return paths, Traversal(tx, plan, nil)
}

// AcyclicTraverseNodes Does a traversal, but includes nodes that are intermediaries and terminals
func AcyclicTraverseNodes(tx graph.Transaction, plan TraversalPlan, nodeFilter NodeFilter) (graph.NodeSet, error) {
	var (
		nodes         = graph.NewNodeSet()
		descentFilter = plan.DescentFilter
		visitedBitmap = roaring64.New()
	)

	// Wrap our descent filter so we can test candidates
	plan.DescentFilter = func(ctx *TraversalContext, segment *graph.PathSegment) bool {
		if terminalID := segment.Node.ID.Uint64(); visitedBitmap.Contains(terminalID) {
			return false
		} else {
			visitedBitmap.Add(terminalID)
		}

		if descentFilter != nil && !descentFilter(ctx, segment) {
			return false
		}

		if (nodeFilter == nil || nodeFilter(segment.Node)) && ctx.LimitSkipTracker.ShouldCollect() {
			nodes.Add(segment.Node)
		}

		return true
	}

	//Remember to test our root node as well
	if nodeFilter == nil || nodeFilter(plan.Root) {
		nodes.Add(plan.Root)
	}

	return nodes, Traversal(tx, plan, nil)
}

func AcyclicTraverseTerminals(tx graph.Transaction, plan TraversalPlan) (graph.NodeSet, error) {
	var (
		terminals     = graph.NewNodeSet()
		descentFilter = plan.DescentFilter
		visitedBitmap = roaring64.New()
	)

	// Wrap the existing descent filter to avoid revisiting nodes during traversal
	plan.DescentFilter = func(ctx *TraversalContext, segment *graph.PathSegment) bool {
		// If the descent filter is nil or if it accepts the given path check against the bitmap to see if we have
		// visited the path terminal node already
		if descentFilter == nil || descentFilter(ctx, segment) {
			terminalID := segment.Node.ID.Uint64()

			if !visitedBitmap.Contains(terminalID) {
				visitedBitmap.Add(terminalID)
				return true
			}
		}

		return false
	}

	// Add the path root to the bitmap; it shouldn't be included in the result set
	visitedBitmap.Add(plan.Root.ID.Uint64())

	return terminals, Traversal(tx, plan, func(ctx *TraversalContext, segment *graph.PathSegment) error {
		if ctx.LimitSkipTracker.ShouldCollect() {
			// Add the path terminal
			terminals.Add(segment.Node)
		}

		return nil
	})
}

func TraversePaths(tx graph.Transaction, plan TraversalPlan) (graph.PathSet, error) {
	var (
		paths         = graph.NewPathSet()
		descentFilter = plan.DescentFilter
	)

	// Wrap the existing descent filter to avoid revisiting nodes during traversal
	plan.DescentFilter = func(ctx *TraversalContext, segment *graph.PathSegment) bool {
		if descentFilter == nil || descentFilter(ctx, segment) {
			return !segment.IsCycle()
		}

		return false
	}

	return paths, Traversal(tx, plan, func(ctx *TraversalContext, segment *graph.PathSegment) error {
		if ctx.LimitSkipTracker.ShouldCollect() {
			paths.AddPath(segment.Path())
		}

		return nil
	})
}
