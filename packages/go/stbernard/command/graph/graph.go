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
	"io/fs"
	"log/slog"

	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/specterops/bloodhound/src/services/upload"
)

const (
	Name  = "graph"
	Usage = "Ingest valid collection files and transform graph data into a generic graph file"
)

type command struct {
	env     environment.Environment
	outfile string
	path    string
}

// Create new instance of command to capture given environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
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
func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ContinueOnError)

	cmd.StringVar(&s.outfile, "outfile", "", "destination path for generic graph file, default is {root}/tmp/graph.json")
	cmd.StringVar(&s.path, "path", "", "directory containing bloodhound collection files")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	if s.path == "" {
		cmd.Usage()

		return fmt.Errorf("path flag is required")
	}

	return nil
}

// Run generate command
func (s *command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if database, err := s.initializeDatabase(ctx); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	} else if ingestFilePaths, err := s.getIngestFilePaths(); err != nil {
		return fmt.Errorf("error getting ingest file paths from test directory %w", err)
	} else if err = ingestData(ctx, ingestFilePaths, database); err != nil {
		return fmt.Errorf("error ingesting data %w", err)
	} else if nodes, edges, err := getNodesAndEdges(ctx, database); err != nil {
		return fmt.Errorf("error retrieving nodes and edges from database %w", err)
	} else if graph, err := transformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph %w", err)
	} else if jsonBytes, err := json.MarshalIndent(generateIngestFile(graph), "", "  "); err != nil {
		return fmt.Errorf("error occurred while marshalling ingest file into bytes %w", err)
	} else if err := os.MkdirAll(filepath.Dir(s.outfile), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	} else {
		return os.WriteFile(s.outfile, jsonBytes, 0644)
	}
}

type Node struct {
	ID         string         `json:"id"`
	Kinds      []string       `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

type Terminal struct {
	MatchBy string `json:"match_by"`
	Value   string `json:"value"`
}

type Edge struct {
	Start      Terminal       `json:"start"`
	End        Terminal       `json:"end"`
	Kind       string         `json:"kind"`
	Properties map[string]any `json:"properties"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type IngestFile struct {
	Graph Graph `json:"graph"`
}

func ingestData(ctx context.Context, filepaths []string, database graph.Database) error {
	var errs []error

	schema, err := upload.LoadIngestSchema()
	if err != nil {
		return fmt.Errorf("error loading ingest schema %w", err)
	}

	for _, filepath := range filepaths {
		err := database.BatchOperation(ctx, func(batch graph.Batch) error {
			timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())

			readOpts := graphify.ReadOptions{IngestSchema: schema, FileType: model.FileTypeJson, ADCSEnabled: true}

			file, err := os.Open(filepath)
			if err != nil {
				return fmt.Errorf("error opening JSON file %s: %w", filepath, err)
			}
			defer file.Close()

			// ingest file into database
			err = graphify.ReadFileForIngest(timestampedBatch, file, readOpts)
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

func generateIngestFile(graph Graph) IngestFile {
	return IngestFile{
		Graph: graph,
	}
}

func transformGraph(nodes []*graph.Node, edges []*graph.Relationship) (Graph, error) {
	var graphNodes = make([]Node, 0, len(nodes))
	var graphEdges = make([]Edge, 0, len(edges))
	var isAZBase bool
	var isBase bool

	var nodeObjectIDs = make(map[graph.ID]string, len(nodes))

	for _, node := range nodes {
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
			return Graph{}, err
		}

		nodeObjectIDs[node.ID] = objectID

		graphNodes = append(graphNodes, Node{
			ID:         objectID,
			Kinds:      kinds,
			Properties: removeNullMapValues(node.Properties.Map),
		})
	}

	for _, edge := range edges {
		graphEdges = append(graphEdges, Edge{
			Start: Terminal{
				MatchBy: "id",
				Value:   nodeObjectIDs[edge.StartID],
			},
			End: Terminal{
				MatchBy: "id",
				Value:   nodeObjectIDs[edge.EndID],
			},
			Kind:       edge.Kind.String(),
			Properties: removeNullMapValues(edge.Properties.Map),
		})
	}

	return Graph{
		Nodes: graphNodes,
		Edges: graphEdges,
	}, nil
}

func removeNullMapValues[K comparable](m map[K]any) map[K]any {
	newMap := make(map[K]any)
	for k, v := range m {
		if v != nil {
			newMap[k] = convert(v)
		}
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
	var nodes []*graph.Node
	var edges []*graph.Relationship
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

func (s *command) getIngestFilePaths() ([]string, error) {
	var paths = make([]string, 0, 16)

	if err := filepath.Walk(s.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
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

func (s *command) initializeDatabase(ctx context.Context) (graph.Database, error) {
	var (
		connection = s.env[environment.PostgresConnectionVarName]
	)

	if pool, err := pg.NewPool(connection); err != nil {
		return nil, fmt.Errorf("error creating postgres connection %w", err)
	} else if database, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		ConnectionString:      connection,
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
