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

//go:build serial_integration
// +build serial_integration

package analysis_test

import (
	"context"
	ad2 "github.com/specterops/bloodhound/analysis/ad"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/src/test"
	"testing"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func FetchNumHarnessNodes(db graph.Database) (int64, error) {
	var numHarnessNodes int64

	return numHarnessNodes, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if fetchedCount, err := tx.Nodes().Count(); err != nil {
			return err
		} else {
			numHarnessNodes = fetchedCount
		}

		return nil
	})
}

func TestClearOrphanedNodes(t *testing.T) {
	const numNodesToCreate = 1000

	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {
		numHarnessNodes, err := FetchNumHarnessNodes(db)
		test.RequireNilErr(t, err)

		test.RequireNilErr(t, db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
			for numCreated := 0; numCreated < numNodesToCreate; numCreated++ {
				if _, err := tx.CreateNode(graph.NewProperties(), ad.Entity); err != nil {
					return err
				}
			}

			return nil
		}))

		numNodesAfterCreation, err := FetchNumHarnessNodes(db)
		test.RequireNilErr(t, err)

		require.Equal(t, numHarnessNodes+numNodesToCreate, numNodesAfterCreation)
		test.RequireNilErr(t, analysis.ClearOrphanedNodes(context.Background(), db))

		numNodesAfterDeletion, err := FetchNumHarnessNodes(db)
		test.RequireNilErr(t, err)
		require.Equal(t, numHarnessNodes, numNodesAfterDeletion)
	})
}

func TestCrossProduct(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.ShortcutHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		firstSet := []*graph.Node{testContext.Harness.ShortcutHarness.Group1}
		secondSet := []*graph.Node{testContext.Harness.ShortcutHarness.Group2}
		groupExpansions, err := ad2.ExpandAllRDPLocalGroups(context.Background(), db)
		require.Nil(t, err)
		results := ad2.CalculateCrossProductNodeSets(groupExpansions, firstSet, secondSet)
		require.True(t, results.Contains(harness.ShortcutHarness.Group3.ID.Uint32()))
	})
}
