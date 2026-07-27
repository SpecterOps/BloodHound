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

package identity_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Run("successfully registers identity routes", func(t *testing.T) {
		var (
			cfg        = config.Configuration{}
			authorizer = auth.NewAuthorizer(nil)
			routerInst = router.NewRouter(cfg, authorizer, "")
			pool       = new(pgxpool.Pool)
		)

		// Should not panic
		require.NotPanics(t, func() {
			identity.Register(&routerInst, pool)
		})

		// Verify routes are registered
		var (
			muxRouter = routerInst.MuxRouter()
			match     mux.RouteMatch
		)

		// Test GET /api/v2/roles/{role_id} route
		getRoleRequest := httptest.NewRequest(http.MethodGet, "/api/v2/roles/1", nil)
		assert.True(t, muxRouter.Match(getRoleRequest, &match), "GET /api/v2/roles/{role_id} route should be registered")

		// Test GET /api/v2/permissions/{permission_id} route
		getPermissionRequest := httptest.NewRequest(http.MethodGet, "/api/v2/permissions/1", nil)
		assert.True(t, muxRouter.Match(getPermissionRequest, &match), "GET /api/v2/permissions/{permission_id} route should be registered")
	})
}
