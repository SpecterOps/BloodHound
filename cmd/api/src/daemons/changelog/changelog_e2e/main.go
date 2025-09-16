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
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
)

type Harness struct {
	DB        graph.Database
	Log       *changelog.Changelog
	Ctx       context.Context
	Cancel    context.CancelFunc
	WaitGroup *sync.WaitGroup
}

var (
	nodeKinds = graph.Kinds{graph.StringKind("NK1")}
	edgeKinds = graph.Kinds{graph.StringKind("EK2")}
)

func schema() graph.Schema {
	defaultGraph := graph.Graph{
		Name:  "default",
		Nodes: nodeKinds,
		Edges: edgeKinds,
		NodeConstraints: []graph.Constraint{{
			Field: "objectid",
			Type:  graph.BTreeIndex,
		}},
	}

	return graph.Schema{
		Graphs:       []graph.Graph{defaultGraph},
		DefaultGraph: defaultGraph,
	}
}

func newHarness() *Harness {
	ctx, cancel := context.WithCancel(context.Background())

	connStr := os.Getenv("PG_CONNECTION_STRING")
	if connStr == "" {
		slog.Error("PG_CONNECTION_STRING env variable is not set")
		cancel()
		os.Exit(1)
	}

	pool, err := pg.NewPool(connStr)
	if err != nil {
		slog.Error("failed to connect", "err", err)
		os.Exit(1)
	}

	dawgsDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{ConnectionString: connStr, Pool: pool})
	if err != nil {
		slog.Error("failed to open", "err", err)
		os.Exit(1)
	}

	gormDB, err := database.OpenDatabase(connStr)
	if err != nil {
		slog.Error("failed to open", "err", err)
		os.Exit(1)
	}

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	// Attempt to truncate but don't care about the error
	dawgsDB.Run(
		ctx,
		`
						do $$ declare
							r record;
						begin
							for r in (select tablename from pg_tables where schemaname = 'public') loop
								execute 'drop table if exists ' || quote_ident(r.tablename) || ' cascade';
							end loop;
						end $$;
		`, nil)

	if err := dawgsDB.AssertSchema(ctx, schema()); err != nil {
		slog.Error("schema validation failed", "err", err)
		os.Exit(1)
	}

	log := changelog.NewChangelog(dawgsDB, db, changelog.DefaultOptions())

	return &Harness{DB: dawgsDB, Log: log, Ctx: ctx, Cancel: cancel, WaitGroup: &sync.WaitGroup{}}
}

func (h *Harness) Close() {
	h.Cancel()
	h.WaitGroup.Wait()
}

func waitForShutdown(cancel func(), wg *sync.WaitGroup) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigC
		slog.Info("Received shutdown signal")
		cancel()
		// os.Exit(0)
	}()
}

// main() is a playground for rapidly testing changelog functionality end-end
// without having to go through the broader Bloodhound application.
func main() {
	harness := newHarness()
	defer harness.Close()

	harness.Log.Start(harness.Ctx)
	waitForShutdown(harness.Cancel, harness.WaitGroup)

	if err := runTest(harness.Ctx, harness.Log, harness.DB); err != nil {
		slog.Error("test failed", "err", err)
		os.Exit(1)
	}

	harness.WaitGroup.Wait()
	slog.Info("Shutdown complete")

	os.Exit(0)
}

func runTest(ctx context.Context, log *changelog.Changelog, db graph.Database) error {
	numNodes := 1_000_000
	diffBegins := 900_000

	if err := runNodeBatch(ctx, log, db, numNodes, diffBegins, false, "node batch 1"); err != nil {
		slog.ErrorContext(ctx, "run node batch 1", "err", err)
	}
	if err := runNodeBatch(ctx, log, db, numNodes, diffBegins, true, "node batch 2"); err != nil {
		slog.ErrorContext(ctx, "run node batch 2", "err", err)
	}

	if err := runEdgeBatch(ctx, log, db, numNodes, diffBegins, false, "edge batch 3"); err != nil {
		slog.ErrorContext(ctx, "run edge batch 1", "err", err)
	}
	if err := runEdgeBatch(ctx, log, db, numNodes, diffBegins, true, "edge batch 4"); err != nil {
		slog.ErrorContext(ctx, "run edge batch 2", "err", err)
	}

	slog.Info("Done with test")
	return nil
}

