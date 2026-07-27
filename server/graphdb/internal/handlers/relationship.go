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

// URIPathVariableRelationshipID is the name of the path variable carrying the
// graph-assigned relationship id.
const URIPathVariableRelationshipID = "relationship_id"

// RelationshipKindView is the JSON shape of the kind embedded in a RelationshipView.
// RelationshipKindID is null when the kind has no schema_relationship_kinds entry.
type RelationshipKindView struct {
	RelationshipKindID *int32 `json:"relationship_kind_id"`
	Name               string `json:"name"`
}

// RelationshipView is the JSON shape returned by the relationship handlers. It is
// decoupled from services.Relationship so the wire format can evolve independently of
// the domain model.
type RelationshipView struct {
	RelationshipID int64                      `json:"relationship_id"`
	SourceNodeID   int64                      `json:"source_node_id"`
	TargetNodeID   int64                      `json:"target_node_id"`
	Kind           RelationshipKindView       `json:"kind"`
	Properties     map[string]any             `json:"properties"`
	Info           []RelationshipKindInfoView `json:"info,omitempty"`
}

// RelationshipKindInfoView is the JSON shape for each kind-info object associated with a relationship.
type RelationshipKindInfoView struct {
	Name               string       `json:"name"`
	Title              string       `json:"title"`
	Position           int32        `json:"position"`
	RelationshipKindID int          `json:"relationship_kind_id"`
	Markdown           MarkdownView `json:"markdown"`
}

// BuildRelationshipView projects a services.Relationship into the view type the handlers
// return in their JSON envelope.
func BuildRelationshipView(relationship services.Relationship, includeKindInfo bool) RelationshipView {
	relView := RelationshipView{
		RelationshipID: relationship.ID,
		SourceNodeID:   relationship.SourceNodeID,
		TargetNodeID:   relationship.TargetNodeID,
		Kind: RelationshipKindView{
			RelationshipKindID: relationship.Kind.ID,
			Name:               relationship.Kind.Name,
		},
		Properties: relationship.Properties,
	}

	if includeKindInfo {
		for _, kindInfo := range relationship.KindInfos {
			relView.Info = append(relView.Info, RelationshipKindInfoView{
				Name:               kindInfo.InfoKey,
				Title:              kindInfo.Title,
				Position:           kindInfo.Position,
				RelationshipKindID: int(*relationship.Kind.ID),
				Markdown:           buildMarkdownView(kindInfo.Content),
			})
		}
	}

	return relView
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s RelationshipView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// GetRelationshipByID returns the details of a graph relationship identified by its
// graph-assigned integer id. Returns 400 when the id is malformed, 404 when the
// relationship cannot be found, and 200 with the relationship details otherwise.
// Kinds that are not registered in schema_relationship_kinds are returned with ID=nil.
func (s Handlers) GetRelationshipByID(response http.ResponseWriter, request *http.Request) {
	var (
		err                error
		ctx                = request.Context()
		rawRelationshipID  = mux.Vars(request)[URIPathVariableRelationshipID]
		includeKindInfoRaw = request.URL.Query().Get("include-info")
		includeKindInfo    = false
	)

	relationshipID, err := strconv.ParseInt(rawRelationshipID, 10, 64)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "relationship id is malformed", response)
		return
	}

	if includeKindInfoRaw != "" {
		includeKindInfo, err = strconv.ParseBool(includeKindInfoRaw)
		if err != nil {
			responses.WriteError(ctx, http.StatusBadRequest, "include-info is malformed", response)
			return
		}
	}

	relationship, err := s.graphDB.GetRelationship(ctx, relationshipID, includeKindInfo)
	if errors.Is(err, services.ErrRelationshipNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, "relationship not found", response)
		return
	}
	if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
		return
	}

	responses.WriteBasic(ctx, BuildRelationshipView(relationship, includeKindInfo), http.StatusOK, response)
}
