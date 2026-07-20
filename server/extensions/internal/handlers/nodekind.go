// Copyright 2026 Specter Ops, Inc.
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
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

// URIPathVariableNodeKindID is the name of the path variable carrying the
// schema_node_kinds row id.
const URIPathVariableNodeKindID = "node_kind_id"

// NodeKindView is the JSON shape returned by the node kind handlers. It is
// decoupled from services.NodeKind so the wire format can evolve independently of
// the domain model.
type NodeKindView struct {
	NodeKindID    int32                   `json:"node_kind_id"`
	Name          string                  `json:"name"`
	DisplayName   string                  `json:"display_name"`
	Description   string                  `json:"description"`
	IsDisplayKind bool                    `json:"is_display_kind"`
	Icon          string                  `json:"icon"`
	Color         string                  `json:"color"`
	Info          map[string]KindInfoView `json:"info"`
}

// KindInfoView is the JSON shape of a kind info embedded in a NodeKindView.
type KindInfoView struct {
	Title    string       `json:"title"`
	Position int32        `json:"position"`
	Markdown MarkdownView `json:"markdown"`
}

// MarkdownView is the JSON shape of the flattened markdown content.
type MarkdownView struct {
	Content string `json:"content"`
}

// kindInfoContentView unwraps the stored content JSON object shaped
// {"markdown": {"content": "..."}} into its inner MarkdownView.
type kindInfoContentView struct {
	Markdown MarkdownView `json:"markdown"`
}

// BuildNodeKindView maps a services.NodeKind onto its wire representation, keying the
// info entries by their info_key and flattening each stored markdown content object.
// Entries whose content fails to parse are retained with an empty markdown view, and
// the joined parse errors are returned so the caller can log them.
func BuildNodeKindView(nodeKind services.NodeKind) (NodeKindView, error) {
	var markdownErr error

	nodeKindView := NodeKindView{
		NodeKindID:    nodeKind.ID,
		Name:          nodeKind.Name,
		DisplayName:   nodeKind.DisplayName,
		Description:   nodeKind.Description,
		IsDisplayKind: nodeKind.IsDisplayKind,
		Icon:          nodeKind.Icon,
		Color:         nodeKind.Color,
		Info:          map[string]KindInfoView{},
	}

	for _, info := range nodeKind.Info {
		markdown, err := buildMarkdownView(info.Content)
		if err != nil {
			markdownErr = errors.Join(markdownErr, err)
		}

		nodeKindView.Info[info.InfoKey] = KindInfoView{
			Title:    info.Title,
			Position: info.Position,
			Markdown: markdown,
		}
	}

	return nodeKindView, markdownErr
}

// buildMarkdownView unwraps the stored content object into its inner MarkdownView,
// returning an empty MarkdownView and an error when the content cannot be parsed.
func buildMarkdownView(content json.RawMessage) (MarkdownView, error) {
	var contentView kindInfoContentView

	if err := json.Unmarshal(content, &contentView); err != nil {
		return MarkdownView{}, fmt.Errorf("unmarshalling markdown content: %w", err)
	}

	return contentView.Markdown, nil
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s NodeKindView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// GetNodeKindByID returns the details of a node kind identified by its schema_node_kinds
// row id. Returns 400 when the id is malformed, 404 when the node kind cannot be found,
// and 200 with the node kind details otherwise.
func (s Handlers) GetNodeKindByID(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
		raw = mux.Vars(request)[URIPathVariableNodeKindID]
	)

	if id, err := strconv.ParseInt(raw, 10, 32); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "node kind id is malformed", response)
	} else if nodeKind, err := s.extensions.GetNodeKind(ctx, int32(id)); errors.Is(err, services.ErrNodeKindNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, "node kind not found", response)
	} else if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
	} else if nodeKindView, markdownErr := BuildNodeKindView(nodeKind); markdownErr != nil {
		slog.WarnContext(ctx, "failed to parse node kind info markdown content", attr.Error(markdownErr))
		responses.WriteBasic(ctx, nodeKindView, http.StatusOK, response)
	} else {
		responses.WriteBasic(ctx, nodeKindView, http.StatusOK, response)
	}
}
