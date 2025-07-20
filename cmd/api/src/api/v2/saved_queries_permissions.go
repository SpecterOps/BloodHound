// Copyright 2024 Specter Ops, Inc.
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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	ctx2 "github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type ShareSavedQueriesResponse []model.SavedQueriesPermissions

type SavedQueryPermissionRequest struct {
	UserIDs []uuid.UUID `json:"user_ids"`
	Public  bool        `json:"public"`
}

var (
	ErrInvalidSelfShare   = errors.New("invalidSelfShare")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidPublicShare = errors.New("invalidPublicShare")
)

func CanUpdateSavedQueriesPermission(user model.User, savedQueryBelongsToUser bool, createRequest SavedQueryPermissionRequest, dbSavedQueryScope database.SavedQueryScopeMap) error {
	if user.Roles.Has(model.Role{Name: auth.RoleAdministrator}) {
		if createRequest.Public && savedQueryBelongsToUser {
			return nil
		} else if len(createRequest.UserIDs) == 0 && (savedQueryBelongsToUser || dbSavedQueryScope[model.SavedQueryScopePublic]) {
			return nil
		} else if len(createRequest.UserIDs) > 0 && !createRequest.Public {
			if dbSavedQueryScope[model.SavedQueryScopePublic] {
				return ErrInvalidPublicShare
			}
			if savedQueryBelongsToUser {
				for _, sharedUserID := range createRequest.UserIDs {
					if sharedUserID == user.ID {
						return ErrInvalidSelfShare
					}
				}
				return nil
			}
		}
	} else if savedQueryBelongsToUser && !dbSavedQueryScope[model.SavedQueryScopePublic] {
		if len(createRequest.UserIDs) > 0 && !createRequest.Public {
			for _, sharedUserID := range createRequest.UserIDs {
				if sharedUserID == user.ID {
					return ErrInvalidSelfShare
				}
			}
		}
		return nil
	}
	return ErrForbidden
}

type SavedQueryPermissionResponse struct {
	QueryID         int64       `json:"query_id"`
	Public          bool        `json:"public"`
	SharedToUserIDs []uuid.UUID `json:"shared_to_user_ids"`
}

func (s *SavedQueryPermissionResponse) AppendUserId(userId uuid.NullUUID) {
	if s.SharedToUserIDs == nil {
		s.SharedToUserIDs = make([]uuid.UUID, 0)
	}
	if userId.Valid {
		s.SharedToUserIDs = append(s.SharedToUserIDs, userId.UUID)
	}
}

// GetSavedQueryPermissions - users or admins can retrieve who queries
// Public queries will return for any user with no attached user ids.
func (s Resources) GetSavedQueryPermissions(response http.ResponseWriter, request *http.Request) {
	var rawSavedQueryID = mux.Vars(request)[api.URIPathVariableSavedQueryID]

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if savedQueryID, err := strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if savedQueryPermissions, err := s.DB.GetSavedQueryPermissions(request.Context(), savedQueryID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if savedQueryPermissions == nil || len(savedQueryPermissions) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "no query permissions exist for saved query", request), response)
	} else if isAccessibleToUser, err := s.canUserAccessSavedQueryPermissions(request.Context(), savedQueryPermissions[0], user); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if !isAccessibleToUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "no query permissions exist for saved query", request), response)
	} else {
		var savedQueryPermissionResponse = SavedQueryPermissionResponse{
			QueryID:         savedQueryID,
			Public:          savedQueryPermissions[0].Public,
			SharedToUserIDs: make([]uuid.UUID, 0),
		}
		if !savedQueryPermissionResponse.Public {
			for _, savedQueryPermission := range savedQueryPermissions {
				savedQueryPermissionResponse.AppendUserId(savedQueryPermission.SharedToUserID)
			}
		}
		api.WriteBasicResponse(request.Context(), savedQueryPermissionResponse, http.StatusOK, response)
	}
}

// canUserAccessSavedQueryPermissions - users can access query permissions if its public, they own the query or are an admin.
func (s Resources) canUserAccessSavedQueryPermissions(ctx context.Context, savedQueryPermissions model.SavedQueriesPermissions, user model.User) (bool, error) {
	if savedQueryPermissions.Public || user.Roles.Has(model.Role{Name: auth.RoleAdministrator}) {
		return true, nil
	}
	if savedQuery, err := s.DB.GetSavedQuery(ctx, savedQueryPermissions.QueryID); err != nil {
		return false, err
	} else {
		return user.ID.String() == savedQuery.UserID, nil
	}
}

