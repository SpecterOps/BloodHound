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
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/auth/bhsaml"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/serde"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/src/utils/validation"
)

const (
	ErrorResponseDetailsNumRoles               = "a user can only have one role"
	ErrorResponseDetailsInvalidCurrentPassword = "unable to verify current password"
	ErrorResponseDetailsMFAActivated           = "multi-factor authentication already active"
	ErrorResponseDetailsMFAEnrollmentRequired  = "multi-factor authentication enrollment is required before activation"
)

type ManagementResource struct {
	config                     config.Configuration
	secretDigester             crypto.SecretDigester
	db                         database.Database
	QueryParameterFilterParser model.QueryParameterFilterParser
	authorizer                 auth.Authorizer
}

func NewManagementResource(authConfig config.Configuration, db database.Database, authorizer auth.Authorizer) ManagementResource {
	return ManagementResource{
		config:                     authConfig,
		secretDigester:             authConfig.Crypto.Argon2.NewDigester(),
		db:                         db,
		QueryParameterFilterParser: model.NewQueryParameterFilterParser(),
		authorizer:                 authorizer,
	}
}

func (s ManagementResource) ListSAMLSignOnEndpoints(response http.ResponseWriter, request *http.Request) {
	if samlProviders, err := bhsaml.GetAllSAMLProviders(s.db, request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		var (
			samlSignOnEndpoints = make([]v2.SAMLSignOnEndpoint, len(samlProviders))
			requestContext      = ctx.Get(request.Context())
		)

		for idx, samlProvider := range samlProviders {
			providerURLs := bhsaml.FormatServiceProviderURLs(*requestContext.Host, samlProviders[idx].Name)

			samlSignOnEndpoints[idx].Name = samlProvider.Name
			samlSignOnEndpoints[idx].InitiationURL = serde.FromURL(providerURLs.SingleSignOnService)
		}

		api.WriteBasicResponse(request.Context(), v2.ListSAMLSignOnEndpointsResponse{
			Endpoints: samlSignOnEndpoints,
		}, http.StatusOK, response)
	}
}

func (s ManagementResource) ListSAMLProviders(response http.ResponseWriter, request *http.Request) {
	if samlProviders, err := bhsaml.GetAllSAMLProviders(s.db, request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), v2.ListSAMLProvidersResponse{SAMLProviders: samlProviders}, http.StatusOK, response)
	}
}

func (s ManagementResource) GetSAMLProvider(response http.ResponseWriter, request *http.Request) {
	pathVars := mux.Vars(request)

	if rawProviderID, hasID := pathVars[api.URIPathVariableSAMLProviderID]; !hasID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, api.ErrorResponseDetailsAuthenticationInvalid, request), response)
	} else if providerID, err := strconv.ParseInt(rawProviderID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if provider, err := s.db.GetSAMLProvider(int32(providerID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), provider, http.StatusOK, response)
	}
}

func (s ManagementResource) CreateSAMLProviderMultipart(response http.ResponseWriter, request *http.Request) {
	var samlIdentityProvider model.SAMLProvider

	if err := request.ParseMultipartForm(api.DefaultAPIPayloadReadLimitBytes); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if providerNames, hasProviderName := request.MultipartForm.Value["name"]; !hasProviderName {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"name\" parameter", request), response)
	} else if numProviderNames := len(providerNames); numProviderNames == 0 || numProviderNames > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"name\" parameter", request), response)
	} else if metadataXMLFileHandles, hasMetadataXML := request.MultipartForm.File["metadata"]; !hasMetadataXML {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "form is missing \"metadata\" parameter", request), response)
	} else if numHeaders := len(metadataXMLFileHandles); numHeaders == 0 || numHeaders > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "expected only one \"metadata\" parameter", request), response)
	} else if metadataXMLReader, err := metadataXMLFileHandles[0].Open(); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else {
		defer metadataXMLReader.Close()

		if metadataXML, err := io.ReadAll(metadataXMLReader); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else if metadata, err := samlsp.ParseMetadata(metadataXML); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else if ssoDescriptor, err := bhsaml.GetIDPSingleSignOnDescriptor(metadata, saml.HTTPPostBinding); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		} else if ssoURL, err := bhsaml.GetIDPSingleSignOnServiceURL(ssoDescriptor, saml.HTTPPostBinding); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "metadata does not have a SSO service that supports HTTP POST binding", request), response)
		} else {
			samlIdentityProvider.Name = providerNames[0]
			samlIdentityProvider.DisplayName = providerNames[0]
			samlIdentityProvider.MetadataXML = metadataXML
			samlIdentityProvider.IssuerURI = metadata.EntityID
			samlIdentityProvider.SingleSignOnURI = ssoURL

			if newSAMLProvider, err := s.db.CreateSAMLIdentityProvider(request.Context(), samlIdentityProvider); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				api.WriteBasicResponse(request.Context(), newSAMLProvider, http.StatusOK, response)
			}
		}
	}
}

