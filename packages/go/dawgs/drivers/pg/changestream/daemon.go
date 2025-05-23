package changestream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

const (
	// Limit batch sizes
	batchSize = 1_000
)

var (
	ignoredPropertiesKeys = map[string]struct{}{
		common.ObjectID.String():      {},
		common.LastSeen.String():      {},
		common.LastCollected.String(): {},
		common.IsInherited.String():   {},
		ad.DomainSID.String():         {},
		ad.IsACL.String():             {},
		azure.TenantID.String():       {},
	}
)

type Log interface {
	LastNodeChange(ctx context.Context, proposedChange *NodeChange) (ChangeLookup, error)
	CachedLastNodeChange(proposedChange *NodeChange) (ChangeLookup, error)
	LastEdgeChange(ctx context.Context, proposedChange *EdgeChange) (ChangeLookup, error)
	CachedLastEdgeChange(proposedChange *EdgeChange) (ChangeLookup, error)
	Submit(ctx context.Context, change Change) bool
}

type Daemon struct {
	writerC          chan<- Change
	readerC          <-chan Change
	pgxPool          *pgxpool.Pool
	kindMapper       pg.KindMapper
	changeCacheLock  *sync.RWMutex
	changeCache      map[string]ChangeLookup
	state            *stateManager
	nodeChangeBuffer []*NodeChange
	edgeChangeBuffer []*EdgeChange
}

func NewDaemon(ctx context.Context, flags appcfg.GetFlagByKeyer, connectionStr string, kindMapper pg.KindMapper) (*Daemon, error) {
	var (
		writerC, readerC = channels.BufferedPipe[Change](ctx)
		changeStream     = &Daemon{
			writerC:         writerC,
			readerC:         readerC,
			kindMapper:      kindMapper,
			changeCacheLock: &sync.RWMutex{},
			changeCache:     make(map[string]ChangeLookup),
			state:           newStateManager(flags),
		}
	)

	// TODO: Paging doctor spaghetti... I don't actually know if this really wants for its own pool
	if pool, err := newPGXPool(ctx, connectionStr); err != nil {
		return nil, err
	} else {
		changeStream.pgxPool = pool
	}

	// Prime the feature flag check for the daemon state manager
	if err := changeStream.state.checkFeatureFlag(ctx, time.Now()); err != nil {
		return nil, err
	}

	return changeStream, nil
}

func (s *Daemon) putCachedChange(cacheKey string, cachedLookup ChangeLookup) {
	s.changeCacheLock.Lock()
	defer s.changeCacheLock.Unlock()

	s.changeCache[cacheKey] = cachedLookup
}

func (s *Daemon) lastCachedChange(cacheKey string) (ChangeLookup, bool) {
	s.changeCacheLock.RLock()
	defer s.changeCacheLock.RUnlock()

	cachedChange, hasCachedChange := s.changeCache[cacheKey]
	return cachedChange, hasCachedChange
}

func (s *Daemon) queueChange(ctx context.Context, now time.Time, nextChange Change) {
	switch typedChange := nextChange.(type) {
	case *NodeChange:
		s.nodeChangeBuffer = append(s.nodeChangeBuffer, typedChange)

		if len(s.nodeChangeBuffer) >= batchSize {
			if err := s.flushNodeChanges(ctx, nodeChangePartitionName(now)); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("writing changes to change stream failed: %v", err))
			}
		}

	case *EdgeChange:
		s.edgeChangeBuffer = append(s.edgeChangeBuffer, typedChange)

		if len(s.edgeChangeBuffer) >= batchSize {
			if err := s.flushEdgeChanges(ctx, edgeChangePartitionName(now)); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("writing edge changes to change stream failed: %v", err))
			}
		}

	default:
		slog.ErrorContext(ctx, fmt.Sprintf("Unexpected change type: %T", nextChange))
	}
}

func (s *Daemon) flushNodeChanges(ctx context.Context, partitionName string) error {
	// Early exit check for empty buffer flushes
	if len(s.nodeChangeBuffer) == 0 {
		return nil
	}

	var (
		numChanges  = len(s.nodeChangeBuffer)
		now         = time.Now()
		copyColumns = []string{
			"target_node",
			"kind_ids",
			"properties_hash",
			"property_fields",
			"change_type",
			"created_at",
		}

		iteratorFunc = func(idx int) ([]any, error) {
			nextNodeChange := s.nodeChangeBuffer[idx]

			if mappedKindIDs, err := s.kindMapper.MapKinds(ctx, nextNodeChange.Kinds); err != nil {
				return nil, fmt.Errorf("node kind ID mapping error: %w", err)
			} else if propertiesHash, err := nextNodeChange.Properties.Hash(ignoredPropertiesKeys); err != nil {
				return nil, fmt.Errorf("node properties hash error: %w", err)
			} else {
				return []any{
					nextNodeChange.TargetNodeID,
					mappedKindIDs,
					propertiesHash,
					[]string{}, // nextNodeChange.Properties.Keys(ignoredPropertiesKeys),
					nextNodeChange.Type(),
					now,
				}, nil
			}
		}
	)

	slog.DebugContext(ctx, fmt.Sprintf("flushing %d node changes", numChanges))

	if _, err := s.pgxPool.CopyFrom(ctx, pgx.Identifier{partitionName}, copyColumns, pgx.CopyFromSlice(numChanges, iteratorFunc)); err != nil {
		slog.Info(fmt.Sprintf("change stream node change insert error: %v", err))
	}

	// TODO: Should the buffer clear be dependent on no errors?
	s.nodeChangeBuffer = s.nodeChangeBuffer[:0]
	return nil
}