// ShareSavedQueries allows a user to share queries between users, as well as share them publicly
func (s Resources) ShareSavedQueries(response http.ResponseWriter, request *http.Request) {
	var (
		rawSavedQueryID = mux.Vars(request)[api.URIPathVariableSavedQueryID]
		createRequest   SavedQueryPermissionRequest
	)

	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if savedQueryID, err := strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&createRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if createRequest.Public && len(createRequest.UserIDs) > 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Public cannot be true while user_ids is populated", request), response)
	} else if savedQueryBelongsToUser, err := s.DB.SavedQueryBelongsToUser(request.Context(), user.ID, savedQueryID); errors.Is(err, database.ErrNotFound) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "Query does not exist", request), response)
	} else if err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if dbSavedQueryScope, err := s.DB.GetScopeForSavedQuery(request.Context(), savedQueryID, user.ID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := CanUpdateSavedQueriesPermission(user, savedQueryBelongsToUser, createRequest, dbSavedQueryScope); err != nil {
		if errors.Is(err, ErrInvalidSelfShare) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Cannot share query to self", request), response)
		} else if errors.Is(err, ErrInvalidPublicShare) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Public query cannot be shared to users. You must set your query to private first", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
		}
	} else {
		// Query set to public
		if createRequest.Public {
			if dbSavedQueryScope[model.SavedQueryScopePublic] {
				response.WriteHeader(http.StatusNoContent)
			} else {
				if savedPermission, err := s.DB.CreateSavedQueryPermissionToPublic(request.Context(), savedQueryID); err != nil {
					api.HandleDatabaseError(request, response, err)
				} else {
					api.WriteBasicResponse(request.Context(), ShareSavedQueriesResponse{savedPermission}, http.StatusCreated, response)
				}
			}
			// Query set to private
		} else if len(createRequest.UserIDs) == 0 {
			if err := s.DB.DeleteSavedQueryPermissionsForUsers(request.Context(), savedQueryID); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				response.WriteHeader(http.StatusNoContent)
			}
			// Sharing a query
		} else if len(createRequest.UserIDs) > 0 && !createRequest.Public {
			if dbSavedQueryScope[model.SavedQueryScopePublic] {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Public query cannot be shared to users. You must set your query to private first", request), response)
			} else {
				if savedPermissions, err := s.DB.CreateSavedQueryPermissionsToUsers(request.Context(), savedQueryID, createRequest.UserIDs...); err != nil {
					api.HandleDatabaseError(request, response, err)
				} else {
					api.WriteBasicResponse(request.Context(), savedPermissions, http.StatusCreated, response)
				}
			}
		}
	}
}

// DeleteSavedQueryPermissionsRequest represents the payload sent to the unshare endpoint
type DeleteSavedQueryPermissionsRequest struct {
	UserIds []uuid.UUID `json:"user_ids"`
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
	} else if savedQueryID, err := strconv.ParseInt(rawSavedQueryID, 10, 64); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsFromMalformed, request), response)
	} else if err = json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		// Check if the user is attempting to unshare a query from themselves
		if slices.Contains(deleteRequest.UserIds, user.ID) {
			if isShared, err := s.DB.IsSavedQuerySharedToUser(request.Context(), savedQueryID, user.ID); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else if !isShared {
				// The user cannot unshare a saved query if a saved query permission does not exist for them. This means a user cannot unshare a query that they don't own, or hasn't been shared with them
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "User cannot unshare a query from themselves that is not shared to them", request), response)
				return
			}
		} else {
			// User is attempting to unshare a query from another user
			isAdmin := user.Roles.Has(model.Role{Name: auth.RoleAdministrator})
			if !isAdmin {
				// If a user is not admin, then they need to own the query in order to unshare it
				if savedQueryBelongsToUser, err := s.DB.SavedQueryBelongsToUser(request.Context(), user.ID, savedQueryID); err != nil {
					api.HandleDatabaseError(request, response, err)
					return
				} else if !savedQueryBelongsToUser {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "Query does not belong to the user", request), response)
					return
				}
			}

		}

		// Unshare the queries
		if err = s.DB.DeleteSavedQueryPermissionsForUsers(request.Context(), savedQueryID, deleteRequest.UserIds...); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	}
}
