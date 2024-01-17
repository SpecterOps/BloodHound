// Copyright 2023 Specter Ops, Inc.
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
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/src/api/middleware"
)

func TestServeHTTP_Success(t *testing.T) {
	handlerFunc := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		api.WriteBasicResponse(request.Context(), "success", http.StatusOK, response)
	})
	wrapper := middleware.NewWrapper(handlerFunc)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "foo/bar", nil)
	require.Nil(t, err)

	req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	response := httptest.NewRecorder()

	wrapper.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "success")
}

func TestContextMiddleware_InvalidWait(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if req, err := http.NewRequestWithContext(ctx, http.MethodOptions, "/foo", nil); err != nil {
		t.Error(err)
	} else {
		q := url.Values{}
		req.Header.Set(headers.Prefer.String(), "wait=1.5")
		req.URL.RawQuery = q.Encode()
		rr := httptest.NewRecorder()

		middleware.ContextMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "header has an invalid value")
	}
}

func TestContextMiddleware(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if req, err := http.NewRequestWithContext(ctx, http.MethodOptions, "/foo", nil); err != nil {
		t.Error(err)
	} else {
		q := url.Values{}
		req.Header.Set(headers.Prefer.String(), "wait=1")
		req.URL.RawQuery = q.Encode()
		rr := httptest.NewRecorder()

		middleware.ContextMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestParseHeaderValues(t *testing.T) {
	values := middleware.ParseHeaderValues("     respond-async, wait=10,other =  z=   ")

	require.Equal(t, "", values["respond-async"])
	require.Equal(t, "10", values["wait"])
	require.Equal(t, "z=", values["other"])
}

func TestCORSMiddleware(t *testing.T) {
	if req, err := http.NewRequest(http.MethodOptions, "/foo", nil); err != nil {
		t.Error(err)
	} else {
		req.Header.Add(headers.Origin.String(), "")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {})
		middleware.CORSMiddleware()(handler).ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Errorf("bad status: got %v want %v", got, want)
		}

		// ACAO header should not be set to prevent browser agent CORS requests
		if got, want := rr.Header().Get(headers.AccessControlAllowOrigin.String()), ""; got != want {
			t.Errorf("bad access-control-allow-origin header: got '%v' want '%v'", got, want)
		}
	}
}
