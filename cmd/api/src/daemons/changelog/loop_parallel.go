package changelog

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

func (s *loop) start_parallel(ctx context.Context) error {
	slog.InfoContext(ctx, "starting changelog workers")

	var wg sync.WaitGroup
	workerCount := 4 // tune for DB pool size and workload

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.runWorker(ctx, workerID)
		}(i)
	}

	<-ctx.Done()
	wg.Wait()

	slog.InfoContext(ctx, "shutting down changelog")
	return nil
}

func (s *loop) runWorker(ctx context.Context, id int) {
	buf := make([]Change, 0, s.batchSize)
	idle := time.NewTimer(s.flushInterval)

	if !idle.Stop() {
		<-idle.C
	}

	flush := func(reason string) {
		if len(buf) == 0 {
			return
		}
		if err := s.changeBuffer.flusher.flush(ctx, buf); err != nil {
			slog.WarnContext(ctx, "flush failed",
				"worker", id, "err", err, "count", len(buf), "reason", reason)
		}
		buf = buf[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush("shutdown")
			return

		case change, ok := <-s.readerC:
			if !ok {
				flush("channel-closed")
				return
			}

			buf = append(buf, change)
			if len(buf) >= s.batchSize {
				flush("size")
			}

			if !idle.Stop() {
				select {
				case <-idle.C:
				default:
				}
			}
			idle.Reset(s.flushInterval)

		case <-idle.C:
			flush("idle")
		}
	}
}
