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
package graph

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"

	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/size"
)

const (
	Name  = "graph"
	Usage = "Ingest valid collection files and transform graph data into a generic graph file"
)

func main() {
	env := environment.NewEnvironment()
	command := Create(env, BHCEGraphService{})

	if err := command.Parse(); err != nil {
		log.Fatal("Failed to parse CLI args: %w", err)
	} else if err := command.Run(); err != nil {
		log.Fatal("Failed to run command: %w", err)
	}
}

type command struct {
	env     environment.Environment
	outpath string
	path    string
	service GraphService
}

// Create new instance of command to capture given environment
func Create(env environment.Environment, service GraphService) *command {
	return &command{
		env:     env,
		service: service,
	}
}

// Usage of command
func (s *command) Usage() string {
	return Usage
}

// Name of command
func (s *command) Name() string {
	return Name
}

// Parse command flags
func (s *command) Parse() error {
	var err error
	cmd := flag.NewFlagSet(Name, flag.ContinueOnError)

	cmd.StringVar(&s.outpath, "outpath", "/tmp/", "destination path for generic graph files, default is {root}/tmp/")
	cmd.StringVar(&s.path, "path", "", "directory containing bloodhound collection files")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	if s.path == "" {
		cmd.Usage()

		return fmt.Errorf("path flag is required")
	}

	s.path, err = filepath.Abs(s.path)
	if err != nil {
		return fmt.Errorf("could not convert path to absolute path: %w", err)
	}

	return nil
}

type GraphService interface {
	Ingest(context.Context, *graphify.TimestampedBatch, io.ReadSeeker) error
	RunAnalysis(context.Context, Database, graph.Database) error
}

type BHCEGraphService struct{}

func (s *BHCEGraphService) Ingest(ctx context.Context, batch *graphify.TimestampedBatch, reader io.ReadSeeker) error {
	schema, err := upload.LoadIngestSchema()
	if err != nil {
		return fmt.Errorf("error loading ingest schema %w", err)
	}

	readOpts := graphify.ReadOptions{IngestSchema: schema, FileType: model.FileTypeJson, ADCSEnabled: true}

	// ingest file into database
	return graphify.ReadFileForIngest(batch, reader, readOpts)
}

func (s *BHCEGraphService) RunAnalysis(ctx context.Context, db Database, graphDB graph.Database) error {

}

// Run generate command
func (s *command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if graphDB, err := initializeGraphDatabase(ctx, s.env[environment.PostgresConnectionVarName]); err != nil {
		return fmt.Errorf("error connecting to graphDB: %w", err)
	} else if db, err := initializeDatabase(ctx, s.env[environment.PostgresConnectionVarName]); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	} else if ingestFilePaths, err := s.getIngestFilePaths(); err != nil {
		return fmt.Errorf("error getting ingest file paths from directory %w", err)
	} else if err = ingestData(ctx, s.service, ingestFilePaths, graphDB); err != nil {
		return fmt.Errorf("error ingesting data %w", err)
	} else if nodes, edges, err := getNodesAndEdges(ctx, graphDB); err != nil {
		return fmt.Errorf("error retrieving nodes and edges from database %w", err)
	} else if graph, err := transformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph %w", err)
	} else if err := writeGraphToFile(&graph, s.outpath, "ingested.json"); err != nil {
		return fmt.Errorf("error writing graph to file %w", err)
	} else if err := datapipe.RunAnalysisOperations(ctx, db, graphDB, config.Configuration{}); err != nil {
		return fmt.Errorf("error running analysis: %w", err)
	} else if nodes, edges, err := getNodesAndEdges(ctx, graphDB); err != nil {
		return fmt.Errorf("error getting nodes and edges: %w", err)
	} else if graph, err := transformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph: %w", err)
	} else if err := writeGraphToFile(&graph, s.outpath, "analyzed.json"); err != nil {
		return fmt.Errorf("error writing graph to file: %w", err)
	} else if err := db.Wipe(ctx); err != nil {
		return fmt.Errorf("error wiping db: %w", err)
	}

	return nil
}

func writeGraphToFile(graph *generic.Graph, folder string, fileName string) error {
	outputFile := filepath.Join(folder, fileName)

	if jsonBytes, err := json.MarshalIndent(generateGenericGraphFile(graph), "", "  "); err != nil {
		return fmt.Errorf("error occurred while marshalling ingest file into bytes %w", err)
	} else if err := os.MkdirAll(folder, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	} else {
		return os.WriteFile(outputFile, jsonBytes, 0644)
	}
}

func generateGenericGraphFile(graph *generic.Graph) generic.GenericObject {
	return generic.GenericObject{
		Graph: *graph,
	}
}

