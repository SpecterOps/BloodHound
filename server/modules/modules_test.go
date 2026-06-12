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

// TestRegister_WiresAnalysisRoutes verifies that the composition root correctly
// attaches the analysis module routes to the shared router. Matching a
// representative route proves that Register successfully delegated to the
// feature module.
func TestRegister_WiresAnalysisRoutes(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		deps       = modules.Deps{
			Router: &routerInst,
			Pool:   new(pgxpool.Pool),
		}
	)

	modules.Register(deps)

	var (
		muxRouter = routerInst.MuxRouter()
		request   = httptest.NewRequest(http.MethodGet, "/api/v2/analysis/status", nil)
		match     mux.RouteMatch
	)

	assert.True(t, muxRouter.Match(request, &match), "analysis route should be registered by Register")
}
