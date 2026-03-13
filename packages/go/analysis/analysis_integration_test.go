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

package analysis_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	ctx context.Context
	db  *database.BloodhoundDB
}

func setupIntegrationTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	// #region Setup for dbs
	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver(), config.Configuration{})

	err = db.Migrate(ctx)
	require.NoError(t, err)

	err = db.PopulateExtensionData(ctx)
	require.NoError(t, err)

	// #endregion
	return &TestSuite{
		ctx: ctx,
		db:  db,
	}
}

func teardownIntegrationTestSuite(t *testing.T, testSuite *TestSuite) {
	t.Helper()
	if testSuite.db != nil {
		testSuite.db.Close(testSuite.ctx)
	}
}

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

func TestGetNodeKindDisplayLabel(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, testSuite)

	extension, err := testSuite.db.CreateGraphSchemaExtension(testSuite.ctx, "test_extension", "test_extension", "1.0.0", "Test")
	require.NoError(t, err)
	// Create 2 OG display kinds
	nodeKindDis, err := testSuite.db.CreateGraphSchemaNodeKind(testSuite.ctx, "TestKindDis", extension.ID, "", "", true, "", "")
	require.NoError(t, err)
	nodeKindNoDis, err := testSuite.db.CreateGraphSchemaNodeKind(testSuite.ctx, "TestKindNoDis", extension.ID, "", "", false, "", "")
	require.NoError(t, err)

	primaryNodeKinds, err := testSuite.db.GetDisplayNodeGraphKinds(testSuite.ctx)
	require.NoError(t, err)

	assert.Equal(t, ad.Entity.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity)), "should return base kind if no other valid kinds are present")
	assert.Equal(t, ad.User.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.User)), "should return valid AD kind when base and kind are present")
	assert.Equal(t, ad.Group.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.Group, ad.LocalGroup)), "should return valid kind other than LocalGroup if one is present")
	assert.Equal(t, ad.LocalGroup.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.LocalGroup)), "should return LocalGroup if no other valid kinds are present")
	assert.Equal(t, azure.Group.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), azure.Entity, azure.Group)), "should return valid Azure kind when base and kind are present")
	assert.Equal(t, analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), unsupportedKind)), "should return Unknown when only an unsupported kind is present")
	assert.Equal(t, ad.Entity.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, unsupportedKind)), "should return valid kind if one is present even if an unsupported kind is also present")
	assert.Equal(t, analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties())), "should return Unknown if no node has no kinds on it")
	assert.Equal(t, nodeKindDis.Name, analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), nodeKindDis.ToKind())), "should return open graph kind if it is a display kind")
	assert.Equal(t, analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), nodeKindNoDis.ToKind())), "should return unknown kind if open graph kind is not a display kind and has no base")
}