func transformGraph(nodes []*graph.Node, edges []*graph.Relationship) (generic.Graph, error) {
	var (
		isAZBase bool
		isBase   bool

		graphNodes    = make([]generic.Node, 0, len(nodes))
		graphEdges    = make([]generic.Edge, 0, len(edges))
		nodeObjectIDs = make(map[graph.ID]string, len(nodes))
	)

	for _, node := range nodes {
		isAZBase = false
		isBase = false

		var kinds = make([]string, 0, len(node.Kinds))

		for _, kind := range node.Kinds {
			if kind == ad.Entity {
				isBase = true
			} else if kind == azure.Entity {
				isAZBase = true
			} else {
				kinds = append(kinds, kind.String())
			}
		}

		if isBase {
			kinds = append(kinds, ad.Entity.String())
		} else if isAZBase {
			kinds = append(kinds, azure.Entity.String())
		}

		objectID, err := node.Properties.Get(common.ObjectID.String()).String()
		if err != nil {
			return generic.Graph{}, err
		}

		nodeObjectIDs[node.ID] = objectID

		graphNodes = append(graphNodes, generic.Node{
			ID:         objectID,
			Kinds:      kinds,
			Properties: removeNullMapValues(node.Properties.Map),
		})
	}

	for _, edge := range edges {
		graphEdges = append(graphEdges, generic.Edge{
			Start: generic.Terminal{
				MatchBy: "id",
				Value:   nodeObjectIDs[edge.StartID],
			},
			End: generic.Terminal{
				MatchBy: "id",
				Value:   nodeObjectIDs[edge.EndID],
			},
			Kind:       edge.Kind.String(),
			Properties: removeNullMapValues(edge.Properties.Map),
		})
	}

	return generic.Graph{
		Nodes: graphNodes,
		Edges: graphEdges,
	}, nil
}

func removeNullMapValues[K comparable](m map[K]any) map[K]any {
	newMap := make(map[K]any, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}

		newMap[k] = convert(v)
	}
	return newMap
}

func convert(val any) any {
	switch v := val.(type) {
	case []any:
		return removeNullSliceValues(v)
	case map[any]any:
		return removeNullMapValues(v)
	default:
		return val
	}
}

func removeNullSliceValues(l []any) []any {
	newSlice := make([]any, 0, len(l))

	for _, val := range l {
		if val == nil {
			continue
		}

		newSlice = append(newSlice, convert(val))
	}

	return newSlice
}

func getNodesAndEdges(ctx context.Context, database graph.Database) ([]*graph.Node, []*graph.Relationship, error) {
	var (
		nodes []*graph.Node
		edges []*graph.Relationship
	)

	if err := database.ReadTransaction(ctx, func(tx graph.Transaction) error {
		err := tx.Nodes().Filter(
			query.Not(query.Kind(query.Node(), common.MigrationData)),
		).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				nodes = append(nodes, node)
			}
			return cursor.Error()
		})
		if err != nil {
			return fmt.Errorf("error fetching nodes %w", err)
		}
		err = tx.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for edge := range cursor.Chan() {
				edges = append(edges, edge)
			}
			return cursor.Error()
		})
		if err != nil {
			return fmt.Errorf("error fetching relationships %w", err)
		}

		return nil
	}); err != nil {
		return nodes, edges, fmt.Errorf("error occurred reading the database %w", err)
	} else {
		return nodes, edges, nil
	}
}

func initializeGraphDatabase(ctx context.Context, postgresConnection string) (graph.Database, error) {

	if pool, err := pg.NewPool(postgresConnection); err != nil {
		return nil, fmt.Errorf("error creating postgres connection %w", err)
	} else if database, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		ConnectionString:      postgresConnection,
		Pool:                  pool,
	}); err != nil {
		return nil, fmt.Errorf("error connecting to database %w", err)
	} else if err = migrations.NewGraphMigrator(database).Migrate(ctx, graphschema.DefaultGraphSchema()); err != nil {
		return nil, fmt.Errorf("error migrating graph %w", err)
	} else if err = database.SetDefaultGraph(ctx, graphschema.DefaultGraph()); err != nil {
		return nil, fmt.Errorf("error setting default graph %w", err)
	} else {
		return database, nil
	}

}

func initializeDatabase(ctx context.Context, connection string) (database.Database, error) {
	var db database.Database

	if gormDB, err := database.OpenDatabase(connection); err != nil {
		return db, fmt.Errorf("error opening database %w", err)
	} else {
		db = database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())
	}

	if err := db.Migrate(ctx); err != nil {
		return db, fmt.Errorf("error migrating database %w", err)
	}

	return db, nil
}

func ingestData(ctx context.Context, service GraphService, filepaths []string, database graph.Database) error {
	var errs []error

	for _, filepath := range filepaths {
		err := database.BatchOperation(ctx, func(batch graph.Batch) error {
			timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())

			file, err := os.Open(filepath)
			if err != nil {
				return fmt.Errorf("error opening JSON file %s: %w", filepath, err)
			}
			defer file.Close()

			// ingest file into database
			err = service.Ingest(ctx, timestampedBatch, file)
			if err != nil {
				errs = append(errs, fmt.Errorf("error ingesting file %s: %w", filepath, err))
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("unrecoverable error occurred during batch operation %w", err)
		}
	}

	if len(errs) > 0 {
		var errStrings []string
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		slog.Warn("errors occurred while ingesting files", "errors", errStrings)
	}

	return nil
}

func (s *command) getIngestFilePaths() ([]string, error) {
	var paths = make([]string, 0, 16)

	if err := filepath.Walk(s.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking filepath %w", err)
		}

		if !info.IsDir() && filepath.Ext(path) == ".json" {
			paths = append(paths, path)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error getting files from directory %w", err)
	}

	return paths, nil
}
