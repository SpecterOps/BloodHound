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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/utils/validation"
)

// Checks that any cypher selectors are valid cypher and not too complex.
func areCypherSelectorsValidCypher(graph queries.Graph, seeds []model.SelectorSeed) (bool, error) {
	for _, seed := range seeds {
		if seed.Type == model.SelectorTypeCypher {
			if _, err := graph.PrepareCypherQuery(seed.Value, queries.MaxSelectorQueryComplexityWeight); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (s *Resources) CreateAssetGroupLabelSelector(response http.ResponseWriter, request *http.Request) {
	var (
		err error
		sel model.AssetGroupLabelSelector
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Label Selector Create")()

	if err = json.NewDecoder(request.Body).Decode(&sel); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if errs := validation.Validate(sel); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if idStr, hasLabelID := mux.Vars(request)[api.URIPathVariableAssetGroupLabelID]; !hasLabelID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoAssetGroupLabelId, request), response)
	} else if id, err := strconv.Atoi(idStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrInvalidAssetGroupLabelId, request), response)
	} else if _, err := s.DB.GetAssetGroupLabel(request.Context(), id); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrInvalidAssetGroupLabelId, request), response)
	} else if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if validCypher, err := areCypherSelectorsValidCypher(s.GraphQuery, sel.Seeds); !validCypher {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("cypher is invalid: %v", err), request), response)
	} else if selector, err := s.DB.CreateAssetGroupLabelSelector(request.Context(), id, actor.ID.String(), sel.Name, sel.Description, false, true, sel.AutoCertify, sel.Seeds); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), selector, http.StatusCreated, response)
	}
}
