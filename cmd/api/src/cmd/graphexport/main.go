// Copyright 2026 Specter Ops, Inc.
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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/neo4j"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/size"
)

type document struct {
	Metadata ein.GenericMetadata `json:"metadata"`
	Graph    graphData           `json:"graph"`
}

type graphData struct {
	Nodes []ein.GenericNode `json:"nodes"`
	Edges []exportedEdge    `json:"edges"`
}

// ID endpoints must omit property_matchers entirely to satisfy the schema.
type exportedEndpoint struct {
	MatchBy string `json:"match_by"`
	Value   string `json:"value"`
}

type exportedEdge struct {
	Start      exportedEndpoint `json:"start"`
	End        exportedEndpoint `json:"end"`
	Kind       string           `json:"kind"`
	Properties map[string]any   `json:"properties"`
}

func main() {
	var (
		configurationPath = flag.String("config", "", "BloodHound configuration file; BHE_* environment overrides also apply")
		sourceKind        = flag.String("source-kind", azure.Entity.String(), "Export nodes of this source kind and their relationships")
		outputPath        = flag.String("output", "-", "Output file, or - for stdout")
		validate          = flag.Bool("validate", true, "Enforce the ingest schema; disable for snapshots containing reserved post-processing tags")
		ctx, cancel       = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	)
	defer cancel()
	flag.Parse()
	if err := run(ctx, *configurationPath, *sourceKind, *outputPath, *validate); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, configurationPath, sourceKind, outputPath string, validate bool) error {
	var (
		configuration config.Configuration
		database      graph.Database
		result        document
		content       []byte
		err           error
	)
	if sourceKind == "" {
		return fmt.Errorf("source-kind must not be empty")
	}
	if configuration, err = config.GetConfiguration(configurationPath, config.NewDefaultConfiguration); err != nil {
		return err
	}
	if database, err = connectGraph(ctx, configuration); err != nil {
		return fmt.Errorf("connect graph: %w", err)
	}
	defer database.Close(context.Background())
	// Select the existing graph namespace without asserting or modifying schema.
	if err = database.SetDefaultGraph(ctx, graphschema.DefaultGraph()); err != nil {
		return fmt.Errorf("select default graph: %w", err)
	}
	if result, err = exportGraph(ctx, database, graph.StringKind(sourceKind)); err != nil {
		return err
	}
	if content, err = json.MarshalIndent(result, "", "  "); err != nil {
		return fmt.Errorf("encode graph: %w", err)
	}
	// Validate the actual JSON representation, including serialized property types.
	if validate {
		if err = validateDocument(content); err != nil {
			return err
		}
	}
	content = append(content, '\n')
	if outputPath == "-" {
		_, err = os.Stdout.Write(content)
		return err
	}
	return os.WriteFile(outputPath, content, 0o600)
}

func validateDocument(content []byte) error {
	var (
		schema, err = payload.LoadSchema()
		validator   payload.Validator
	)
	if err != nil {
		return err
	}
	validator = payload.NewValidator(bytes.NewReader(content), schema)
	if _, report, err := validator.ParseAndValidate(); err != nil {
		return fmt.Errorf("graph cannot be represented as valid OpenGraph: %w; diagnostics: %+v", err, report)
	}
	return nil
}

