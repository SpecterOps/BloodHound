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
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

func newPropertiesWithFirstSeen() *graph.Properties {
	newProperties := graph.NewProperties()
	newProperties.Set(common.FirstSeen.String(), time.Now().UTC())

	return newProperties
}

// FilteredRelationshipSink is an asynchronous graph relationship writer that ensures only new relationships are
// inserted and removes unused ones. It uses a delta tracker to track changes between graphs and avoids reinserting
// existing edges. Any edge not visited during successful processing is treated as obsolete and will be deleted
// by Done. Failed processing must call Abort to preserve existing edges.
type FilteredRelationshipSink struct {
	operationName string
	db            graph.Database
	edgeTracker   *Tracker
	jobC          chan EnsureRelationshipJob
	stats         AtomicPostProcessingStats
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	submitLock    sync.RWMutex
	submitDone    chan struct{}
	shutdownOnce  sync.Once
	errorLock     sync.Mutex
	workerError   error
	doneError     error
}

// NewFilteredRelationshipSink creates a new filtered relationship sink initialized with a given database, delta tracker, and operation name.
func NewFilteredRelationshipSink(ctx context.Context, operationName string, db graph.Database, deltaSubgraph *Tracker) *FilteredRelationshipSink {
	var (
		workerContext, cancel = context.WithCancel(ctx)
		newSink               = &FilteredRelationshipSink{
			db:            db,
			edgeTracker:   deltaSubgraph,
			operationName: operationName,
			jobC:          make(chan EnsureRelationshipJob),
			stats:         NewAtomicPostProcessingStats(),
			ctx:           workerContext,
			cancel:        cancel,
			submitDone:    make(chan struct{}),
		}
	)

	newSink.start(workerContext)
	return newSink
}

// insertWorker processes incoming jobs by inserting them into the database using batch operations. It uses common properties
// (with first seen timestamp) and applies custom relationship properties if provided.
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
					s.recordError(fmt.Errorf("%s: creating relationship: %w", s.operationName, err))
					return err
				} else {
					s.stats.AddRelationshipsCreated(nextJob.Kind, 1)

					postOperationsVec.With(map[string]string{
						"kind":      nextJob.Kind.String(),
						"operation": "edge_insert",
					}).Add(1)
				}
			}
		}

		return ctx.Err()
	}); err != nil {
		s.recordError(fmt.Errorf("%s: inserting relationships: %w", s.operationName, err))
	}
}

// recordError cancels the pipeline so submissions cannot block after the writer stops.
func (s *FilteredRelationshipSink) recordError(err error) {
	if contextError := s.ctx.Err(); contextError != nil && errors.Is(err, contextError) {
		return
	}

	s.errorLock.Lock()
	if s.workerError == nil {
		s.workerError = err
		slog.ErrorContext(s.ctx, "FilteredRelationshipSink Error", attr.Error(err))
	}
	s.errorLock.Unlock()
	s.cancel()
}

// deltaFilterWorker filters out duplicate edges before they reach the insert worker. It checks whether
// an edge has already been tracked in the delta subgraph; if not, it forwards it to the insert channel.
func (s *FilteredRelationshipSink) deltaFilterWorker(ctx context.Context, filterC, insertC chan EnsureRelationshipJob) {
	for {
		nextJob, shouldContinue := channels.Receive(ctx, filterC)

		if !shouldContinue {
			break
		}

		if !s.edgeTracker.HasEdge(nextJob.FromID.Uint64(), nextJob.ToID.Uint64(), nextJob.Kind, nextJob.RelProperties) {
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

// worker coordinates filtering and insertion. Pruning is reserved for successful completion in Done.
func (s *FilteredRelationshipSink) worker(ctx context.Context) {
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
		s.insertWorker(ctx, newPropertiesWithFirstSeen(), insertC)
	}()

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
	s.submitLock.RLock()
	defer s.submitLock.RUnlock()

	if ctx.Err() != nil || s.ctx.Err() != nil {
		return false
	}
	select {
	case <-s.submitDone:
		return false
	default:
	}

	select {
	case <-ctx.Done():
		return false
	case <-s.ctx.Done():
		return false
	case <-s.submitDone:
		return false
	case s.jobC <- nextJob:
		return true
	}
}

// shutdown stops submissions and waits for workers. The first shutdown call determines whether to prune.
func (s *FilteredRelationshipSink) shutdown(prune bool) {
	s.shutdownOnce.Do(func() {
		defer s.cancel()
		if !prune {
			s.cancel()
		}

		close(s.submitDone)
		s.submitLock.Lock()
		close(s.jobC)
		s.submitLock.Unlock()
		s.wg.Wait()

		s.errorLock.Lock()
		s.doneError = s.workerError
		s.errorLock.Unlock()

		if s.doneError == nil {
			if err := s.ctx.Err(); err != nil {
				s.doneError = err
			} else if prune {
				s.doneError = s.deleteMissingEdges(s.ctx)
				if s.doneError != nil {
					slog.ErrorContext(s.ctx, "Error deleting stale relationships", attr.Error(s.doneError))
				}
			}
		}
	})
}

// Done drains submitted jobs and removes obsolete relationships only if processing succeeded.
// Repeated calls return the same result. Producers must complete successfully before calling Done.
func (s *FilteredRelationshipSink) Done() error {
	s.shutdown(true)
	return s.doneError
}

// Abort cancels processing and waits for workers without deleting obsolete relationships.
// It returns any writer error, excluding cancellation caused by shutdown. Repeated calls are safe.
// Calling Done after Abort returns a cancellation or worker error.
func (s *FilteredRelationshipSink) Abort() error {
	s.shutdown(false)
	s.errorLock.Lock()
	defer s.errorLock.Unlock()
	return s.workerError
}
