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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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

// Run generate command
func (s *command) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if graph, err := unmarshalGraph(s.infile); err != nil {
		return fmt.Errorf("error unmarshalling graph: %w", err)
	}
	// TODO: continue logic, use WriteGraphToDatabase

	return nil
}

func unmarshalGraph(filePath string) (shared.Graph, error) {
	var graphFile shared.GenericGraphFile

	if bytes, err := os.ReadFile(filePath); err != nil {
		return graphFile.Graph, err
	} else if err := json.Unmarshal(bytes, graphFile); err != nil {
		return graphFile.Graph, err
	}

	return graphFile.Graph, nil
}
