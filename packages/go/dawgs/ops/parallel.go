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
	"github.com/specterops/bloodhound/dawgs/util"
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
