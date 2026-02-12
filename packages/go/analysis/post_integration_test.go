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

package analysis_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis"
	azureAnalysis "github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// This is a test to validate when we have a situ such
// There exists an AD user and an Azure user that represent the same principal (the same user identity)
//
// This connection is made by correlating properties that are inserted when data from Active Directory or Azure is ingested
// into the system. These properties are referenced in the function in bhce/packages/go/analysis/hybrid/hybrid.go - hasOnPremUser(...)
// and then mapped to AD users for creation of the SyncedToEntraUser and SyncedToADUser edges.
//
// Hybrid post-processing is driven by https://learn.microsoft.com/en-us/azure/architecture/reference-architectures/identity/azure-ad - current
// limitations of the implementation in MS means that the relationship between User and AZUser is 1:* where a AZUser may only be connected to
// one AD principal.
func TestDeleteTransitEdges(t *testing.T) {
	var (
		// This creates a new live integration test context with the graph database
		// This call will load whatever BHE configuration the environment variable `INTEGRATION_CONFIG_PATH` points to.
		testCtx = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		// For this test we need to validate BED-4954 - this requires, at minimum, an AD user and an Entra (Azure) user. The lines below
		// will utilize the test context to put the data directly into the graph.

		// AD user first
		adUser = testCtx.NewNode(graph.AsProperties(map[string]any{
			"name":     "ad_user",
			"objectid": "1234",
		}), ad.Entity, ad.User)

		// Azure user second
		azureUser = testCtx.NewNode(graph.AsProperties(map[string]any{
			"name":     "azure_user",
			"objectid": "4321",
		}), azure.Entity, azure.User)
	)

	// In order to validate that DeleteTransitEdges and the updated PostProcessedRelationships for both AD and Azure are correct, we need to simulate
	// the completion of post-processing in: bhce/cmd/api/src/analysis/azure/post.go
	//
	// The specific function that is responsible for creating the edges below can be found in bhce/packages/go/analysis/hybrid/hybrid.go - PostHybrid(...)
	//
	// Here, we are choosing to create these edges such that the data describes what we would expect to see after a successful execution of the logic
	// in bhce/cmd/api/src/analysis/azure/post.go.
	testCtx.NewRelationship(adUser, azureUser, ad.SyncedToEntraUser)
	testCtx.NewRelationship(azureUser, adUser, azure.SyncedToADUser)

	// The way post-processing operates is that all edges created during post-processing are deleted before each analysis run. This helps keep the graph consistent
	// where certain graph conditions (edges, node properties, etc.) that once existed were removed or modified due to the user's environment changing.

	// This first run removes all Azure post-processed relationships - expected outcome is that SyncedToADUser is removed at this stage
	_, err := analysis.DeleteTransitEdges(context.Background(), testCtx.Graph.Database, graph.Kinds{ad.Entity, azure.Entity}, azure.PostProcessedRelationships())

	// Deleting transit edges must not return an error
	require.Nil(t, err)

	err = testCtx.Graph.Database.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		numEdges, err := tx.Relationships().Filter(query.Kind(query.Relationship(), azure.SyncedToADUser)).Count()

		// This must be true which would mean that the above created SyncedToADUser was correctly deleted by the DeleteTransitEdges call
		require.Equal(t, int64(0), numEdges)
		return err
	})

	// The DB must not return any errors
	require.Nil(t, err)

	// This first run removes all AD post-processed relationships - expected outcome is that SyncedToEntraUser is removed at this stage
	_, err = analysis.DeleteTransitEdges(context.Background(), testCtx.Graph.Database, graph.Kinds{ad.Entity, azure.Entity}, ad.PostProcessedRelationships())
	// Deleting transit edges must not return an error
	require.Nil(t, err)

	err = testCtx.Graph.Database.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		numEdges, err := tx.Relationships().Filter(query.Kind(query.Relationship(), ad.SyncedToEntraUser)).Count()

		// This must be true which would mean that the above created SyncedToADUser was correctly deleted by the DeleteTransitEdges call
		require.Equal(t, int64(0), numEdges)
		return err
	})

	// The DB must not return any errors
	require.Nil(t, err)
}

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

	err := azureAnalysis.FixManagementGroupNames(context.Background(), testCtx.Graph.Database)
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
