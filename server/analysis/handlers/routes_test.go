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

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/analysis/handlers"
	"github.com/specterops/bloodhound/server/analysis/handlers/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
		mock       = mocks.NewMockAnalysis(t)
		handlerSet = handlers.NewHandlersContainer(mock)
	)

	// Register the routes - this should not panic
	handlers.Register(&routerInst, handlerSet)

	// Verify routes were registered by attempting to match them via the underlying mux router
	muxRouter := routerInst.MuxRouter()

	// Test GET route
	getReq := httptest.NewRequest(http.MethodGet, "/api/v2/analysis", nil)
	var getMatch mux.RouteMatch
	foundGet := muxRouter.Match(getReq, &getMatch)

	// Test PUT route
	putReq := httptest.NewRequest(http.MethodPut, "/api/v2/analysis", nil)
	var putMatch mux.RouteMatch
	foundPut := muxRouter.Match(putReq, &putMatch)

	assert.True(t, foundGet, "GET /api/v2/analysis route should be registered")
	assert.True(t, foundPut, "PUT /api/v2/analysis route should be registered")
}
