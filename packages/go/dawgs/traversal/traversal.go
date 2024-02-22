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

package traversal

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/graphcache"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/dawgs/util/atomics"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/log"
)

// Driver is a function that drives sending queries to the graph and retrieving vertexes and edges. Traversal
// drivers are expected to operate on a cactus tree representation of path space using the graph.PathSegment data
// structure. Path segments returned by a traversal driver are considered extensions of path space that require
// further expansion. If a traversal driver returns no descending path segments then the given segment may be
// considered terminal.
type Driver = func(ctx context.Context, tx graph.Transaction, segment *graph.PathSegment) ([]*graph.PathSegment, error)

type PatternMatchDelegate = func(terminal *graph.PathSegment) error

// PatternContinuation is an openCypher inspired fluent pattern for defining parallel chained expansions. After
// building the pattern the user may call the Do(...) function and pass it a delegate for handling paths that match
// the pattern.
//
// The return value of the Do(...) function may be passed directly to a Traversal via a Plan as the Plan.Driver field.
type PatternContinuation interface {
	Outbound(criteria ...graph.Criteria) PatternContinuation
	OutboundWithDepth(min, max int, criteria ...graph.Criteria) PatternContinuation
	Inbound(criteria ...graph.Criteria) PatternContinuation
	InboundWithDepth(min, max int, criteria ...graph.Criteria) PatternContinuation
	Do(delegate PatternMatchDelegate) Driver
}

// expansion is an internal representation of a path expansion step.
type expansion struct {
	criteria  []graph.Criteria
	direction graph.Direction
	minDepth  int
	maxDepth  int
}

func (s expansion) PrepareCriteria(segment *graph.PathSegment) (graph.Criteria, error) {
	var (
		criteria = s.criteria
	)

	switch s.direction {
	case graph.DirectionOutbound:
		criteria = append([]graph.Criteria{
			query.Equals(query.StartID(), segment.Node.ID),
		}, criteria...)

	case graph.DirectionInbound:
		criteria = append([]graph.Criteria{
			query.Equals(query.EndID(), segment.Node.ID),
		}, criteria...)

	default:
		return nil, fmt.Errorf("unsupported direction %v", s.direction)
	}

	return query.And(criteria...), nil
}

type patternTag struct {
	patternIdx int
	depth      int
}

func popSegmentPatternTag(segment *graph.PathSegment) *patternTag {
	var tag *patternTag

	if typedTag, typeOK := segment.Tag.(*patternTag); typeOK && typedTag != nil {
		tag = typedTag
		segment.Tag = nil
	} else {
		tag = &patternTag{
			patternIdx: 0,
			depth:      0,
		}
	}

	return tag
}

type pattern struct {
	expansions []expansion
	delegate   PatternMatchDelegate
}

// Do assigns the PatterMatchDelegate internally before returning a function pointer to the Driver receiver function.
func (s *pattern) Do(delegate PatternMatchDelegate) Driver {
	s.delegate = delegate
	return s.Driver
}

// OutboundWithDepth specifies the next outbound expansion step for this pattern with depth parameters.
func (s *pattern) OutboundWithDepth(min, max int, criteria ...graph.Criteria) PatternContinuation {
	if min < 0 {
		min = 1
		log.Warnf("Negative mindepth not allowed. Setting min depth for expansion to 1")
	}

	if max < 0 {
		max = 0
		log.Warnf("Negative maxdepth not allowed. Setting max depth for expansion to 0")
	}

	s.expansions = append(s.expansions, expansion{
		criteria:  criteria,
		direction: graph.DirectionOutbound,
		minDepth:  min,
		maxDepth:  max,
	})

	return s
}

// Outbound specifies the next outbound expansion step for this pattern. By default, this expansion will use a minimum
// depth of 1 to make the expansion required and a maximum depth of 0 to expand indefinitely.
func (s *pattern) Outbound(criteria ...graph.Criteria) PatternContinuation {
	return s.OutboundWithDepth(1, 0, criteria...)
}

