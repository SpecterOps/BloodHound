// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/stretchr/testify/assert"
)

type filterTestSchema struct{}

func (s filterTestSchema) ValidFilters() map[string]params.FilterableField {
	equality := []params.FilterOperator{params.Equals, params.NotEquals}

	return map[string]params.FilterableField{
		"foo": {Operators: equality},
		"bar": {Operators: equality},
	}
}

func TestFilterMiddleware(t *testing.T) {
	tests := []struct {
		name                        string
		requestTarget               string
		filterable                  params.Filterable
		additionalIgnoredParameters []string
		wantStatus                  int
		wantNextCalled              bool
		wantQueryParameters         url.Values
		wantFilters                 params.Filters
	}{
		{
			name:           "parses a supported filter",
			requestTarget:  "/endpoint?foo=eq:value",
			filterable:     filterTestSchema{},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantFilters: params.Filters{
				"foo": {
					{
						Field:       "foo",
						Operator:    params.Equals,
						Value:       "value",
						SetOperator: params.FilterAnd,
					},
				},
			},
		},
		{
			name:                        "parses a supported filter and skips an ignored filter",
			requestTarget:               "/endpoint?foo=eq:value&environment_id=tenant:prod",
			filterable:                  filterTestSchema{},
			additionalIgnoredParameters: []string{"environment_id"},
			wantStatus:                  200,
			wantNextCalled:              true,
			wantQueryParameters: url.Values{
				"foo":            {"eq:value"},
				"environment_id": {"tenant:prod"},
			},
			wantFilters: params.Filters{
				"foo": {
					{
						Field:       "foo",
						Operator:    params.Equals,
						Value:       "value",
						SetOperator: params.FilterAnd,
					},
				},
			},
		},
		{
			name:           "rejects a non-filterable param",
			requestTarget:  "/endpoint?baz=eq:value",
			filterable:     filterTestSchema{},
			wantStatus:     400,
			wantNextCalled: false,
		},
		{
			name:           "rejects a malformed predicate",
			requestTarget:  "/endpoint?bar=invalid:value",
			filterable:     filterTestSchema{},
			wantStatus:     400,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				bloodhoundContext = &bhctx.Context{}
				nextCalled        = false
			)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				assert.Equal(t, tt.wantFilters, bhctx.Get(r.Context()).Filters)

				if tt.wantQueryParameters != nil {
					assert.Equal(t, tt.wantQueryParameters, r.URL.Query())
				}

				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.FilterMiddleware(tt.filterable, tt.additionalIgnoredParameters...)(next)

			request := httptest.NewRequest(http.MethodGet, tt.requestTarget, nil)
			request = request.WithContext(bhctx.Set(request.Context(), bloodhoundContext))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			assert.Equal(t, tt.wantNextCalled, nextCalled)
			assert.Equal(t, tt.wantStatus, response.Code)
		})
	}
}
