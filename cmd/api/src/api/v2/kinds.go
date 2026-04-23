// Copyright 2025 Specter Ops, Inc.
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

package v2

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

func isExtendedNodeKind(kind graph.Kind) bool {
	return strings.HasPrefix(kind.String(), model.AssetGroupTagKindPrefix) || kind.Is(graph.StringKind("Meta"), graph.StringKind("MetaDetail"), common.MigrationData)
}

type ListKindsResponse struct {
	Kinds graph.Kinds `json:"kinds"`
}

// ListKinds returns all node kinds, edge kinds, and tier tags present in the system.
// It is a comprehensive view of the various kinds the graph currently recognizes.
func (s Resources) ListKinds(response http.ResponseWriter, request *http.Request) {
	if kinds, err := s.Graph.FetchKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.Kind{}, name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for _, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
				}
			}
		}

		if len(queryFilters) > 0 {
			if filters, ok := queryFilters["type"]; ok {
				// This gets both custom node kinds (schemaless) and schema based node kinds
				// Only kinds that have been synced to the kinds table will be included in this.
				displayKinds, err := s.DB.GetPrimaryDisplayKinds(request.Context())
				if err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				}

				// Source kinds are node kinds
				sourceKindsByKind := make(map[graph.Kind]bool)
				if sourceKinds, err := s.DB.GetSourceKinds(request.Context()); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else {
					for _, kind := range sourceKinds {
						sourceKindsByKind[kind.ToKind()] = true
					}
				}

				// Filter down kinds
				var filteredKinds = graph.Kinds{}
				for _, filter := range filters {
					for _, kind := range kinds {
						var isNodeKind bool
						// Asset group tags are node kinds as well as meta / migrationData kinds
						if _, ok := displayKinds[kind]; ok || sourceKindsByKind[kind] || isExtendedNodeKind(kind) {
							isNodeKind = true
						}

						switch filter.Value {
						case "node":
							if (isNodeKind && filter.Operator == model.Equals) || (!isNodeKind && filter.Operator == model.NotEquals) {
								filteredKinds = append(filteredKinds, kind)
							}
						case "edge":
							if (!isNodeKind && filter.Operator == model.Equals) || (isNodeKind && filter.Operator == model.NotEquals) {
								filteredKinds = append(filteredKinds, kind)
							}
						}
					}
				}
				kinds = filteredKinds
			}
		}

		// Alpha sort
		slices.SortFunc(kinds, func(a, b graph.Kind) int {
			return strings.Compare(a.String(), b.String())
		})

		api.WriteBasicResponse(request.Context(), ListKindsResponse{Kinds: kinds}, http.StatusOK, response)
	}
}

type ListSourceKindsResponse struct {
	Kinds []model.SourceKind `json:"kinds"`
}

// ListSourceKinds returns only the subset of kinds that are registered as source kinds.
//
// Source kinds typically represent the origin of ingested data, such as Base, AZBase,
// or OpenGraph-related node kinds.
func (s Resources) ListSourceKinds(response http.ResponseWriter, request *http.Request) {
	if kinds, err := s.DB.GetSourceKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// inject 0, Sourceless into the payload. We don't track this as an official kind
		// but it will facilitate delete requests for data that isn't associated with a kind.
		kinds = append(kinds, model.SourceKind{ID: 0, Name: "Sourceless"})
		api.WriteBasicResponse(request.Context(), ListSourceKindsResponse{Kinds: kinds}, http.StatusOK, response)
	}
}
