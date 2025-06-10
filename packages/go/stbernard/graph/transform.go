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
	"log/slog"
	"strconv"
	"strings"

	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/slicesext"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/specterops/bloodhound/src/services/upload"
)

const (
	Name  = "graph"
	Usage = "Ingest test files and transform graph data into arrows.app compatible JSON files"
)

type command struct {
	env  environment.Environment
	path string
	root string
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
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)
	workspacePaths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("failed to find workspace paths: %w", err)
	}
	s.root = workspacePaths.Root

	path := cmd.String("path", workspacePaths.Root, "destination path for arrows.json file, default is root")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	if *path != "" {
		s.path = *path
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
		return fmt.Errorf("error getting ingest file paths from test directory %v", err)
	} else if err = ingestData(ctx, ingestFilePaths, database); err != nil {
		return fmt.Errorf("error ingesting data %v", err)
	} else if nodes, edges, err := getNodesAndEdges(database); err != nil {
		return fmt.Errorf("error retrieving nodes and edges from database %w", err)
	} else if arrows, err := transformToArrows(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to arrows.app format %w", err)
	} else if jsonBytes, err := json.MarshalIndent(arrows, "", "  "); err != nil {
		return fmt.Errorf("error occurred while marshalling arrows into bytes %w", err)
	} else {
		outPath := filepath.Join(s.path, "arrows.json")
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		return os.WriteFile(outPath, jsonBytes, 0o644)
	}
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Node struct {
	ID         string            `json:"id"`
	Position   Position          `json:"position"`
	Caption    string            `json:"caption"` // name of node -- if not, use object_id
	Label      []string          `json:"label"`   // kinds
	Properties map[string]string `json:"properties"`
}

type Relationship struct {
	ID         string            `json:"id"`
	Label      string            `json:"label"` // kind
	From       string            `json:"fromId"`
	To         string            `json:"toId"`
	Properties map[string]string `json:"properties"`
}

type Arrows struct {
	Nodes         []Node         `json:"nodes"`
	Relationships []Relationship `json:"relationships"`
}

func ingestData(ctx context.Context, filepaths []string, database graph.Database) error {
	var errs []error

	schema, err := upload.LoadIngestSchema()
	if err != nil {
		return fmt.Errorf("error loading ingest schema %v", err)
	}

	for _, filepath := range filepaths {
		err := database.BatchOperation(ctx, func(batch graph.Batch) error {
			timestampedBatch := graphify.NewTimestampedBatch(batch, time.Now().UTC())

			readOpts := graphify.ReadOptions{IngestSchema: schema, FileType: model.FileTypeJson, ADCSEnabled: true}

			file, err := os.Open(filepath)
			if err != nil {
				return fmt.Errorf("error opening JSON file %s: %v", filepath, err)
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

func transformToArrows(nodes []*graph.Node, edges []*graph.Relationship) (Arrows, error) {
	var arrowNodes = make([]Node, 0, len(nodes))
	var arrowEdges = make([]Relationship, 0, len(edges))

	for _, node := range nodes {
		name, err := node.Properties.Get(common.Name.String()).String()
		if err != nil || name == "" {
			name, err = node.Properties.Get(common.ObjectID.String()).String()
			if err != nil {
				return Arrows{}, err
			}
		}

		var labels = make([]string, 0, 4)
		for _, kind := range node.Kinds {
			labels = append(labels, kind.String())
		}

		arrowNodes = append(arrowNodes, Node{
			ID: node.ID.String(),
			Position: Position{
				X: 0,
				Y: 0,
			},
			Caption:    name,
			Label:      labels,
			Properties: convertProperties(node.Properties.Map),
		})
	}

	for _, edge := range edges {
		arrowEdges = append(arrowEdges, Relationship{
			ID:         edge.ID.String(),
			From:       edge.StartID.String(),
			To:         edge.EndID.String(),
			Label:      edge.Kind.String(),
			Properties: convertProperties(edge.Properties.Map),
		})
	}

	return Arrows{
		Nodes:         arrowNodes,
		Relationships: arrowEdges,
	}, nil
}

func convertProperties(input map[string]any) map[string]string {
	output := make(map[string]string)
	for key, value := range input {
		output[key] = convertProperty(value)
	}
	return output
}

func convertProperty(input any) string {
   switch v := input.(type) {
		case string:
			return v
		case int:
			return strconv.Itoa(v)
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(v)
		case []any:
			return strings.Join(slicesext.Map(v, convertProperty), ",")
		case nil:
			return "null"
		default:
			slog.Warn("unknown type encountered", slog.String("type", fmt.Sprintf("%T", v)))
			return ""
		}
}

func getNodesAndEdges(database graph.Database) ([]*graph.Node, []*graph.Relationship, error) {
	var nodes []*graph.Node
	var edges []*graph.Relationship
	if err := database.ReadTransaction(context.TODO(), func(tx graph.Transaction) error {
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
	var (
		testFilePath    = filepath.Join(strings.Split("cmd/api/src/test/fixtures/fixtures/v6/ingest", "/")...)
		ingestDirectory = filepath.Join(s.root, testFilePath)
	)
	if ingestFiles, err := os.ReadDir(filepath.Join(s.root, testFilePath)); err != nil {
		return []string{}, err
	} else {
		var paths = make([]string, 0, len(ingestFiles))
		for _, path := range ingestFiles {
			paths = append(paths, filepath.Join(ingestDirectory, path.Name()))
		}
		return paths, nil
	}
}

func (s *command) initializeDatabase(ctx context.Context) (graph.Database, error) {
	var (
		connection = s.env[environment.PostgresConnectionVarName]
	)

	if pool, err := pg.NewPool(connection); err != nil {
		return nil, err
	} else if database, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		ConnectionString:      connection,
		Pool:                  pool,
	}); err != nil {
		return nil, err
	} else if err = migrations.NewGraphMigrator(database).Migrate(ctx, graphschema.DefaultGraphSchema()); err != nil {
		return nil, err
	} else if err = database.SetDefaultGraph(ctx, graphschema.DefaultGraph()); err != nil {
		return nil, err
	} else {
		return database, nil
	}
}
