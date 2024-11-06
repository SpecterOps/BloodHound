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
	"context"
	"errors"
	"sync"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/dawgs/util/channels"
)

var (
	ErrOperationDone = errors.New("parallel operation context has expired")
)

type ReaderFunc[T any] func(ctx context.Context, tx graph.Transaction, outC chan<- T) error
type WriterFunc[T any] func(ctx context.Context, batch graph.Batch, inC <-chan T) error

type readerJob[T any] struct {
	logic ReaderFunc[T]
	outC  chan<- T
}

type writerJob[T any] struct {
	logic WriterFunc[T]
	inC   <-chan T
}

type worker struct {
	ctx         context.Context
	ctxDoneFunc func()
	error       error
}

type reader[T any] struct {
	worker
	jobC <-chan readerJob[T]
}

func (s *reader[T]) Start(waitGroup *sync.WaitGroup, db graph.Database) {
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		if err := db.ReadTransaction(s.ctx, func(tx graph.Transaction) error {
			for {
				select {
				case nextJob, hasJob := <-s.jobC:
					if !hasJob {
						return nil
					}

					if err := nextJob.logic(s.ctx, tx, nextJob.outC); err != nil {
						return err
					}

				case <-s.ctx.Done():
					return nil
				}
			}
		}); err != nil {
			// If we can no longer accept jobs, this could produce a starvation effect on the writer side that must be
			// mitigated by closing the worker context after collecting the error
			s.error = err
			s.ctxDoneFunc()
		}
	}()
}

type writer[T any] struct {
	worker
	jobC <-chan writerJob[T]
}

func (s *writer[T]) Start(waitGroup *sync.WaitGroup, db graph.Database) {
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		if err := db.BatchOperation(s.ctx, func(batch graph.Batch) error {
			for {
				select {
				case nextJob, hasJob := <-s.jobC:
					if !hasJob {
						return nil
					}

					if err := nextJob.logic(s.ctx, batch, nextJob.inC); err != nil {
						return err
					}

				case <-s.ctx.Done():
					return nil
				}
			}
		}); err != nil {
			// If we can no longer accept jobs, this could produce a starvation effect on the reader side that must be
			// mitigated by closing the worker context after collecting the error
			s.error = err
			s.ctxDoneFunc()
		}
	}()
}

type OperationContext struct {
	Parent            context.Context
	DB                graph.Database
	NumReaders        int
	NumWriters        int
	ReaderJobCapacity int
	WriterJobCapacity int
}

func (s OperationContext) GetParent() context.Context {
	if s.Parent == nil {
		return context.Background()
	}

	return s.Parent
}

func (s OperationContext) GetReaderJobCapacity() int {
	if s.ReaderJobCapacity == 0 {
		return s.NumReaders * 2
	}

	return s.ReaderJobCapacity
}

func (s OperationContext) GetWriterJobCapacity() int {
	if s.WriterJobCapacity == 0 {
		return s.NumWriters * 2
	}

	return s.WriterJobCapacity
}

type Operation[T any] struct {
	opCtx              OperationContext
	readers            []*reader[T]
	readerJobC         chan readerJob[T]
	writers            []*writer[T]
	writerJobC         chan writerJob[T]
	readerWriterValueC chan T
	workerCtx          context.Context
	workerCtxDoneFunc  func()
	readerWG           *sync.WaitGroup
	writerWG           *sync.WaitGroup
}

func NewOperation[T any](opCtx OperationContext) *Operation[T] {
	return &Operation[T]{
		opCtx:    opCtx,
		readerWG: &sync.WaitGroup{},
		writerWG: &sync.WaitGroup{},
	}
}

func StartNewOperation[T any](opCtx OperationContext) *Operation[T] {
	newOp := NewOperation[T](opCtx)
	newOp.Start()

	return newOp
}

