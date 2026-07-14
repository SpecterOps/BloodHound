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

package middleware

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
)

// PagingMiddleware parses the skip and limit query parameters from the request according to the supplied
// params.PagingConfig and enriches the BloodHound context with the resulting values. An absent skip
// yields zero and an absent limit yields the configured default. If either parameter is malformed, the
// middleware writes a 400 response and halts the chain.
func PagingMiddleware(config params.PagingConfig) mux.MiddlewareFunc {
	parser := params.NewPagingParser(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if paging, err := parser.ParseAndValidate(request.URL.Query()); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, pagingErrorMessage(err), request), response)
				return
			} else {
				bhCtx := bhctx.Get(request.Context())
				bhCtx.Skip = paging.Skip
				bhCtx.Limit = paging.Limit
			}

			next.ServeHTTP(response, request)
		})
	}
}

// pagingErrorMessage maps a paging validation failure to the appropriate API error response detail,
// preserving the offending field and operator in the message where available.
func pagingErrorMessage(err error) string {
	if validationErr, ok := errors.AsType[*params.PagingValidationError](err); !ok {
		return api.ErrorResponseDetailsBadQueryParameterFilters
	} else {
		switch {
		case errors.Is(validationErr, params.ErrInvalidInteger):
			return validationErr.Error()
		default:
			return api.ErrorResponseDetailsBadQueryParameterFilters
		}
	}
}
