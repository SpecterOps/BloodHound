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

package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/size"
)

func WriteGraphToFile(graph *generic.Graph, path string) error {
	if jsonBytes, err := json.MarshalIndent(generateGenericGraphFile(graph), "", "  "); err != nil {
		return fmt.Errorf("error occurred while marshalling ingest file into bytes %w", err)
	} else if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	} else {
		return os.WriteFile(path, jsonBytes, 0644)
	}
}

func generateGenericGraphFile(graph *generic.Graph) generic.GenericObject {
	return generic.GenericObject{
		Graph: *graph,
	}
}

func TransformGraph(nodes []*graph.Node, edges []*graph.Relationship) (generic.Graph, error) {
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

func GetNodesAndEdges(ctx context.Context, database graph.Database) ([]*graph.Node, []*graph.Relationship, error) {
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

func InitializeGraphDatabase(ctx context.Context, postgresConnection string) (graph.Database, error) {

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

func InitializeDatabase(ctx context.Context, connection string) (database.Database, error) {
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
