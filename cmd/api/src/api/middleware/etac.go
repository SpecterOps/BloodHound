package middleware

import (
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
 			} else if !etacFlag.Enabled{
				next.ServeHTTP(response, request)
			} else if bhCtx := ctx.FromRequest(request); !bhCtx.AuthCtx.Authenticated() {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnauthorized, "not authenticated", request), response)
			} else if currentUser, found := auth.GetUserFromAuthCtx(bhCtx.AuthCtx); !found {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated user found with request", request), response)
			} else if currentUser.AllEnvironments {
				next.ServeHTTP(response, request)
			} else if domainsid, hasDomainID := mux.Vars(request)[api.URIPathVariableObjectID]; !hasDomainID {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorNoDomainId, request), response)
			} else if hasAccess, err := v2.CheckUserAccessToEnvironments(request.Context(), db, currentUser, domainsid); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "error checking user's environment access control", request), response)
			} else if !hasAccess {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "User does not have permission to access this domain", request), response)
			} else {
				next.ServeHTTP(response, request)
			}
		})
	}
}
