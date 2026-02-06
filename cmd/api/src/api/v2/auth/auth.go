// Copyright 2023 Specter Ops, Inc.
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

package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/serde"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/services/oidc"
	"github.com/specterops/bloodhound/cmd/api/src/services/saml"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
	"github.com/specterops/bloodhound/packages/go/crypto"
)

const (
	ErrResponseDetailsNumRoles               = "a user can only have one role"
	ErrResponseDetailsInvalidCurrentPassword = "unable to verify current password"
	ErrResponseDetailsMFAActivated           = "multi-factor authentication already active"
	ErrResponseDetailsMFAEnrollmentRequired  = "multi-factor authentication enrollment is required before activation"
)

type ManagementResource struct {
	config                     config.Configuration
	secretDigester             crypto.SecretDigester
	db                         database.Database
	QueryParameterFilterParser model.QueryParameterFilterParser
	authorizer                 auth.Authorizer   // Used for Permissions
	authenticator              api.Authenticator // Used for secrets
	OIDC                       oidc.Service
	SAML                       saml.Service
	GraphQuery                 queries.Graph
	DogTags                    dogtags.Service
}

func NewManagementResource(authConfig config.Configuration, db database.Database, authorizer auth.Authorizer, authenticator api.Authenticator, graphQuery queries.Graph, dogTagsService dogtags.Service) ManagementResource {
	return ManagementResource{
		config:                     authConfig,
		secretDigester:             authConfig.Crypto.Argon2.NewDigester(),
		db:                         db,
		QueryParameterFilterParser: model.NewQueryParameterFilterParser(),
		authorizer:                 authorizer,
		authenticator:              authenticator,
		OIDC:                       &oidc.Client{},
		SAML:                       &saml.Client{},
		GraphQuery:                 graphQuery,
		DogTags:                    dogTagsService,
	}
}

func (s ManagementResource) ListPermissions(response http.ResponseWriter, request *http.Request) {
	var (
		order         []string
		permissions   model.Permissions
		sortByColumns = request.URL.Query()[api.QueryParameterSortBy]
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !permissions.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsNotSortable, column), request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(permissions.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := permissions.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = permissions.IsString(filter.Name)
				}
			}
		}

		// ignoring the error here as this would've failed at ParseQueryParameterFilters before getting here
		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if permissions, err = s.db.GetAllPermissions(request.Context(), strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		} else {
			api.WriteBasicResponse(request.Context(), v2.ListPermissionsResponse{Permissions: permissions}, http.StatusOK, response)
		}
	}
}

func (s ManagementResource) GetPermission(response http.ResponseWriter, request *http.Request) {
	var (
		pathVars        = mux.Vars(request)
		rawPermissionID = pathVars[api.URIPathVariablePermissionID]
	)

	if permissionID, err := strconv.Atoi(rawPermissionID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if permission, err := s.db.GetPermission(request.Context(), permissionID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), permission, http.StatusOK, response)
	}
}

func (s ManagementResource) ListRoles(response http.ResponseWriter, request *http.Request) {
	var (
		order         []string
		roles         model.Roles
		sortByColumns = request.URL.Query()[api.QueryParameterSortBy]
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !roles.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsNotSortable, column), request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(roles.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := roles.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = roles.IsString(filter.Name)
				}
			}
		}

		// ignoring the error here as this would've failed at ParseQueryParameterFilters before getting here
		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if roles, err = s.db.GetAllRoles(request.Context(), strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), v2.ListRolesResponse{Roles: roles}, http.StatusOK, response)
		}
	}
}

func (s ManagementResource) GetRole(response http.ResponseWriter, request *http.Request) {
	var (
		pathVars  = mux.Vars(request)
		rawRoleID = pathVars[api.URIPathVariableRoleID]
	)

	if roleID, err := strconv.ParseInt(rawRoleID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if role, err := s.db.GetRole(request.Context(), int32(roleID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), role, http.StatusOK, response)
	}
}

