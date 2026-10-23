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
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context context.Context
	GraphDB graph.Database
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests
func setupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	pool, err := dbpool.NewDawgsPool(cfg.Database)
	require.NoError(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	err = migrations.NewGraphMigrator(graphDB).Migrate(ctx)
	require.NoError(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	return IntegrationTestSuite{
		Context: ctx,
		GraphDB: graphDB,
	}
}

// getPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	config, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(config.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func teardownIntegrationTestSuite(t *testing.T, suite *IntegrationTestSuite) {
	t.Helper()

	if suite.GraphDB != nil {
		if err := suite.GraphDB.Close(suite.Context); err != nil {
			t.Logf("Failed to close GraphDB: %v", err)
		}
	}
}

func NewNode(t *testing.T, suite *IntegrationTestSuite, properties *graph.Properties, kinds ...graph.Kind) *graph.Node {
	var (
		node *graph.Node
		err  error
	)

	err = suite.GraphDB.WriteTransaction(suite.Context, func(tx graph.Transaction) error {
		if node, err = tx.CreateNode(properties, kinds...); err != nil {
			return err
		}

		return nil
	})

	require.NoError(t, err)

	return node
}

func NewRelationship(t *testing.T, suite *IntegrationTestSuite, startNode, endNode *graph.Node, kind graph.Kind, propertyBags ...*graph.Properties) *graph.Relationship {
	var (
		relationshipProperties = graph.NewPropertiesRed()
		relationship           *graph.Relationship
		err                    error
	)

	for _, additionalProperties := range propertyBags {
		relationshipProperties.Merge(additionalProperties)
	}

	err = suite.GraphDB.WriteTransaction(suite.Context, func(tx graph.Transaction) error {
		relationship, err = tx.CreateRelationshipByIDs(startNode.ID, endNode.ID, kind, relationshipProperties)
		return err
	})

	require.NoError(t, err)

	return relationship
}

func NewAzureTenant(t *testing.T, suite *IntegrationTestSuite, tenantID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     "New Tenant",
		common.ObjectID: tenantID,
		azure.TenantID:  tenantID,
		azure.License:   "license",
	}), azure.Entity, azure.Tenant)
}

func NewAzureApplication(t *testing.T, suite *IntegrationTestSuite, name, objectID, tenantID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.App)
}

func NewAzureServicePrincipal(t *testing.T, suite *IntegrationTestSuite, name, objectID, tenantID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: objectID,
		azure.TenantID:  tenantID,
	}), azure.Entity, azure.ServicePrincipal)
}

func NewAzureRole(t *testing.T, suite *IntegrationTestSuite, name, objectID, roleTemplateID, tenantID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:          name,
		common.ObjectID:      objectID,
		azure.RoleTemplateID: roleTemplateID,
		azure.TenantID:       tenantID,
	}), azure.Entity, azure.Role)
}

type appRoleAssignmentFixture struct {
	tenant              *graph.Node
	microsoftGraph      *graph.Node
	source              *graph.Node
	otherSource         *graph.Node
	application         *graph.Node
	role                *graph.Node
	ordinaryGroup       *graph.Node
	roleAssignableGroup *graph.Node
}

type appRoleAssignmentEdge struct {
	target graph.ID
	kind   graph.Kind
}

func newAppRoleAssignmentObjectID(t *testing.T) string {
	t.Helper()
	var objectID, err = uuid.NewV4()
	require.NoError(t, err)
	return objectID.String()
}

func newAppRoleAssignmentFixture(t *testing.T, suite *IntegrationTestSuite) appRoleAssignmentFixture {
	t.Helper()
	var (
		tenantID = newAppRoleAssignmentObjectID(t)
		fixture  = appRoleAssignmentFixture{
			tenant:         NewAzureTenant(t, suite, tenantID),
			microsoftGraph: NewAzureServicePrincipal(t, suite, "Microsoft Graph", newAppRoleAssignmentObjectID(t), tenantID),
			source:         NewAzureServicePrincipal(t, suite, "Source", newAppRoleAssignmentObjectID(t), tenantID),
			otherSource:    NewAzureServicePrincipal(t, suite, "Other source", newAppRoleAssignmentObjectID(t), tenantID),
			application:    NewAzureApplication(t, suite, "Application", newAppRoleAssignmentObjectID(t), tenantID),
			role:           NewAzureRole(t, suite, "Role", newAppRoleAssignmentObjectID(t), newAppRoleAssignmentObjectID(t), tenantID),
		}
	)
	fixture.ordinaryGroup = NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.ObjectID:          newAppRoleAssignmentObjectID(t),
		azure.TenantID:           tenantID,
		azure.IsAssignableToRole: false,
	}), azure.Entity, azure.Group)
	fixture.roleAssignableGroup = NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.ObjectID:          newAppRoleAssignmentObjectID(t),
		azure.TenantID:           tenantID,
		azure.IsAssignableToRole: true,
	}), azure.Entity, azure.Group)
	for _, node := range []*graph.Node{fixture.microsoftGraph, fixture.source, fixture.otherSource, fixture.application, fixture.role, fixture.ordinaryGroup, fixture.roleAssignableGroup} {
		NewRelationship(t, suite, fixture.tenant, node, azure.Contains)
	}
	return fixture
}

func requireAppRoleAssignmentEdges(t *testing.T, suite *IntegrationTestSuite, source *graph.Node, expected []appRoleAssignmentEdge) {
	t.Helper()
	var actual []appRoleAssignmentEdge
	require.NoError(t, suite.GraphDB.ReadTransaction(suite.Context, func(tx graph.Transaction) error {
		relationships, err := ops.FetchRelationships(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), source.ID),
			query.KindIn(query.Relationship(), azure.AZMGGrantRole, azure.AZMGGrantAppRoles, azure.AZMGAddSecret, azure.AZMGAddOwner, azure.AZMGAddMember),
		)))
		if err != nil {
			return err
		}
		for _, relationship := range relationships {
			actual = append(actual, appRoleAssignmentEdge{target: relationship.EndID, kind: relationship.Kind})
		}
		return nil
	}))
	require.ElementsMatch(t, expected, actual)
}
