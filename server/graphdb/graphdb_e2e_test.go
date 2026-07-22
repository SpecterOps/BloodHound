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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	graphschema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/server/etac"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/authz"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// graphDBHarness bundles the wired HTTP handler and the test data IDs
// so E2E test cases can assert the full response contract.
type graphDBHarness struct {
	handler                  *mux.Router
	nodeID                   int64
	relationshipID           int64
	relationshipKindID       int32
	relationshipInfoMarkdown string
	sourceNodeID             int64
	targetNodeID             int64
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

func withAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		ctx := &bhctx.Context{
			AuthCtx: auth.Context{
				Owner: model.User{
					AllEnvironments: true,
				},
			},
		}

		next.ServeHTTP(response, bhctx.SetRequestContext(request, ctx))
	})
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
	seedNode(t, ctx, graphDatabase, &harness)
	seedRelationship(t, ctx, graphDatabase, &harness)
	seedRelationshipKindInfo(t, ctx, bhDatabase, dbPool, &harness)

	store := appdb.NewStore(graphDatabase, dbPool)
	etacService := etac.Register(dbPool, dogtags.NewDefaultService())
	nodeAuthorizer := authz.NewNodeAuthorizer(etacService)
	handlerSet := handlers.NewHandlersContainer(services.NewService(store), nodeAuthorizer)

	harness.handler = mux.NewRouter()
	harness.handler.Use(withAuthenticatedUser)
	harness.handler.HandleFunc(
		fmt.Sprintf("/api/v2/nodes/{%s}", handlers.URIPathVariableNodeID),
		handlerSet.GetNodeByID,
	).Methods("GET")
	harness.handler.HandleFunc(
		fmt.Sprintf("/api/v2/relationships/{%s}", handlers.URIPathVariableRelationshipID),
		handlerSet.GetRelationshipByID,
	).Methods("GET")

	return harness
}

// seedNode creates a test node in the graph with multiple kinds and properties,
// recording the node ID in the harness.
func seedNode(t *testing.T, ctx context.Context, graphDB graph.Database, harness *graphDBHarness) {
	t.Helper()

	err := graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		node, err := tx.CreateNode(
			graph.AsProperties(graph.PropertyMap{
				common.Name:     "Administrator",
				common.ObjectID: uuid.Must(uuid.NewV4()).String(),
			}),
			ad.Entity,
			ad.User,
		)
		if err != nil {
			return err
		}
		harness.nodeID = int64(node.ID)
		return nil
	})
	require.NoError(t, err)
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

func seedRelationshipKindInfo(t *testing.T, ctx context.Context, bloodhoundDB *database.BloodhoundDB, dbPool *pgxpool.Pool, harness *graphDBHarness) {
	t.Helper()

	var backingKindID int32
	require.NoError(t, dbPool.QueryRow(ctx, `
		SELECT k.id, rk.id
		FROM kind k
		JOIN schema_relationship_kinds rk ON rk.kind_id = k.id
		WHERE k.name = $1`, ad.MemberOf.String()).Scan(&backingKindID, &harness.relationshipKindID))

	harness.relationshipInfoMarkdown = "relationship kind info markdown"
	_, err := bloodhoundDB.CreateKindInfo(ctx, backingKindID, nil, &harness.relationshipKindID, model.KindInfoInput{
		InfoKey:  "relationship_test_info",
		Title:    "Relationship Test Info",
		Position: 1,
		Content:  json.RawMessage(`{"markdown":{"content":"relationship kind info markdown"}}`),
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
		assert.NotContains(t, recorder.Body.String(), `"info"`)
	})

	t.Run("returns 200 without relationship info when include-info is false", func(t *testing.T) {
		var (
			recorder = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", basePath+strconv.FormatInt(harness.relationshipID, 10)+"?include-info=false", nil)
		)

		harness.handler.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.NotContains(t, recorder.Body.String(), `"info"`)
	})

	t.Run("returns 200 with relationship info when include-info is true", func(t *testing.T) {
		var (
			recorder = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", basePath+strconv.FormatInt(harness.relationshipID, 10)+"?include-info=true", nil)
		)

		harness.handler.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)

		var envelope struct {
			Data handlers.RelationshipView `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))

		require.Len(t, envelope.Data.Info, 1)
		assert.Equal(t, "relationship_test_info", envelope.Data.Info[0].Name)
		assert.Equal(t, "Relationship Test Info", envelope.Data.Info[0].Title)
		assert.Equal(t, int32(1), envelope.Data.Info[0].Position)
		assert.Equal(t, int(harness.relationshipKindID), envelope.Data.Info[0].RelationshipKindID)
		assert.Equal(t, harness.relationshipInfoMarkdown, envelope.Data.Info[0].Markdown.Content)
	})

	t.Run("returns 400 when include-info is malformed", func(t *testing.T) {
		var (
			recorder = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", basePath+strconv.FormatInt(harness.relationshipID, 10)+"?include-info=wat", nil)
		)

		harness.handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
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

func TestGetNodeByID(t *testing.T) {
	var (
		harness  = newGraphDBHarness(t)
		basePath = "/api/v2/nodes/"
	)

	tests := []struct {
		name           string
		path           string
		wantStatusCode int
		assertBody     func(t *testing.T, body []byte)
	}{
		{
			name:           "returns 200 with the node view for an existing node",
			path:           basePath + strconv.FormatInt(harness.nodeID, 10),
			wantStatusCode: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				t.Helper()

				var envelope struct {
					Data handlers.NodeView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))

				assert.Equal(t, harness.nodeID, envelope.Data.NodeID)
				assert.Len(t, envelope.Data.Kinds, 2)
				assert.Equal(t, "Administrator", envelope.Data.Properties[common.Name.String()])
				assert.NotEmpty(t, envelope.Data.Properties[common.ObjectID.String()])

				// Verify at least one kind has a non-nil ID (Entity and User should both be registered)
				var hasNonNilKindID bool
				for _, kind := range envelope.Data.Kinds {
					if kind.NodeKindID != nil {
						hasNonNilKindID = true
						break
					}
				}
				assert.True(t, hasNonNilKindID, "at least one kind should have a resolved ID")
			},
		},
		{
			name:           "returns 400 when the id is malformed",
			path:           basePath + "not-a-number",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "returns 404 when the node does not exist",
			path:           basePath + strconv.FormatInt(harness.nodeID+100000, 10),
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest("GET", tt.path, nil)
			)

			harness.handler.ServeHTTP(recorder, request)

			assert.Equal(t, tt.wantStatusCode, recorder.Code)

			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}
