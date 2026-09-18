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

package middleware_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressionMiddleware_CompressesResponseWhenGzipIsAccepted(t *testing.T) {
	var (
		router       = mux.NewRouter()
		responseBody = "compressible body"
	)

	router.Use(middleware.CompressionMiddleware)
	router.HandleFunc("/teapot", func(response http.ResponseWriter, request *http.Request) {
		_, err := response.Write([]byte(responseBody))
		require.NoError(t, err)
	})

	request := httptest.NewRequest(http.MethodGet, "/teapot", nil)
	request.Header.Set(headers.AcceptEncoding.String(), "gzip")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "gzip", recorder.Header().Get(headers.ContentEncoding.String()))

	gzipReader, err := gzip.NewReader(recorder.Body)
	require.NoError(t, err)
	defer gzipReader.Close()

	decompressedBody, err := io.ReadAll(gzipReader)
	require.NoError(t, err)
	assert.Equal(t, responseBody, string(decompressedBody))
}

func TestCompressionMiddleware_SkipsCompressionForNamedRoute(t *testing.T) {
	var (
		router       = mux.NewRouter()
		responseBody = "download data"
	)

	router.Use(middleware.CompressionMiddleware)
	router.HandleFunc("/teapot", func(response http.ResponseWriter, request *http.Request) {
		_, err := response.Write([]byte(responseBody))
		require.NoError(t, err)
	}).Name(middleware.SkipCompressionMiddleware)

	request := httptest.NewRequest(http.MethodGet, "/teapot", nil)
	request.Header.Set(headers.AcceptEncoding.String(), "gzip")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Header().Get(headers.ContentEncoding.String()))
	assert.Equal(t, responseBody, recorder.Body.String())
}