func (s ManagementResource) ListUsers(response http.ResponseWriter, request *http.Request) {
	var (
		order         []string
		users         model.Users
		sortByColumns = request.URL.Query()[api.QueryParameterSortBy]
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !users.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsNotSortable, column), request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(users.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := users.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = users.IsString(filter.Name)
				}
			}
		}

		// ignoring the error here as this would've failed at ParseQueryParameterFilters before getting here
		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if users, err = s.db.GetAllUsers(request.Context(), strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), v2.ListUsersResponse{Users: users}, http.StatusOK, response)
		}
	}
}

func (s ManagementResource) CreateUser(response http.ResponseWriter, request *http.Request) {
	var (
		createUserRequest v2.CreateUserRequest
		userTemplate      model.User
	)

	if err := api.ReadJSONRequestPayloadLimited(&createUserRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if len(createUserRequest.Roles) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrResponseDetailsNumRoles, request), response)
	} else if roles, err := s.db.GetRoles(request.Context(), createUserRequest.Roles); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		userTemplate.Roles = roles
		userTemplate.FirstName = null.StringFrom(createUserRequest.FirstName)
		userTemplate.LastName = null.StringFrom(createUserRequest.LastName)
		userTemplate.EmailAddress = null.StringFrom(createUserRequest.EmailAddress)
		userTemplate.PrincipalName = createUserRequest.Principal
		// EULA Acceptance does not pertain to Bloodhound Community Edition; this flag is used for Bloodhound Enterprise users.
		userTemplate.EULAAccepted = true

		if createUserRequest.Secret != "" {
			if errs := validation.Validate(createUserRequest.SetUserSecretRequest); errs != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
				return
			} else if secretDigest, err := s.secretDigester.Digest(createUserRequest.Secret); err != nil {
				slog.ErrorContext(request.Context(), fmt.Sprintf("Error while attempting to digest secret for user: %v", err))
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
				return
			} else {
				passwordExpiration := appcfg.GetPasswordExpiration(request.Context(), s.db)

				userTemplate.AuthSecret = &model.AuthSecret{
					Digest:       secretDigest.String(),
					DigestMethod: s.secretDigester.Method(),
					ExpiresAt:    time.Now().Add(passwordExpiration).UTC(),
				}

				if createUserRequest.NeedsPasswordReset {
					userTemplate.AuthSecret.ExpiresAt = time.Time{}
				}
			}
		}

		if createUserRequest.SAMLProviderID != "" {
			if samlProviderID, err := serde.ParseInt32(createUserRequest.SAMLProviderID); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("SAML Provider ID must be a number: %v", err.Error()), request), response)
				return
			} else if samlProvider, err := s.db.GetSAMLProvider(request.Context(), samlProviderID); err != nil {
				slog.ErrorContext(request.Context(), fmt.Sprintf("Error while attempting to fetch SAML provider %s: %v", createUserRequest.SAMLProviderID, err))
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				userTemplate.SSOProviderID = samlProvider.SSOProviderID
			}
		} else if createUserRequest.SSOProviderID.Valid {
			if _, err := s.db.GetSSOProviderById(request.Context(), createUserRequest.SSOProviderID.Int32); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				userTemplate.SSOProviderID = createUserRequest.SSOProviderID
			}
		}

		// ETAC DogTags
		// This is to handle an edge case where GORM defaults this value to false on user creation
		// Once ETAC is available to GA, this can be removed
		userTemplate.AllEnvironments = true
		if etacEnabled := s.DogTags.GetFlagAsBool(dogtags.ETAC_ENABLED); etacEnabled {
			// Access to all environments will be denied by default
			// The migration sets the default for all_environments to true, which will enable all users to have access to all environments until ETAC is explicitly enabled
			userTemplate.AllEnvironments = false

			if err := handleETACRequest(request.Context(), createUserRequest.UpdateUserRequest, roles, &userTemplate, s.GraphQuery); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			}
		}

		if newUser, err := s.db.CreateUser(request.Context(), userTemplate); err != nil {
			if errors.Is(err, database.ErrDuplicateUserPrincipal) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseUserDuplicatePrincipal, request), response)
			} else if errors.Is(err, database.ErrDuplicateEmail) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseUserDuplicateEmail, request), response)
			} else {
				api.HandleDatabaseError(request, response, err)
			}
		} else {
			api.WriteBasicResponse(request.Context(), newUser, http.StatusOK, response)
		}

	}
}

