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

package graphdb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	graphschema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// graphDBHarness bundles the wired HTTP handler and the ids of the relationship
// seeded into the graph so the e2e cases can assert the full response contract.
type graphDBHarness struct {
	handler        *mux.Router
	relationshipID int64
	sourceNodeID   int64
	targetNodeID   int64
}

// getPostgresConfig reads the integration test configuration from the environment
// and returns a pgtestdb.Config for the graphdb e2e tests.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for entry := range strings.FieldsSeq(cfg.Database.Connection) {
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

// newGraphDBHarness spins up an isolated postgres database via pgtestdb, applies
// the relational and graph migrations, seeds a single MemberOf relationship into
// the graph, and wires the appdb -> services -> handlers stack onto a mux router.
func newGraphDBHarness(t *testing.T) graphDBHarness {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	dawgsPool, err := dbpool.NewDawgsPool(cfg.Database)
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	bhDatabase := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, bhDatabase.Migrate(ctx))
	require.NoError(t, bhDatabase.PopulateExtensionData(ctx))
	t.Cleanup(func() { bhDatabase.Close(ctx) })

	graphDatabase, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  dawgsPool,
	})
	require.NoError(t, err)
	t.Cleanup(func() { graphDatabase.Close(ctx) })

	require.NoError(t, migrations.NewGraphMigrator(graphDatabase).Migrate(ctx))
	require.NoError(t, graphDatabase.AssertSchema(ctx, graphschema.DefaultGraphSchema()))

	harness := graphDBHarness{}
	seedRelationship(t, ctx, graphDatabase, &harness)

	store := appdb.NewStore(graphDatabase, dbPool)
	handlerSet := handlers.NewHandlersContainer(services.NewService(store))

	harness.handler = mux.NewRouter()
	harness.handler.HandleFunc(
		fmt.Sprintf("/api/v2/relationships/{%s}", handlers.URIPathVariableRelationshipID),
		handlerSet.GetRelationshipByID,
	).Methods("GET")

	return harness
}

// seedRelationship writes two AD nodes and a MemberOf relationship between them
// into the graph, recording the resulting ids on the harness. MemberOf is a
// built-in kind populated into schema_relationship_kinds by PopulateExtensionData,
// so the handler's kind resolution succeeds for the happy-path case.
func seedRelationship(t *testing.T, ctx context.Context, graphDB graph.Database, harness *graphDBHarness) {
	t.Helper()

	err := graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		user, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     "user@test.local",
			common.ObjectID: uuid.Must(uuid.NewV4()).String(),
		}), ad.Entity, ad.User)
		if err != nil {
			return err
		}

		group, err := tx.CreateNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     "group@test.local",
			common.ObjectID: uuid.Must(uuid.NewV4()).String(),
		}), ad.Entity, ad.Group)
		if err != nil {
			return err
		}

		relationship, err := tx.CreateRelationshipByIDs(user.ID, group.ID, ad.MemberOf, graph.NewProperties())
		if err != nil {
			return err
		}

		harness.relationshipID = int64(relationship.ID)
		harness.sourceNodeID = int64(user.ID)
		harness.targetNodeID = int64(group.ID)
		return nil
	})
	require.NoError(t, err)
}

func TestGetRelationshipByID(t *testing.T) {
	var (
		harness  = newGraphDBHarness(t)
		basePath = "/api/v2/relationships/"
	)

	t.Run("returns 200 with the relationship view for an existing relationship", func(t *testing.T) {
		var (
			recorder = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", basePath+strconv.FormatInt(harness.relationshipID, 10), nil)
		)

		harness.handler.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)

		var envelope struct {
			Data handlers.RelationshipView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))

		assert.Equal(t, harness.relationshipID, envelope.Data.RelationshipID)
		assert.Equal(t, harness.sourceNodeID, envelope.Data.SourceNodeID)
		assert.Equal(t, harness.targetNodeID, envelope.Data.TargetNodeID)
		assert.Equal(t, "MemberOf", envelope.Data.Kind.Name)
		require.NotNil(t, envelope.Data.Kind.RelationshipKindID)
		assert.Greater(t, *envelope.Data.Kind.RelationshipKindID, int32(0))
	})

	t.Run("returns 400 when the id is malformed", func(t *testing.T) {
		var (
			recorder = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", basePath+"not-a-number", nil)
		)

		harness.handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("returns 404 when the relationship does not exist", func(t *testing.T) {
		var (
			recorder = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", basePath+strconv.FormatInt(harness.relationshipID+100000, 10), nil)
		)

		harness.handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
}
