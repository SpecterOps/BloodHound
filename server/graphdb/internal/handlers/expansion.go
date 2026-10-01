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

package handlers

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"slices"

	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util"
)

type GraphExpansionPayload struct {
	NodeID            *int64 `json:"node_id"`
	Direction         string `json:"direction"`
	Limit             int    `json:"limit,omitempty"`
	IncludeProperties bool   `json:"include_properties,omitempty"`
}

type GraphExpansionResponse struct {
	NodeKeys  []string                     `json:"node_keys,omitempty"`
	EdgeKeys  []string                     `json:"edge_keys,omitempty"`
	Nodes     map[string]model.UnifiedNode `json:"nodes"`
	Edges     []model.UnifiedEdge          `json:"edges"`
	Literals  graph.Literals               `json:"literals"`
	Limit     int                          `json:"limit"`
	Truncated bool                         `json:"truncated"`
}

func (s GraphExpansionResponse) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

func (s Handlers) ExpandGraph(response http.ResponseWriter, request *http.Request) {
	var payload GraphExpansionPayload

	user, isUser := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx)
	if !isUser {
		slog.Error("Unable to get user from auth context")
		responses.WriteError(request.Context(), http.StatusInternalServerError, "unknown user", response)
		return
	}

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		responses.WriteError(request.Context(), http.StatusBadRequest, "JSON malformed.", response)
		return
	}

	if payload.NodeID == nil {
		responses.WriteError(request.Context(), http.StatusBadRequest, "node_id is required", response)
		return
	}

	if *payload.NodeID < 0 {
		responses.WriteError(request.Context(), http.StatusBadRequest, "node_id must be greater than or equal to 0", response)
		return
	}

	expansion, err := s.graphDB.ExpandGraph(request.Context(), services.GraphExpansionRequest{
		NodeID:            *payload.NodeID,
		Direction:         payload.Direction,
		Limit:             payload.Limit,
		IncludeProperties: payload.IncludeProperties,
		User:              user,
	})
	if err == nil {
		responses.WriteBasic(request.Context(), buildGraphExpansionResponse(expansion, payload.IncludeProperties), http.StatusOK, response)
	} else if errors.Is(err, services.ErrInvalidGraphExpansionLimit) ||
		errors.Is(err, services.ErrInvalidGraphExpansionDirection) ||
		errors.Is(err, services.ErrInvalidGraphExpansionRelationshipKind) {
		responses.WriteError(request.Context(), http.StatusBadRequest, err.Error(), response)
	} else if errors.Is(err, services.ErrNodeNotFound) {
		responses.WriteError(request.Context(), http.StatusNotFound, "node not found", response)
	} else if errors.Is(err, services.ErrNodeAccessDenied) {
		responses.WriteError(request.Context(), http.StatusForbidden, "forbidden", response)
	} else if util.IsNeoTimeoutError(err) {
		responses.WriteError(request.Context(), http.StatusInternalServerError, "transaction timed out, reduce query complexity or try again later", response)
	} else {
		responses.WriteInternalServerError(request.Context(), err, response)
	}
}

func buildGraphExpansionResponse(expansion services.GraphExpansion, includeProperties bool) GraphExpansionResponse {
	if includeProperties {
		nodeKeys := map[string]struct{}{}
		edgeKeys := map[string]struct{}{}

		for _, node := range expansion.Graph.Nodes {
			for key := range node.Properties {
				nodeKeys[key] = struct{}{}
			}
		}

		for _, edge := range expansion.Graph.Edges {
			for key := range edge.Properties {
				edgeKeys[key] = struct{}{}
			}
		}

		return GraphExpansionResponse{
			NodeKeys:  slices.Sorted(maps.Keys(nodeKeys)),
			EdgeKeys:  slices.Sorted(maps.Keys(edgeKeys)),
			Nodes:     expansion.Graph.Nodes,
			Edges:     expansion.Graph.Edges,
			Literals:  expansion.Graph.Literals,
			Limit:     expansion.Limit,
			Truncated: expansion.Truncated,
		}
	}

	return GraphExpansionResponse{
		Nodes:     expansion.Graph.Nodes,
		Edges:     expansion.Graph.Edges,
		Literals:  expansion.Graph.Literals,
		Limit:     expansion.Limit,
		Truncated: expansion.Truncated,
	}
}
