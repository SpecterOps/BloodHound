// Copyright 2023 Specter Ops, Inc.
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
// +build integration

package ad_test

import (
	"context"
	"github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	adSchema "github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/server"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

// TODO: This should be refactored as it uses a lot of lower level functions from utils and server
// TODO: This integration test does not have a valid harness
func TestGetADCSESC3EdgeComposition(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		t.Fatalf("%v", err)
	} else if bhDB, graphDB, err := server.ConnectDatabases(cfg); err != nil {
		t.Fatalf("%v", err)
	} else {
		defer graphDB.Close()

		// Run analysis
		require.Nil(t, datapipe.RunAnalysisOperations(ctx, bhDB, graphDB, cfg))

		// Graph the ESC3 edges
		var esc3Edges []*graph.Relationship

		require.Nil(t, graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
			return tx.Relationships().Filter(query.KindIn(query.Relationship(), adSchema.ADCSESC3)).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
				for edge := range cursor.Chan() {
					esc3Edges = append(esc3Edges, edge)
				}

				return cursor.Error()
			})
		}))

		require.Equal(t, 1, len(esc3Edges))

		// Validate composition
		paths, err := ad.GetADCSESC3EdgeComposition(ctx, graphDB, esc3Edges[0])

		require.Nil(t, err)
		require.Equal(t, 4, paths.Len())
	}
}
