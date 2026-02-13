// Copyright 2026 Specter Ops, Inc.
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
package ad

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util"
	"github.com/specterops/dawgs/util/channels"
)

func PostCanRDP(parentCtx context.Context, graphDB graph.Database, localGroupData *LocalGroupData, enforceURA bool, citrixEnabled bool) (*analysis.AtomicPostProcessingStats, error) {
	var (
		ctx, done             = context.WithCancel(parentCtx)
		stats                 = analysis.NewAtomicPostProcessingStats()
		numComputersProcessed = &atomic.Uint64{}
		workC                 = make(chan uint64)
		workerWG              sync.WaitGroup
		computerC             = make(chan *CanRDPComputerData)
		computerWG            sync.WaitGroup
		postC                 = make(chan analysis.CreatePostRelationshipJob, 4096)
		postWG                sync.WaitGroup
		submitStatusf         = util.SLogSampleRepeated("PostCanRDP")

		// Requirement for any CanRDP processing
		canRDPData, err = localGroupData.FetchCanRDPData(ctx, graphDB)
	)

	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"PostCanRDP",
		attr.Namespace("analysis"),
		attr.Function("PostCanRDP"),
		attr.Scope("process"),
	)()

	// Ensure the internal operation context is closed out
	defer done()

	// If we didn't get the canRDPData then we can't run post
	if err != nil {
		return nil, err
	}

	postWG.Add(1)

	go func() {
		defer postWG.Done()

		relProperties := analysis.NewPropertiesWithLastSeen()

		if err := graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			for {
				nextPost, shouldContinue := channels.Receive(ctx, postC)

				if !shouldContinue {
					break
				}

				// Because reach is calculated using a compontent graph the code must exclude
				// any self references
				if nextPost.FromID != nextPost.ToID {
					if err := batch.CreateRelationshipByIDs(nextPost.FromID, nextPost.ToID, nextPost.Kind, relProperties); err != nil {
						return err
					}
				}

				stats.AddRelationshipsCreated(nextPost.Kind, 1)
			}

			return nil
		}); err != nil {
			slog.Error("Write Computer CanRDP Post Processed Edge", attr.Error(err))
			done()
		}
	}()

	for workerID := 0; workerID < runtime.NumCPU()/2+1; workerID++ {
		computerWG.Add(1)

		go func(workerID int) {
			defer computerWG.Done()

			for {
				nextComputerRDPJob, shouldContinue := channels.Receive(ctx, computerC)

				if !shouldContinue {
					break
				}

				rdpEntities, err := FetchCanRDPEntityBitmapForComputer(nextComputerRDPJob, enforceURA, citrixEnabled)

				if err != nil {
					slog.Error("FetchCanRDPEntityBitmapForComputer Error", attr.Error(err))
					done()
				} else {
					rdpEntities.Each(func(fromID uint64) bool {
						return channels.Submit(ctx, postC, analysis.CreatePostRelationshipJob{
							FromID: graph.ID(fromID),
							ToID:   nextComputerRDPJob.Computer,
							Kind:   ad.CanRDP,
						})
					})
				}
			}
		}(workerID)
	}

	for workerID := 0; workerID < analysis.MaximumDatabaseParallelWorkers; workerID++ {
		workerWG.Add(1)

		go func(workerID int) {
			defer workerWG.Done()

			if err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
				for {
					var (
						nextComputer, shouldContinue = channels.Receive(ctx, workC)
						nextComputerID               = graph.ID(nextComputer)
					)

					if !shouldContinue {
						break
					}

					if computerCanRDPData, err := canRDPData.FetchCanRDPComputerData(tx, nextComputerID); err != nil {
						if !graph.IsErrNotFound(err) {
							return err
						}
					} else if !channels.Submit(ctx, computerC, computerCanRDPData) {
						break
					}

					if numComputersProcessed.Add(1)%10_000 == 0 {
						cacheStats := canRDPData.GroupMembershipCache.Stats()

						submitStatusf(
							slog.Uint64("num_computers", numComputersProcessed.Load()),
							slog.Uint64("num_cached", cacheStats.Cached),
							slog.Uint64("cache_hits", cacheStats.Hits),
						)
					}
				}

				return nil
			}); err != nil {
				slog.Error("Fetching CanRDP Computer Data", attr.Error(err))
				done()
			}
		}(workerID)
	}

	localGroupData.Computers.Each(func(nextComputer uint64) bool {
		return channels.Submit(ctx, workC, nextComputer)
	})

	close(workC)
	workerWG.Wait()

	close(computerC)
	computerWG.Wait()

	close(postC)
	postWG.Wait()

	return &stats, nil
}