func (s *Daemon) flushEdgeChanges(ctx context.Context, partitionName string) error {
	// Early exit check for empty buffer flushes
	if len(s.edgeChangeBuffer) == 0 {
		return nil
	}

	var (
		numChanges  = len(s.edgeChangeBuffer)
		now         = time.Now()
		copyColumns = []string{
			"target_node",
			"related_node",
			"kind_id",
			"properties_hash",
			"property_fields",
			"identity_hash",
			"change_type",
			"created_at",
		}

		iteratorFunc = func(idx int) ([]any, error) {
			nextEdgeChange := s.edgeChangeBuffer[idx]

			if edgeIdentity, err := nextEdgeChange.IdentityHash(); err != nil {
				return nil, fmt.Errorf("edge identity hash error: %w", err)
			} else if mappedKindID, err := s.kindMapper.MapKind(ctx, nextEdgeChange.Kind); err != nil {
				return nil, fmt.Errorf("edge kind ID mapping error: %w", err)
			} else if propertiesHash, err := nextEdgeChange.Properties.Hash(ignoredPropertiesKeys); err != nil {
				return nil, fmt.Errorf("failed creating edge change property hash: %w", err)
			} else {
				return []any{
					nextEdgeChange.TargetNodeID,
					nextEdgeChange.RelatedNodeID,
					mappedKindID,
					propertiesHash,
					[]string{}, // nextEdgeChange.Properties.Keys(ignoredPropertiesKeys),
					edgeIdentity,
					nextEdgeChange.Type(),
					now,
				}, nil
			}
		}
	)

	slog.DebugContext(ctx, fmt.Sprintf("flushing %d edge changes", numChanges))

	if _, err := s.pgxPool.CopyFrom(ctx, pgx.Identifier{partitionName}, copyColumns, pgx.CopyFromSlice(numChanges, iteratorFunc)); err != nil {
		slog.Info(fmt.Sprintf("change stream node change insert error: %v", err))
	}

	// TODO: Should the buffer clear be dependent on no errors?
	s.edgeChangeBuffer = s.edgeChangeBuffer[:0]
	return nil
}

func (s *Daemon) RunLoop(ctx context.Context) error {
	// Assert the table and the current partition before entering the async loop
	if _, err := s.pgxPool.Exec(ctx, assertTableSQL); err != nil {
		return fmt.Errorf("failed asserting changelog tablespace: %w", err)
	}

	if err := assertChangelogPartition(ctx, s.pgxPool); err != nil {
		return fmt.Errorf("failed asserting changelog partition: %w", err)
	}

	go func() {
		var (
			// Attempt to flush every 5 seconds
			flushTicker             = time.NewTicker(5 * time.Second)
			lastNodeBufferWatermark = 0
			lastEdgeBufferWatermark = 0
		)

		// This goroutine now owns the lifecycle for the PGX connection pool
		defer s.pgxPool.Close()
		defer close(s.writerC)
		defer flushTicker.Stop()

		slog.InfoContext(ctx, "Starting change stream")
		defer slog.InfoContext(ctx, "Shutting down change stream")

		for {
			now := time.Now()

			select {
			case nextChange := <-s.readerC:
				if err := s.state.checkFeatureFlag(ctx, now); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("change log feature flag check: %v", err))
				} else if s.state.isEnabled() {
					if err := s.state.checkChangelogPartitions(ctx, now, s.pgxPool); err != nil {
						slog.ErrorContext(ctx, fmt.Sprintf("change log schema check: %v", err))
					} else {
						s.queueChange(ctx, now, nextChange)
					}
				}

			case <-flushTicker.C:
				if err := s.state.checkFeatureFlag(ctx, now); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("change log flush feature flag check error: %v", err))
				} else if s.state.isEnabled() {
					var (
						numNodeChangesPending      = len(s.nodeChangeBuffer)
						numEdgeChangesPending      = len(s.edgeChangeBuffer)
						hasNodeChangesPendingFlush = numNodeChangesPending > 0 && numNodeChangesPending == lastNodeBufferWatermark
						hasEdgeChangesPendingFlush = numEdgeChangesPending > 0 && numEdgeChangesPending == lastEdgeBufferWatermark
					)

					if hasNodeChangesPendingFlush || hasEdgeChangesPendingFlush {
						if err := s.state.checkChangelogPartitions(ctx, now, s.pgxPool); err != nil {
							slog.ErrorContext(ctx, fmt.Sprintf("change log state check error: %v", err))
						} else {
							if hasNodeChangesPendingFlush {
								if err := s.flushNodeChanges(ctx, nodeChangePartitionName(now)); err != nil {
									slog.ErrorContext(ctx, fmt.Sprintf("writing changes to change stream failed: %v", err))
								}
							}

							if hasEdgeChangesPendingFlush {
								if err := s.flushEdgeChanges(ctx, edgeChangePartitionName(now)); err != nil {
									slog.ErrorContext(ctx, fmt.Sprintf("writing edge changes to change stream failed: %v", err))
								}
							}
						}
					}

					// Update the buffer watermarks
					lastNodeBufferWatermark = len(s.nodeChangeBuffer)
					lastEdgeBufferWatermark = len(s.edgeChangeBuffer)
				}

			case <-ctx.Done():
				// Context canceled, exit right away
				return
			}
		}
	}()

	return nil
}

