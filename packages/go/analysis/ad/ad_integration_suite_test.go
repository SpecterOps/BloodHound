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

package ad_test

import (
	"context"
	"fmt"
	"math/rand"
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
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context context.Context
	GraphDB graph.Database
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests in the ad package.
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

	require.NoError(t, migrations.NewGraphMigrator(graphDB).Migrate(ctx))
	require.NoError(t, graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema()))

	return IntegrationTestSuite{
		Context: ctx,
		GraphDB: graphDB,
	}
}

// getPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	integrationConfig, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(integrationConfig.Database.Connection) {
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
		node, err = tx.CreateNode(properties, kinds...)
		return err
	})

	require.NoError(t, err)

	return node
}

func UpdateNode(t *testing.T, suite *IntegrationTestSuite, node *graph.Node) {
	err := suite.GraphDB.WriteTransaction(suite.Context, func(tx graph.Transaction) error {
		return tx.UpdateNode(node)
	})
	require.NoError(t, err)
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

// RandomDomainSID returns a randomly generated S-1-5-21-* domain SID for use
// in integration test fixtures.
func RandomDomainSID() string {
	return fmt.Sprintf("S-1-5-21-%d-%d-%d", rand.Int31(), rand.Int31(), rand.Int31())
}

// newRandomObjectID returns a new UUIDv4 string for use as the ObjectID of an
// Active Directory entity in test fixtures.
func newRandomObjectID(t *testing.T) string {
	t.Helper()
	objectID, err := uuid.NewV4()
	require.NoError(t, err)
	return objectID.String()
}

func NewActiveDirectoryDomain(t *testing.T, suite *IntegrationTestSuite, name, objectID string, blocksInheritance, collected bool) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:          name,
		common.ObjectID:      objectID,
		ad.DomainSID:         objectID,
		common.Collected:     collected,
		ad.BlocksInheritance: blocksInheritance,
	}), ad.Entity, ad.Domain)
}

func NewActiveDirectoryUser(t *testing.T, suite *IntegrationTestSuite, name, domainSID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: strings.ToUpper(newRandomObjectID(t)),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.User)
}

func NewActiveDirectoryComputer(t *testing.T, suite *IntegrationTestSuite, name, domainSID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: newRandomObjectID(t),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.Computer)
}

func NewActiveDirectoryNTAuthStore(t *testing.T, suite *IntegrationTestSuite, name, domainSID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:        name,
		common.ObjectID:    newRandomObjectID(t),
		ad.DomainSID:       domainSID,
		ad.CertThumbprints: []string{"a", "b", "c"},
	}), ad.Entity, ad.NTAuthStore)
}

func NewActiveDirectoryEnterpriseCA(t *testing.T, suite *IntegrationTestSuite, name, domainSID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: newRandomObjectID(t),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.EnterpriseCA)
}

func NewActiveDirectoryRootCA(t *testing.T, suite *IntegrationTestSuite, name, domainSID string) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:     name,
		common.ObjectID: newRandomObjectID(t),
		ad.DomainSID:    domainSID,
	}), ad.Entity, ad.RootCA)
}

// CertTemplateProperties captures the subset of cert-template properties used
// by integration tests in the ad package.
type CertTemplateProperties struct {
	RequiresManagerApproval bool
	AuthenticationEnabled   bool
	EnrolleeSuppliesSubject bool
	NoSecurityExtension     bool
	SchemaVersion           float64
	AuthorizedSignatures    float64
	SubjectAltRequireDNS    bool
	SubjectAltRequireUPN    bool
	SubjectAltRequireSPN    bool
	EffectiveEKUs           []string
	ApplicationPolicies     []string
}

func NewActiveDirectoryCertTemplate(t *testing.T, suite *IntegrationTestSuite, name, domainSID string, data CertTemplateProperties) *graph.Node {
	return NewNode(t, suite, graph.AsProperties(graph.PropertyMap{
		common.Name:                name,
		common.ObjectID:            newRandomObjectID(t),
		ad.DomainSID:               domainSID,
		ad.RequiresManagerApproval: data.RequiresManagerApproval,
		ad.AuthenticationEnabled:   data.AuthenticationEnabled,
		ad.EnrolleeSuppliesSubject: data.EnrolleeSuppliesSubject,
		ad.NoSecurityExtension:     data.NoSecurityExtension,
		ad.SchemaVersion:           data.SchemaVersion,
		ad.AuthorizedSignatures:    data.AuthorizedSignatures,
		ad.EffectiveEKUs:           data.EffectiveEKUs,
		ad.ApplicationPolicies:     data.ApplicationPolicies,
		ad.SubjectAltRequireDNS:    data.SubjectAltRequireDNS,
		ad.SubjectAltRequireUPN:    data.SubjectAltRequireUPN,
		ad.SubjectAltRequireSPN:    data.SubjectAltRequireSPN,
	}), ad.Entity, ad.CertTemplate)
}
