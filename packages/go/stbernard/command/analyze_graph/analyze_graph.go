// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package analyzegraph

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/shared"
)

const (
	Name  = "analyze-graph"
	Usage = "Run analysis on a generic graph (from the graph command) and write new graph to file"
)

type command struct {
	env     environment.Environment
	outfile string
	infile  string
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

	cmd.StringVar(&s.outfile, "outfile", "/tmp/graph.json", "destination path for the analyzed graph file, default is {root}/tmp/graph.json")
	cmd.StringVar(&s.infile, "infile", "", "file path for generic graph file")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	if s.infile == "" {
		cmd.Usage()

		return fmt.Errorf("path flag is required")
	}

	s.infile, err = filepath.Abs(s.infile)
	if err != nil {
		return fmt.Errorf("could not convert path to absolute path: %w", err)
	}

	return nil
}

// This is because fs.FS Open does not like abosulte paths prefixed with root
func removeRoot(path string) string {
	trimmed, _ := strings.CutPrefix(path, "/")

	return trimmed
}

// Run generate command
func (s *command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if graphDB, err := shared.InitializeGraphDatabase(ctx, s.env[environment.PostgresConnectionVarName]); err != nil {
		return fmt.Errorf("error initializing graph database: %w", err)
	} else if db, err := shared.InitializeDatabase(ctx, s.env[environment.PostgresConnectionVarName]); err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	} else if graph, err := generic.LoadGraphFromFile(os.DirFS("/"), removeRoot(s.infile)); err != nil {
		return fmt.Errorf("error loading graph from file: %w", err)
	} else if err := generic.WriteGraphToDatabase(graphDB, &graph); err != nil {
		return fmt.Errorf("error writing graph to database: %w", err)
	} else if err := datapipe.RunAnalysisOperations(ctx, db, graphDB, config.Configuration{}); err != nil {
		return fmt.Errorf("error running analysis: %w", err)
	} else if nodes, edges, err := shared.GetNodesAndEdges(ctx, graphDB); err != nil {
		return fmt.Errorf("error getting nodes and edges: %w", err)
	} else if graph, err := shared.TransformGraph(nodes, edges); err != nil {
		return fmt.Errorf("error transforming nodes and edges to graph: %w", err)
	} else if err := shared.WriteGraphToFile(&graph, s.outfile); err != nil {
		return fmt.Errorf("error writing graph to file: %w", err)
	} else if err := db.Wipe(ctx); err != nil {
		return fmt.Errorf("error wiping db: %w", err)
	}

	return nil
}