func (s ManagementResource) UpdateUser(response http.ResponseWriter, request *http.Request) {
	var (
		updateUserRequest v2.UpdateUserRequest
		pathVars          = mux.Vars(request)
		rawUserID         = pathVars[api.URIPathVariableUserID]
		authCtx           = *ctx.FromRequest(request)
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err := s.db.GetUser(request.Context(), userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := api.ReadJSONRequestPayloadLimited(&updateUserRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if len(updateUserRequest.Roles) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "a user can only have one role", request), response)
	} else if roles, err := s.db.GetRoles(request.Context(), updateUserRequest.Roles); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// PATCH requests may not contain every field, only conditionally update if fields exist
		if updateUserRequest.FirstName != "" {
			user.FirstName = null.StringFrom(updateUserRequest.FirstName)
		}

		if updateUserRequest.LastName != "" {
			user.LastName = null.StringFrom(updateUserRequest.LastName)
		}

		if updateUserRequest.EmailAddress != "" {
			user.EmailAddress = null.StringFrom(updateUserRequest.EmailAddress)
		}

		if updateUserRequest.Principal != "" {
			user.PrincipalName = updateUserRequest.Principal
		}

		if updateUserRequest.IsDisabled != nil {
			user.IsDisabled = *updateUserRequest.IsDisabled
		}

		loggedInUser, _ := auth.GetUserFromAuthCtx(authCtx.AuthCtx)

		if user.IsDisabled {
			if user.ID == loggedInUser.ID {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseUserSelfDisable, request), response)
				return
			} else if userSessions, err := s.db.LookupActiveSessionsByUser(request.Context(), user); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				for _, session := range userSessions {
					s.db.EndUserSession(request.Context(), session)
				}
			}
		}

		if updateUserRequest.SAMLProviderID != "" {
			// We're setting a SAML provider. If the user has an associated secret the secret will be removed.
			if samlProviderID, err := serde.ParseInt32(updateUserRequest.SAMLProviderID); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("SAML Provider ID must be a number: %v", err.Error()), request), response)
				return
			} else if provider, err := s.db.GetSAMLProvider(request.Context(), samlProviderID); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else if ssoProvider, err := s.db.GetSSOProviderById(request.Context(), provider.SSOProviderID.Int32); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				// Ensure that the AuthSecret reference is nil and the SSO provider is set
				user.AuthSecret = nil // Required or the below updateUser will re-add the authSecret
				user.SSOProvider = &ssoProvider
				user.SSOProviderID = provider.SSOProviderID
			}
		} else if updateUserRequest.SSOProviderID.Valid {
			if ssoProvider, err := s.db.GetSSOProviderById(request.Context(), updateUserRequest.SSOProviderID.Int32); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				user.AuthSecret = nil // Required or the below updateUser will re-add the authSecret
				user.SSOProvider = &ssoProvider
				user.SSOProviderID = updateUserRequest.SSOProviderID
			}
		} else {
			// Default SSOProviderID to null if the update request contains no SSOProviderID
			user.SSOProvider = nil
			user.SSOProviderID = null.NewInt32(0, false)
		}

		// Prevent a user from modifying their own roles/permissions
		if user.ID == loggedInUser.ID {
			if !slices.Equal(roles.IDs(), loggedInUser.Roles.IDs()) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseUserSelfRoleChange, request), response)
				return
			} else if !user.SSOProviderID.Equal(loggedInUser.SSOProviderID) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseUserSelfSSOProviderChange, request), response)
				return
			}
		}

		// We have to wait until after SSOProvider updates are handled above to validate roles can be safely updated.
		if user.SSOProviderHasRoleProvisionEnabled() && !slices.Equal(roles.IDs(), user.Roles.IDs()) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseUserSSOProviderRoleProvisionChange, request), response)
			return
		} else if updateUserRequest.Roles != nil {
			user.Roles = roles
		}

		// ETAC DogTags
		if etacEnabled := s.DogTags.GetFlagAsBool(dogtags.ETAC_ENABLED); etacEnabled {
			// Use the request's roles if it is being sent, otherwise use the user's current role to determine if an ETAC list may be applied
			effectiveRoles := user.Roles
			if updateUserRequest.Roles != nil {
				effectiveRoles = roles
			}

			if err := handleETACRequest(request.Context(), updateUserRequest, effectiveRoles, &user, s.GraphQuery); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
				return
			}
		}

		if err := s.db.UpdateUser(request.Context(), user); err != nil {
			if errors.Is(err, database.ErrDuplicateUserPrincipal) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseUserDuplicatePrincipal, request), response)
			} else if errors.Is(err, database.ErrDuplicateEmail) {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, api.ErrorResponseUserDuplicateEmail, request), response)
			} else {
				api.HandleDatabaseError(request, response, err)
			}
		} else {
			response.WriteHeader(http.StatusOK)
		}
	}
}

