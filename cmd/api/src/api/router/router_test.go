// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	databaseMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_UsePostroutingAppliesRateLimitToRoutesRegisteredAfterMiddleware(t *testing.T) {
	var (
		mockController = gomock.NewController(t)
		mockDatabase   = databaseMocks.NewMockDatabase(mockController)
		routerInst     = router.NewRouter(config.Configuration{}, auth.NewAuthorizer(nil), "")
	)

	mockDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.TrustedProxiesConfig).Return(appcfg.Parameter{}, nil).AnyTimes()
	routerInst.UsePostrouting(middleware.RateLimitMiddleware(mockDatabase, 1))
	routerInst.GET("/api/v2/test", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})

	for requestNumber, expectedStatus := range []int{http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/api/v2/test", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()

		routerInst.Handler().ServeHTTP(response, request)

		require.Equal(t, expectedStatus, response.Code, "request %d should have the expected rate-limit result", requestNumber+1)
	}
}
