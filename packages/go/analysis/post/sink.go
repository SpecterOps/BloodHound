// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package post

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/specterops/bloodhound/packages/go/analysis/delta"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/metrics"
	"github.com/specterops/bloodhound/packages/go/trace"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

var (
	postOperationsVec = metrics.CounterVec("post_processing_ops", "analysis", map[string]string{}, []string{
		"kind",
		"operation",
	})
)

func newPropertiesWithLastSeen() *graph.Properties {
	newProperties := graph.NewProperties()
	newProperties.Set(common.LastSeen.String(), time.Now().UTC())

	return newProperties
}

// FilteredRelationshipSink is an asynchronous graph relationship writer that ensures only new relationships are
// inserted and removes unused ones. It uses a delta tracker to track changes between graphs and avoids reinserting
// existing edges. Any edge not visited during processing is treated as obsolete and will be deleted after the
// operation completes.
type FilteredRelationshipSink struct {
	operationName string
	db            graph.Database
	edgeTracker   *delta.Tracker
	jobC          chan EnsureRelationshipJob
	stats         AtomicPostProcessingStats
	wg            sync.WaitGroup
}

// NewFilteredRelationshipSink creates a new filtered relationship sink initialized with a given database, delta tracker, and operation name.
func NewFilteredRelationshipSink(ctx context.Context, operationName string, db graph.Database, deltaSubgraph *delta.Tracker) *FilteredRelationshipSink {
	newSink := &FilteredRelationshipSink{
		db:            db,
		edgeTracker:   deltaSubgraph,
		operationName: operationName,
		jobC:          make(chan EnsureRelationshipJob),
		stats:         NewAtomicPostProcessingStats(),
	}

	newSink.start(ctx)
	return newSink
}

// insertWorker processes incoming jobs by inserting them into the database using batch operations. It uses common properties
// (with last seen timestamp) and applies custom relationship properties if provided.
func (s *FilteredRelationshipSink) insertWorker(ctx context.Context, commonProps *graph.Properties, insertC chan EnsureRelationshipJob) {
	if err := s.db.BatchOperation(ctx, func(batch graph.Batch) error {
		for {
			if nextJob, shouldContinue := channels.Receive(ctx, insertC); !shouldContinue {
				break
			} else {
				relProps := commonProps

				if len(nextJob.RelProperties) > 0 {
					relProps = commonProps.Clone()

					for key, val := range nextJob.RelProperties {
						relProps.Set(key, val)
					}
				}

				if err := batch.CreateRelationshipByIDs(nextJob.FromID, nextJob.ToID, nextJob.Kind, relProps); err != nil {
					slog.Error("Create Relationship Error", slog.String("err", err.Error()))
				}

				s.stats.AddRelationshipsCreated(nextJob.Kind, 1)

				postOperationsVec.With(map[string]string{
					"kind":      nextJob.Kind.String(),
					"operation": "edge_insert",
				}).Add(1)
			}
		}

		return nil
	}); err != nil {
		slog.Error("FilteredRelationshipSink Error", attr.Error(err))
	}
}

// deltaFilterWorker filters out duplicate edges before they reach the insert worker. It checks whether
// an edge has already been tracked in the delta subgraph; if not, it forwards it to the insert channel.
func (s *FilteredRelationshipSink) deltaFilterWorker(ctx context.Context, filterC, insertC chan EnsureRelationshipJob) {
	for {
		nextJob, shouldContinue := channels.Receive(ctx, filterC)

		if !shouldContinue {
			break
		}

		if !s.edgeTracker.HasEdge(nextJob.FromID.Uint64(), nextJob.ToID.Uint64(), nextJob.Kind) {
			if !channels.Submit(ctx, insertC, nextJob) {
				break
			}
		} else {
			postOperationsVec.With(map[string]string{
				"kind":      nextJob.Kind.String(),
				"operation": "filtered",
			}).Add(1)
		}
	}
}

// deleteMissingEdges removes any lingering edges that were not part of the current operation. This ensures
// that only valid relationships remain after the sink completes its work.
func (s *FilteredRelationshipSink) deleteMissingEdges(ctx context.Context) error {
	deletedEdges := s.edgeTracker.Deleted()

	defer trace.Method(ctx, "FilteredRelationshipSink", "deleteMissingEdges", slog.Int("num_edges", len(deletedEdges)))()

	if err := s.db.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, deletedEdge := range deletedEdges {
			if err := batch.DeleteRelationship(graph.ID(deletedEdge)); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	postOperationsVec.With(map[string]string{
		"kind":      "all",
		"operation": "edge_delete",
	}).Add(float64(len(deletedEdges)))

	return nil
}

// worker is the main goroutine responsible for managing the entire lifecycle of the sink. It
// coordinates between filtering, insertion, and deletion phases and handles shutdown gracefully.
func (s *FilteredRelationshipSink) worker(ctx context.Context) error {
	defer trace.Method(ctx, "FilteredRelationshipSink", "worker", slog.String("operation", s.operationName))()
	defer s.wg.Done()

	var (
		filterC  = make(chan EnsureRelationshipJob)
		insertC  = make(chan EnsureRelationshipJob)
		filterWG sync.WaitGroup
		insertWG sync.WaitGroup
	)

	insertWG.Add(1)

	go func() {
		defer insertWG.Done()
		s.insertWorker(ctx, newPropertiesWithLastSeen(), insertC)
	}()

	// FIXME: Really, really need a better CPU heuristic or config value
	for workerID := 0; workerID < runtime.NumCPU()/2+1; workerID += 1 {
		filterWG.Add(1)

		go func(workerID int) {
			defer filterWG.Done()
			s.deltaFilterWorker(ctx, filterC, insertC)
		}(workerID)
	}

	for {
		if nextJob, shouldContinue := channels.Receive(ctx, s.jobC); !shouldContinue {
			break
		} else if !channels.Submit(ctx, filterC, nextJob) {
			break
		}
	}

	close(filterC)
	filterWG.Wait()

	close(insertC)
	insertWG.Wait()

	// Remove any lingering edges after the operation completes
	return s.deleteMissingEdges(ctx)
}

// Stats returns a pointer to the atomic statistics structure tracking processed relationships.
func (s *FilteredRelationshipSink) Stats() *AtomicPostProcessingStats {
	return &s.stats
}

// start begins execution of the sink's main worker loop.
func (s *FilteredRelationshipSink) start(ctx context.Context) {
	s.wg.Add(1)
	go s.worker(ctx)
}

// Submit submits a new job to be processed by the sink.
func (s *FilteredRelationshipSink) Submit(ctx context.Context, nextJob EnsureRelationshipJob) bool {
	return channels.Submit(ctx, s.jobC, nextJob)
}

// Done signals the end of processing and waits for all workers to complete.
func (s *FilteredRelationshipSink) Done() {
	close(s.jobC)
	s.wg.Wait()
}