func (s *Daemon) CachedLastNodeChange(proposedChange *NodeChange) (ChangeLookup, error) {
	var (
		lastChange     ChangeLookup
		changeCacheKey = proposedChange.CacheKey()
	)

	if propertiesHash, err := proposedChange.Properties.Hash(ignoredPropertiesKeys); err != nil {
		return lastChange, err
	} else {
		// Track the properties hash and kind IDs
		lastChange.PropertiesHash = propertiesHash
	}

	if cachedChange, hasCachedChange := s.lastCachedChange(changeCacheKey); hasCachedChange {
		lastChange.Changed = !bytes.Equal(lastChange.PropertiesHash, cachedChange.PropertiesHash)
		lastChange.Exists = true
	} else {
		// If the change log is disabled then mark every non-cached lookup as changed
		lastChange.Changed = !s.state.isEnabled()
	}

	// Ensure this makes it into the cache before returning
	s.putCachedChange(changeCacheKey, lastChange)
	return lastChange, nil
}

func (s *Daemon) LastNodeChange(ctx context.Context, proposedChange *NodeChange) (ChangeLookup, error) {
	lastChange, err := s.CachedLastNodeChange(proposedChange)

	if err != nil || lastChange.Exists {
		return lastChange, err
	}

	if s.state.isEnabled() {
		var (
			lastChangeRow = s.pgxPool.QueryRow(ctx, lastNodeChangeSQL, proposedChange.TargetNodeID, lastChange.PropertiesHash)
			err           = lastChangeRow.Scan(&lastChange.Changed, &lastChange.Type)
		)

		// Assume that the change that exists in some form and error inspect for the negative case
		lastChange.Exists = true

		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				// Exit here as this is an unexpected error
				return lastChange, err
			}

			// No rows found means the change does not exist
			lastChange.Exists = false
		}

		// Ensure this makes it into the cache before returning
		s.putCachedChange(proposedChange.CacheKey(), lastChange)
	}

	return lastChange, nil
}

func (s *Daemon) CachedLastEdgeChange(proposedChange *EdgeChange) (ChangeLookup, error) {
	var (
		lastChange     = ChangeLookup{}
		changeCacheKey = proposedChange.CacheKey()
	)

	if propertiesHash, err := proposedChange.Properties.Hash(ignoredPropertiesKeys); err != nil {
		return lastChange, err
	} else {
		// Track the properties hash
		lastChange.PropertiesHash = propertiesHash
	}

	if cachedChange, hasCachedChange := s.lastCachedChange(changeCacheKey); hasCachedChange {
		lastChange.Changed = !bytes.Equal(lastChange.PropertiesHash, cachedChange.PropertiesHash)
		lastChange.Exists = true
	} else {
		// If the change log is disabled then mark every non-cached lookup as changed
		lastChange.Changed = !s.state.isEnabled()
	}

	// Ensure this makes it into the cache before returning
	s.putCachedChange(changeCacheKey, lastChange)
	return lastChange, nil
}

func (s *Daemon) LastEdgeChange(ctx context.Context, proposedChange *EdgeChange) (ChangeLookup, error) {
	lastChange, err := s.CachedLastEdgeChange(proposedChange)

	if err != nil || lastChange.Exists {
		return lastChange, err
	}

	if s.state.isEnabled() {
		if edgeIdentity, err := proposedChange.IdentityHash(); err != nil {
			return lastChange, err
		} else {
			var (
				lastChangeRow = s.pgxPool.QueryRow(ctx, lastEdgeChangeSQL, edgeIdentity, lastChange.PropertiesHash)
				err           = lastChangeRow.Scan(&lastChange.Changed, &lastChange.Type)
			)

			// Assume that the change that exists in some form and error inspect for the negative case
			lastChange.Exists = true

			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					// Exit here as this is an unexpected error
					return lastChange, err
				}

				// No rows found means the change does not exist
				lastChange.Exists = false
			}

			// Ensure this makes it into the cache before returning
			s.putCachedChange(proposedChange.CacheKey(), lastChange)
		}
	}

	// Ensure this makes it into the cache before returning
	return lastChange, nil
}

func (s *Daemon) Submit(ctx context.Context, change Change) bool {
	return channels.Submit(ctx, s.writerC, change)
}