func (s ManagementResource) GetUser(response http.ResponseWriter, request *http.Request) {
	var (
		pathVars  = mux.Vars(request)
		rawUserID = pathVars[api.URIPathVariableUserID]
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err := s.db.GetUser(request.Context(), userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), user, http.StatusOK, response)
	}
}

func (s ManagementResource) GetSelf(response http.ResponseWriter, request *http.Request) {
	bhCtx := ctx.FromRequest(request)
	api.WriteBasicResponse(request.Context(), bhCtx.AuthCtx.Owner, http.StatusOK, response)
}

func (s ManagementResource) DeleteUser(response http.ResponseWriter, request *http.Request) {
	var (
		user      model.User
		pathVars  = mux.Vars(request)
		rawUserID = pathVars[api.URIPathVariableUserID]
		bhCtx     = ctx.FromRequest(request)
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err = s.db.GetUser(request.Context(), userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if currentUser, found := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !found {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found with request", request), response)
	} else if userID == currentUser.ID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "User cannot delete themselves", request), response)
	} else if err := s.db.DeleteUser(request.Context(), user); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (s ManagementResource) setUserSecret(ctx context.Context, user model.User, authSecret model.AuthSecret) error {
	if user.AuthSecret != nil {
		user.AuthSecret.Digest = authSecret.Digest
		user.AuthSecret.DigestMethod = authSecret.DigestMethod
		user.AuthSecret.ExpiresAt = authSecret.ExpiresAt.UTC()

		if err := s.db.UpdateAuthSecret(ctx, *user.AuthSecret); err != nil {
			return api.FormatDatabaseError(err)
		} else {
			return nil
		}
	} else {
		if _, err := s.db.CreateAuthSecret(ctx, authSecret); err != nil {
			return api.FormatDatabaseError(err)
		} else {
			return nil
		}
	}
}

func (s ManagementResource) PutUserAuthSecret(response http.ResponseWriter, request *http.Request) {
	var (
		authSecret           model.AuthSecret
		setUserSecretRequest v2.SetUserSecretRequest
		pathVars             = mux.Vars(request)
		rawUserID            = pathVars[api.URIPathVariableUserID]
		bhCtx                = ctx.FromRequest(request)
	)

	if loggedInUser, found := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !found {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found", request), response)
	} else if targetUserID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&setUserSecretRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if errs := validation.Validate(setUserSecretRequest); errs != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, errs.Error(), request), response)
	} else if targetUser, err := s.db.GetUser(request.Context(), targetUserID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if targetUser.SSOProviderID.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid operation, user is SSO", request), response)
	} else {
		if loggedInUser.ID == targetUserID {
			if targetUser.AuthSecret == nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrNoUserSecret.Error(), request), response)
			} else if err := s.authenticator.ValidateSecret(request.Context(), setUserSecretRequest.CurrentSecret, *targetUser.AuthSecret); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "Invalid current password", request), response)
				return
			}
		}

		passwordExpiration := appcfg.GetPasswordExpiration(request.Context(), s.db)
		if secretDigest, err := s.secretDigester.Digest(setUserSecretRequest.Secret); err != nil {
			slog.ErrorContext(request.Context(), fmt.Sprintf("Error while attempting to digest secret for user: %v", err))
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		} else {
			authSecret.UserID = targetUser.ID
			authSecret.Digest = secretDigest.String()
			authSecret.DigestMethod = s.secretDigester.Method()
			authSecret.ExpiresAt = time.Now().Add(passwordExpiration).UTC()

			if setUserSecretRequest.NeedsPasswordReset {
				authSecret.ExpiresAt = time.Time{}
			}

			if err := s.setUserSecret(request.Context(), targetUser, authSecret); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				response.WriteHeader(http.StatusOK)
			}
		}
	}
}

