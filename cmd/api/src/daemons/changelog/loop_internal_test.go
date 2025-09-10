// Copyright 2025 Specter Ops, Inc.
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
package changelog

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/specterops/dawgs/util/channels"
	"github.com/stretchr/testify/require"
)

func TestLoop(t *testing.T) {
	t.Run("flushes nodes on batch size", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db := &mockFlusher{}
		loop := newLoop(ctx, db, 2, 5*time.Second)

		// Inject two changes. explicitly cast the NodeChange bc generics jank
		require.True(t, channels.Submit(ctx, loop.writerC, Change(NodeChange{NodeID: "1"})))
		require.True(t, channels.Submit(ctx, loop.writerC, Change(NodeChange{NodeID: "2"})))

		// Run one iteration
		go func() { _ = loop.start(ctx) }()

		require.Eventually(t, func() bool {
			return db.flushedLen() == 2
		}, 500*time.Millisecond, 5*time.Millisecond)
	})

	t.Run("flushes edges on batch size", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db := &mockFlusher{}
		loop := newLoop(ctx, db, 2, 5*time.Second)

		// queue up nodes < batchSize, edges >= batchSize
		require.True(t, channels.Submit(ctx, loop.writerC, Change(EdgeChange{})))
		require.True(t, channels.Submit(ctx, loop.writerC, Change(EdgeChange{})))

		// Run one iteration
		go func() { _ = loop.start(ctx) }()

		require.Eventually(t, func() bool {
			return db.flushedLen() == 2
		}, 500*time.Millisecond, 5*time.Millisecond)
	})

	t.Run("no flush happens before batch size", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db := &mockFlusher{}
		loop := newLoop(ctx, db, 3, 5*time.Second)

		// Inject two changes. explicitly cast the NodeChange bc generics jank
		require.True(t, channels.Submit(ctx, loop.writerC, Change(NodeChange{NodeID: "1"})))
		require.True(t, channels.Submit(ctx, loop.writerC, Change(NodeChange{NodeID: "2"})))

		// Run one iteration
		go func() { _ = loop.start(ctx) }()

		// nothing was flushed because buffer never reached batch_size
		require.Eventually(t, func() bool {
			return db.flushedLen() == 0
		}, 500*time.Millisecond, 5*time.Millisecond)
	})

	t.Run("timer triggers flush after inactivity", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db := &mockFlusher{}
		loop := newLoop(ctx, db, 3, 20*time.Millisecond)

		// Inject two changes. explicitly cast the NodeChange bc generics jank
		require.True(t, channels.Submit(ctx, loop.writerC, Change(NodeChange{NodeID: "1"})))

		go func() { _ = loop.start(ctx) }()

		require.Eventually(t, func() bool { // wait longer than flush interval
			return db.flushedLen() == 1
		}, 500*time.Millisecond, 25*time.Millisecond)

	})
}

func TestChangeBuffer(t *testing.T) {
	t.Run("buffer not full, no flush", func(t *testing.T) {
		var (
			batchSize   = 5
			ctx         = context.Background()
			mockFlusher = &mockFlusher{}
			buffer      = newChangeBuffer(mockFlusher, batchSize)
		)

		buffer.add(Change(NodeChange{}))

		require.NoError(t, buffer.tryFlush(ctx, batchSize))
		require.Equal(t, 0, mockFlusher.flushedLen())
	})

	t.Run("buffer full, flush", func(t *testing.T) {
		var (
			batchSize   = 3
			ctx         = context.Background()
			mockFlusher = &mockFlusher{}
			buffer      = newChangeBuffer(mockFlusher, batchSize)
		)
		// x3
		buffer.add(Change(NodeChange{}))
		buffer.add(Change(NodeChange{}))
		buffer.add(Change(NodeChange{}))

		require.NoError(t, buffer.tryFlush(ctx, batchSize))
		require.Equal(t, 3, mockFlusher.flushedLen())
	})

	t.Run("buffer above threshold, flush", func(t *testing.T) {
		var (
			batchSize   = 2
			ctx         = context.Background()
			mockFlusher = &mockFlusher{}
			buffer      = newChangeBuffer(mockFlusher, batchSize)
		)
		// x3
		buffer.add(Change(NodeChange{}))
		buffer.add(Change(NodeChange{}))
		buffer.add(Change(NodeChange{}))

		require.NoError(t, buffer.tryFlush(ctx, batchSize))
		require.Equal(t, 3, mockFlusher.flushedLen())
	})

	t.Run("force flush", func(t *testing.T) {
		var (
			batchSize   = 5
			ctx         = context.Background()
			mockFlusher = &mockFlusher{}
			buffer      = newChangeBuffer(mockFlusher, batchSize)
		)

		buffer.add(Change(NodeChange{}))

		require.NoError(t, buffer.tryFlush(ctx, 0))
		require.Equal(t, 1, mockFlusher.flushedLen())
	})

	t.Run("no flush on empty buffer ", func(t *testing.T) {
		var (
			batchSize   = 5
			ctx         = context.Background()
			mockFlusher = &mockFlusher{}
			buffer      = newChangeBuffer(mockFlusher, batchSize)
		)

		require.NoError(t, buffer.tryFlush(ctx, 0))
		require.Equal(t, 0, mockFlusher.flushedLen())
		require.Equal(t, 0, mockFlusher.called)
	})
}

type mockFlusher struct {
	mu             sync.Mutex
	flushedChanges []Change
	called         int
}

func (s *mockFlusher) flush(_ context.Context, changes []Change) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushedChanges = append(s.flushedChanges, changes...)
	s.called++
	return nil
}

// coderabbit suggestion. "directly reading flushedChanges without locking can race with the goroutine appending to it"
func (s *mockFlusher) flushedLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.flushedChanges)
}
