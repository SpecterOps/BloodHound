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

package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/dataquality/internal/handlers"
	handlermocks "github.com/specterops/bloodhound/server/dataquality/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/dataquality/internal/routes"
	"github.com/specterops/bloodhound/server/dataquality/internal/routes/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRequiresAuthentication(t *testing.T) {
	t.Parallel()
	var (
		routerInstance = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
		service        = handlermocks.NewMockDataQuality(t)
		featureFlags   = mocks.NewMockFeatureFlags(t)
		request        = httptest.NewRequest(http.MethodGet, "/api/v2/data-quality-stats-aggregations", nil)
		recorder       = httptest.NewRecorder()
		match          mux.RouteMatch
	)
	routes.Register(&routerInstance, handlers.NewHandlersContainer(service), featureFlags)
	assert.True(t, routerInstance.MuxRouter().Match(request, &match))
	routerInstance.Handler().ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