// InboundWithDepth specifies the next inbound expansion step for this pattern with depth parameters.
func (s *pattern) InboundWithDepth(min, max int, criteria ...graph.Criteria) PatternContinuation {
	if min < 0 {
		min = 1
		log.Warnf("Negative mindepth not allowed. Setting min depth for expansion to 1")
	}

	if max < 0 {
		max = 0
		log.Warnf("Negative maxdepth not allowed. Setting max depth for expansion to 0")
	}

	s.expansions = append(s.expansions, expansion{
		criteria:  criteria,
		direction: graph.DirectionInbound,
		minDepth:  min,
		maxDepth:  max,
	})

	return s
}

// Inbound specifies the next inbound expansion step for this pattern. By default, this expansion will use a minimum
// depth of 1 to make the expansion required and a maximum depth of 0 to expand indefinitely.
func (s *pattern) Inbound(criteria ...graph.Criteria) PatternContinuation {
	return s.InboundWithDepth(1, 0, criteria...)
}

// NewPattern returns a new PatternContinuation for building a new pattern.
func NewPattern() PatternContinuation {
	return &pattern{}
}

func (s *pattern) Driver(ctx context.Context, tx graph.Transaction, segment *graph.PathSegment) ([]*graph.PathSegment, error) {
	var (
		nextSegments []*graph.PathSegment

		// The patternTag lives on the current terminal segment of each path. Once popped the pointer reference for
		// this segment is set to nil.
		tag              = popSegmentPatternTag(segment)
		currentExpansion = s.expansions[tag.patternIdx]

		// fetchFunc handles directional results from the graph database and is called twice to fetch segment
		// expansions.
		fetchFunc = func(cursor graph.Cursor[graph.DirectionalResult]) error {
			for next := range cursor.Chan() {
				nextSegment := segment.Descend(next.Node, next.Relationship)

				// Don't emit cycles out of the fetch
				if !nextSegment.IsCycle() {
					nextSegment.Tag = &patternTag{
						// Use the tag's patternIdx and depth since this is a continuation of the expansions
						patternIdx: tag.patternIdx,
						depth:      tag.depth + 1,
					}

					nextSegments = append(nextSegments, nextSegment)
				}
			}

			return cursor.Error()
		}
	)

	// The fetch direction is the reverse intent of the expansion direction
	if fetchDirection, err := currentExpansion.direction.Reverse(); err != nil {
		return nil, err
	} else {
		// If no max depth was set or if a max depth was set expand the current step further
		if currentExpansion.maxDepth == 0 || tag.depth < currentExpansion.maxDepth {
			// Perform the current expansion.
			if criteria, err := currentExpansion.PrepareCriteria(segment); err != nil {
				return nil, err
			} else if err := tx.Relationships().Filter(criteria).FetchDirection(fetchDirection, fetchFunc); err != nil {
				return nil, err
			}
		}

		// Check first if this current segment was fetched using the current expansion (i.e. non-optional)
		if tag.depth > 0 && currentExpansion.minDepth == 0 || tag.depth >= currentExpansion.minDepth {
			// No further expansions means this pattern segment is complete. Increment the pattern index to select the
			// next pattern expansion. Additionally, set the depth back to zero for the tag since we are leaving the
			// current expansion.
			tag.patternIdx++
			tag.depth = 0

			// Perform the next expansion if there is one.
			if tag.patternIdx < len(s.expansions) {
				nextExpansion := s.expansions[tag.patternIdx]

				// Expand the next segments
				if criteria, err := nextExpansion.PrepareCriteria(segment); err != nil {
					return nil, err
				} else if err := tx.Relationships().Filter(criteria).FetchDirection(fetchDirection, fetchFunc); err != nil {
					return nil, err
				}

				// If the next expansion is optional, make sure to preserve the current traversal branch
				if nextExpansion.minDepth == 0 {
					// Reattach the tag to the segment before adding it to the returned segments for the next expansion
					segment.Tag = tag
					nextSegments = append(nextSegments, segment)
				}
			} else if len(nextSegments) == 0 {
				// If there are no expanded segments and there are no remaining expansions, this is a terminal segment.
				// Hand it off to the delegate and handle any returned error.
				if err := s.delegate(segment); err != nil {
					return nil, err
				}
			}
		}

		// If the above condition does not match then this current expansion is non-terminal and non-continuable
	}

	// Return any collected segments
	return nextSegments, nil
}

