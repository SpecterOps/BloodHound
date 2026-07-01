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

package featureflags_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/featureflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFeatureFlagRequestAdapter(t *testing.T) {
	t.Run("returns non-nil adapter", func(t *testing.T) {
		pool := new(pgxpool.Pool)
		adapter := featureflags.NewFeatureFlagRequestAdapter(pool)

		require.NotNil(t, adapter)
	})
}

func TestRegister(t *testing.T) {
	t.Run("successfully registers featureflags routes", func(t *testing.T) {
		var (
			cfg        = config.Configuration{}
			authorizer = auth.NewAuthorizer(nil)
			routerInst = router.NewRouter(cfg, authorizer, "")
			pool       = new(pgxpool.Pool)
		)

		// Should not panic
		require.NotPanics(t, func() {
			featureflags.Register(&routerInst, pool)
		})

		// Verify routes are registered
		var (
			muxRouter = routerInst.MuxRouter()
			match     mux.RouteMatch
		)

		// Test GET /api/v2/features route
		getRequest := httptest.NewRequest(http.MethodGet, "/api/v2/features", nil)
		assert.True(t, muxRouter.Match(getRequest, &match), "GET /api/v2/features route should be registered")

		// Test PUT /api/v2/features/{feature_id}/toggle route
		putRequest := httptest.NewRequest(http.MethodPut, "/api/v2/features/1/toggle", nil)
		assert.True(t, muxRouter.Match(putRequest, &match), "PUT /api/v2/features/{feature_id}/toggle route should be registered")
	})
}
