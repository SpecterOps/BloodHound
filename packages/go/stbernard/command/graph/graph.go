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
	"flag"
	"fmt"
	"io/fs"
	"log/slog"

	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/shared"
	"github.com/specterops/dawgs/graph"
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
	var err error
	cmd := flag.NewFlagSet(Name, flag.ContinueOnError)

	cmd.StringVar(&s.outfile, "outfile", "/tmp/graph.json", "destination path for generic graph file, default is {root}/tmp/graph.json")
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

	s.path, err = filepath.Abs(s.path)
	if err != nil {
		return fmt.Errorf("could not convert path to absolute path: %w", err)
	}

	return nil
}

// Run generate command
func (s *command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if database, err := shared.InitializeGraphDatabase(ctx, s.env[environment.PostgresConnectionVarName]); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	} else if ingestFilePaths, err := s.getIngestFilePaths(); err != nil {
		return fmt.Errorf("error getting ingest file paths from directory %w", err)
	} else if err = ingestData(ctx, ingestFilePaths, database); err != nil {
		return fmt.Errorf("error ingesting data %w", err)
	} else if nodes, edges, err := shared.GetNodesAndEdges(ctx, database); err != nil {
		return fmt.Errorf("error retrieving nodes and edges from database %w", err)
	} else if graph, err := shared.TransformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph %w", err)
	} else if err := shared.WriteGraphToFile(&graph, s.outfile); err != nil {
		return fmt.Errorf("error writing graph to file %w", err)
	}

	return nil
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
