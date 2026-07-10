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
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

// URIPathVariableNodeID is the name of the path variable carrying the
// graph-assigned node id.
const URIPathVariableNodeID = "node_id"

// NodeKindView is the JSON shape of a kind embedded in a NodeView.
// NodeKindID is nil when the kind is not registered in schema_node_kinds.
type NodeKindView struct {
	NodeKindID *int32 `json:"node_kind_id"`
	Name       string `json:"name"`
}

// NodeView is the JSON shape returned by the node handlers. It is
// decoupled from services.Node so the wire format can evolve independently of
// the domain model.
type NodeView struct {
	NodeID     int64          `json:"node_id"`
	Kinds      []NodeKindView `json:"kinds"`
	Properties map[string]any `json:"properties"`
}

// BuildNodeView projects a services.Node into the view type the handlers
// return in their JSON envelope.
func BuildNodeView(node services.Node) NodeView {
	var kinds []NodeKindView

	for _, kind := range node.Kinds {
		kinds = append(kinds, NodeKindView{
			NodeKindID: kind.ID,
			Name:       kind.Name,
		})
	}

	return NodeView{
		NodeID:     node.ID,
		Kinds:      kinds,
		Properties: node.Properties,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s NodeView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// GetNodeByID returns the details of a graph node identified by its
// graph-assigned integer id. Returns 400 when the id is malformed, 404 when the
// node or its kinds cannot be found, and 200 with the node details otherwise.
func (s Handlers) GetNodeByID(response http.ResponseWriter, request *http.Request) {
	var (
		ctx       = request.Context()
		rawNodeID = mux.Vars(request)[URIPathVariableNodeID]
	)

	if nodeID, err := strconv.ParseInt(rawNodeID, 10, 64); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "node id is malformed", response)
	} else if node, err := s.graphDB.GetNode(ctx, nodeID); errors.Is(err, services.ErrNodeNotFound) || errors.Is(err, services.ErrKindNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, "node not found", response)
	} else if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
	} else if !s.nodeAuthorizer.CanAccessNode(ctx, node) {
		responses.WriteError(ctx, http.StatusForbidden, "forbidden", response)
	} else {
		responses.WriteBasic(ctx, BuildNodeView(node), http.StatusOK, response)
	}
}
