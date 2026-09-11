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
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
)

// SortMiddleware parses the sort_by query parameters from the request, validates them against the supplied
// params.Sortable definition, and enriches the BloodHound context with the resulting params.SortItems.
//
// When sortable is nil the middleware performs no parsing and passes the request through unchanged. If a
// sort field is empty or references a field that cannot be sorted, the middleware writes a 400 response
// and halts the chain.
func SortMiddleware(sortable params.Sortable) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if sortable == nil {
				next.ServeHTTP(response, request)
				return
			}

			parser := params.NewQueryParameterSortParser()
			if parsedSort, err := parser.ParseAndValidate(request.URL.Query(), sortable); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, sortErrorMessage(err), request), response)
				return
			} else {
				bhCtx := bhctx.Get(request.Context())
				bhCtx.Sort = parsedSort
			}

			next.ServeHTTP(response, request)
		})
	}
}

// sortErrorMessage maps a sort validation failure to the appropriate API error response detail, preserving
// the offending field in the message where available.
func sortErrorMessage(err error) string {
	if validationErr, ok := errors.AsType[*params.SortValidationError](err); !ok {
		return api.ErrorResponseDetailsNotSortable
	} else {
		switch {
		case errors.Is(validationErr, params.ErrFieldEmpty):
			return api.ErrorResponseEmptySortParameter
		case errors.Is(validationErr, params.ErrFieldNotSortable):
			return fmt.Sprintf("%s: %s", api.ErrorResponseDetailsNotSortable, validationErr.Field)
		default:
			return api.ErrorResponseDetailsNotSortable
		}
	}
}
