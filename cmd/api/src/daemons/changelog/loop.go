package changelog

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

// loop coordinates the ingestion of deduplicated graph changes.
// It buffers Change values in memory, and flushes them
// to the backing flusher either when:
//   - the buffer reaches the configured batch size, or
//   - no new changes arrive within the flush interval.
type loop struct {
	readerC       <-chan Change
	writerC       chan<- Change
	flushInterval time.Duration
	batchSize     int

	changeBuffer *changeBuffer
}

func newLoop(ctx context.Context, flusher flusher, batchSize int, flushInterval time.Duration) loop {
	writerC, readerC := channels.BufferedPipe[Change](ctx)

	return loop{
		writerC:       writerC,
		readerC:       readerC,
		flushInterval: flushInterval,
		batchSize:     batchSize,
		changeBuffer:  newChangeBuffer(flusher),
	}
}

func (s *loop) start(ctx context.Context) error {
	idle := time.NewTimer(s.flushInterval) // fires once, when we've been idle for flushInterval
	idle.Stop()                            // if nothing is buffered, keep the timer stopped

	defer func() {
		idle.Stop()
		slog.InfoContext(ctx, "shutting down changelog")
	}()

	slog.InfoContext(ctx, "starting changelog")

	for {
		select {
		case <-ctx.Done():
			// flush any leftovers
			_ = s.changeBuffer.tryFlush(ctx, 0)
			return nil

		case change, ok := <-s.readerC:
			if !ok {
				// input closed; try flushing
				_ = s.changeBuffer.tryFlush(ctx, 0)
				return nil
			}

			s.changeBuffer.add(change)
			// sized-base flush
			if err := s.changeBuffer.tryFlush(ctx, s.batchSize); err != nil {
				slog.WarnContext(ctx, "flush failed", "err", err)
			}

			// everytime we append to the buffer, we want the flush timer to start
			// ticking FROM NOW
			idle.Reset(s.flushInterval)

		case <-idle.C:
			slog.InfoContext(ctx, "idle flush", "timestamp", time.Now())
			// idle-based flush
			if err := s.changeBuffer.tryFlush(ctx, 0); err != nil {
				slog.WarnContext(ctx, "idle flush failed", "err", err)
			}
		}

	}
}

// this interface helps unit testing the loop logic
type flusher interface {
	flush(ctx context.Context, changes []Change) error
}

type dbFlusher struct {
	conn graph.Database
}

// newDBFlusher wraps a graph.Database into a Flusher.
func newDBFlusher(conn graph.Database) flusher {
	return &dbFlusher{conn: conn}
}

// flush implements Flusher by using BatchOperation.
func (s *dbFlusher) flush(ctx context.Context, changes []Change) error {
	if len(changes) == 0 {
		return nil
	}

	return s.conn.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, ch := range changes {
			if err := ch.Apply(batch); err != nil {
				return err
			}
		}
		return nil
	})
}

// changeBuffer is a small helper that accumulates changes and flushes them
// when size thresholds are hit.
type changeBuffer struct {
	buf     []Change
	flusher flusher
}

func newChangeBuffer(flusher flusher) *changeBuffer {
	return &changeBuffer{
		buf:     make([]Change, 0),
		flusher: flusher,
	}
}

func (s *changeBuffer) add(change Change) {
	s.buf = append(s.buf, change)
}

// tryFlush flushes the buffer if batchSize == 0 (force flush) or
// if the buffer length meets/exceeds batchSize. It clears the buffer
// regardless of flush success.
func (s *changeBuffer) tryFlush(ctx context.Context, batchSize int) error {
	if len(s.buf) == 0 {
		return nil
	}

	if batchSize == 0 || len(s.buf) >= batchSize {
		if err := s.flusher.flush(ctx, s.buf); err != nil {
			return err
		}
		s.buf = s.buf[:0]
	}
	return nil
}
