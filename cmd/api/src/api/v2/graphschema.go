// Copyright 2025 Specter Ops, Inc.
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
package v2

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (s *Resources) ListEdgeTypes(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()
	if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)

	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.GraphSchemaRelationshipKind{}, name); err != nil {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
					queryFilters[name][i].IsStringData = model.GraphSchemaRelationshipKind{}.IsStringColumn(filter.Name)
				}
			}
		}

		translatedQueryFilters := translateQueryFilters(queryFilters)

		if edges, _, err := s.DB.GetGraphSchemaRelationshipKindsWithSchemaName(ctx, translatedQueryFilters, model.Sort{}, 0, 0); err != nil {
			api.HandleDatabaseError(request, response, err)

		} else {
			api.WriteBasicResponse(ctx, edges, http.StatusOK, response)
		}
	}
}

// translateQueryFilters takes the queryParameterFilterMap filters and translates them into the column names and struct type the database expects. Any future filters added will need to be added to the Filters struct here.
func translateQueryFilters(queryFilters model.QueryParameterFilterMap) model.Filters {
	translatedFilters := make(model.QueryParameterFilterMap)
	if len(queryFilters["is_traversable"]) > 0 {
		translatedFilters["is_traversable"] = queryFilters["is_traversable"]
	}

	if len(queryFilters["schema_names"]) > 0 {
		translatedFilters["schema.name"] = make([]model.QueryParameterFilter, len(queryFilters["schema_names"]))
		for i, schema_name_filter := range queryFilters["schema_names"] {
			schema_name_filter.SetOperator = model.FilterOr
			translatedFilters["schema.name"][i] = schema_name_filter
		}
	}

	return translatedFilters.ToFiltersModel()
}
