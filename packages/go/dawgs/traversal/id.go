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

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/dawgs/util/channels"
)

// IDDriver is a function that drives sending queries to the graph and retrieving vertexes and edges. Traversal
// drivers are expected to operate on a cactus tree representation of path space using the graph.PathSegment data
// structure. Path segments returned by a traversal drivers are considered extensions of path space that require
// further expansion. If a traversal drivers returns no descending path segments then the given segment may be
// considered terminal.
type IDDriver func(ctx context.Context, tx graph.Transaction, segment *graph.IDSegment) ([]*graph.IDSegment, error)

type IDPlan struct {
	Root     graph.ID
	Delegate IDDriver
}

type IDTraversal struct {
	db         graph.Database
	numWorkers int
}

func NewIDTraversal(db graph.Database, numParallelWorkers int) IDTraversal {
	return IDTraversal{
		db:         db,
		numWorkers: numParallelWorkers,
	}
}

func (s IDTraversal) BreadthFirst(ctx context.Context, plan IDPlan) error {
	var (
		// workerWG keeps count of background workers launched in goroutines
		workerWG = &sync.WaitGroup{}

		// descentWG keeps count of in-flight traversal work. When this wait group reaches a count of 0 the traversal
		// is considered complete.
		completionC  = make(chan struct{}, s.numWorkers*2)
		descentCount = &atomic.Int64{}

		errors                         = util.NewErrorCollector()
		pathTree                       = graph.NewRootIDSegment(plan.Root)
		traversalCtx, doneFunc         = context.WithCancel(ctx)
		segmentWriterC, segmentReaderC = channels.BufferedPipe[*graph.IDSegment](traversalCtx)
	)

	// Defer calling the cancellation function of the context to ensure that all workers join, no matter what
	defer doneFunc()

	// Close the writer channel to the buffered pipe
	defer close(segmentWriterC)

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
						if descendingSegments, err := plan.Delegate(traversalCtx, tx, nextDescent); err != nil {
							return err
						} else if len(descendingSegments) > 0 {
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

					if !channels.Submit(traversalCtx, completionC, struct{}{}) {
						return nil
					}

					// Mark descent for this segment as complete
					descentCount.Add(-1)
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
	segmentWriterC <- pathTree

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