func PostLocalGroups(parentCtx context.Context, graphDB graph.Database, localGroupData *LocalGroupData) (*analysis.AtomicPostProcessingStats, error) {
	const (
		adminGroupSuffix    = "-544"
		psRemoteGroupSuffix = "-580"
		dcomGroupSuffix     = "-562"
	)

	type reachJob struct {
		targetComputer uint64
		targetGroup    uint64
		groupSuffix    string
	}

	var (
		ctx, done             = context.WithCancel(parentCtx)
		stats                 = analysis.NewAtomicPostProcessingStats()
		computerC             = make(chan uint64)
		reachC                = make(chan reachJob, 4096)
		postC                 = make(chan analysis.CreatePostRelationshipJob, 4096)
		numGroupsProcessed    = &atomic.Uint64{}
		numComputersProcessed = &atomic.Uint64{}
		submitStatusf         = util.SLogSampleRepeated("PostLocalGroups")

		postWG  sync.WaitGroup
		reachWG sync.WaitGroup
		fetchWG sync.WaitGroup
	)

	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"PostLocalGroups",
		attr.Namespace("analysis"),
		attr.Function("PostLocalGroups"),
		attr.Scope("process"),
	)()

	// Ensure the internal operation context is closed out
	defer done()

	postWG.Add(1)

	go func() {
		defer postWG.Done()

		relProperties := analysis.NewPropertiesWithLastSeen()

		if err := graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
			for {
				nextPost, shouldContinue := channels.Receive(ctx, postC)

				if !shouldContinue {
					break
				}

				if err := batch.CreateRelationshipByIDs(nextPost.FromID, nextPost.ToID, nextPost.Kind, relProperties); err != nil {
					return err
				}

				stats.AddRelationshipsCreated(nextPost.Kind, 1)
			}

			return nil
		}); err != nil {
			slog.Error("Write Computer Local Group Post Processed Edge", attr.Error(err))
			done()
		}
	}()

	// Graph path workers
	for workerID := 0; workerID < runtime.NumCPU()/2+1; workerID += 1 {
		reachWG.Add(1)

		go func(workerID int) {
			defer reachWG.Done()

			for {
				var (
					nextJob, shouldContinue = channels.Receive(ctx, reachC)
					edgeKind                graph.Kind
				)

				if !shouldContinue {
					break
				}

				switch nextJob.groupSuffix {
				case adminGroupSuffix:
					edgeKind = ad.AdminTo
				case psRemoteGroupSuffix:
					edgeKind = ad.CanPSRemote
				case dcomGroupSuffix:
					edgeKind = ad.ExecuteDCOM
				default:
					continue
				}

				localGroupData.LocalGroupMembershipDigraph.EachAdjacentNode(nextJob.targetGroup, graph.DirectionInbound, func(fromID uint64) bool {
					return channels.Submit(ctx, postC, analysis.CreatePostRelationshipJob{
						FromID: graph.ID(fromID),
						ToID:   graph.ID(nextJob.targetComputer),
						Kind:   edgeKind,
					})
				})
			}
		}(workerID)
	}

	for workerID := 0; workerID < analysis.MaximumDatabaseParallelWorkers; workerID += 1 {
		fetchWG.Add(1)

		go func(workerID int) {
			defer fetchWG.Done()

			if err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
				for {
					computerID, shouldContinue := channels.Receive(ctx, computerC)

					if !shouldContinue {
						break
					}

					if localAdminGroup, err := FetchComputerLocalGroupIDBySIDSuffix(tx, graph.ID(computerID), adminGroupSuffix); err != nil {
						if !graph.IsErrNotFound(err) {
							return err
						}
					} else {
						numGroupsProcessed.Add(1)

						channels.Submit(ctx, reachC, reachJob{
							targetComputer: computerID,
							targetGroup:    localAdminGroup.Uint64(),
							groupSuffix:    adminGroupSuffix,
						})
					}

					if localPSRemoteGroup, err := FetchComputerLocalGroupIDBySIDSuffix(tx, graph.ID(computerID), psRemoteGroupSuffix); err != nil {
						if !graph.IsErrNotFound(err) {
							return err
						}
					} else {
						numGroupsProcessed.Add(1)

						channels.Submit(ctx, reachC, reachJob{
							targetComputer: computerID,
							targetGroup:    localPSRemoteGroup.Uint64(),
							groupSuffix:    psRemoteGroupSuffix,
						})
					}

					if localDCOMGroup, err := FetchComputerLocalGroupIDBySIDSuffix(tx, graph.ID(computerID), dcomGroupSuffix); err != nil {
						if !graph.IsErrNotFound(err) {
							return err
						}
					} else {
						numGroupsProcessed.Add(1)

						channels.Submit(ctx, reachC, reachJob{
							targetComputer: computerID,
							targetGroup:    localDCOMGroup.Uint64(),
							groupSuffix:    dcomGroupSuffix,
						})
					}

					if numComputersProcessed.Add(1)%10000 == 0 {
						submitStatusf(slog.Uint64("num_computers", numComputersProcessed.Load()))
					}
				}

				return nil
			}); err != nil {
				slog.Error("Read Computer Local Groups", attr.Error(err))
				done()
			}
		}(workerID)
	}

	localGroupData.Computers.Each(func(value uint64) bool {
		return channels.Submit(ctx, computerC, value)
	})

	close(computerC)
	fetchWG.Wait()

	close(reachC)
	reachWG.Wait()

	close(postC)
	postWG.Wait()

	return &stats, nil
}