type Plan struct {
	Root        *graph.Node
	RootSegment *graph.PathSegment
	Driver      Driver
}

type Traversal struct {
	db         graph.Database
	numWorkers int
}

func New(db graph.Database, numParallelWorkers int) Traversal {
	return Traversal{
		db:         db,
		numWorkers: numParallelWorkers,
	}
}

func (s Traversal) BreadthFirst(ctx context.Context, plan Plan) error {
	defer log.Measure(log.LevelDebug, "BreadthFirst - %d workers", s.numWorkers)()

	var (
		// workerWG keeps count of background workers launched in goroutines
		workerWG = &sync.WaitGroup{}

		// descentWG keeps count of in-flight traversal work. When this wait group reaches a count of 0 the traversal
		// is considered complete.
		completionC                    = make(chan struct{}, s.numWorkers*2)
		descentCount                   = &atomic.Int64{}
		errors                         = util.NewErrorCollector()
		traversalCtx, doneFunc         = context.WithCancel(ctx)
		segmentWriterC, segmentReaderC = channels.BufferedPipe[*graph.PathSegment](traversalCtx)
		pathTree                       graph.Tree
	)

	// Defer calling the cancellation function of the context to ensure that all workers join, no matter what
	defer doneFunc()

	// Close the writer channel to the buffered pipe
	defer close(segmentWriterC)

	if plan.Root != nil {
		pathTree = graph.NewTree(plan.Root)
	} else if plan.RootSegment != nil {
		pathTree = graph.Tree{
			Root: plan.RootSegment,
		}
	} else {
		return fmt.Errorf("no root specified")
	}

	// Launch the background traversal workers
	for workerID := 0; workerID < s.numWorkers; workerID++ {
		workerWG.Add(1)

		go func(workerID int) {
			defer workerWG.Done()

			if err := s.db.ReadTransaction(ctx, func(tx graph.Transaction) error {
				for {
					if nextDescent, ok := channels.Receive(traversalCtx, segmentReaderC); !ok {
						return nil
					} else if pathTreeSize := pathTree.SizeOf(); pathTreeSize < tx.TraversalMemoryLimit() {
						// Traverse the descending relationships of the current segment
						if descendingSegments, err := plan.Driver(traversalCtx, tx, nextDescent); err != nil {
							return err
						} else {
							for _, descendingSegment := range descendingSegments {
								// Add to the descent count before submitting to the channel
								descentCount.Add(1)
								channels.Submit(traversalCtx, segmentWriterC, descendingSegment)
							}
						}
					} else {
						// Did we encounter a memory limit?
						errors.Add(fmt.Errorf("%w - Limit: %.2f MB - Memory In-Use: %.2f MB", ops.ErrTraversalMemoryLimit, tx.TraversalMemoryLimit().Mebibytes(), pathTree.SizeOf().Mebibytes()))
					}

					// Mark descent for this segment as complete
					descentCount.Add(-1)

					if !channels.Submit(traversalCtx, completionC, struct{}{}) {
						return nil
					}
				}
			}); err != nil && err != graph.ErrContextTimedOut {
				// A worker encountered a fatal error, kill the traversal context
				doneFunc()

				errors.Add(fmt.Errorf("reader %d failed: %w", workerID, err))
			}
		}(workerID)
	}

	// Add to the descent wait group and then queue the root of the path tree for traversal
	descentCount.Add(1)
	segmentWriterC <- pathTree.Root

	for {
		if _, ok := channels.Receive(traversalCtx, completionC); !ok || descentCount.Load() == 0 {
			break
		}
	}

	// Actively cancel the traversal context to force any idle workers to join and exit
	doneFunc()

	// Wait for all workers to exit
	workerWG.Wait()

	return errors.Combined()
}

