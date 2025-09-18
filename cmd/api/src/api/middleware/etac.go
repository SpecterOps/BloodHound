package middleware

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
)

// SupportsETACMiddleware will check a user's environment access control to determine if they have access to the environment provided in the url
// If a user has the AllEnvironments flag set to true, they will be given access to all environments
func SupportsETACMiddleware(db database.Database) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if etacFlag, err := db.GetFlagByKey(request.Context(), appcfg.FeatureEnvironmentAccessControl); err != nil {
				api.HandleDatabaseError(request, response, err)
			} else if !etacFlag.Enabled {
				next.ServeHTTP(response, request)
			} else if bhCtx := ctx.FromRequest(request); !bhCtx.AuthCtx.Authenticated() {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "not authenticated", request), response)
			} else if currentUser, found := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !found {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found with request", request), response)
			} else if currentUser.AllEnvironments {
				next.ServeHTTP(response, request)
			} else if envvironmentID, err := getEnvironmentIdFromRequest(request); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorNoDomainId, request), response)
			} else if hasAccess, err := v2.CheckUserAccessToEnvironments(request.Context(), db, currentUser, envvironmentID); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error checking user's environment access control", request), response)
			} else if !hasAccess {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "User does not have permission to access this domain", request), response)
			} else {
				next.ServeHTTP(response, request)
			}
		})
	}
}

// getEnvironmentIdFromRequest will pull the environment id from the request's path variables where the environment id can be equal to an objectid, tenantid, or domainsid
func getEnvironmentIdFromRequest(request *http.Request) (string, error) {
	if domainSID, hasDomainSID := mux.Vars(request)[api.URIPathVariableDomainID]; hasDomainSID {
		return domainSID, nil
	} else if objectID, hasObjectID := mux.Vars(request)[api.URIPathVariableObjectID]; hasObjectID {
		return objectID, nil
	} else if tenantID, hasTenantID := mux.Vars(request)[api.URIPathVariableTenantID]; hasTenantID {
		return tenantID, nil
	} else {
		return "", errors.New("environment id could not be found in url")
	}
}
