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

package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/model"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/headers"
)

func parseAuthorizationHeader(request *http.Request) (string, string, *api.ErrorWrapper) {
	if authorizationHeader := request.Header.Get(headers.Authorization.String()); authorizationHeader == "" {
		return "", "", nil
	} else if authorizationValues := strings.Split(authorizationHeader, " "); len(authorizationValues) != 2 {
		return "", "", api.
			BuildErrorResponse(http.StatusBadRequest, "Expected only two components for the Authorization header.", request)
	} else {
		return strings.ToLower(authorizationValues[0]), authorizationValues[1], nil
	}
}

// AuthMiddleware is a middleware func generator that returns a http.Handler which closes around an instances of the
// v2.Authenticator struct.
//
// On request, the middleware attempts to parse the Authorization HTTP header if it exists. If the header does not
// exist then the middleware sets the auth.Context of the request context to "unauthenticated." If the header exists
// the scheme of the Authorization header is interpreted next to identify which authorization method to utilize.
//
// BloodHound Auth supports the following Authorization schemes:
//
//	`bearer`
//	   Bearer token scheme that contains the user's authenticated session JWT as its parameter.
//	`bhesignature`
//	   Request signing scheme that contains the BloodHound token ID as its parameter. See: `src/api/v2/signature.go`
func AuthMiddleware(authenticator api.Authenticator) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if authScheme, schemeParameter, err := parseAuthorizationHeader(request); err != nil {
				api.WriteErrorResponse(request.Context(), err, response)
				return
			} else {
				switch authScheme {
				case api.AuthorizationSchemeBearer:
					if authContext, err := authenticator.ValidateBearerToken(request.Context(), schemeParameter); err != nil {
						slog.ErrorContext(request.Context(), "Error while authenticating bearer token in AuthMiddleware", attr.Error(err))
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "Token Authorization failed.", request), response)
						return
					} else {
						bhCtx := ctx.Get(request.Context())
						bhCtx.AuthCtx = authContext
					}

				case api.AuthorizationSchemeBHESignature:
					if tokenID, err := uuid.FromString(schemeParameter); err != nil {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Token ID is malformed.", request), response)
						return
					} else if userAuth, responseCode, err := authenticator.ValidateRequestSignature(tokenID, request, time.Now()); errors.Is(err, api.ErrApiKeysDisabled) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(responseCode, err.Error(), request), response)
						return
					} else if err != nil {
						msg := fmt.Errorf("unable to validate request signature for client: %w", err).Error()
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(responseCode, msg, request), response)
						return
					} else {
						bhCtx := ctx.Get(request.Context())
						bhCtx.AuthCtx = userAuth
					}

				case "":
				default:
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Unexpected authorization scheme: %s.", authScheme), request), response)
				}
			}

			next.ServeHTTP(response, request)
		})
	}
}

// PermissionsCheckAll is a middleware func generator that returns a http.Handler which closes around a list of
// permissions that an actor must have in the request auth context to access the wrapped http.Handler.
func PermissionsCheckAll(authorizer auth.Authorizer, permissions ...model.Permission) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if bhCtx := ctx.FromRequest(request); !bhCtx.AuthCtx.Authenticated() {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "not authenticated", request), response)
			} else if !authorizer.AllowsAllPermissions(bhCtx.AuthCtx, permissions) {
				authorizer.AuditLogUnauthorizedAccess(request)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "not authorized", request), response)
			} else {
				next.ServeHTTP(response, request)
			}
		})
	}
}

// PermissionsCheckAtLeastOne is a middleware func generator that returns a http.Handler which closes around a list of
// permissions that an actor must have at least one in the request auth context to access the wrapped http.Handler.
func PermissionsCheckAtLeastOne(authorizer auth.Authorizer, permissions ...model.Permission) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if bhCtx := ctx.FromRequest(request); !bhCtx.AuthCtx.Authenticated() {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "not authenticated", request), response)
			} else if !authorizer.AllowsAtLeastOnePermission(bhCtx.AuthCtx, permissions) {
				authorizer.AuditLogUnauthorizedAccess(request)
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "not authorized", request), response)
			} else {
				next.ServeHTTP(response, request)
			}
		})
	}
}

// Helper function to pull the userID from the path variable.
func getUserId(request *http.Request) (string, bool) {
	if mux.Vars(request)[api.URIPathVariableUserID] != "" {
		return mux.Vars(request)[api.URIPathVariableUserID], true
	}

	return "", false
}

// RequireUserId is a middleware func generator that returns a http.Handler which checks to see if a user_id parameter has been included.
// There are a number of handlers that expect this parameter to be present, so this middleware can be applied to those to validate the required
// parameter has been included in the request.
func RequireUserId() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if _, hasUserId := getUserId(request); !hasUserId {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "user_id is required", request), response)
			} else {
				next.ServeHTTP(response, request)
			}
		})
	}
}

// AuthorizeAuthManagementAccess is a middleware func generator that returns a http.Handler which closes around a
// permission set to reference permission definitions and validate API access to user management related calls.
//
// An actor may operate on other actor entities if they have the permission "permission://auth/ManageUsers." If not the
// actor may only exercise auth management calls on auth entities that are explicitly owned by the actor.
func AuthorizeAuthManagementAccess(permissions auth.PermissionSet, authorizer auth.Authorizer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			bhCtx := ctx.FromRequest(request)

			if !bhCtx.AuthCtx.Authenticated() {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "not authorized", request), response)
			} else {
				authorized := false
				userID, hasUserId := getUserId(request)

				if user, isUser := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); isUser {
					// Use the auth session's user info if no parameter was passed.
					if !hasUserId {
						userID = user.ID.String()
					}

					if userID == user.ID.String() {
						// If we're operating on our own user context then check to make sure we have the self permission
						authorized = authorizer.AllowsPermission(bhCtx.AuthCtx, permissions.AuthManageSelf)
					} else {
						// If we're operating on a user account that is different from our own user context then check to make sure we
						// have the other permission
						authorized = authorizer.AllowsPermission(bhCtx.AuthCtx, permissions.AuthManageUsers)
					}
				}

				if !authorized {
					authorizer.AuditLogUnauthorizedAccess(request)
					api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, fmt.Sprintf("not authorized for %s", userID), request), response)
				} else {
					next.ServeHTTP(response, request)
				}
			}
		})
	}
}

const (
	loginMinimum   = time.Second + 500*time.Millisecond
	loginVariation = 500 * time.Millisecond
)

// LoginTimer is a middleware to protect against time-based user enumeration on the Login route. It does this by
// starting a timer before the actual login procedure to normalize the duration of this procedure to be within 1.5s and
// 2s.
func LoginTimer() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			timer := time.NewTimer(loginMinimum + time.Duration(rand.Int64N(loginVariation.Nanoseconds())))

			next.ServeHTTP(response, request)

			select {
			case <-timer.C:
			case <-request.Context().Done():
			}
		})
	}
}
