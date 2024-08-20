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
	"errors"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	ctx2 "github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

type ShareSavedQueriesResponse []model.SavedQueriesPermissions

type SavedQueryPermissionRequest struct {
	UserIDs []uuid.UUID `json:"user_ids"`
	Public  bool        `json:"public"`
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
	} else if dbSavedQueryScope, err := s.DB.GetScopeForSavedQuery(request.Context(), int64(savedQueryID), user.ID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if isSavedQueryShared, err := s.DB.IsSavedQueryShared(request.Context(), int64(savedQueryID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		isAdmin := user.Roles.Has(model.Role{Name: auth.RoleAdministrator})

		if isAdmin {
			// Query set to public
			if createRequest.Public {
				if savedQueryBelongsToUser {
					if dbSavedQueryScope[model.SavedQueryScopePublic] {
						response.WriteHeader(http.StatusNoContent)
					} else {
						if isSavedQueryShared {
							if savedPermission, err := s.DB.CreateSavedQueryPermissionToPublic(request.Context(), int64(savedQueryID)); err != nil {
								api.HandleDatabaseError(request, response, err)
							} else if savedQueryPermissions, err := s.DB.GetPermissionsForSavedQuery(request.Context(), int64(savedQueryID)); err != nil {
								api.HandleDatabaseError(request, response, err)
							} else {
								for _, permission := range savedQueryPermissions {
									sharedToUserID := permission.SharedToUserID

									if err := s.DB.DeleteSavedQueryPermissionsForUser(request.Context(), int64(savedQueryID), sharedToUserID.UUID); err != nil {
										api.HandleDatabaseError(request, response, err)
									}
								}
								api.WriteBasicResponse(request.Context(), ShareSavedQueriesResponse{savedPermission}, http.StatusCreated, response)
							}
						} else {
							if savedPermission, err := s.DB.CreateSavedQueryPermissionToPublic(request.Context(), int64(savedQueryID)); err != nil {
								api.HandleDatabaseError(request, response, err)
							} else {
								api.WriteBasicResponse(request.Context(), ShareSavedQueriesResponse{savedPermission}, http.StatusCreated, response)
							}
						}
					}
				} else {
					if dbSavedQueryScope[model.SavedQueryScopePublic] {
						response.WriteHeader(http.StatusNoContent)
					} else {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
						return
					}
				}
				// Query set to private
			} else if len(createRequest.UserIDs) == 0 {
				if savedQueryBelongsToUser {
					if dbSavedQueryScope[model.SavedQueryScopePublic] {
						if err := s.DB.DeleteSavedQueryPermissionPublic(request.Context(), int64(savedQueryID)); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							response.WriteHeader(http.StatusNoContent)
						}
					} else {
						if isSavedQueryShared {
							if savedQueryPermissions, err := s.DB.GetPermissionsForSavedQuery(request.Context(), int64(savedQueryID)); err != nil {
								api.HandleDatabaseError(request, response, err)
							} else {
								for _, permission := range savedQueryPermissions {
									sharedToUserID := permission.SharedToUserID

									if err := s.DB.DeleteSavedQueryPermissionsForUser(request.Context(), int64(savedQueryID), sharedToUserID.UUID); err != nil {
										api.HandleDatabaseError(request, response, err)
									}
								}
								response.WriteHeader(http.StatusNoContent)
							}
						} else {
							response.WriteHeader(http.StatusNoContent)
						}
					}
				} else {
					if dbSavedQueryScope[model.SavedQueryScopePublic] {
						if err := s.DB.DeleteSavedQueryPermissionPublic(request.Context(), int64(savedQueryID)); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							response.WriteHeader(http.StatusNoContent)
						}
					} else {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
						return
					}
				}
				// Sharing a query
			} else if len(createRequest.UserIDs) > 0 && !createRequest.Public {
				if savedQueryBelongsToUser {
					if dbSavedQueryScope[model.SavedQueryScopePublic] {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Public query cannot be shared to users. You must set your query to private first", request), response)
					} else {
						var newPermissions []model.SavedQueriesPermissions
						for _, sharedUserID := range createRequest.UserIDs {
							if sharedUserID != user.ID {
								newPermissions = append(newPermissions, model.SavedQueriesPermissions{
									QueryID:        int64(savedQueryID),
									Public:         false,
									SharedToUserID: database.NullUUID(sharedUserID),
								})
							} else {
								api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Cannot share query to self", request), response)
								return
							}
						}
						// Save the permissions to the database
						if savedPermissions, err := s.DB.CreateSavedQueryPermissionsBatch(request.Context(), newPermissions); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							api.WriteBasicResponse(request.Context(), savedPermissions, http.StatusCreated, response)
						}
					}
				} else {
					if dbSavedQueryScope[model.SavedQueryScopePublic] {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Public query cannot be shared to users. You must set your query to private first", request), response)
					} else {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
						return
					}
				}
			}
		} else if !isAdmin {
			if !savedQueryBelongsToUser {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
				return
				// Query set to public
			} else if createRequest.Public {
				if dbSavedQueryScope[model.SavedQueryScopePublic] {
					response.WriteHeader(http.StatusNoContent)
				} else {
					if isSavedQueryShared {
						if savedPermission, err := s.DB.CreateSavedQueryPermissionToPublic(request.Context(), int64(savedQueryID)); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else if savedQueryPermissions, err := s.DB.GetPermissionsForSavedQuery(request.Context(), int64(savedQueryID)); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							for _, permission := range savedQueryPermissions {
								sharedToUserID := permission.SharedToUserID

								if err := s.DB.DeleteSavedQueryPermissionsForUser(request.Context(), int64(savedQueryID), sharedToUserID.UUID); err != nil {
									api.HandleDatabaseError(request, response, err)
								}
							}
							api.WriteBasicResponse(request.Context(), ShareSavedQueriesResponse{savedPermission}, http.StatusCreated, response)
						}
					} else {
						if savedPermission, err := s.DB.CreateSavedQueryPermissionToPublic(request.Context(), int64(savedQueryID)); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							api.WriteBasicResponse(request.Context(), ShareSavedQueriesResponse{savedPermission}, http.StatusCreated, response)
						}
					}
				}
				// Query set to private
			} else if len(createRequest.UserIDs) == 0 {
				if dbSavedQueryScope[model.SavedQueryScopePublic] {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
					return
				} else {
					if isSavedQueryShared {
						if savedQueryPermissions, err := s.DB.GetPermissionsForSavedQuery(request.Context(), int64(savedQueryID)); err != nil {
							api.HandleDatabaseError(request, response, err)
						} else {
							for _, permission := range savedQueryPermissions {
								sharedToUserID := permission.SharedToUserID

								if err := s.DB.DeleteSavedQueryPermissionsForUser(request.Context(), int64(savedQueryID), sharedToUserID.UUID); err != nil {
									api.HandleDatabaseError(request, response, err)
								}
							}
							response.WriteHeader(http.StatusNoContent)
						}
					} else {
						response.WriteHeader(http.StatusNoContent)
					}
				}
				// Sharing a query
			} else if len(createRequest.UserIDs) > 0 && !createRequest.Public {
				if dbSavedQueryScope[model.SavedQueryScopePublic] {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
					return
				} else {
					var newPermissions []model.SavedQueriesPermissions
					for _, sharedUserID := range createRequest.UserIDs {
						if sharedUserID != user.ID {
							newPermissions = append(newPermissions, model.SavedQueriesPermissions{
								QueryID:        int64(savedQueryID),
								Public:         false,
								SharedToUserID: database.NullUUID(sharedUserID),
							})
						} else {
							api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Cannot share query to self", request), response)
							return
						}
					}
					// Save the permissions to the database
					if savedPermissions, err := s.DB.CreateSavedQueryPermissionsBatch(request.Context(), newPermissions); err != nil {
						api.HandleDatabaseError(request, response, err)
					} else {
						api.WriteBasicResponse(request.Context(), savedPermissions, http.StatusCreated, response)
					}
				}
			}
		}
	}
}
