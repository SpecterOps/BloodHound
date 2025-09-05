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

		time.Sleep(50 * time.Millisecond)

		require.Len(t, db.flushedChanges, 2)
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
		time.Sleep(50 * time.Millisecond)

		require.Len(t, db.flushedChanges, 2)
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
		time.Sleep(50 * time.Millisecond)

		require.Len(t, db.flushedChanges, 0) // nothing was flushed because buffer never reached batch_size
	})

	t.Run("timer triggers flush after inactivity", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db := &mockFlusher{}
		loop := newLoop(ctx, db, 3, 5*time.Second)
		loop.flushInterval = 20 * time.Millisecond // best effort

		// Inject two changes. explicitly cast the NodeChange bc generics jank
		require.True(t, channels.Submit(ctx, loop.writerC, Change(NodeChange{NodeID: "1"})))

		go func() { _ = loop.start(ctx) }()
		time.Sleep(50 * time.Millisecond) // wait longer than flush interval

		require.Len(t, db.flushedChanges, 1)
	})
}

type mockFlusher struct {
	mu             sync.Mutex
	flushedChanges []Change
}

func (m *mockFlusher) flush(_ context.Context, changes []Change) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushedChanges = append(m.flushedChanges, changes...)
	return nil
}
