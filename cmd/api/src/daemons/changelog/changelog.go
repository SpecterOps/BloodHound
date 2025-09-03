package changelog

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

type Changelog struct {
	cache   cache
	conn    graph.Database
	loop    loop
	options Options

	// clean shutdown
	cancel context.CancelFunc
	done   chan struct{}
}

type Options struct {
	BatchSize     int
	FlushInterval time.Duration
}

func DefaultOptions() Options {
	return Options{
		BatchSize:     1_000,
		FlushInterval: 5 * time.Second,
	}
}

func NewChangelog(dawgsDB graph.Database, opts Options) *Changelog {
	var (
		cache = newCache()
	)

	return &Changelog{
		cache:   cache,
		options: opts,
		conn:    dawgsDB,
	}
}

// Start begins a long-running loop that buffers and flushes node/edge updates
// TODO: right-size the map here?
// will need to pass a bool
func (s *Changelog) Start(ctx context.Context) {
	var cctx context.Context
	cctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})

	// FF check, db check for size of nodes + edge table. default to 1mil size
	// s.cache = newCache()

	s.loop = newLoop(cctx, newDBFlusher(s.conn), s.options.BatchSize, s.options.FlushInterval)
	// s.enabled := FF
	go func() {
		defer close(s.done)
		if err := s.loop.start(cctx); err != nil {
			slog.ErrorContext(cctx, "changelog loop exited with error", "err", err)
		}
	}()
}

func (s *Changelog) Stop(ctx context.Context) error {
	if s.cancel == nil {
		return nil // never started
	}

	// tell loop to stop
	s.cancel()

	// wait until loop exits or context times out
	select {
	case <-s.done:
		slog.Info("changelog shutdown complete")
		return nil
	case <-ctx.Done():
		return ctx.Err() // caller's timeout
	}
}

func (s *Changelog) Name() string {
	return "Changelog Daemon"
}

func (s *Changelog) GetStats() cacheStats {
	return s.cache.getStats()
}

func (s *Changelog) FlushStats() {
	stats := s.cache.resetStats()
	slog.Info("changelog metrics",
		"hits", stats.Hits,
		"misses", stats.Misses,
	)
}

func (s *Changelog) ResolveChange(change Change) (bool, error) {
	return s.cache.shouldSubmit(change)
}

func (s *Changelog) Submit(ctx context.Context, change Change) bool {
	return channels.Submit(ctx, s.loop.writerC, change)
}