func (s ManagementResource) ExpireUserAuthSecret(response http.ResponseWriter, request *http.Request) {
	var (
		rawUserID = mux.Vars(request)[api.URIPathVariableUserID]
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if targetUser, err := s.db.GetUser(request.Context(), userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if targetUser.SSOProviderID.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "user has SAML auth enabled", request), response)
	} else {
		authSecret := targetUser.AuthSecret
		authSecret.ExpiresAt = time.Time{}

		if err := s.db.UpdateAuthSecret(request.Context(), *authSecret); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			// NOTE: This "should" be a 204 since we're not returning a payload but am returning a 200 to retain
			// uniformity.
			response.WriteHeader(http.StatusOK)
		}
	}
}

func (s ManagementResource) ListAuthTokens(response http.ResponseWriter, request *http.Request) {
	var (
		order         []string
		authTokens    = model.AuthTokens{}
		sortByColumns = request.URL.Query()[api.QueryParameterSortBy]
	)

	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !authTokens.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsNotSortable, column), request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(authTokens.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := authTokens.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = authTokens.IsString(filter.Name)
				}
			}
		}

		// Only show the user their tokens unless they have permission to manage other users
		bhCtx := ctx.FromRequest(request)
		if user, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); isUser {
			if !s.authorizer.AllowsPermission(bhCtx.AuthCtx, auth.Permissions().AuthManageUsers) {
				if queryFilters.IsFiltered("user_id") {
					if len(queryFilters["user_id"]) > 0 && queryFilters["user_id"][0].Value != user.ID.String() {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "only admins are able to filter tokens by user_id", request), response)
						return
					}
				} else {
					queryFilters.AddFilter(model.QueryParameterFilter{Name: "user_id", Operator: model.Equals, Value: user.ID.String()})
				}
			}
		}

		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if authTokens, err := s.db.GetAllAuthTokens(request.Context(), strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), v2.ListTokensResponse{Tokens: authTokens.StripKeys()}, http.StatusOK, response)
		}
	}
}

func (s ManagementResource) CreateAuthToken(response http.ResponseWriter, request *http.Request) {
	var (
		createUserTokenRequest = v2.CreateUserToken{}
		bhCtx                  = ctx.FromRequest(request)
	)

	if !appcfg.GetAPITokensParameter(request.Context(), s.db) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "API key creation is disabled", request), response)
		return
	} else if user, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&createUserTokenRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if user, err := s.db.GetUser(request.Context(), user.ID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := verifyUserID(&createUserTokenRequest, user, bhCtx, s.authorizer); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, err.Error(), request), response)
	} else if authToken, err := auth.NewUserAuthToken(createUserTokenRequest.UserID, createUserTokenRequest.TokenName, auth.HMAC_SHA2_256); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if newAuthToken, err := s.db.CreateAuthToken(request.Context(), authToken); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), newAuthToken, http.StatusOK, response)
	}
}

