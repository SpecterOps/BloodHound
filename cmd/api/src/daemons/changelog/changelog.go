package changelog

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/channels"
)

// ChangeManager represents the ingestion-facing API for the changelog daemon.
//
// It provides three responsibilities:
//   - Deduplication: ResolveChange determines whether a proposed change is new or modified
//     and therefore requires persistence, or whether it has already been seen.
//   - Submission: Submit enqueues a change for asynchronous processing by the changelog loop.
//   - Metrics: FlushStats logs and resets internal cache hit/miss statistics,
//     allowing callers to observe deduplication efficiency over time.
//
// Typical usage in ingestion pipelines is:
//  1. Call ResolveChange to decide if the update should be applied.
//  2. If ResolveChange returns true, apply the update to the batch/DB.
//  3. If ResolveChange returns false, Submit the change to the changelog. (this updates an entities lastseen prop for reconciliation)
//
// To generate mocks for this interface for unit testing seams in the application
// please use:
//
//	mockgen -source=changelog.go -destination=mocks/changelog.go -package=mocks
type ChangeManager interface {
	ResolveChange(change Change) (bool, error)
	Submit(ctx context.Context, change Change) bool
	FlushStats()
}

type Changelog struct {
	cache   *cache // pointer so we can replace it atmoically
	conn    graph.Database
	loop    loop
	options Options

	mu         sync.RWMutex // protects cache swaps
	flagGetter func(context.Context) (bool, int, error)

	// clean shutdown
	cancel context.CancelFunc
	done   chan struct{}
}

type Options struct {
	BatchSize     int
	FlushInterval time.Duration // interval for force flush when ingest is quiet (to clear out leftovers in buffer)
	PollInterval  time.Duration // interval for FF check
}

func DefaultOptions() Options {
	return Options{
		BatchSize:     1_000,
		FlushInterval: 5 * time.Second,
		PollInterval:  10 * time.Second,
	}
}

func NewChangelog(dawgsDB graph.Database, flagProvider appcfg.GetFlagByKeyer, opts Options) *Changelog {
	return &Changelog{
		flagGetter: dbFlagGetter(dawgsDB, flagProvider),
		options:    opts,
		conn:       dawgsDB,
	}
}

// Start begins a long-running loop that buffers and flushes node/edge updates
func (s *Changelog) Start(ctx context.Context) {
	var cctx context.Context
	cctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})

	s.loop = newLoop(cctx, newDBFlusher(s.conn), s.options.BatchSize, s.options.FlushInterval)

	go func() {
		defer close(s.done)
		if err := s.loop.start(cctx); err != nil {
			slog.ErrorContext(cctx, "changelog loop exited with error", "err", err)
		}
	}()

	// Poll feature flag in a separate goroutine
	go func() {
		ticker := time.NewTicker(s.options.PollInterval)
		defer ticker.Stop()

		var isEnabled bool // track last seen state

		for {
			select {
			case <-cctx.Done():
				return

			case <-ticker.C:
				flagEnabled, size, err := s.flagGetter(cctx)
				if err != nil {
					slog.WarnContext(cctx, "feature flag check failed", "err", err)
					continue
				}

				switch {
				case flagEnabled && !isEnabled:
					s.Enable(cctx, size)
					isEnabled = true
				case !flagEnabled && isEnabled:
					s.Disable(cctx)
					isEnabled = false
				}
			}
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

func dbFlagGetter(dawgsDB graph.Database, flagProvider appcfg.GetFlagByKeyer) func(context.Context) (bool, int, error) {
	return func(ctx context.Context) (bool, int, error) {
		flag, err := flagProvider.GetFlagByKey(ctx, appcfg.FeatureChangelog)
		if err != nil {
			return false, 0, fmt.Errorf("getting changelog flag: %w", err)
		}

		if !flag.Enabled {
			return false, 0, nil
		}

		var graphSize int64
		if err := dawgsDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if nodeCount, err := tx.Nodes().Count(); err != nil {
				return err
			} else if edgeCount, err := tx.Relationships().Count(); err != nil {
				return err
			} else {
				graphSize = nodeCount + edgeCount
				return nil
			}
		}); err != nil {
			return false, 0, fmt.Errorf("counting nodes and relationships in graph: %w ", err)
		}

		return true, int(graphSize), nil
	}
}

// Enable sets up the cache with the calculated size.
func (s *Changelog) Enable(ctx context.Context, size int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.InfoContext(ctx, "enabling changelog", "cache size", size)
	cache := newCache(size)
	s.cache = &cache
}

// Disable resets the cache to free memory.
func (s *Changelog) Disable(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.InfoContext(ctx, "disabling changelog, clearing cache")
	s.cache = nil
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