func (s *Operation[T]) Start() {
	// Bind a new context for the workers that we can cancel, make a new chan for the worker value type and
	// then launch the workers
	s.workerCtx, s.workerCtxDoneFunc = context.WithCancel(s.opCtx.GetParent())
	s.readerWriterValueC = make(chan T)

	// Create new work channels for the workers
	s.readerJobC = make(chan readerJob[T], s.opCtx.GetReaderJobCapacity())
	s.writerJobC = make(chan writerJob[T], s.opCtx.GetWriterJobCapacity())

	// Start up the workers
	for i := 0; i < s.opCtx.NumReaders; i++ {
		reader := &reader[T]{
			worker: worker{
				ctx:         s.workerCtx,
				ctxDoneFunc: s.workerCtxDoneFunc,
			},
			jobC: s.readerJobC,
		}

		s.readers = append(s.readers, reader)
		reader.Start(s.readerWG, s.opCtx.DB)
	}

	for i := 0; i < s.opCtx.NumWriters; i++ {
		writer := &writer[T]{
			worker: worker{
				ctx:         s.workerCtx,
				ctxDoneFunc: s.workerCtxDoneFunc,
			},
			jobC: s.writerJobC,
		}

		s.writers = append(s.writers, writer)
		writer.Start(s.writerWG, s.opCtx.DB)
	}
}

func (s *Operation[T]) Done() error {
	// Ensure the context is closed
	defer s.workerCtxDoneFunc()

	// Close the reader work channel and then wait for all readers to exit
	close(s.readerJobC)
	s.readerWG.Wait()

	// Close the writer work channel and the reader writer value channel next to signal to any lingering workers that
	// the operation is finished
	close(s.writerJobC)
	close(s.readerWriterValueC)

	// Wait for workers to exit
	s.writerWG.Wait()

	// Clear the old worker context
	s.workerCtx = nil
	s.workerCtxDoneFunc = nil
	s.readerWriterValueC = nil
	s.readerJobC = nil
	s.writerJobC = nil

	// TODO this should be split up from the rest of the function
	// Collect any errors from the workers
	errorCollector := util.NewErrorCollector()

	for _, reader := range s.readers {
		if reader.error != nil {
			errorCollector.Add(reader.error)
		}
	}

	for _, writer := range s.writers {
		if writer.error != nil {
			errorCollector.Add(writer.error)
		}
	}

	// Clear the slices of reader and worker references
	s.readers = nil
	s.writers = nil

	return errorCollector.Combined()
}

func (s *Operation[T]) SubmitReader(reader ReaderFunc[T]) error {
	job := readerJob[T]{
		logic: reader,
		outC:  s.readerWriterValueC,
	}

	select {
	case <-s.workerCtx.Done():
		return ErrOperationDone
	case s.readerJobC <- job:
		return nil
	}
}

func (s *Operation[T]) SubmitWriter(writer WriterFunc[T]) error {
	job := writerJob[T]{
		logic: writer,
		inC:   s.readerWriterValueC,
	}

	select {
	case <-s.workerCtx.Done():
		return ErrOperationDone
	case s.writerJobC <- job:
		return nil
	}
}

func parallelNodeQuery(ctx context.Context, db graph.Database, numWorkers int, criteria graph.Criteria, largestNodeID graph.ID, queryDelegate func(query graph.NodeQuery) error) error {
	const stride = 20_000

	var (
		rangeC   = make(chan graph.ID)
		errorC   = make(chan error)
		workerWG = &sync.WaitGroup{}
		errorWG  = &sync.WaitGroup{}
		errs     []error
	)

	// Query workers
	for workerID := 0; workerID < numWorkers; workerID++ {
		workerWG.Add(1)

		go func() {
			defer workerWG.Done()

			if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
				// Create a slice of criteria to join the node ID range constraints to any passed user criteria
				var criteriaSlice []graph.Criteria

				if criteria != nil {
					criteriaSlice = append(criteriaSlice, criteria)
				}

				// Select the next node ID range floor while honoring context cancellation and channel closure
				nextRangeFloor, channelOpen := channels.Receive(ctx, rangeC)

				for channelOpen {
					nextQuery := tx.Nodes().Filter(query.And(
						append(criteriaSlice,
							query.GreaterThanOrEquals(query.NodeID(), nextRangeFloor),
							query.LessThan(query.NodeID(), nextRangeFloor+stride),
						)...,
					))

					if err := queryDelegate(nextQuery); err != nil {
						return err
					}

					nextRangeFloor, channelOpen = channels.Receive(ctx, rangeC)
				}

				return nil
			}); err != nil {
				channels.Submit(ctx, errorC, err)
			}
		}()
	}

	// Merge goroutine for collected errors
	errorWG.Add(1)

	go func() {
		defer errorWG.Done()

		for {
			select {
			case <-ctx.Done():
				// Bail if the context is canceled
				return

			case nextErr, channelOpen := <-errorC:
				if !channelOpen {
					// Channel closure indicates completion of work and join of the parallel workers
					return
				}

				errs = append(errs, nextErr)
			}
		}
	}()

	// Iterate through node ID ranges up to the maximum ID by the stride constant
	for nextRangeFloor := graph.ID(0); nextRangeFloor <= largestNodeID; nextRangeFloor += stride {
		channels.Submit(ctx, rangeC, nextRangeFloor)
	}

	// Stop the fetch workers
	close(rangeC)
	workerWG.Wait()

	// Wait for the merge routine to join to ensure that both the nodes instance and the errs instance contain
	// everything to be collected from the parallel workers
	close(errorC)
	errorWG.Wait()

	// Return the joined errors lastly
	return errors.Join(errs...)
}