func runNodeBatch(ctx context.Context, log *changelog.Changelog, db graph.Database, numNodes, diffBegins int, simulateDiff bool, label string) error {
	return db.BatchOperation(ctx, func(batch graph.Batch) error {

		start := time.Now()
		slog.Info(fmt.Sprintf("%s starting", label), "timestamp", start)
		defer func() {
			slog.Info(fmt.Sprintf("%s finished", label), "duration", time.Since(start))
		}()

		for idx := 0; idx < numNodes; idx++ {
			if idx%1000 == 0 {
				if ctx.Err() != nil {
					slog.Info("batch interrupted", "idx", idx)
					return ctx.Err()
				}
			}

			props := graph.NewProperties().
				Set("objectid", strconv.Itoa(idx)).
				Set("node_index", idx).
				Set("lastseen", start)

			if simulateDiff && idx > diffBegins {
				props.Set("different_prop", idx)
			}

			change := changelog.NewNodeChange(strconv.Itoa(idx), nodeKinds, props)
			if shouldSubmit, err := log.ResolveChange(change); err != nil {
				slog.Error("ResolveChange failed", "err", err)
			} else if shouldSubmit {
				if err := batch.UpdateNodeBy(graph.NodeUpdate{
					Node:               graph.PrepareNode(props, nodeKinds...),
					IdentityProperties: []string{"objectid"},
				}); err != nil {
					return fmt.Errorf("UpdatedNodeBy failed at idx=%d: %w", idx, err)
				}
			} else {
				log.Submit(ctx, change)
			}
		}

		log.FlushStats()
		return nil
	})

}

func runEdgeBatch(ctx context.Context, log *changelog.Changelog, db graph.Database, numNodes, diffBegins int, simulateDiff bool, label string) error {
	return db.BatchOperation(ctx, func(batch graph.Batch) error {
		start := time.Now()
		slog.Info(fmt.Sprintf("%s starting", label), "timestamp", start)
		defer func() {
			slog.Info(fmt.Sprintf("%s finished", label), "duration", time.Since(start))
		}()

		batchCount := 0
		submittedCount := 0
		for idx := 0; idx < numNodes; idx++ {
			if idx%1000 == 0 {
				if ctx.Err() != nil {
					slog.Info("batch interrupted", "idx", idx)
					return ctx.Err()
				}
			}

			startObjID := strconv.Itoa(idx)
			endObjID := strconv.Itoa(idx + 1)

			props := graph.NewProperties().
				Set("startID", startObjID).
				Set("endID", endObjID).
				Set("lastseen", start)

			if simulateDiff && idx > diffBegins {
				props.Set("different_prop", idx)
			}

			change := changelog.NewEdgeChange(startObjID, endObjID, edgeKinds[0], props)
			if shouldSubmit, err := log.ResolveChange(change); err != nil {
				slog.Error("ResolveChange failed", "err", err)
			} else if shouldSubmit {
				batchCount++
				update := graph.RelationshipUpdate{
					Start:                   graph.PrepareNode(graph.NewProperties().SetAll(map[string]any{"objectid": startObjID, "lastseen": start}), nodeKinds...),
					StartIdentityProperties: []string{"objectid"},
					End:                     graph.PrepareNode(graph.NewProperties().SetAll(map[string]any{"objectid": endObjID, "lastseen": start}), nodeKinds...),
					EndIdentityProperties:   []string{"objectid"},
					Relationship:            graph.PrepareRelationship(props, edgeKinds[0]),
				}
				if err := batch.UpdateRelationshipBy(update); err != nil {
					return fmt.Errorf("UpdateRelationshipBy failed at idx=%d: %w", idx, err)
				}
			} else {
				submittedCount++
				log.Submit(ctx, change)
			}
		}
		fmt.Printf("batchCount: %d, submittedCount: %d \n", batchCount, submittedCount)

		log.FlushStats()
		return nil
	})
}
