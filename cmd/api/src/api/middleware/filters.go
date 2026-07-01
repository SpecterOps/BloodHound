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
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/filters"
)

// FilterMiddleware parses query parameter filters from the request, validates them against the supplied
// filters.Filterable definition, and enriches the BloodHound context with the resulting filters.Filters map.
//
// When filterable is nil the middleware performs no parsing and passes the request through unchanged. If
// the filters are malformed, reference a column that cannot be filtered, or use an operator the column
// does not support, the middleware writes a 400 response and halts the chain.
func FilterMiddleware(filterable filters.Filterable) mux.MiddlewareFunc {
	if filterable == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	parser := filters.NewQueryParameterFilterParser(append(model.IgnoreFilters(), model.AllPaginationQueryParameters()...)...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if parsedFilters, err := parser.ParseAndValidate(request.URL.Query(), filterable); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, filterErrorMessage(err), request), response)
				return
			} else {
				bhCtx := bhctx.Get(request.Context())
				bhCtx.Filters = parsedFilters
			}

			next.ServeHTTP(response, request)
		})
	}
}

// filterErrorMessage maps a filter validation failure to the appropriate API error response detail,
// preserving the offending field and operator in the message where available.
func filterErrorMessage(err error) string {
	var validationErr *filters.ValidationError
	if !errors.As(err, &validationErr) {
		return api.ErrorResponseDetailsBadQueryParameterFilters
	}

	switch {
	case errors.Is(validationErr, filters.ErrFieldNotFilterable):
		return fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, validationErr.Field)
	case errors.Is(validationErr, filters.ErrOperatorNotSupported):
		return fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, validationErr.Field, validationErr.Operator)
	default:
		return api.ErrorResponseDetailsBadQueryParameterFilters
	}
}
