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

// Changelog is a long-running daemon that manages change deduplication
// and buffering for graph ingestion.
type Changelog struct {
	Cache *cache
	db    graph.Database
	loop  loop

	options Options

	// protects cache swaps. runloop and runpoller need access to cache syncrhonized
	mu         sync.RWMutex
	flagGetter func(context.Context) (bool, int, error)

	// clean shutdown
	cancel context.CancelFunc
	done   chan struct{}
}

type Options struct {
	BatchSize     int
	FlushInterval time.Duration // interval for force flush when ingest is quiet (clears out leftovers in buffer)
	PollInterval  time.Duration // interval for Feature Flag check
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
		db:         dawgsDB,
	}
}

// Start begins a long-running loop that buffers and flushes node/edge updates
func (s *Changelog) Start(ctx context.Context) {
	var cctx context.Context
	cctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})

	s.loop = newLoop(cctx, newDBFlusher(s.db), s.options.BatchSize, s.options.FlushInterval)

	go func() {
		defer close(s.done)
		// this loop owns updating the lastseen property
		s.runLoop(cctx)
	}()

	// Poll feature flag in a separate goroutine
	go s.runPoller(cctx)
}

// runLoop owns the changelog’s inner ingestion loop.
func (s *Changelog) runLoop(ctx context.Context) {
	if err := s.loop.start(ctx); err != nil {
		slog.ErrorContext(ctx, "changelog loop exited with error", "err", err)
	}
}

// runPoller periodically checks the feature flag and sizes the cache accordingly.
func (s *Changelog) runPoller(ctx context.Context) {
	ticker := time.NewTicker(s.options.PollInterval)
	defer ticker.Stop()

	var isEnabled bool // track last seen state

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			flagEnabled, size, err := s.flagGetter(ctx)
			if err != nil {
				slog.WarnContext(ctx, "feature flag check failed", "err", err)
				continue
			}

			switch {
			case flagEnabled && !isEnabled:
				s.enable(ctx, size)
				isEnabled = true
			case !flagEnabled && isEnabled:
				s.disable(ctx)
				isEnabled = false
			}
		}
	}
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

// dbFlagGetter returns a closure that checks whether the changelog feature
// is enabled and, if so, reports the current graph size (nodes + edges).
// This allows the changelog to size its in-memory cache relative to the graph.
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

// enable sets up the cache with the calculated size.
func (s *Changelog) enable(ctx context.Context, size int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.InfoContext(ctx, "enabling changelog", "cache size", size)
	cache := newCache(size)
	s.Cache = cache
}

// disable resets the cache to free memory.
func (s *Changelog) disable(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.InfoContext(ctx, "disabling changelog, clearing cache")
	s.Cache = nil
}

// dont fiddle with FF for test hook.
func (s *Changelog) InitCacheForTest(ctx context.Context) {
	s.enable(ctx, 10)
}

func (s *Changelog) GetStats() cacheStats {
	s.mu.RLock()
	c := s.Cache
	s.mu.RUnlock()

	if c == nil { // cache may be nil when feature is disabled.
		return cacheStats{}
	}
	return s.Cache.getStats()
}

func (s *Changelog) FlushStats() {
	s.mu.RLock()
	c := s.Cache
	s.mu.RUnlock()

	if c == nil { // cache may be nil when feature is disabled.
		return
	}

	stats := s.Cache.resetStats()
	slog.Info("changelog metrics",
		"hits", stats.Hits,
		"misses", stats.Misses,
	)
}

func (s *Changelog) ResolveChange(change Change) (bool, error) {
	s.mu.RLock()
	c := s.Cache
	s.mu.RUnlock()

	if c == nil { // treat as pass-through when disabled.
		return true, nil
	}

	return s.Cache.shouldSubmit(change)
}

func (s *Changelog) Submit(ctx context.Context, change Change) bool {
	return channels.Submit(ctx, s.loop.writerC, change)
}
