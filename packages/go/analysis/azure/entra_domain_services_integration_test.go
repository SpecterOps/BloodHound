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

package azure_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis/azure"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	graphAzure "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainServiceEntityDetails(t *testing.T) {
	t.Parallel()

	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		domainServiceObjectID = integration.RandomObjectID(t)
		tenantID              = integration.RandomObjectID(t)
		controller            = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:         "Controller",
			common.ObjectID:     integration.RandomObjectID(t),
			graphAzure.TenantID: tenantID,
		}), graphAzure.Entity, graphAzure.User)
		domainService = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:         "Domain Service",
			common.ObjectID:     domainServiceObjectID,
			graphAzure.TenantID: tenantID,
		}), graphAzure.Entity, graphAzure.EntraDS)
	)
	NewRelationship(t, &suite, controller, domainService, graphAzure.ManageEntraDS)

	details, err := azure.DomainServiceEntityDetails(suite.Context, suite.GraphDB, schema.ValidKinds, domainServiceObjectID, false)
	require.NoError(t, err)
	assert.Equal(t, domainServiceObjectID, details.Properties[common.ObjectID.String()])
	assert.Zero(t, details.InboundObjectControl)

	details, err = azure.DomainServiceEntityDetails(suite.Context, suite.GraphDB, schema.ValidKinds, domainServiceObjectID, true)
	require.NoError(t, err)
	assert.Equal(t, 1, details.InboundObjectControl)
}

func TestManageEntraDSDCA(t *testing.T) {
	t.Parallel()

	suite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &suite)

	var (
		tenantID        = integration.RandomObjectID(t)
		tenant          = NewAzureTenant(t, &suite, tenantID)
		appAdminRole    = NewAzureRole(t, &suite, "Application Administrator", integration.RandomObjectID(t), graphAzure.ApplicationAdministratorRole, tenantID)
		groupsAdminRole = NewAzureRole(t, &suite, "Groups Administrator", integration.RandomObjectID(t), graphAzure.GroupsAdministratorRole, tenantID)
		domainService   = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:         "Managed Domain",
			common.ObjectID:     integration.RandomObjectID(t),
			graphAzure.TenantID: tenantID,
		}), graphAzure.Entity, graphAzure.EntraDS)
		controller = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:         "Qualified Controller",
			common.ObjectID:     integration.RandomObjectID(t),
			graphAzure.TenantID: tenantID,
		}), graphAzure.Entity, graphAzure.User)
	)

	NewRelationship(t, &suite, tenant, controller, graphAzure.Contains)
	NewRelationship(t, &suite, tenant, domainService, graphAzure.Contains)
	NewRelationship(t, &suite, controller, domainService, graphAzure.EntraDSContributor)
	NewRelationship(t, &suite, controller, appAdminRole, graphAzure.HasRole)
	groupsAdminAssignment := NewRelationship(t, &suite, controller, groupsAdminRole, graphAzure.HasRole)

	firstRunStats, err := azure.ManageEntraDS(context.Background(), suite.GraphDB, true)
	require.NoError(t, err)
	require.NotNil(t, firstRunStats.RelationshipsCreated[graphAzure.ManageEntraDS])
	assert.Equal(t, int32(1), *firstRunStats.RelationshipsCreated[graphAzure.ManageEntraDS])

	firstRunEdges := fetchManageEntraDSEdges(t, suite.Context, suite.GraphDB)
	require.Len(t, firstRunEdges, 1)
	firstRunEdge := firstRunEdges[0]
	firstRunFirstSeen, err := firstRunEdge.Properties.Get(common.FirstSeen.String()).Time()
	require.NoError(t, err)

	secondRunStats, err := azure.ManageEntraDS(context.Background(), suite.GraphDB, true)
	require.NoError(t, err)
	assert.NotContains(t, secondRunStats.RelationshipsCreated, graphAzure.ManageEntraDS)

	secondRunEdges := fetchManageEntraDSEdges(t, suite.Context, suite.GraphDB)
	require.Len(t, secondRunEdges, 1)
	assert.Equal(t, firstRunEdge.ID, secondRunEdges[0].ID)
	secondRunFirstSeen, err := secondRunEdges[0].Properties.Get(common.FirstSeen.String()).Time()
	require.NoError(t, err)
	assert.Equal(t, firstRunFirstSeen, secondRunFirstSeen)

	err = suite.GraphDB.WriteTransaction(suite.Context, func(tx graph.Transaction) error {
		return tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.StartID(), groupsAdminAssignment.StartID),
				query.InIDs(query.EndID(), groupsAdminAssignment.EndID),
				query.Kind(query.Relationship(), groupsAdminAssignment.Kind),
			)
		}).Delete()
	})
	require.NoError(t, err)

	_, err = azure.ManageEntraDS(context.Background(), suite.GraphDB, true)
	require.NoError(t, err)
	assert.Empty(t, fetchManageEntraDSEdges(t, suite.Context, suite.GraphDB))

	NewRelationship(t, &suite, controller, groupsAdminRole, graphAzure.HasRole)
	_, err = azure.ManageEntraDS(context.Background(), suite.GraphDB, true)
	require.NoError(t, err)
	require.Len(t, fetchManageEntraDSEdges(t, suite.Context, suite.GraphDB), 1)

	_, err = azure.ManageEntraDS(context.Background(), suite.GraphDB, false)
	require.NoError(t, err)
	assert.Empty(t, fetchManageEntraDSEdges(t, suite.Context, suite.GraphDB))
}

func fetchManageEntraDSEdges(t *testing.T, ctx context.Context, db graph.Database) []*graph.Relationship {
	t.Helper()

	var relationships []*graph.Relationship
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		relationships, err = ops.FetchRelationships(tx.Relationships().Filter(query.Kind(query.Relationship(), graphAzure.ManageEntraDS)))
		return err
	})
	require.NoError(t, err)

	return relationships
}