// ParallelNodeQuery will first look up the largest node database identifier. The function will then spin up to
// numWorkers parallel read transactions. Each transaction will apply the user passed criteria to this function to a
// range of node database identifiers to avoid parallel worker collisions.
func ParallelNodeQuery(ctx context.Context, db graph.Database, criteria graph.Criteria, numWorkers int, queryDelegate func(query graph.NodeQuery) error) error {
	if largestNodeID, err := FetchLargestNodeID(ctx, db); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}

		return err
	} else {
		return parallelNodeQuery(ctx, db, numWorkers, criteria, largestNodeID, queryDelegate)
	}
}

// ParallelNodeQueryBuilder is a type that can be used to construct a dawgs node query that is run in parallel. The
// Stream(...) function commits the query to as many workers as specified and then submits all results to a single
// channel that can be safely ranged over. Context cancellation is taken into consideration and the channel will close
// upon exit of the parallel query's context.
type ParallelNodeQueryBuilder[T any] struct {
	db            graph.Database
	wg            *sync.WaitGroup
	err           error
	criteria      graph.Criteria
	queryDelegate func(query graph.NodeQuery, outC chan<- T) error
}

func NewParallelNodeQuery[T any](db graph.Database) *ParallelNodeQueryBuilder[T] {
	return &ParallelNodeQueryBuilder[T]{
		db: db,
		wg: &sync.WaitGroup{},
	}
}

// UsingQuery specifies the execution and marshalling of results from the database. All results written to the outC
// channel parameter will be received by the Stream(...) caller.
func (s *ParallelNodeQueryBuilder[T]) UsingQuery(queryDelegate func(query graph.NodeQuery, outC chan<- T) error) *ParallelNodeQueryBuilder[T] {
	s.queryDelegate = queryDelegate
	return s
}

// WithCriteria specifies the criteria being used to filter this query.
func (s *ParallelNodeQueryBuilder[T]) WithCriteria(criteria graph.Criteria) *ParallelNodeQueryBuilder[T] {
	s.criteria = criteria
	return s
}

// Error returns any error that may have occurred during the parallel operation. This error may be a joined error.
func (s *ParallelNodeQueryBuilder[T]) Error() error {
	return s.err
}

// Join blocks the current thread and waits for the parallel node query to complete.
func (s *ParallelNodeQueryBuilder[T]) Join() {
	s.wg.Wait()
}

// Stream commits the query to the database in parallel and writes all results to the returned output channel.
func (s *ParallelNodeQueryBuilder[T]) Stream(ctx context.Context, numWorkers int) <-chan T {
	mergeC := make(chan T)

	s.wg.Add(1)

	go func() {
		defer close(mergeC)
		defer s.wg.Done()

		if err := ParallelNodeQuery(ctx, s.db, s.criteria, numWorkers, func(query graph.NodeQuery) error {
			return s.queryDelegate(query, mergeC)
		}); err != nil {
			s.err = err
		}
	}()

	return mergeC
}
