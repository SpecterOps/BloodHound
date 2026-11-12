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

package tiering_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestLegacyTierFiltersIncludeUntaggedNodesOutsideTierZero(t *testing.T) {
	var (
		graphDatabase = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema()).Graph.Database
		criteria      = tiering.NewSearchTierNodesCtx(false, true, 1, tiering.KindTagTierZero)
	)

	require.NoError(t, graphDatabase.WriteTransaction(t.Context(), func(tx graph.Transaction) error {
		target, err := tx.CreateNode(graph.NewProperties(), ad.Domain)
		if err != nil {
			return err
		}
		for _, properties := range []map[string]any{
			{},
			{common.SystemTags.String(): nil},
			{common.SystemTags.String(): ""},
			{common.SystemTags.String(): "other_tag"},
			{common.SystemTags.String(): ad.AdminTierZero},
		} {
			node, err := tx.CreateNode(graph.AsProperties(properties), ad.User)
			if err != nil {
				return err
			}
			if _, err := tx.CreateRelationshipByIDs(node.ID, target.ID, ad.MemberOf, graph.NewProperties()); err != nil {
				return err
			}
		}
		return nil
	}))

	for _, testCase := range []struct {
		name          string
		criterion     graph.Criteria
		relationships bool
		wantCount     int64
	}{
		{name: "tier nodes", criterion: criteria.SearchPrimaryTierNodes, wantCount: 1},
		{name: "non-tier nodes", criterion: query.Not(criteria.SearchPrimaryTierNodes), wantCount: 4},
		{name: "tier relationships", criterion: criteria.SearchPrimaryTierNodesRel, relationships: true, wantCount: 1},
		{name: "non-tier relationships", criterion: query.Not(criteria.SearchPrimaryTierNodesRel), relationships: true, wantCount: 4},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var count int64
			require.NoError(t, graphDatabase.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
				var err error
				if testCase.relationships {
					count, err = tx.Relationships().Filter(testCase.criterion).Count()
				} else {
					count, err = tx.Nodes().Filter(query.And(query.Kind(query.Node(), ad.User), testCase.criterion)).Count()
				}
				return err
			}))
			require.Equal(t, testCase.wantCount, count)
		})
	}
}
