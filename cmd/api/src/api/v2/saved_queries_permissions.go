/*
 * Copyright 2024 Specter Ops, Inc.
 *
 * Licensed under the Apache License, Version 2.0
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v2

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	ctx2 "github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"net/http"
	"strconv"
)

// DeleteSavedQueryPermissionsRequest represents the payload sent to the unshare endpoint
type DeleteSavedQueryPermissionsRequest struct {
	UserIds []uuid.UUID `json:"user_ids"`
	Self    bool        `json:"self"`
}

// DeleteSavedQueryPermissions allows an owner of a shared query, a user that has a saved query shared to them, or an admin, to remove sharing privileges.
// A user who owns a query may unshare a query from anyone they have shared to
// A user who had a query shared to them may unshare that query from themselves
// And admins may unshare queries that have been shared to other users
func (s Resources) DeleteSavedQueryPermissions(response http.ResponseWriter, request *http.Request) {
	var (
		rawSavedQueryID = mux.Vars(request)[api.URIPathVariableSavedQueryID]
		deleteRequest   DeleteSavedQueryPermissionsRequest
	)

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
		return
	} else if savedQueryID, err := strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsFromMalformed, request), response)
		return
	} else if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else {
		if deleteRequest.Self {
			if isShared, err := s.DB.IsSavedQuerySharedToUser(request.Context(), savedQueryID, user.ID); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else if !isShared {
				// The user cannot unshare a saved query if a saved query permission does not exist for them. This means a user cannot unshare a query they own or have not been shared
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "User cannot unshare a query from themselves that is not shared to them", request), response)
				return
			} else {
				// User is trying to unshare from themselves
				if err := s.DB.DeleteSavedQueryPermissionsForUser(request.Context(), savedQueryID, user.ID); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				}
			}
		} else {
			isAdmin := user.Roles.Has(model.Role{Name: auth.RoleAdministrator})
			if !isAdmin {
				// If a user is not admin, then they need to own the query in order to unshare it
				if savedQueryBelongsToUser, err := s.DB.SavedQueryBelongsToUser(request.Context(), user.ID, savedQueryID); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else if !savedQueryBelongsToUser {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "Query does not belong to the user", request), response)
					return
				}
			}

			// Unshare the queries
			if err := s.DB.DeleteSavedQueryPermissionsForUsers(request.Context(), savedQueryID, deleteRequest.UserIds); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		}

		response.WriteHeader(http.StatusNoContent)
	}
}