func (s ManagementResource) disassociateUsersFromSAMLProvider(request *http.Request, providerUsers model.Users) error {
	for _, user := range providerUsers {
		user.SAMLProvider = nil
		user.SAMLProviderID = null.NewInt32(0, false)

		if err := s.db.UpdateUser(request.Context(), user); err != nil {
			return api.FormatDatabaseError(err)
		}
	}

	return nil
}

func (s ManagementResource) DeleteSAMLProvider(response http.ResponseWriter, request *http.Request) {
	var (
		identityProvider model.SAMLProvider
		rawProviderID    = mux.Vars(request)[api.URIPathVariableSAMLProviderID]
		requestContext   = ctx.FromRequest(request)
	)

	if providerID, err := strconv.ParseInt(rawProviderID, 10, 32); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if identityProvider, err = s.db.GetSAMLProvider(int32(providerID)); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if user, isUser := auth.GetUserFromAuthCtx(requestContext.AuthCtx); isUser && int64(user.SAMLProviderID.Int32) == providerID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusConflict, "user may not delete their own SAML auth provider", request), response)
	} else if providerUsers, err := s.db.GetSAMLProviderUsers(identityProvider.ID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := s.disassociateUsersFromSAMLProvider(request, providerUsers); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := s.db.DeleteSAMLProvider(request.Context(), identityProvider); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteBasicResponse(request.Context(), v2.DeleteSAMLProviderResponse{
			AffectedUsers: providerUsers,
		}, http.StatusOK, response)
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
		} else if permissions, err = s.db.GetAllPermissions(strings.Join(order, ", "), sqlFilter); err != nil {
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
	} else if permission, err := s.db.GetPermission(permissionID); err != nil {
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
		} else if roles, err = s.db.GetAllRoles(strings.Join(order, ", "), sqlFilter); err != nil {
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
	} else if role, err := s.db.GetRole(int32(roleID)); err != nil {
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
		} else if users, err = s.db.GetAllUsers(strings.Join(order, ", "), sqlFilter); err != nil {
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorResponseDetailsNumRoles, request), response)
	} else if roles, err := s.db.GetRoles(createUserRequest.Roles); err != nil {
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
				msg := strings.Join(utils.Errors(errs).AsStringSlice(), ", ")
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, msg, request), response)
				return
			} else if secretDigest, err := s.secretDigester.Digest(createUserRequest.Secret); err != nil {
				log.Errorf("Error while attempting to digest secret for user: %v", err)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
				return
			} else if passwordExpiration, err := appcfg.GetPasswordExpiration(s.db); err != nil {
				log.Errorf("Error while attempting to fetch password expiration window: %v", err)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			} else {
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
			} else if samlProvider, err := s.db.GetSAMLProvider(samlProviderID); err != nil {
				log.Errorf("Error while attempting to fetch SAML provider %d: %v", createUserRequest.SAMLProviderID, err)
				api.HandleDatabaseError(request, response, err)
			} else {
				userTemplate.SAMLProviderID = null.Int32From(samlProvider.ID)
			}
		}

		if newUser, err := s.db.CreateUser(request.Context(), userTemplate); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteBasicResponse(request.Context(), newUser, http.StatusOK, response)
		}

	}
}

