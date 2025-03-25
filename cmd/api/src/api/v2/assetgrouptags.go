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
	"slices"
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

const (
	ErrInvalidAssetGroupTagId = "invalid asset group tag id specified in url"
)

func (s Resources) GetAssetGroupTags(response http.ResponseWriter, request *http.Request) {
	const (
		pnameTagType       = "type"
		pnameIncludeCounts = "includeCounts"
	)
	var (
		pvalsTagType       = []string{"tag", "tier"}
		pvalsIncludeCounts = []string{"false", "true"}
	)

	var params = request.URL.Query()
	// set defaults. This works since .Get() only retrives the first value
	params.Add(pnameTagType, pvalsTagType[0])
	params.Add(pnameIncludeCounts, pvalsIncludeCounts[0])

	if paramTagType := params.Get(pnameTagType); !slices.Contains(pvalsTagType, paramTagType) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specifed for tag type", request), response)
	} else if paramIncludeCounts := params.Get(pnameIncludeCounts); !slices.Contains(pvalsIncludeCounts, paramIncludeCounts) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid value specifed for include counts", request), response)
	} else if tags, err := s.DB.GetAssetGroupTags(request.Context(), paramTagType); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		resp := map[string]any{"asset_group_tags": tags}
		if paramIncludeCounts == "true" {
		}
		api.WriteBasicResponse(request.Context(), resp, http.StatusOK, response)
	}
}

// Checks that the selector seeds are valid.
func validateSelectorSeeds(graph queries.Graph, seeds []model.SelectorSeed) error {
	// all seeds must be of the same type
	seedType := seeds[0].Type

	if seedType != model.SelectorTypeObjectId && seedType != model.SelectorTypeCypher {
		return fmt.Errorf("invalid seed type %v", seedType)
	}

	for _, seed := range seeds {
		if seed.Type != seedType {
			return fmt.Errorf("all seeds must be of the same type")
		}
		if seed.Type == model.SelectorTypeCypher {
			if _, err := graph.PrepareCypherQuery(seed.Value, queries.QueryComplexityLimitSelector); err != nil {
				return fmt.Errorf("cypher is invalid: %v", err)
			}
		}
	}
	return nil
}

func (s *Resources) CreateAssetGroupTagSelector(response http.ResponseWriter, request *http.Request) {
	var (
		sel           model.AssetGroupTagSelector
		assetTagIdStr = mux.Vars(request)[api.URIPathVariableAssetGroupTagID]
	)
	defer measure.ContextMeasure(request.Context(), slog.LevelDebug, "Asset Group Tag Selector Create")()

	if assetTagId, err := strconv.Atoi(assetTagIdStr); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, ErrInvalidAssetGroupTagId, request), response)
	} else if _, err := s.DB.GetAssetGroupTag(request.Context(), assetTagId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := json.NewDecoder(request.Body).Decode(&sel); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if errs := validation.Validate(sel); len(errs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if actor, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		slog.Error("Unable to get user from auth context")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "unknown user", request), response)
	} else if err := validateSelectorSeeds(s.GraphQuery, sel.Seeds); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if selector, err := s.DB.CreateAssetGroupTagSelector(request.Context(), assetTagId, actor.ID.String(), sel.Name, sel.Description, false, true, sel.AutoCertify, sel.Seeds); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), selector, http.StatusCreated, response)
	}
}
