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
	tests := []struct {
		name string
		pool *pgxpool.Pool
	}{
		{
			name: "returns non-nil adapter",
			pool: new(pgxpool.Pool),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := featureflags.NewFeatureFlagRequestAdapter(tt.pool)

			require.NotNil(t, adapter)
		})
	}
}

func TestRegister(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		pool       = new(pgxpool.Pool)
		tests      = []struct {
			name   string
			method string
			path   string
		}{
			{
				name:   "registers the GET features route",
				method: http.MethodGet,
				path:   "/api/v2/features",
			},
			{
				name:   "registers the PUT toggle route",
				method: http.MethodPut,
				path:   "/api/v2/features/1/toggle",
			},
		}
	)

	// Should not panic
	require.NotPanics(t, func() {
		featureflags.Register(&routerInst, pool)
	})

	muxRouter := routerInst.MuxRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var match mux.RouteMatch
			request := httptest.NewRequest(tt.method, tt.path, nil)

			assert.True(t, muxRouter.Match(request, &match), "%s %s route should be registered", tt.method, tt.path)
		})
	}
}