func (s ManagementResource) updateUser(response http.ResponseWriter, request *http.Request, user model.User) {
	if err := s.db.UpdateUser(request.Context(), user); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (s ManagementResource) ensureUserHasNoAuthSecret(ctx context.Context, user model.User) error {
	if user.AuthSecret != nil {
		if err := s.db.DeleteAuthSecret(ctx, *user.AuthSecret); err != nil {
			return api.FormatDatabaseError(err)
		} else {
			return nil
		}
	}

	return nil
}

func (s ManagementResource) UpdateUser(response http.ResponseWriter, request *http.Request) {
	var (
		updateUserRequest v2.UpdateUserRequest
		pathVars          = mux.Vars(request)
		rawUserID         = pathVars[api.URIPathVariableUserID]
		context           = *ctx.FromRequest(request)
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err := s.db.GetUser(userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if err := api.ReadJSONRequestPayloadLimited(&updateUserRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if len(updateUserRequest.Roles) > 1 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "a user can only have one role", request), response)
	} else if roles, err := s.db.GetRoles(updateUserRequest.Roles); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		user.Roles = roles
		user.FirstName = null.StringFrom(updateUserRequest.FirstName)
		user.LastName = null.StringFrom(updateUserRequest.LastName)
		user.EmailAddress = null.StringFrom(updateUserRequest.EmailAddress)
		user.PrincipalName = updateUserRequest.Principal
		user.IsDisabled = updateUserRequest.IsDisabled

		if user.IsDisabled {
			if loggedInUser, _ := auth.GetUserFromAuthCtx(context.AuthCtx); user.ID == loggedInUser.ID {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseUserSelfDisable, request), response)
				return
			} else if userSessions, err := s.db.LookupActiveSessionsByUser(user); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				for _, session := range userSessions {
					s.db.EndUserSession(session)
				}
			}
		}

		if updateUserRequest.SAMLProviderID != "" {
			// We're setting a SAML provider. If the user has an associated secret the secret will be removed.
			if samlProviderID, err := serde.ParseInt32(updateUserRequest.SAMLProviderID); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("SAML Provider ID must be a number: %v", err.Error()), request), response)
			} else if err := s.ensureUserHasNoAuthSecret(request.Context(), user); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else if provider, err := s.db.GetSAMLProvider(samlProviderID); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else {
				// Ensure that the AuthSecret reference is nil and that the SAML provider is set
				user.AuthSecret = nil
				user.SAMLProvider = &provider
				user.SAMLProviderID = null.Int32From(samlProviderID)

				s.updateUser(response, request, user)
			}
		} else {
			// Default SAMLProviderID to null if the update request contains no SAMLProviderID
			user.SAMLProviderID = null.NewInt32(0, false)
			user.SAMLProvider = nil

			s.updateUser(response, request, user)
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
	} else if user, err := s.db.GetUser(userID); err != nil {
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
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err = s.db.GetUser(userID); err != nil {
		api.HandleDatabaseError(request, response, err)
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
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&setUserSecretRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if errs := validation.Validate(setUserSecretRequest); errs != nil {
		msg := strings.Join(utils.Errors(errs).AsStringSlice(), ", ")
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, msg, request), response)
	} else if targetUser, err := s.db.GetUser(userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if targetUser.SAMLProviderID.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid operation, user is SSO", request), response)
	} else if passwordExpiration, err := appcfg.GetPasswordExpiration(s.db); err != nil {
		log.Errorf("Error while attempting to fetch password expiration window: %v", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseCodeInternalServerError, request), response)
	} else if secretDigest, err := s.secretDigester.Digest(setUserSecretRequest.Secret); err != nil {
		log.Errorf("Error while attempting to digest secret for user: %v", err)
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

func (s ManagementResource) ExpireUserAuthSecret(response http.ResponseWriter, request *http.Request) {
	var (
		rawUserID = mux.Vars(request)[api.URIPathVariableUserID]
	)

	if userID, err := uuid.FromString(rawUserID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if targetUser, err := s.db.GetUser(userID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if targetUser.SAMLProviderID.Valid {
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
		} else if authTokens, err := s.db.GetAllAuthTokens(strings.Join(order, ", "), sqlFilter); err != nil {
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

	if user, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&createUserTokenRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
	} else if user, err := s.db.GetUser(user.ID); err != nil {
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
		pathVars   = mux.Vars(request)
		rawTokenID = pathVars[api.URIPathVariableTokenID]
		bhCtx      = ctx.FromRequest(request)
	)

	if user, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if tokenID, err := uuid.FromString(rawTokenID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if token, err := s.db.GetAuthToken(tokenID); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if token.UserID.Valid && token.UserID.UUID != user.ID && !s.authorizer.AllowsPermission(bhCtx.AuthCtx, auth.Permissions().AuthManageUsers) {
		log.Errorf("Bad user ID: %s != %s", token.UserID.UUID.String(), user.ID.String())
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else if err := s.db.DeleteAuthToken(request.Context(), token); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, v2.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorContentTypeJson.Error(), request), response)
	} else if user, err := s.db.GetUser(userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if user.SAMLProviderID.Valid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Invalid operation, user is SSO", request), response)
	} else if user.AuthSecret.TOTPActivated {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorResponseDetailsMFAActivated, request), response)
	} else if err := api.ValidateSecret(s.secretDigester, payload.Secret, *user.AuthSecret); err != nil {
		// In this context an authenticated user revalidating their password for mfa enrollment should get a 400 bad request
		// b/c the bearer token is valid despite the secret in the request payload being invalid
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorResponseDetailsInvalidCurrentPassword, request), response)
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, v2.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorContentTypeJson.Error(), request), response)
	} else if user, err := s.db.GetUser(userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Default the password to check against to the user from the path param
		secretToValidate := *user.AuthSecret
		bhCtx := ctx.FromRequest(request)
		if authedUser, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); isUser {
			if authedUser.ID != userId {
				// If the operation is being performed on a different user than who is logged in then we need to ensure they have proper permission
				if s.authorizer.AllowsPermission(bhCtx.AuthCtx, auth.Permissions().AuthManageUsers) {
					// Compare passed password against the logged in user's password instead
					secretToValidate = *authedUser.AuthSecret
				} else {
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "must be an admin to disable MFA for another user", request), response)
					return
				}
			}
		}

		// Check the password
		if err := api.ValidateSecret(s.secretDigester, payload.Secret, secretToValidate); err != nil {
			// In this context an authenticated user revalidating their password for mfa enrollment should get a 400 bad request
			// b/c the bearer token is valid despite the secret in the request payload being invalid
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorResponseDetailsInvalidCurrentPassword, request), response)
			return
		}

		user.AuthSecret.TOTPSecret = ""
		user.AuthSecret.TOTPActivated = false

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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, v2.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if user, err := s.db.GetUser(userId); err != nil {
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
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, v2.ErrorParseParams, request), response)
	} else if userId, err := uuid.FromString(rawUserId); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorContentTypeJson.Error(), request), response)
	} else if user, err := s.db.GetUser(userId); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if user.AuthSecret.TOTPSecret == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrorResponseDetailsMFAEnrollmentRequired, request), response)
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
