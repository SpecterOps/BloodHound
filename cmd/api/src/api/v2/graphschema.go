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
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
)

func (s *Resources) ListEdgeTypes(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	if openGraphSearchFeatureFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureOpenGraphSearch); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if !openGraphSearchFeatureFlag.Enabled {
		api.WriteBasicResponse(ctx, model.GraphSchemaEdgeKindsWithNamedSchema{}, http.StatusOK, response)
	} else {
		if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)

		} else {
			for name, filters := range queryFilters {
				if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.GraphSchemaEdgeKind{}, name); err != nil {
					api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
					return
				} else {
					for i, filter := range filters {
						if !slices.Contains(validPredicates, string(filter.Operator)) {
							api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
							return
						}
						queryFilters[name][i].IsStringData = model.GraphSchemaEdgeKind{}.IsStringColumn(filter.Name)
					}
				}
			}

			// translate schema_names filter into the column format the DB expects
			if len(queryFilters["schema_names"]) > 0 {
				queryFilters["schema.name"] = queryFilters["schema_names"]
				delete(queryFilters, "schema_names")
				for i, schema_name_filter := range queryFilters["schema.name"] {
					schema_name_filter.SetOperator = model.FilterOr
					queryFilters["schema.name"][i] = schema_name_filter
				}
			}

			if edges, _, err := s.DB.GetGraphSchemaEdgeKindsWithSchemaName(ctx, queryFilters.ToFiltersModel(), model.Sort{}, 0, 0); err != nil {
				api.HandleDatabaseError(request, response, err)

			} else {
				api.WriteBasicResponse(ctx, edges, http.StatusOK, response)
			}
		}
	}

}