// This is a helper function that selects the correct user_id to use for the token being created.
// If no user_id is passed in the request, use the authed user's ID and proceed.
// If the request contains a user_id other than their own, check to make sure they have permissions to create tokens for other users and reject.
func verifyUserID(createUserTokenRequest *v2.CreateUserToken, user model.User, bhCtx *ctx.Context, authorizer auth.Authorizer) error {
	if createUserTokenRequest.UserID == "" {
		createUserTokenRequest.UserID = user.ID.String()
		return nil
	}
	if createUserTokenRequest.UserID != user.ID.String() && !authorizer.AllowsPermission(bhCtx.AuthCtx, auth.Permissions().AuthManageUsers) {
		return errors.New("missing permission to create tokens for other users")
	}

	return nil
}

func (s ManagementResource) DeleteAuthToken(response http.ResponseWriter, request *http.Request) {
	var (
		pathVars      = mux.Vars(request)
		rawTokenID    = pathVars[api.URIPathVariableTokenID]
		bhCtx         = ctx.FromRequest(request)
		auditLogEntry model.AuditEntry
	)

	if user, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if tokenID, err := uuid.FromString(rawTokenID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if token, err := s.db.GetAuthToken(request.Context(), tokenID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Log Intent to delete auth token for target user
		if auditLogEntry, err = model.NewAuditEntry(model.AuditLogActionDeleteAuthToken, model.AuditLogStatusIntent, model.AuditData{"target_user_id": token.UserID.UUID, "id": token.ID.String()}); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			return
		} else if err = s.db.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		} else if token.UserID.Valid && token.UserID.UUID != user.ID && !s.authorizer.AllowsPermission(bhCtx.AuthCtx, auth.Permissions().AuthManageUsers) {
			auditLogEntry.Status = model.AuditLogStatusFailure
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, api.ErrorResponseDetailsForbidden, request), response)
		} else if err := s.db.DeleteAuthToken(request.Context(), token); err != nil {
			auditLogEntry.Status = model.AuditLogStatusFailure
			api.HandleDatabaseError(request, response, err)
		} else {
			auditLogEntry.Status = model.AuditLogStatusSuccess
			response.WriteHeader(http.StatusOK)
		}

		// Audit Log Result to delete auth token for target user
		if err := s.db.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
			// We want to keep err scoped because response trumps this error
			if errors.Is(err, database.ErrNotFound) {
				slog.ErrorContext(request.Context(), fmt.Sprintf("resource not found: %v", err))
			} else if errors.Is(err, context.DeadlineExceeded) {
				slog.ErrorContext(request.Context(), fmt.Sprintf("context deadline exceeded: %v", err))
			} else {
				slog.ErrorContext(request.Context(), fmt.Sprintf("unexpected database error: %v", err))
			}
		}
	}
}

type MFAEnrollmentRequest struct {
	Secret string `json:"secret"`
}

type MFAEnrollmentReponse struct {
	QrCode     string `json:"qr_code"`
	TOTPSecret string `json:"totp_secret"`
}

type MFAActivationRequest struct {
	OTP string `json:"otp"`
}

type MFAActivationStatus string

const (
	MFAActivated   MFAActivationStatus = "activated"
	MFADeactivated MFAActivationStatus = "deactivated"
	MFAPending     MFAActivationStatus = "pending"
)

type MFAStatusResponse struct {
	Status MFAActivationStatus `json:"status"`
}

func (s ManagementResource) EnrollMFA(response http.ResponseWriter, request *http.Request) {
	rawUserId := mux.Vars(request)[api.URIPathVariableUserID]
	host := *ctx.Get(request.Context()).Host

	payload := MFAEnrollmentRequest{}

	if err := request.ParseForm(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrContentTypeJson.Error(), request), response)
	} else if user, err := s.db.GetUser(request.Context(), userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if user.SSOProviderID.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid operation, user is SSO", request), response)
	} else if user.AuthSecret.TOTPActivated {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrResponseDetailsMFAActivated, request), response)
	} else if err := api.ValidateSecret(s.secretDigester, payload.Secret, *user.AuthSecret); err != nil {
		// In this context an authenticated user revalidating their password for mfa enrollment should get a 400 bad request
		// b/c the bearer token is valid despite the secret in the request payload being invalid
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrResponseDetailsInvalidCurrentPassword, request), response)
	} else if totpSecret, err := auth.GenerateTOTPSecret(host.String(), user.PrincipalName); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		user.AuthSecret.TOTPSecret = totpSecret.Secret()

		if err := s.db.UpdateAuthSecret(request.Context(), *user.AuthSecret); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else if qrCode, err := auth.GenerateQRCodeBase64(*totpSecret); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		} else {
			responseBody := MFAEnrollmentReponse{
				qrCode,
				totpSecret.Secret(),
			}
			api.WriteBasicResponse(request.Context(), responseBody, http.StatusOK, response)
		}
	}
}

