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

package ops_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/ops"
)

func TestOperation_DriverFailures(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = graph_mocks.NewMockDatabase(mockCtrl)
		mockTx    = graph_mocks.NewMockTransaction(mockCtrl)
		mockBatch = graph_mocks.NewMockBatch(mockCtrl)
		testCtx   = context.Background()
		operation = ops.NewOperation[int](ops.OperationContext{
			Parent:            testCtx,
			DB:                mockDB,
			NumReaders:        1,
			NumWriters:        1,
			ReaderJobCapacity: 1,
			WriterJobCapacity: 1,
		})
	)

	mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(tx graph.Transaction) error, options ...graph.TransactionOption) error {
		return logic(mockTx)
	}).AnyTimes()

	mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(batch graph.Batch) error) error {
		return errors.New("error")
	})

	mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(batch graph.Batch) error) error {
		return logic(mockBatch)
	})

	operation.Start()

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- int) error {
		for itr := 0; itr < 100_000; itr++ {
			select {
			case outC <- itr:
			case <-ctx.Done():
				return nil
			}
		}

		return nil
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan int) error {
		return nil
	})

	require.NotNil(t, operation.Done())

	operation.Start()

	operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- int) error {
		for itr := 0; itr < 100_000; itr++ {
			select {
			case outC <- itr:
			case <-ctx.Done():
				return nil
			}
		}

		return nil
	})

	operation.SubmitWriter(func(ctx context.Context, batch graph.Batch, inC <-chan int) error {
		return errors.New("err")
	})

	require.NotNil(t, operation.Done())
}

func TestNewOperation(t *testing.T) {
	var (
		mockCtrl     = gomock.NewController(t)
		mockDB       = graph_mocks.NewMockDatabase(mockCtrl)
		mockTx       = graph_mocks.NewMockTransaction(mockCtrl)
		mockBatch    = graph_mocks.NewMockBatch(mockCtrl)
		testCtx      = context.Background()
		operationCtx = ops.OperationContext{
			Parent:            testCtx,
			DB:                mockDB,
			NumReaders:        2,
			NumWriters:        2,
			ReaderJobCapacity: 2,
			WriterJobCapacity: 2,
		}

		operation  = ops.NewOperation[int](operationCtx)
		writerFunc = func(ctx context.Context, batch graph.Batch, inC <-chan int) error {
			for done := false; !done; {
				select {
				case _, ok := <-inC:
					done = !ok
				case <-ctx.Done():
					done = true
				}
			}

			return nil
		}

		readerFunc = func(ctx context.Context, tx graph.Transaction, outC chan<- int) error {
			for itr := 0; itr < 100_000; itr++ {
				select {
				case outC <- itr:
				case <-ctx.Done():
					return nil
				}
			}

			return nil
		}
	)

	mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(tx graph.Transaction) error, options ...graph.TransactionOption) error {
		return logic(mockTx)
	})

	mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(tx graph.Transaction) error, options ...graph.TransactionOption) error {
		if err := logic(mockTx); err != nil {
			return err
		}

		return graph.ErrNoResultsFound
	}).AnyTimes()

	mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(batch graph.Batch) error) error {
		return logic(mockBatch)
	})

	mockDB.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(batch graph.Batch) error) error {
		if err := logic(mockBatch); err != nil {
			return err
		}

		return graph.ErrNoResultsFound
	}).AnyTimes()

	operation.Start()

	require.Nil(t, operation.SubmitWriter(writerFunc))
	require.Nil(t, operation.SubmitWriter(writerFunc))

	for itr := 0; itr < 10; itr++ {
		require.Nil(t, operation.SubmitReader(readerFunc))
	}

	require.NotNil(t, operation.Done())

	//
	var (
		timeoutCtx, timeoutCtxDoneFunc = context.WithTimeout(testCtx, time.Millisecond*100)
		then                           = time.Now()
		successfullyExited             = false
	)

	defer timeoutCtxDoneFunc()

	operationCtx.Parent = timeoutCtx
	operation = ops.StartNewOperation[int](operationCtx)

	for time.Since(then) < time.Second*2 {
		if err := operation.SubmitReader(readerFunc); err != nil {
			if err == ops.ErrOperationDone {
				successfullyExited = true
				break
			} else if err != graph.ErrNoResultsFound {
				t.Fatalf("Unexpected reader error: %v", err)
			}
		}
	}

	require.True(t, successfullyExited)
	require.NotNil(t, operation.Done())
}