func exportGraph(ctx context.Context, database graph.Database, sourceKind graph.Kind) (document, error) {
	var (
		result = document{
			Metadata: ein.GenericMetadata{SourceKind: sourceKind.String()},
			Graph:    graphData{Nodes: []ein.GenericNode{}, Edges: []exportedEdge{}},
		}
		identifiers     = map[graph.ID]string{}
		seenIdentifiers = map[string]struct{}{}
		err             = database.ReadTransaction(ctx, func(transaction graph.Transaction) error {
			if err := transaction.Nodes().Filter(query.Kind(query.Node(), sourceKind)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for node := range cursor.Chan() {
					exportedNode, err := exportNode(node, sourceKind)
					if err != nil {
						return err
					}
					if _, exists := seenIdentifiers[exportedNode.ID]; exists {
						return fmt.Errorf("duplicate objectid %q in source %s", exportedNode.ID, sourceKind)
					}
					seenIdentifiers[exportedNode.ID] = struct{}{}
					identifiers[node.ID] = exportedNode.ID
					result.Graph.Nodes = append(result.Graph.Nodes, exportedNode)
				}
				return cursor.Error()
			}); err != nil {
				return err
			}
			return transaction.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
				for relationship := range cursor.Chan() {
					var (
						start, hasStart = identifiers[relationship.StartID]
						end, hasEnd     = identifiers[relationship.EndID]
					)
					if !hasStart && !hasEnd {
						continue
					}
					if !hasStart || !hasEnd {
						return fmt.Errorf("relationship %s crosses source-kind boundary; cannot export a self-contained graph", relationship.ID)
					}
					result.Graph.Edges = append(result.Graph.Edges, exportedEdge{
						Start:      exportedEndpoint{MatchBy: "id", Value: start},
						End:        exportedEndpoint{MatchBy: "id", Value: end},
						Kind:       relationship.Kind.String(),
						Properties: relationship.Properties.Map,
					})
				}
				return cursor.Error()
			})
		})
	)
	if err != nil {
		return document{}, err
	}
	slices.SortFunc(result.Graph.Nodes, func(left, right ein.GenericNode) int {
		return bytes.Compare([]byte(left.ID), []byte(right.ID))
	})
	// JSON map keys are sorted by encoding/json. Full edge serialization also
	// orders parallel edges by properties without discarding multiplicity.
	slices.SortFunc(result.Graph.Edges, func(left, right exportedEdge) int {
		leftJSON, _ := json.Marshal(left)
		rightJSON, _ := json.Marshal(right)
		return bytes.Compare(leftJSON, rightJSON)
	})
	return result, nil
}

func exportNode(node *graph.Node, sourceKind graph.Kind) (ein.GenericNode, error) {
	var (
		objectID, err = node.Properties.Get(common.ObjectID.String()).String()
		properties    = make(map[string]any, len(node.Properties.Map))
		kinds         = node.Kinds.Strings()
		primaryKind   string
	)
	if err != nil || objectID == "" {
		return ein.GenericNode{}, fmt.Errorf("node %s has no usable objectid", node.ID)
	}
	for key, value := range node.Properties.Map {
		if key != common.ObjectID.String() {
			properties[key] = value
		}
	}
	slices.Sort(kinds)
	// OpenGraph overwrites primarykind with kinds[0]. Preserve an existing
	// primary kind; otherwise prefer a concrete kind over the source kind.
	if value, ok := properties[common.PrimaryKind.String()]; ok {
		primaryKind = fmt.Sprint(value)
		if !slices.Contains(kinds, primaryKind) {
			return ein.GenericNode{}, fmt.Errorf("node %q has primarykind %q absent from its kinds", objectID, primaryKind)
		}
	} else {
		primaryKind = sourceKind.String()
		for _, kind := range kinds {
			if kind != sourceKind.String() {
				primaryKind = kind
				break
			}
		}
	}
	if index := slices.Index(kinds, primaryKind); index > 0 {
		kinds = append([]string{primaryKind}, append(kinds[:index:index], kinds[index+1:]...)...)
	}
	return ein.GenericNode{ID: objectID, Kinds: kinds, Properties: properties}, nil
}

// Connect directly to the configured driver. The server connection helper also
// resolves migration state and may create a SQL table; an exporter must not.
func connectGraph(ctx context.Context, configuration config.Configuration) (graph.Database, error) {
	var (
		connectionString string
		pool             *pgxpool.Pool
		database         graph.Database
		err              error
	)
	switch configuration.GraphDriver {
	case neo4j.DriverName:
		connectionString = configuration.Neo4J.Neo4JConnectionString()
	case pg.DriverName:
		connectionString = configuration.Database.PostgreSQLConnectionString()
		if pool, err = dbpool.NewDawgsPool(configuration.Database); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported graph_driver %q", configuration.GraphDriver)
	}
	database, err = dawgs.Open(ctx, configuration.GraphDriver, dawgs.Config{
		ConnectionString:      connectionString,
		Pool:                  pool,
		GraphQueryMemoryLimit: size.Size(configuration.GraphQueryMemoryLimit) * size.Gibibyte,
	})
	if err != nil && pool != nil {
		pool.Close()
	}
	return database, err
}