func (s ManagementResource) DisenrollMFA(response http.ResponseWriter, request *http.Request) {
	rawUserId := mux.Vars(request)[api.URIPathVariableUserID]

	payload := MFAEnrollmentRequest{}

	if err := request.ParseForm(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrContentTypeJson.Error(), request), response)
	} else if user, err := s.db.GetUser(request.Context(), userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		bhCtx := ctx.FromRequest(request)
		if authedUser, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); isUser {
			// Default the password to check against to the user from the path param
			secretToValidate := *user.AuthSecret

			if authedUser.ID != userId {
				// If the operation is being performed on a different user than who is logged in then we need to ensure they have proper permission
				if s.authorizer.AllowsPermission(bhCtx.AuthCtx, auth.Permissions().AuthManageUsers) {
					// Compare passed password against the logged in user's password instead
					if authedUser.AuthSecret != nil {
						secretToValidate = *authedUser.AuthSecret
					}
				} else {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "must be an admin to disable MFA for another user", request), response)
					return
				}
			}

			// Check the password only if the current authed user is not using SSO
			if !authedUser.SSOProviderID.Valid {
				if err := api.ValidateSecret(s.secretDigester, payload.Secret, secretToValidate); err != nil {
					// In this context an authenticated user revalidating their password for mfa enrollment should get a 400 bad request
					// b/c the bearer token is valid despite the secret in the request payload being invalid
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrResponseDetailsInvalidCurrentPassword, request), response)
					return
				}
			}
		}

		if user.AuthSecret != nil {
			user.AuthSecret.TOTPSecret = ""
			user.AuthSecret.TOTPActivated = false
		}

		if err := s.db.UpdateAuthSecret(request.Context(), *user.AuthSecret); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			responseBody := MFAStatusResponse{MFADeactivated}
			api.WriteBasicResponse(request.Context(), responseBody, http.StatusOK, response)
		}
	}
}

func (s ManagementResource) GetMFAActivationStatus(response http.ResponseWriter, request *http.Request) {
	rawUserId := mux.Vars(request)[api.URIPathVariableUserID]

	if err := request.ParseForm(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err := s.db.GetUser(request.Context(), userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		responseBody := MFAStatusResponse{}
		if user.AuthSecret.TOTPActivated && user.AuthSecret.TOTPSecret != "" {
			responseBody.Status = MFAActivated
		} else if user.AuthSecret.TOTPSecret != "" {
			responseBody.Status = MFAPending
		} else {
			responseBody.Status = MFADeactivated
		}
		api.WriteBasicResponse(request.Context(), responseBody, http.StatusOK, response)
	}
}

func (s ManagementResource) ActivateMFA(response http.ResponseWriter, request *http.Request) {
	rawUserId := mux.Vars(request)[api.URIPathVariableUserID]

	payload := MFAActivationRequest{}

	if err := request.ParseForm(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrContentTypeJson.Error(), request), response)
	} else if user, err := s.db.GetUser(request.Context(), userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if user.AuthSecret.TOTPSecret == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrResponseDetailsMFAEnrollmentRequired, request), response)
	} else if !totp.Validate(payload.OTP, user.AuthSecret.TOTPSecret) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsOTPInvalid, request), response)
	} else {
		user.AuthSecret.TOTPActivated = true

		if err := s.db.UpdateAuthSecret(request.Context(), *user.AuthSecret); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			responseBody := MFAStatusResponse{MFAActivated}
			api.WriteBasicResponse(request.Context(), responseBody, http.StatusOK, response)
		}
	}
}