func newVisitorFilter(direction graph.Direction, userFilter graph.Criteria) func(segment *graph.PathSegment) graph.Criteria {
	return func(segment *graph.PathSegment) graph.Criteria {
		var filters []graph.Criteria

		if userFilter != nil {
			filters = append(filters, userFilter)
		}

		switch direction {
		case graph.DirectionOutbound:
			filters = append(filters, query.Equals(query.StartID(), segment.Node.ID))

		case graph.DirectionInbound:
			filters = append(filters, query.Equals(query.EndID(), segment.Node.ID))
		}

		return query.And(filters...)
	}
}

func shallowFetchRelationships(direction graph.Direction, segment *graph.PathSegment, graphQuery graph.RelationshipQuery) ([]*graph.Relationship, error) {
	var (
		relationships  []*graph.Relationship
		returnCriteria graph.Criteria
	)

	switch direction {
	case graph.DirectionOutbound:
		returnCriteria = query.Returning(
			query.EndID(),
			query.KindsOf(query.End()),
			query.RelationshipID(),
			query.KindsOf(query.Relationship()),
		)

	case graph.DirectionInbound:
		returnCriteria = query.Returning(
			query.StartID(),
			query.KindsOf(query.Start()),
			query.RelationshipID(),
			query.KindsOf(query.Relationship()),
		)

	default:
		return nil, fmt.Errorf("bi-directional or non-directed edges are not supported")
	}

	if err := graphQuery.Query(func(results graph.Result) error {
		defer results.Close()

		var (
			nodeID    graph.ID
			nodeKinds graph.Kinds
			edgeID    graph.ID
			edgeKind  graph.Kind
		)

		for results.Next() {
			if err := results.Scan(&nodeID, &nodeKinds, &edgeID, &edgeKind); err != nil {
				return err
			}

			switch direction {
			case graph.DirectionOutbound:
				relationships = append(relationships, graph.NewRelationship(edgeID, segment.Node.ID, nodeID, nil, edgeKind))

			case graph.DirectionInbound:
				relationships = append(relationships, graph.NewRelationship(edgeID, nodeID, segment.Node.ID, nil, edgeKind))
			}
		}

		return results.Error()
	}, returnCriteria); err != nil {
		return nil, err
	}

	return relationships, nil
}

// SegmentFilter is a function type that takes a given path segment and returns true if further descent into the path
// is allowed.
type SegmentFilter = func(next *graph.PathSegment) bool

// SegmentVisitor is a function that receives a path segment as part of certain traversal strategies.
type SegmentVisitor = func(next *graph.PathSegment)

// UniquePathSegmentFilter is a SegmentFilter constructor that will allow a traversal to all unique paths. This is done
// by tracking edge IDs traversed in a bitmap.
func UniquePathSegmentFilter(delegate SegmentFilter) SegmentFilter {
	traversalBitmap := cardinality.ThreadSafeDuplex(cardinality.NewBitmap32())

	return func(next *graph.PathSegment) bool {
		// Bail on cycles
		if next.IsCycle() {
			return false
		}

		// Return if we've seen this edge before
		if !traversalBitmap.CheckedAdd(next.Edge.ID.Uint32()) {
			return false
		}

		// Pass this segment to the delegate if we've never seen it before
		return delegate(next)
	}
}

// AcyclicNodeFilter is a SegmentFilter constructor that will allow traversal to a node only once. It will ignore all
// but the first inbound or outbound edge that traverses to it.
func AcyclicNodeFilter(filter SegmentFilter) SegmentFilter {
	return func(next *graph.PathSegment) bool {
		// Bail on counting ourselves
		if next.IsCycle() {
			return false
		}

		// Descend only if we've never seen this node before.
		return filter(next)
	}
}

// A SkipLimitFilter is a function that represents a collection and descent filter for PathSegments. This function must
// return two boolean values:
//
// The first boolean value in the return tuple communicates to the FilteredSkipLimit SegmentFilter if the given
// PathSegment is eligible for collection and therefore should be counted when considering the traversal's skip and
// limit parameters.
//
// The second boolean value in the return tuple communicates to the FilteredSkipLimit SegmentFilter if the given
// PathSegment is eligible for further descent. When this value is true the path will be expanded further during
// traversal.
type SkipLimitFilter = func(next *graph.PathSegment) (bool, bool)

