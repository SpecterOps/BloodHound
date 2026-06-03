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
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database"
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
	ctx := request.Context()
	if kinds, err := s.Graph.FetchKinds(ctx); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if queryFilters, err := model.NewQueryParameterFilterParser().ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else {
		for name, filters := range queryFilters {
			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(model.Kind{}, name); err != nil {
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for _, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
				}
			}
		}

		if len(queryFilters) > 0 {
			if filters, ok := queryFilters["type"]; ok {
				validNodeKinds := make(map[graph.Kind]bool)

				// Schema based node kinds
				if schemaNodeKinds, _, err := s.DB.GetGraphSchemaNodeKinds(ctx, model.Filters{}, model.Sort{}, 0, 0); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else {
					for _, kind := range schemaNodeKinds {
						validNodeKinds[kind.ToKind()] = true
					}
				}

				// Schemaless customnode kinds
				if customNodeKinds, err := s.DB.GetCustomNodeKinds(ctx); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else {
					var customNames []string
					for _, kind := range customNodeKinds {
						customNames = append(customNames, kind.KindName)
					}
					// Until work is complete to ensure custom_node_kinds are properly kind backed, this will filter out invalid kinds
					if kinds, err := s.DB.GetKindsByNames(ctx, customNames...); err != nil && !errors.Is(err, database.ErrNotFound) {
						api.HandleDatabaseError(request, response, err)
						return
					} else {
						for _, kind := range kinds {
							validNodeKinds[kind.ToKind()] = true
						}
					}
				}

				// Source kinds
				if sourceKinds, err := s.DB.GetSourceKinds(ctx); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else {
					for _, kind := range sourceKinds {
						validNodeKinds[kind.ToKind()] = true
					}
				}

				// Filter down kinds
				var filteredKinds = graph.Kinds{}
				var kindsSeen = make(map[graph.Kind]bool)
				for _, filter := range filters {
					for _, kind := range kinds {
						var isNodeKind bool
						// Asset group tags are node kinds as well as meta / migrationData kinds
						if validNodeKinds[kind] || isExtendedNodeKind(kind) {
							isNodeKind = true
						}

						switch filter.Value {
						case "node":
							if (isNodeKind && filter.Operator == model.Equals) || (!isNodeKind && filter.Operator == model.NotEquals) {
								if !kindsSeen[kind] {
									filteredKinds = append(filteredKinds, kind)
									kindsSeen[kind] = true
								}
							}
						case "edge":
							if (!isNodeKind && filter.Operator == model.Equals) || (isNodeKind && filter.Operator == model.NotEquals) {
								if !kindsSeen[kind] {
									filteredKinds = append(filteredKinds, kind)
									kindsSeen[kind] = true
								}
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
