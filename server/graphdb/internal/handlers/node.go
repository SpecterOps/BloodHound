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

type KindInfoView struct {
	Name       string       `json:"name"`
	Title      string       `json:"title"`
	Position   int32        `json:"position"`
	NodeKindID int          `json:"node_kind_id"`
	Markdown   MarkdownView `json:"markdown"`
}

type MarkdownView struct {
	Content string `json:"content"`
}

type kindInfoContentView struct {
	Markdown MarkdownView `json:"markdown"`
}

// NodeView is the JSON shape returned by the node handlers. It is
// decoupled from services.Node so the wire format can evolve independently of
// the domain model.
type NodeView struct {
	NodeID     int64          `json:"node_id"`
	Kinds      []NodeKindView `json:"kinds"`
	Properties map[string]any `json:"properties"`
	KindInfos  []KindInfoView `json:"info,omitempty"`
}

func BuildNodeView(node services.Node, includeInfo bool) NodeView {
	kinds := []NodeKindView{}

	for _, kind := range node.Kinds {
		kinds = append(kinds, NodeKindView{
			NodeKindID: kind.ID,
			Name:       kind.Name,
		})
	}

	nodeView := NodeView{
		NodeID:     node.ID,
		Kinds:      kinds,
		Properties: node.Properties,
	}

	if includeInfo {
		for _, kindInfo := range node.KindInfos {
			nodeView.KindInfos = append(nodeView.KindInfos, KindInfoView{
				Name:       kindInfo.InfoKey,
				Title:      kindInfo.Title,
				Position:   kindInfo.Position,
				NodeKindID: int(*kindInfo.NodeKindID),
				Markdown:   buildMarkdownView(kindInfo.Content),
			})
		}
	}

	return nodeView
}

func buildMarkdownView(content json.RawMessage) MarkdownView {
	var contentView kindInfoContentView

	if err := json.Unmarshal(content, &contentView); err != nil {
		return MarkdownView{}
	}

	return contentView.Markdown
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
		ctx            = request.Context()
		nodeIDRaw      = mux.Vars(request)[URIPathVariableNodeID]
		includeInfoRaw = request.URL.Query().Get("include-info")
		includeInfo    bool
		err            error
	)

	if includeInfoRaw != "" {
		if includeInfo, err = strconv.ParseBool(includeInfoRaw); err != nil {
			responses.WriteError(ctx, http.StatusBadRequest, "include-info is malformed", response)
			return
		}
	}

	if nodeID, err := strconv.ParseInt(nodeIDRaw, 10, 64); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "node id is malformed", response)
	} else if node, err := s.graphDB.GetNode(ctx, nodeID, includeInfo); errors.Is(err, services.ErrNodeNotFound) || errors.Is(err, services.ErrKindNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, "node not found", response)
	} else if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
	} else {
		responses.WriteBasic(ctx, BuildNodeView(node, includeInfo), http.StatusOK, response)
	}
}
