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
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	Usage = "Ingest valid collection files, analyzes ingested nodes and edges, and transform graph data into a generic graph file"
)

type Command struct {
	env     environment.Environment
	path    string
	service GraphService
}

// Create new instance of command to capture given environment
func Create(env environment.Environment, service GraphService) *Command {
	return &Command{
		env:     env,
		service: service,
	}
}

// Usage of command
func (s *Command) Usage() string {
	return Usage
}

// Name of command
func (s *Command) Name() string {
	return Name
}

// Parse command flags
func (s *Command) Parse() error {
	var (
		err error
		cmd = flag.NewFlagSet(Name, flag.ContinueOnError)
	)

	cmd.StringVar(&s.path, "path", "", "working directory containing a raw folder with ingest files")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n",
			Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	flagsIdx := 0

	for idx, arg := range os.Args {
		if strings.HasPrefix(arg, "-") {
			flagsIdx = idx
			break
		}
	}

	if err := cmd.Parse(os.Args[flagsIdx:]); err != nil {
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
	TeardownService(context.Context)
	InitializeService(context.Context, string, graph.Database) error
	Ingest(context.Context, *graphify.TimestampedBatch, io.ReadSeeker) error
	RunAnalysis(context.Context, graph.Database) error
}

type CommunityGraphService struct {
	db       database.Database
	readOpts graphify.ReadOptions
}

func NewCommunityGraphService() (*CommunityGraphService, error) {
	schema, err := upload.LoadIngestSchema()
	if err != nil {
		return nil, fmt.Errorf("error loading ingest schema: %w", err)
	}

	readOpts := graphify.ReadOptions{IngestSchema: schema, FileType: model.FileTypeJson}

	return &CommunityGraphService{readOpts: readOpts}, nil
}

func (s *CommunityGraphService) TeardownService(ctx context.Context) {
	if s.db != nil {
		err := s.db.Wipe(ctx)
		if err != nil {
			slog.Error("Failed to wipe database after command completion", slog.String("error", err.Error()))
		} else {
			slog.Info("Successfully wiped database")
		}
	}
}

func (s *CommunityGraphService) InitializeService(ctx context.Context, connection string, graphDB graph.Database) error {
	gormDB, err := database.OpenDatabase(connection)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	s.db = database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	if s.db != nil {
		err := s.db.Wipe(ctx)
		if err != nil {
			return fmt.Errorf("precommand wipe database: %w", err)
		} else {
			slog.Info("Successfully wiped database during initialization")
		}
	}

	if err := s.db.Migrate(ctx); err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	} else if err := migrations.NewGraphMigrator(graphDB).Migrate(ctx, graphschema.DefaultGraphSchema()); err != nil {
		return fmt.Errorf("error migrating graph schema: %w", err)
	} else if err = graphDB.SetDefaultGraph(ctx, graphschema.DefaultGraph()); err != nil {
		return fmt.Errorf("error setting default graph: %w", err)
	}

	return nil
}

func (s *CommunityGraphService) Ingest(ctx context.Context, batch *graphify.TimestampedBatch, reader io.ReadSeeker) error {
	return graphify.ReadFileForIngest(batch, reader, s.readOpts)
}

func (s *CommunityGraphService) RunAnalysis(ctx context.Context, graphDB graph.Database) error {
	return datapipe.RunAnalysisOperations(ctx, s.db, graphDB, config.Configuration{})
}