// FilteredSkipLimit is a SegmentFilter constructor that allows a caller to inform the skip-limit algorithm when a
// result was collected and if the traversal should continue to descend further during traversal.
func FilteredSkipLimit(filter SkipLimitFilter, visitorFilter SegmentVisitor, skip, limit int) SegmentFilter {
	var (
		shouldCollect = atomics.NewCounter(uint64(skip))
		atLimit       = atomics.NewCounter(uint64(limit))
	)

	return func(next *graph.PathSegment) bool {
		canCollect, shouldDescend := filter(next)

		if canCollect {
			// Check to see if this result should be skipped
			if skip == 0 || shouldCollect() {
				// If we should collect this result, check to see if we're already at a limit for the number of results
				if limit > 0 && atLimit() {
					log.Debugf("At collection limit, rejecting path: %s", graph.FormatPathSegment(next))
					return false
				}

				log.Debugf("Collected path: %s", graph.FormatPathSegment(next))
				visitorFilter(next)
			} else {
				log.Debugf("Skipping path visit: %s", graph.FormatPathSegment(next))
			}
		}

		if shouldDescend {
			log.Debugf("Descending into path: %s", graph.FormatPathSegment(next))
		} else {
			log.Debugf("Rejecting further descent into path: %s", graph.FormatPathSegment(next))
		}

		return shouldDescend
	}
}

// LightweightDriver is a Driver constructor that fetches only IDs and Kind information from vertexes and
// edges stored in the database. This cuts down on network transit and is appropriate for traversals that may involve
// a large number of or all vertexes within a target graph.
func LightweightDriver(direction graph.Direction, cache graphcache.Cache, criteria graph.Criteria, filter SegmentFilter, terminalVisitors ...SegmentVisitor) Driver {
	filterProvider := newVisitorFilter(direction, criteria)

	return func(ctx context.Context, tx graph.Transaction, nextSegment *graph.PathSegment) ([]*graph.PathSegment, error) {
		var (
			nextSegments []*graph.PathSegment
			nextQuery    = tx.Relationships().Filter(filterProvider(nextSegment)).OrderBy(
				// Order by relationship ID so that skip and limit behave somewhat predictably - cost of this is pretty
				// small even for large result sets
				query.Order(query.Identity(query.Relationship()), query.Ascending()),
			)
		)

		if relationships, err := shallowFetchRelationships(direction, nextSegment, nextQuery); err != nil {
			return nil, err
		} else {
			// Reconcile the start and end nodes of the fetched relationships with the graph cache
			nodesToFetch := cardinality.NewBitmap32()

			for _, nextRelationship := range relationships {
				if nextID, err := direction.PickReverse(nextRelationship); err != nil {
					return nil, err
				} else {
					nodesToFetch.Add(nextID.Uint32())
				}
			}

			// Shallow fetching the nodes achieves the same result as shallowFetchRelationships(...) but with the added
			// benefit of interacting with the graph cache. Any nodes not already in the cache are fetched just-in-time
			// from the database and stored back in the cache for later.
			if cachedNodes, err := graphcache.ShallowFetchNodesByID(tx, cache, cardinality.DuplexToGraphIDs(nodesToFetch)); err != nil {
				return nil, err
			} else {
				cachedNodeSet := graph.NewNodeSet(cachedNodes...)

				for _, nextRelationship := range relationships {
					if targetID, err := direction.PickReverse(nextRelationship); err != nil {
						return nil, err
					} else {
						nextSegment := nextSegment.Descend(cachedNodeSet[targetID], nextRelationship)

						if filter(nextSegment) {
							nextSegments = append(nextSegments, nextSegment)
						}
					}
				}
			}
		}

		// If this segment has no further descent paths, render it as a path if we have a path visitor specified
		if len(nextSegments) == 0 && len(terminalVisitors) > 0 {
			for _, terminalVisitor := range terminalVisitors {
				terminalVisitor(nextSegment)
			}
		}

		return nextSegments, nil
	}
}
