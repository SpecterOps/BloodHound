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
//go:build integration

package endpoint_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/test"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nodeCount    = 100
	nodeCountMod = nodeCount / (nodeCount * .1)
)

var nodeKind = graph.StringKind("NodeKind1")

func newNodeData(index int) (string, *graph.Properties) {
	id := fmt.Sprintf("node-%d-%d", time.Now().UnixNano(), index)

	return id, graph.AsProperties(map[string]any{
		"objectid":   id,
		"name":       fmt.Sprintf("Node-%d", index),
		"score":      float64(index * 10),
		"is_active":  index%2 == 0, // Alternating boolean
		"created_at": time.Now().Format(time.RFC3339),
		"category":   []string{"cat-a", "cat-b"}[index%2],
	})
}

func buildIngestibleRelationshipsFromExistingNodes(ctx context.Context, db graph.Database) ([]ein.IngestibleRelationship, error) {
	var (
		sources []*graph.Node
		targets []*graph.Node
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		fetched, err := ops.FetchNodes(
			tx.Nodes().Filter(
				query.And(
					query.Kind(query.Node(), nodeKind),
					query.Exists(query.NodeProperty("name")),
					query.Exists(query.NodeProperty("is_active")),
					query.Exists(query.NodeProperty("score")),
				),
			).OrderBy(
				query.Order(query.NodeID(), query.Descending()),
			).Limit(nodeCount),
		)

		if err != nil {
			return err
		}

		sources = fetched[:len(fetched)/2]
		targets = fetched[len(fetched)/2:]

		return nil
	}); err != nil {
		return nil, err
	}

	var ingestEntries []ein.IngestibleRelationship

	for idx := range sources {
		var ingestEntry ein.IngestibleRelationship

		if idx%nodeCountMod == 0 {
			ingestEntry = ein.IngestibleRelationship{
				Source: ein.IngestibleEndpoint{
					Kind:    nodeKind,
					MatchBy: ein.MatchByProperty,
					Matchers: []ein.MatchExpression{
						{
							Key:      "name",
							Operator: ein.OperatorEquals,
							Value:    sources[idx].Properties.GetOrDefault("name", "NO NAME").Any(),
						},
					},
				},

				Target: ein.IngestibleEndpoint{
					Kind:    nodeKind,
					MatchBy: ein.MatchByProperty,
					Matchers: []ein.MatchExpression{
						{
							Key:      "name",
							Operator: ein.OperatorEquals,
							Value:    targets[idx].Properties.GetOrDefault("name", "NO NAME").Any(),
						},
						{
							Key:      "is_active",
							Operator: ein.OperatorEquals,
							Value:    targets[idx].Properties.GetOrDefault("is_active", "false").Any(),
						},
						{
							Key:      "score",
							Operator: ein.OperatorEquals,
							Value:    targets[idx].Properties.GetOrDefault("score", "NO SCORE").Any(),
						},
					},
				},
			}
		} else {
			ingestEntry = ein.IngestibleRelationship{
				Source: ein.IngestibleEndpoint{
					Kind:    nodeKind,
					MatchBy: ein.MatchByID,
					Value:   sources[idx].Properties.GetOrDefault("objectid", "NO ID").Any().(string),
				},

				Target: ein.IngestibleEndpoint{
					Kind:    nodeKind,
					MatchBy: ein.MatchByID,
					Value:   targets[idx].Properties.GetOrDefault("objectid", "NO ID").Any().(string),
				},
			}
		}

		ingestEntries = append(ingestEntries, ingestEntry)
	}

	return ingestEntries, nil
}

func Test_FetchAllNodesByMatchers(t *testing.T) {
	var (
		suite         = test.SetupIntegrationTestSuite(t)
		nodeObjectIDs = map[string]struct{}{}
	)

	require.NoError(t, suite.GraphDB.BatchOperation(suite.Context, func(batch graph.Batch) error {
		for idx := range nodeCount {
			id, props := newNodeData(idx)

			nodeObjectIDs[id] = struct{}{}

			// Create the node with the generated properties
			if err := batch.CreateNode(graph.PrepareNode(props, nodeKind)); err != nil {
				return fmt.Errorf("failed to create node with ID %s: %w", id, err)
			}
		}

		return nil
	}))

	ingestBatch, ingestBatchErr := buildIngestibleRelationshipsFromExistingNodes(suite.Context, suite.GraphDB)
	require.NoError(t, ingestBatchErr)

	var (
		endpointResolver             = endpoint.NewResolver(suite.GraphDB)
		resolvedBatch, resolutionErr = endpoint.ResolveAll(suite.Context, endpointResolver, ingestBatch)
	)

	assert.NoError(t, resolutionErr)

	for _, ingestibleRel := range resolvedBatch {
		var (
			_, hasSource = nodeObjectIDs[ingestibleRel.Source.Value]
			_, hasTarget = nodeObjectIDs[ingestibleRel.Target.Value]
		)

		assert.Truef(t, hasSource, "missing source: %s", ingestibleRel.Source.Value)
		assert.Truef(t, hasTarget, "missing target: %s", ingestibleRel.Target.Value)
	}
}
