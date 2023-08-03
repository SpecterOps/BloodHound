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

package channels_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/util/channels"
)

const (
	numValuesToSend = 1000
)

func TestBufferedPipe_ContextCancel(t *testing.T) {
	var (
		ctx, done        = context.WithCancel(context.Background())
		writerC, readerC = channels.BufferedPipe[uint32](ctx)
	)

	// Spin up the writer as part of the background workers and do not close writerC
	workerWG := &sync.WaitGroup{}
	workerWG.Add(1)

	go func() {
		defer workerWG.Done()

		for {
			if !channels.Submit(ctx, writerC, 1) {
				break
			}
		}
	}()

	for workerID := 0; workerID < 10; workerID++ {
		workerWG.Add(1)

		go func() {
			defer workerWG.Done()

			for {
				if _, ok := channels.Receive(ctx, readerC); !ok {
					break
				}
			}
		}()
	}

	// Cancel the context and wait for workers to join
	done()
	workerWG.Wait()
}

func TestBufferedPipe_DumpOnContextCancel(t *testing.T) {
	var (
		ctx, done  = context.WithTimeout(context.Background(), time.Second*5)
		writerC, _ = channels.BufferedPipe[uint32](ctx)
	)

	// Submit the values first to demonstrate buffering
	for i := uint32(0); i < numValuesToSend; i++ {
		require.True(t, channels.Submit(ctx, writerC, i))
	}

	// Close the writer after submitting all 100 values
	close(writerC)

	// Canceling the context should dump all background channel workers regardless of what's in the buffer
	done()
}

func TestBufferedPipe_HappyPath(t *testing.T) {
	var (
		ctx, done        = context.WithTimeout(context.Background(), time.Second*5)
		writerC, readerC = channels.BufferedPipe[uint32](ctx)
	)

	// Ensure that the context done function is always called
	defer done()

	// Submit the values first to demonstrate buffering
	for i := uint32(0); i < numValuesToSend; i++ {
		require.True(t, channels.Submit(ctx, writerC, i))
	}

	// Close the writer after submitting all 100 values
	close(writerC)

	var (
		workerWG = &sync.WaitGroup{}
		seen     = cardinality.ThreadSafeDuplex(cardinality.NewBitmap32())
	)

	for workerID := 0; workerID < 10; workerID++ {
		workerWG.Add(1)

		go func() {
			defer workerWG.Done()

			for {
				if value, ok := channels.Receive(ctx, readerC); !ok {
					break
				} else {
					seen.Add(value)
				}
			}
		}()
	}

	workerWG.Wait()

	for i := uint32(0); i < numValuesToSend; i++ {
		require.True(t, seen.Contains(i))
	}
}