// Run generate command
func (s *Command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		s.service.TeardownService(ctx)
		cancel()
	}()

	if graphDB, err := initializeGraphDatabase(ctx, s.env[environment.PostgresConnectionVarName]); err != nil {
		return fmt.Errorf("error connecting to graphDB: %w", err)
	} else if err := s.service.InitializeService(ctx, s.env[environment.PostgresConnectionVarName], graphDB); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	} else if ingestFilePaths, err := s.getIngestFilePaths(); err != nil {
		return fmt.Errorf("error getting ingest file paths from directory: %w", err)
	} else if err = ingestData(ctx, s.service, ingestFilePaths, graphDB); err != nil {
		return fmt.Errorf("error ingesting data: %w", err)
	} else if nodes, edges, err := getNodesAndEdges(ctx, graphDB); err != nil {
		return fmt.Errorf("error retrieving nodes and edges from database: %w", err)
	} else if graph, err := transformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph: %w", err)
	} else if err := writeGraphToFile(&graph, s.path, "ingest", "ingested.json"); err != nil {
		return fmt.Errorf("error writing graph to file: %w", err)
	} else if err := s.service.RunAnalysis(ctx, graphDB); err != nil {
		return fmt.Errorf("error running analysis: %w", err)
	} else if nodes, edges, err := getNodesAndEdges(ctx, graphDB); err != nil {
		return fmt.Errorf("error getting nodes and edges: %w", err)
	} else if graph, err := transformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph: %w", err)
	} else if err := writeGraphToFile(&graph, s.path, "analysis", "analyzed.json"); err != nil {
		return fmt.Errorf("error writing graph to file: %w", err)
	}

	return nil
}

func writeGraphToFile(graph *generic.Graph, root, folder, fileName string) error {
	outputDir := filepath.Join(root, folder)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("making output directory %s: %w", outputDir, err)
	}
	outputFile := filepath.Join(outputDir, fileName)

	if jsonBytes, err := json.MarshalIndent(generateGenericGraphFile(graph), "", "  "); err != nil {
		return fmt.Errorf("error occurred while marshalling ingest file into bytes: %w", err)
	} else if err := os.MkdirAll(root, 0755); err != nil {
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
			switch kind {
			case ad.Entity:
				isBase = true
			case azure.Entity:
				isAZBase = true
			default:
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
		nodes   []*graph.Node
		edges   []*graph.Relationship
		nodeIDs []graph.ID
	)

	if err := database.ReadTransaction(ctx, func(tx graph.Transaction) error {
		err := tx.Nodes().Filter(
			query.And(
				query.Not(query.Kind(query.Node(), common.MigrationData)),
				query.IsNotNull(query.NodeProperty(common.ObjectID.String())),
			),
		).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				nodes = append(nodes, node)
				nodeIDs = append(nodeIDs, node.ID)
			}
			return cursor.Error()
		})
		if err != nil {
			return fmt.Errorf("error fetching nodes: %w", err)
		}
		err = tx.Relationships().Filter(
			query.And(
				query.InIDs(query.StartID(), nodeIDs...),
				query.InIDs(query.EndID(), nodeIDs...),
			),
		).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for edge := range cursor.Chan() {
				edges = append(edges, edge)
			}
			return cursor.Error()
		})
		if err != nil {
			return fmt.Errorf("error fetching relationships: %w", err)
		}

		return nil
	}); err != nil {
		return nodes, edges, fmt.Errorf("error occurred reading the database: %w", err)
	} else {
		return nodes, edges, nil
	}
}

func initializeGraphDatabase(ctx context.Context, postgresConnection string) (graph.Database, error) {
	if pool, err := pg.NewPool(postgresConnection); err != nil {
		return nil, fmt.Errorf("error creating postgres connection: %w", err)
	} else if database, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		ConnectionString:      postgresConnection,
		Pool:                  pool,
	}); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	} else {
		return database, nil
	}

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
			return fmt.Errorf("unrecoverable error occurred during batch operation: %w", err)
		}
	}

	if len(errs) > 0 {
		var errStrings []string
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		slog.Warn("errors occurred while ingesting files", slog.Any("errors", errStrings))
	}

	return nil
}

func (s *Command) getIngestFilePaths() ([]string, error) {
	var paths = make([]string, 0, 16)

	if err := filepath.Walk(filepath.Join(s.path, "raw"), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking filepath: %w", err)
		}

		if !info.IsDir() && filepath.Ext(path) == ".json" {
			paths = append(paths, path)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error getting files from directory: %w", err)
	}

	return paths, nil
}
