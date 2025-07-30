// Copyright 2024 Specter Ops, Inc.
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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetCollectorManifest(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		manifests = config.CollectorManifests{"sharphound": config.CollectorManifest{}, "azurehound": config.CollectorManifest{}}
		resources = v2.Resources{
			CollectorManifests: manifests,
		}
	)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/collectors/%s"

	t.Run("sharphound", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "sharphound"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("azurehound", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "azurehound"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("invalid", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "invalid"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		resources := v2.Resources{CollectorManifests: map[string]config.CollectorManifest{}}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf(endpoint, "azurehound"), nil)
		require.NoError(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc("/api/v2/collectors/{collector_type}", resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}
