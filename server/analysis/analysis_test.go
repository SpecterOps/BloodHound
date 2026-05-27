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

package analysis_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/server/analysis"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	var (
		cfg        = config.Configuration{}
		authorizer = auth.NewAuthorizer(nil)
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	// This should not panic and should register routes
	// Pass nil pool since we're only testing route registration, not actual DB operations
	analysis.Register(&routerInst, nil)

	// Verify that analysis routes were registered by attempting to match them
	muxRouter := routerInst.MuxRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/analysis", nil)
	var match mux.RouteMatch
	foundAnalysisRoute := muxRouter.Match(req, &match)

	assert.True(t, foundAnalysisRoute, "analysis routes should be registered")
}
