// Copyright 2024 Specter Ops, Inc.
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

package post_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	azureAnalysis "github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestFixManagementGroupNames(t *testing.T) {
	var (
		// This creates a new live integration test context with the graph database
		// This call will load whatever BHE configuration the environment variable `INTEGRATION_CONFIG_PATH` points to.
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	)

	// Management Group created with barebone details
	testCtx.NewNode(graph.AsProperties(map[string]any{
		common.DisplayName.String(): "MANAGEMENT GROUP",
		common.ObjectID.String():    "1234",
		azure.TenantID.String():     "ABC123",
	}), azure.Entity, azure.ManagementGroup)

	// Tenant
	testCtx.NewNode(graph.AsProperties(map[string]any{
		common.Name.String():     "SPECTERDEV",
		common.ObjectID.String(): "ABC123",
	}), azure.Entity, azure.Tenant)

	err := azureAnalysis.FixManagementGroupNames(context.Background(), testCtx.Graph.Database, false)
	require.NoError(t, err)

	err = testCtx.Graph.Database.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		return tx.Nodes().Filter(query.Kind(query.Node(), azure.ManagementGroup)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			count := 0
			for node := range cursor.Chan() {
				if name, err := node.Properties.Get(common.Name.String()).String(); err != nil {
					return err
				} else {
					count++
					require.Equal(t, "MANAGEMENT GROUP@SPECTERDEV", name)
				}
			}

			require.Equal(t, 1, count)

			return nil
		})
	})

	require.NoError(t, err)
}

func TestFixManagementGroupNames_UseRawObjectIDs(t *testing.T) {
	var (
		// This creates a new live integration test context with the graph database
		// This call will load whatever BHE configuration the environment variable `INTEGRATION_CONFIG_PATH` points to.
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	)

	// Management Group created with barebone details, using mixed casing to verify it is preserved
	testCtx.NewNode(graph.AsProperties(map[string]any{
		common.DisplayName.String(): "Management Group",
		common.ObjectID.String():    "1234",
		azure.TenantID.String():     "ABC123",
	}), azure.Entity, azure.ManagementGroup)

	// Tenant
	testCtx.NewNode(graph.AsProperties(map[string]any{
		common.Name.String():     "SpecterDev",
		common.ObjectID.String(): "ABC123",
	}), azure.Entity, azure.Tenant)

	err := azureAnalysis.FixManagementGroupNames(context.Background(), testCtx.Graph.Database, true)
	require.NoError(t, err)

	err = testCtx.Graph.Database.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		return tx.Nodes().Filter(query.Kind(query.Node(), azure.ManagementGroup)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			count := 0
			for node := range cursor.Chan() {
				if name, err := node.Properties.Get(common.Name.String()).String(); err != nil {
					return err
				} else {
					count++
					require.Equal(t, "Management Group@SpecterDev", name)
				}
			}

			require.Equal(t, 1, count)

			return nil
		})
	})

	require.NoError(t, err)
}
