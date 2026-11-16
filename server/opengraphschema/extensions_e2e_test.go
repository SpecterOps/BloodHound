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

package opengraphschema_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/featureflags"
	"github.com/specterops/bloodhound/server/opengraphschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListExtensions(t *testing.T) {
	var (
		ctx                               = context.Background()
		connection                        = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
		configuration, configurationError = config.NewDefaultConnectionConfiguration(connection.URL())
	)
	require.NoError(t, configurationError)
	gormDatabase, pool, err := database.OpenDatabase(configuration.Database)
	require.NoError(t, err)
	var databaseInterface = database.NewBloodhoundDB(gormDatabase, pool, auth.NewIdentityResolver(), configuration)
	t.Cleanup(func() { databaseInterface.Close(ctx) })
	require.NoError(t, databaseInterface.Migrate(ctx))
	var (
		routerInst = router.NewRouter(configuration, auth.NewAuthorizer(nil), "")
		tests      = []struct {
			name        string
			enabled     bool
			permissions model.Permissions
			setup       func(t *testing.T)
			wantStatus  int
			wantBody    string
		}{
			{name: "orders by name", enabled: true, permissions: model.Permissions{auth.Permissions().OpenGraphRead}, setup: func(t *testing.T) {
				_, err := pool.Exec(ctx, `INSERT INTO schema_extensions (id, name, display_name, version, is_builtin, namespace) VALUES
     (10002, 'zeta', 'Zeta', 'v2', false, 'Z'),
     (10001, 'alpha', 'Alpha', 'v1', true, 'A')`)
				require.NoError(t, err)
			}, wantStatus: http.StatusOK, wantBody: `{"data":{"extensions":[{"id":10001,"name":"Alpha","namespace":"A","version":"v1","is_builtin":true},{"id":10002,"name":"Zeta","namespace":"Z","version":"v2","is_builtin":false}]}}`},
			{name: "empty", enabled: true, permissions: model.Permissions{auth.Permissions().OpenGraphRead}, wantStatus: http.StatusOK, wantBody: `{"data":{"extensions":[]}}`},
			{name: "feature disabled", permissions: model.Permissions{auth.Permissions().OpenGraphRead}, wantStatus: http.StatusNotFound},
			{name: "permission denied", enabled: true, permissions: model.Permissions{auth.Permissions().GraphDBRead}, wantStatus: http.StatusForbidden},
		}
	)
	opengraphschema.Register(&routerInst, pool, featureflags.NewFeatureFlagRequestAdapter(pool))
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, "DELETE FROM schema_extensions")
			require.NoError(t, err)
			_, err = pool.Exec(ctx, "UPDATE feature_flags SET enabled = $1 WHERE key = $2", testCase.enabled, featureflags.FeatureOpenGraphExtensionManagement)
			require.NoError(t, err)
			if testCase.setup != nil {
				testCase.setup(t)
			}
			var (
				user = model.User{
					Unique: model.Unique{ID: uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")},
					Roles:  model.Roles{{Permissions: testCase.permissions}},
				}
				request  = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/extensions", nil), &bhctx.Context{AuthCtx: auth.Context{Owner: user}})
				recorder = httptest.NewRecorder()
			)
			routerInst.Handler().ServeHTTP(recorder, request)
			assert.Equal(t, testCase.wantStatus, recorder.Code)
			if testCase.wantBody != "" {
				assert.JSONEq(t, testCase.wantBody, recorder.Body.String())
			}
		})
	}
}
