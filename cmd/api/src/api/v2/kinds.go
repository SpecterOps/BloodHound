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
	"net/http"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

func isExtendedNodeKind(kind graph.Kind) bool {
	return strings.HasPrefix(kind.String(), model.AssetGroupTagKindPrefix) || kind.Is(graph.StringKind("Meta"), graph.StringKind("MetaDetails"), common.MigrationData)
}

type ListKindsResponse struct {
	Kinds     graph.Kinds `json:"kinds"`
	NodeKinds graph.Kinds `json:"node_kinds"`
	EdgeKinds graph.Kinds `json:"edge_kinds"`
}

// ListKinds returns all node kinds, edge kinds, and tier tags present in the system.
// It is a comprehensive view of the various kinds the graph currently recognizes.
func (s Resources) ListKinds(response http.ResponseWriter, request *http.Request) {
	var (
		resp = ListKindsResponse{}
		err  error
	)

	if resp.Kinds, err = s.Graph.FetchKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Alpha sort kind list
		slices.SortFunc(resp.Kinds, func(a, b graph.Kind) int {
			return strings.Compare(a.String(), b.String())
		})

		// This gets both custom node kinds (schemaless) and schema based node kinds
		// Note: currently this is not an exhaustive list until all the custom-node sync work is complete
		if displayKinds, err := s.DB.GetPrimaryDisplayKinds(request.Context()); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			// Add nodes and edge kinds for UI
			for _, kind := range resp.Kinds {
				if _, ok := displayKinds[kind]; ok {
					resp.NodeKinds = append(resp.NodeKinds, kind)
					// Asset group tags are node kinds as well as meta kinds
				} else if isExtendedNodeKind(kind) {
					resp.NodeKinds = append(resp.NodeKinds, kind)
				} else {
					resp.EdgeKinds = append(resp.EdgeKinds, kind)
				}
			}
		}

		api.WriteBasicResponse(request.Context(), resp, http.StatusOK, response)
	}
}

type ListDisplayKindsResponse struct {
	NodeKinds map[string]graphschema.DisplayKind `json:"node_kinds"`
	EdgeKinds map[string]graphschema.DisplayKind `json:"edge_kinds"`
}

// ListDisplayKinds returns kinds to be mapped to enriched data like display name, icon, etc
func (s Resources) ListDisplayKinds(response http.ResponseWriter, request *http.Request) {
	var resp = ListDisplayKindsResponse{NodeKinds: make(map[string]graphschema.DisplayKind), EdgeKinds: make(map[string]graphschema.DisplayKind)}

	if kinds, err := s.Graph.FetchKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Alpha sort kind list
		slices.SortFunc(kinds, func(a, b graph.Kind) int {
			return strings.Compare(a.String(), b.String())
		})

		// This gets both custom node kinds (schemaless) and schema based node kinds
		// Note: currently this is not an exhaustive list until all the custom-node sync work is complete
		if displayKinds, err := s.DB.GetPrimaryDisplayKinds(request.Context()); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if extensions, _, err := s.DB.GetGraphSchemaExtensions(request.Context(), model.Filters{}, model.Sort{}, 0, 0); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			// Build map to link extension display name to id to enrich the display kind
			extensionNameById := make(map[int]string)
			for _, extension := range extensions {
				extensionNameById[int(extension.ID)] = extension.DisplayName
			}

			// Sort nodes and edge kinds for UI
			for _, kind := range kinds {
				if nodeKind, ok := displayKinds[kind]; ok {
					// Enrich display kind with extension display name so nodes can be labeled "ExtensionX | NodeKindY"
					if extensionDisplayName, ok := extensionNameById[nodeKind.ExtensionId]; ok {
						nodeKind.ExtensionDisplayName = extensionDisplayName
					} else {
						nodeKind.ExtensionDisplayName = "Open Graph" // Fallback to Open graph when schemaless
					}
					resp.NodeKinds[kind.String()] = nodeKind
					// Asset group tags are node kinds as well as meta kinds, they do not have icons
				} else if isExtendedNodeKind(kind) {
					resp.NodeKinds[kind.String()] = graphschema.DisplayKind{
						Name:        kind.String(),
						DisplayName: kind.String(),
					}
				} else {
					resp.EdgeKinds[kind.String()] = graphschema.DisplayKind{
						Name:        kind.String(),
						DisplayName: kind.String(),
					}
				}
			}
		}

		api.WriteBasicResponse(request.Context(), resp, http.StatusOK, response)
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
