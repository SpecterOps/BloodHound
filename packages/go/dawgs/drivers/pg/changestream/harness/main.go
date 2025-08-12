package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/changestream"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

type flagger struct {
}

func (f flagger) GetFlagByKey(ctx context.Context, s string) (appcfg.FeatureFlag, error) {
	return appcfg.FeatureFlag{
		Key:           s,
		Enabled:       true,
		UserUpdatable: true,
	}, nil
}

func newFlagger() appcfg.GetFlagByKeyer {
	return &flagger{}
}

var (
	nodeKinds = graph.Kinds{graph.StringKind("NK1")}
	edgeKinds = graph.Kinds{graph.StringKind("EK2")}
)

func test(ctx context.Context, csDaemon *changestream.Daemon) error {
	for idx := 0; idx < 100_000; idx++ {
		var (
			nodeObjectID   = strconv.Itoa(idx)
			nodeProperties = graph.NewProperties()
		)

		//if idx%1000 == 0 {
		//	time.Sleep(1 * time.Second)
		//}

		nodeProperties.Set("objectid", nodeObjectID)
		nodeProperties.Set("node_index", idx)

		// Ingest File --> This function
		csDaemon.Submit(ctx, changestream.NewNodeChange(
			changestream.ChangeTypeAdd,
			nodeObjectID,
			nodeKinds,
			nodeProperties,
		))
	}

	time.Sleep(5 * time.Second)

	for idx := 0; idx < 100_000; idx++ {
		var (
			nodeObjectID   = strconv.Itoa(idx)
			nodeProperties = graph.NewProperties()
		)

		//if idx%1000 == 0 {
		//	time.Sleep(1 * time.Second)
		//}

		nodeProperties.Set("objectid", nodeObjectID)
		nodeProperties.Set("node_index", idx)
		nodeProperties.Set("new_value", 41240-idx)

		csDaemon.Submit(ctx, changestream.NewNodeChange(
			changestream.ChangeTypeUpdate,
			nodeObjectID,
			nodeKinds,
			nodeProperties,
		))
	}

	slog.Info("Done with test")

	return nil
}

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

func main() {
	const pgConnStr = "user=postgres dbname=bhe password=bhe4eva host=localhost"

	var (
		ctx, done = context.WithCancel(context.Background())
	)

	defer done()

	if pool, err := pg.NewPool(pgConnStr); err != nil {
		fmt.Printf("failed to connect to database: %v\n", err)
		os.Exit(1)
	} else if dawgsDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		ConnectionString: pgConnStr,
		Pool:             pool,
	}); err != nil {
		fmt.Printf("Failed to open database connection: %v\n", err)
		os.Exit(1)
	} else {
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
			fmt.Printf("Failed to validate schema: %v\n", err)
			os.Exit(1)
		}

		notificationC := make(chan changestream.Notification)

		if schemaManager, err := pg.SchemaManagerFromGraphDatabase(dawgsDB); err != nil {
			fmt.Printf("Failed to create kind mapper from graph database: %v\n", err)
			os.Exit(1)
		} else if csDaemon, err := changestream.NewDaemon(ctx, newFlagger(), "user=postgres dbname=bhe password=bhe4eva host=localhost", schemaManager, notificationC); err != nil {
			fmt.Printf("Failed to create daemon: %v\n", err)
			os.Exit(1)
		} else if err := csDaemon.RunLoop(ctx, 5*time.Second); err != nil {
			fmt.Printf("Failed to run daemon: %v\n", err)
			os.Exit(1)
		} else {
			// Ingest routine
			wg := &sync.WaitGroup{}

			go func() {
				var (
					encoder                = pg.NewInt2ArrayEncoder()
					lastNodeChangeID int64 = 0
					//lastEdgeChangeID int64 = 0
				)

				// Ingest routine
				wg.Add(1)

				defer wg.Done()

				for nextNotification := range notificationC {
					switch nextNotification.Type {
					case changestream.NotificationNode:
						slog.Info("received notification to update nodes to changelog revision", slog.Int64("rev_id", lastNodeChangeID), slog.Int64("next_rev_id", nextNotification.RevisionID))

						if defaultGraph, hasDefaultGraph := schemaManager.DefaultGraph(); !hasDefaultGraph {
							slog.ErrorContext(ctx, "default graph not set in schema manager")
						} else if err := csDaemon.ReplayNodeChanges(ctx, lastNodeChangeID, func(change changestream.NodeChange) {
							if propertyJSON, err := json.Marshal(change.Properties.MapOrEmpty()); err != nil {
								slog.ErrorContext(ctx, "failed to marshal change properties", slog.Any("error", err))
							} else if kindIDs, err := schemaManager.MapKinds(ctx, change.Kinds); err != nil {
								slog.ErrorContext(ctx, "failed to map kinds", slog.Any("error", err))
							} else {
								switch change.Type() {
								case changestream.ChangeTypeAdd:
									if _, err := pool.Exec(ctx, changestream.InsertNodeFromChangeSQL, defaultGraph.ID, encoder.EncodeInt16(kindIDs), propertyJSON); err != nil {
										slog.ErrorContext(ctx, "failed to insert node to graph", slog.Any("error", err))
									}

								case changestream.ChangeTypeUpdate:
									if _, err := pool.Exec(ctx, changestream.UpdateNodeFromChangeSQL, change.TargetNodeID, encoder.EncodeInt16(kindIDs), propertyJSON); err != nil {
										slog.ErrorContext(ctx, "failed to insert node to graph", slog.Any("error", err))
									}

								case changestream.ChangeTypeRemove:
									slog.Error("not implemented")
								}
							}
						}); err != nil {
							slog.ErrorContext(ctx, "failed to replay node changes", slog.Any("error", err))
						}

						lastNodeChangeID = nextNotification.RevisionID

					case changestream.NotificationEdge:
						slog.Info("received notification to update edges to changelog revision", slog.Int64("rev_id", nextNotification.RevisionID))
						//lastEdgeChangeID = nextNotification.RevisionID
					}
				}

				slog.Info("notification channel closed")
			}()

			if err := test(ctx, csDaemon); err != nil {
				fmt.Printf("Test failed: %v\n", err)
				os.Exit(1)
			}

			slog.Info("Waiting for dawgs to finish")
			wg.Wait()
		}

		time.Sleep(time.Minute)
		done()
	}
}
