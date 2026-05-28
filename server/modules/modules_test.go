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

package modules_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/modules"
	"github.com/stretchr/testify/assert"
)

func TestRegister_PanicsOnNilRouter(t *testing.T) {
	assert.Panics(t, func() {
		modules.Register(modules.Deps{
			Router: nil,
			Pool:   new(pgxpool.Pool),
		})
	})
}

func TestRegister_PanicsOnNilPool(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	assert.Panics(t, func() {
		modules.Register(modules.Deps{
			Router: &routerInst,
			Pool:   nil,
		})
	})
}

func TestRegister_RegistersAnalysisRoutes(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		// new(pgxpool.Pool) is a non-nil zero-value pool. Register only stores
		// the pointer; no pool methods are called during route registration.
		pool = new(pgxpool.Pool)
	)

	assert.NotPanics(t, func() {
		modules.Register(modules.Deps{
			Router: &routerInst,
			Pool:   pool,
		})
	})

	muxRouter := routerInst.MuxRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v2/analysis"},
		{http.MethodPut, "/api/v2/analysis"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		var match mux.RouteMatch
		assert.True(t, muxRouter.Match(req, &match), "%s %s route should be registered after modules.Register", tc.method, tc.path)
	}
}
