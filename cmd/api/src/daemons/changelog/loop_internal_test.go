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

type mockFlusher struct {
	mu             sync.Mutex
	flushedChanges []Change
}

func (s *mockFlusher) flush(_ context.Context, changes []Change) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushedChanges = append(s.flushedChanges, changes...)
	return nil
}

// coderabbit suggestion. "directly reading flushedChanges without locking can race with the goroutine appending to it"
func (s *mockFlusher) flushedLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.flushedChanges)
}
