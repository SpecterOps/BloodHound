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
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

const URIPathVariableRelationshipKindID = "relationship_kind_id"

func (s Handlers) GetRelationshipKindByID(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
		raw = mux.Vars(request)[URIPathVariableRelationshipKindID]
	)

	if id, err := strconv.ParseInt(raw, 10, 32); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "relationship kind id is malformed", response)
		return
	} else if relationshipKind, err := s.extensions.GetRelationshipKind(ctx, int32(id)); errors.Is(err, services.ErrRelationshipKindNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, "relationship kind not found", response)
		return
	} else if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
		return
	} else if view, err := buildRelationshipKindView(relationshipKind); err != nil {
		slog.WarnContext(ctx, "Failed to parse relationship kind info markdown content", attr.Error(err))
		responses.WriteBasic(ctx, view, http.StatusOK, response)
		return
	} else {
		responses.WriteBasic(ctx, view, http.StatusOK, response)
	}
}

type RelationshipKindView struct {
	RelationshipKindID int32                   `json:"relationship_kind_id"`
	Name               string                  `json:"name"`
	Description        string                  `json:"description"`
	IsTraversable      bool                    `json:"is_traversable"`
	Info               map[string]KindInfoView `json:"info"`
	Extension          ExtensionView           `json:"extension"`
}

func (s RelationshipKindView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

func buildRelationshipKindView(relationshipKind services.RelationshipKind) (RelationshipKindView, error) {

	view := RelationshipKindView{
		RelationshipKindID: relationshipKind.ID,
		Name:               relationshipKind.Name,
		Description:        relationshipKind.Description,
		IsTraversable:      relationshipKind.IsTraversable,
		Extension: ExtensionView{
			ExtensionID: relationshipKind.Extension.ID,
			Name:        relationshipKind.Extension.Name,
			DisplayName: relationshipKind.Extension.DisplayName,
			Namespace:   relationshipKind.Extension.Namespace,
			Version:     relationshipKind.Extension.Version,
		},
		Info: map[string]KindInfoView{},
	}

	var markdownErr error

	for _, info := range relationshipKind.Info {
		markdown, err := buildMarkdownView(info.Content)
		if err != nil {
			markdownErr = errors.Join(markdownErr, err)
		}

		view.Info[info.InfoKey] = KindInfoView{
			Title:    info.Title,
			Position: info.Position,
			Markdown: markdown,
		}
	}

	return view, markdownErr
}
