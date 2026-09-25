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
	t.Parallel()

	type args struct {
		pool *pgxpool.Pool
	}
	type want struct {
		nonNil bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Success: returns a non-nil adapter",
			args: args{pool: new(pgxpool.Pool)},
			want: want{nonNil: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			adapter := featureflags.NewFeatureFlagRequestAdapter(tt.args.pool)

			if tt.want.nonNil {
				require.NotNil(t, adapter)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	type args struct {
		method string
		path   string
	}
	type want struct {
		routeRegistered bool
	}

	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		pool       = new(pgxpool.Pool)
		tests      = []struct {
			name string
			args args
			want want
		}{
			{
				name: "Success: registers the GET features route",
				args: args{method: http.MethodGet, path: "/api/v2/features"},
				want: want{routeRegistered: true},
			},
			{
				name: "Success: registers the PUT toggle route",
				args: args{method: http.MethodPut, path: "/api/v2/features/1/toggle"},
				want: want{routeRegistered: true},
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
			t.Parallel()
			var match mux.RouteMatch
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)

			assert.Equal(t, tt.want.routeRegistered, muxRouter.Match(request, &match), "%s %s route should be registered", tt.args.method, tt.args.path)
		})
	}
}
