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

package registration

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	authapi "github.com/specterops/bloodhound/cmd/api/src/api/v2/auth"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/openapi"
	"github.com/specterops/bloodhound/packages/go/params"
)

func registerV2Auth(resources v2.Resources, routerInst *router.Router, permissions auth.PermissionSet) {
	var (
		loginResource      = authapi.NewLoginResource(resources.Config, resources.Authenticator, resources.DB)
		managementResource = authapi.NewManagementResource(resources.Config, resources.DB, resources.Authorizer, resources.Authenticator, resources.GraphQuery, resources.DogTags)
	)

	router.With(func() mux.MiddlewareFunc {
		return middleware.RateLimitMiddleware(resources.DB, 1)
	},
		// Login resource
		routerInst.POST("/api/v2/login", func(response http.ResponseWriter, request *http.Request) {
			middleware.LoginTimer()(http.HandlerFunc(loginResource.Login)).ServeHTTP(response, request)
		}),
	)

	router.With(func() mux.MiddlewareFunc {
		return middleware.DefaultRateLimitMiddleware(resources.DB)
	},
		// Login resources
		routerInst.GET("/api/v2/self", managementResource.GetSelf),
		routerInst.POST("/api/v2/logout", loginResource.Logout),

		// Login path prefix matcher for SAML providers, order matters here due to PathPrefix
		routerInst.POST(fmt.Sprintf("/api/{version}/login/saml/{%s}/acs", api.URIPathVariableSSOProviderSlug), managementResource.SAMLCallbackRedirect),
		routerInst.GET(fmt.Sprintf("/api/{version}/login/saml/{%s}/metadata", api.URIPathVariableSSOProviderSlug), managementResource.ServeMetadata),
		routerInst.PathPrefix(fmt.Sprintf("/api/{version}/login/saml/{%s}", api.URIPathVariableSSOProviderSlug), http.HandlerFunc(managementResource.SAMLLoginRedirect)),

		// SAML resources
		// DEPRECATED as of v6.4.0: Please use /api/v2/sso-providers/* endpoints instead of /api/v2/saml/*
		routerInst.GET("/api/v2/saml", managementResource.ListSAMLProviders).RequirePermissions(permissions.AuthManageProviders),
		routerInst.GET("/api/v2/saml/sso", managementResource.ListSAMLSignOnEndpoints),
		routerInst.POST("/api/v2/saml/providers", managementResource.CreateSAMLProviderMultipart).RequirePermissions(permissions.AuthManageProviders),
		routerInst.GET(fmt.Sprintf("/api/v2/saml/providers/{%s}", api.URIPathVariableSAMLProviderID), managementResource.GetSAMLProvider).RequirePermissions(permissions.AuthManageProviders),
		routerInst.DELETE(fmt.Sprintf("/api/v2/saml/providers/{%s}", api.URIPathVariableSAMLProviderID), managementResource.DeleteSAMLProvider).RequirePermissions(permissions.AuthManageProviders),

		// SSO
		routerInst.GET("/api/v2/sso-providers", managementResource.ListAuthProviders),
		routerInst.POST("/api/v2/sso-providers/saml", managementResource.CreateSAMLProviderMultipart).RequirePermissions(permissions.AuthManageProviders),
		routerInst.POST("/api/v2/sso-providers/oidc", managementResource.CreateOIDCProvider).CheckFeatureFlag(resources.DB, appcfg.FeatureOIDCSupport).RequirePermissions(permissions.AuthManageProviders),
		routerInst.DELETE(fmt.Sprintf("/api/v2/sso-providers/{%s}", api.URIPathVariableSSOProviderID), managementResource.DeleteSSOProvider).RequirePermissions(permissions.AuthManageProviders),
		routerInst.PATCH(fmt.Sprintf("/api/v2/sso-providers/{%s}", api.URIPathVariableSSOProviderID), managementResource.UpdateSSOProvider).RequirePermissions(permissions.AuthManageProviders),
		routerInst.GET(fmt.Sprintf("/api/v2/sso-providers/{%s}/signing-certificate", api.URIPathVariableSSOProviderID), managementResource.ServeSigningCertificate).RequirePermissions(permissions.AuthManageProviders),

		routerInst.GET(fmt.Sprintf("/api/v2/sso/{%s}/login", api.URIPathVariableSSOProviderSlug), managementResource.SSOLoginHandler),
		routerInst.PathPrefix(fmt.Sprintf("/api/v2/sso/{%s}/callback", api.URIPathVariableSSOProviderSlug), http.HandlerFunc(managementResource.SSOCallbackHandler)),
		routerInst.GET(fmt.Sprintf("/api/v2/sso/{%s}/metadata", api.URIPathVariableSSOProviderSlug), managementResource.ServeMetadata),

		// Permissions
		routerInst.GET("/api/v2/permissions", managementResource.ListPermissions).RequirePermissions(permissions.AuthManageSelf),
		routerInst.GET(fmt.Sprintf("/api/v2/permissions/{%s}", api.URIPathVariablePermissionID), managementResource.GetPermission).RequirePermissions(permissions.AuthManageSelf),

		// Roles
		routerInst.GET("/api/v2/roles", managementResource.ListRoles).RequirePermissions(permissions.AuthManageSelf),
		routerInst.GET(fmt.Sprintf("/api/v2/roles/{%s}", api.URIPathVariableRoleID), managementResource.GetRole).RequirePermissions(permissions.AuthManageSelf),

		// User management for all BloodHound users
		routerInst.GET("/api/v2/bloodhound-users", managementResource.ListUsers).RequirePermissions(permissions.AuthManageUsers),
		routerInst.POST("/api/v2/bloodhound-users", managementResource.CreateUser).RequirePermissions(permissions.AuthManageUsers),
		routerInst.GET("/api/v2/bloodhound-users-minimal", managementResource.ListActiveUsersMinimal).RequirePermissions(permissions.AuthReadUsers), // returns user data without any sensitive information.

		routerInst.GET(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), managementResource.GetUser).RequirePermissions(permissions.AuthManageUsers),
		routerInst.PATCH(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), managementResource.UpdateUser).RequirePermissions(permissions.AuthManageUsers),
		routerInst.DELETE(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), managementResource.DeleteUser).RequirePermissions(permissions.AuthManageUsers),

		routerInst.PUT(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/secret", api.URIPathVariableUserID), managementResource.PutUserAuthSecret).AuthorizeUserManagementAccess().RequireUserId(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/secret", api.URIPathVariableUserID), managementResource.ExpireUserAuthSecret).AuthorizeUserManagementAccess().RequireUserId(),

		routerInst.POST(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa", api.URIPathVariableUserID), managementResource.EnrollMFA).AuthorizeUserManagementAccess().RequireUserId(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa", api.URIPathVariableUserID), managementResource.DisenrollMFA).AuthorizeUserManagementAccess().RequireUserId(),
		routerInst.GET(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa-activation", api.URIPathVariableUserID), managementResource.GetMFAActivationStatus).AuthorizeUserManagementAccess().RequireUserId(),
		routerInst.POST(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa-activation", api.URIPathVariableUserID), managementResource.ActivateMFA).AuthorizeUserManagementAccess().RequireUserId(),

		routerInst.POST("/api/v2/tokens", managementResource.CreateAuthToken).RequirePermissions(permissions.AuthCreateToken).AuthorizeUserManagementAccess(),
		routerInst.GET("/api/v2/tokens", managementResource.ListAuthTokens).RequirePermissions(permissions.AuthCreateToken).AuthorizeUserManagementAccess(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/tokens/{%s}", api.URIPathVariableTokenID), managementResource.DeleteAuthToken).RequirePermissions(permissions.AuthCreateToken).AuthorizeUserManagementAccess(),
	)
}

// NewV2API sets up dependencies, authorization and a router, and then defines the BloodHound V2 API endpoints on said router
func NewV2API(resources v2.Resources, routerInst *router.Router) {
	var permissions = auth.Permissions()

	// Register the auth API endpoints
	registerV2Auth(resources, routerInst, permissions)

	// Collector APIs
	routerInst.GET(fmt.Sprintf("/api/v2/collectors/{%s}", v2.CollectorTypePathParameterName), resources.GetCollectorManifest).RequireAuth()
	routerInst.GET(fmt.Sprintf("/api/v2/collectors/{%s}/{%s:v[0-9]+.[0-9]+.[0-9]+|latest}", v2.CollectorTypePathParameterName, v2.CollectorReleaseTagPathParameterName), resources.DownloadCollectorByVersion).RequireAuth()
	routerInst.GET(fmt.Sprintf("/api/v2/collectors/{%s}/{%s:v[0-9]+.[0-9]+.[0-9]+|latest}/checksum", v2.CollectorTypePathParameterName, v2.CollectorReleaseTagPathParameterName), resources.DownloadCollectorChecksumByVersion).RequireAuth()

	// Generic Ingest + File Upload API
	routerInst.GET("/api/v2/file-upload", resources.ListIngestJobs).RequireAuth()
	routerInst.GET("/api/v2/file-upload/accepted-types", resources.ListAcceptedFileUploadTypes).RequireAuth()
	routerInst.POST("/api/v2/file-upload/start", resources.StartIngestJob).RequirePermissions(permissions.GraphDBIngest)
	routerInst.POST(fmt.Sprintf("/api/v2/file-upload/{%s}", v2.FileUploadJobIdPathParameterName), resources.ProcessIngestTask).RequirePermissions(permissions.GraphDBIngest)
	routerInst.GET(fmt.Sprintf("/api/v2/file-upload/{%s}/completed-tasks", v2.FileUploadJobIdPathParameterName), resources.GetCompletedTasks).RequirePermissions(permissions.GraphDBIngest)
	routerInst.POST(fmt.Sprintf("/api/v2/file-upload/{%s}/end", v2.FileUploadJobIdPathParameterName), resources.EndIngestJob).RequirePermissions(permissions.GraphDBIngest)

	router.With(func() mux.MiddlewareFunc {
		return middleware.DefaultRateLimitMiddleware(resources.DB)
	},
		// Version API
		routerInst.GET("/api/version", v2.GetVersion).RequireAuth(),

		// DogTags API
		routerInst.GET("/api/v2/dog-tags", resources.GetDogTags).RequireAuth(),

		// API Spec
		routerInst.PathPrefix("/api/v2/swagger", http.HandlerFunc(openapi.HttpHandler)),
		routerInst.GET("/api/v2/spec", openapi.HttpHandler),

		// Search API
		routerInst.GET("/api/v2/search", resources.SearchHandler).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET("/api/v2/available-domains", resources.ListAvailableEnvironments).RequirePermissions(permissions.GraphDBRead),

		// Audit API
		routerInst.GET("/api/v2/audit", resources.ListAuditLogs).RequirePermissions(permissions.AuditLogRead),

		// App Config API
		routerInst.GET("/api/v2/config", resources.GetApplicationConfigurations).RequirePermissions(permissions.AppReadApplicationConfiguration),
		routerInst.PUT("/api/v2/config", resources.SetApplicationConfiguration).RequirePermissions(permissions.AppWriteApplicationConfiguration),

		routerInst.GET("/api/v2/features", resources.GetFlags).RequirePermissions(permissions.AppReadApplicationConfiguration),
		routerInst.PUT("/api/v2/features/{feature_id}/toggle", resources.ToggleFlag).RequirePermissions(permissions.AppWriteApplicationConfiguration),

		routerInst.POST("/api/v2/clear-database", resources.HandleDatabaseWipe).RequirePermissions(permissions.WipeDB),

		// Asset Groups API
		routerInst.GET("/api/v2/asset-groups", resources.ListAssetGroups).RequirePermissions(permissions.GraphDBRead),
		routerInst.POST("/api/v2/asset-groups", resources.CreateAssetGroup).RequirePermissions(permissions.GraphDBWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}", api.URIPathVariableAssetGroupID), resources.GetAssetGroup).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/custom-selectors", api.URIPathVariableAssetGroupID), resources.GetAssetGroupCustomMemberCount).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.DELETE(fmt.Sprintf("/api/v2/asset-groups/{%s}", api.URIPathVariableAssetGroupID), resources.DeleteAssetGroup).RequirePermissions(permissions.GraphDBWrite),
		routerInst.PUT(fmt.Sprintf("/api/v2/asset-groups/{%s}", api.URIPathVariableAssetGroupID), resources.UpdateAssetGroup).RequirePermissions(permissions.GraphDBWrite),
		routerInst.DELETE(fmt.Sprintf("/api/v2/asset-groups/{%s}/selectors/{%s}", api.URIPathVariableAssetGroupID, api.URIPathVariableAssetGroupSelectorID), resources.DeleteAssetGroupSelector).RequirePermissions(permissions.GraphDBWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/collections", api.URIPathVariableAssetGroupID), resources.ListAssetGroupCollections).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/members", api.URIPathVariableAssetGroupID), resources.ListAssetGroupMembers).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/members/counts", api.URIPathVariableAssetGroupID), resources.ListAssetGroupMemberCountsByKind).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.PUT(fmt.Sprintf("/api/v2/asset-groups/{%s}/selectors", api.URIPathVariableAssetGroupID), resources.UpdateAssetGroupSelectors).RequirePermissions(permissions.GraphDBWrite),
		// DEPRECATED: this has been changed to a PUT endpoint above, and must be removed for API V3
		routerInst.POST(fmt.Sprintf("/api/v2/asset-groups/{%s}/selectors", api.URIPathVariableAssetGroupID), resources.UpdateAssetGroupSelectors).RequirePermissions(permissions.GraphDBWrite),

		// Asset group management API
		// tags
		routerInst.GET("/api/v2/asset-group-tags", resources.GetAssetGroupTags).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead),
		routerInst.PATCH(fmt.Sprintf("/api/v2/asset-group-tags/{%s}", api.URIPathVariableAssetGroupTagID), resources.UpdateAssetGroupTag).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBWrite),
		routerInst.POST("/api/v2/asset-group-tags/search", resources.SearchAssetGroupTags).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}", api.URIPathVariableAssetGroupTagID), resources.GetAssetGroupTag).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/members", api.URIPathVariableAssetGroupTagID), resources.GetAssetGroupMembersByTag).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/members/counts", api.URIPathVariableAssetGroupTagID), resources.GetAssetGroupTagMemberCountsByKind).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/members/{%s}", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagMemberID), resources.GetAssetGroupTagMemberInfo).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),

		// selectors
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors", api.URIPathVariableAssetGroupTagID), resources.GetAssetGroupTagSelectors).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.POST(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors", api.URIPathVariableAssetGroupTagID), resources.CreateAssetGroupTagSelector).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors/{%s}", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagSelectorID), resources.GetAssetGroupTagSelector).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.PATCH(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors/{%s}", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagSelectorID), resources.UpdateAssetGroupTagSelector).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBWrite),
		routerInst.DELETE(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors/{%s}", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagSelectorID), resources.DeleteAssetGroupTagSelector).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBWrite),
		routerInst.POST("/api/v2/asset-group-tags/preview-selectors", resources.PreviewSelectors).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors/{%s}/members", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagSelectorID), resources.GetAssetGroupMembersBySelector).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-group-tags/{%s}/selectors/{%s}/members/counts", api.URIPathVariableAssetGroupTagID, api.URIPathVariableAssetGroupTagSelectorID), resources.GetAssetGroupSelectorMemberCountsByKind).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),

		// history
		routerInst.GET("/api/v2/asset-group-tags-history", resources.GetAssetGroupTagHistory).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.POST("/api/v2/asset-group-tags-history", resources.SearchAssetGroupTagHistory).CheckFeatureFlag(resources.DB, appcfg.FeatureTierManagement).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),

		// QA API
		routerInst.GET("/api/v2/completeness", resources.GetDatabaseCompleteness).RequirePermissions(permissions.GraphDBRead),

		routerInst.GET("/api/v2/pathfinding", resources.GetPathfindingResult).Queries("start_node", "{start_node}", "end_node", "{end_node}").RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET("/api/v2/graphs/kinds", resources.ListKinds).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET("/api/v2/graphs/source-kinds", resources.ListSourceKinds).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET("/api/v2/graphs/shortest-path", resources.GetShortestPath).Queries(params.StartNode.String(), params.StartNode.RouteMatcher(), params.EndNode.String(), params.EndNode.RouteMatcher()).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET("/api/v2/graphs/edge-composition", resources.GetEdgeComposition).RequirePermissions(permissions.GraphDBRead).RequireAllEnvironmentAccess(resources.DogTags),
		routerInst.GET("/api/v2/graphs/relay-targets", resources.GetEdgeRelayTargets).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET("/api/v2/graphs/acl-inheritance", resources.GetEdgeACLInheritancePath).RequirePermissions(permissions.GraphDBRead),

		// TODO discuss if this should be a post endpoint
		routerInst.GET("/api/v2/graph-search", resources.GetSearchResult).RequirePermissions(permissions.GraphDBRead),

		// Cypher Queries API
		routerInst.POST("/api/v2/graphs/cypher", resources.CypherQuery).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET("/api/v2/saved-queries", resources.ListSavedQueries).RequirePermissions(permissions.SavedQueriesRead),
		routerInst.POST("/api/v2/saved-queries", resources.CreateSavedQuery).RequirePermissions(permissions.SavedQueriesWrite),
		routerInst.GET("/api/v2/saved-queries/export", resources.ExportSavedQueries).RequirePermissions(permissions.SavedQueriesRead),
		routerInst.POST("/api/v2/saved-queries/import", resources.ImportSavedQueries).RequirePermissions(permissions.SavedQueriesWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/saved-queries/{%s}", api.URIPathVariableSavedQueryID), resources.GetSavedQuery).RequirePermissions(permissions.SavedQueriesRead),
		routerInst.PUT(fmt.Sprintf("/api/v2/saved-queries/{%s}", api.URIPathVariableSavedQueryID), resources.UpdateSavedQuery).RequirePermissions(permissions.SavedQueriesWrite),
		routerInst.DELETE(fmt.Sprintf("/api/v2/saved-queries/{%s}", api.URIPathVariableSavedQueryID), resources.DeleteSavedQuery).RequirePermissions(permissions.SavedQueriesWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/saved-queries/{%s}/permissions", api.URIPathVariableSavedQueryID), resources.GetSavedQueryPermissions).RequirePermissions(permissions.SavedQueriesRead),
		routerInst.DELETE(fmt.Sprintf("/api/v2/saved-queries/{%s}/permissions", api.URIPathVariableSavedQueryID), resources.DeleteSavedQueryPermissions).RequirePermissions(permissions.SavedQueriesWrite),
		routerInst.PUT(fmt.Sprintf("/api/v2/saved-queries/{%s}/permissions", api.URIPathVariableSavedQueryID), resources.ShareSavedQueries).RequirePermissions(permissions.SavedQueriesWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/saved-queries/{%s}/export", api.URIPathVariableSavedQueryID), resources.ExportSavedQuery).RequirePermissions(permissions.SavedQueriesRead),

		// Azure Entity API
		routerInst.GET("/api/v2/azure/{entity_type}", resources.GetAZEntity).RequirePermissions(permissions.GraphDBRead),

		// Base Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/base/{%s}", api.URIPathVariableObjectID), resources.GetBaseEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/base/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/base/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(permissions.GraphDBRead),

		// Computer Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}", api.URIPathVariableObjectID), resources.GetComputerEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/sessions", api.URIPathVariableObjectID), resources.ListADComputerSessions).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/admin-users", api.URIPathVariableObjectID), resources.ListADComputerAdmins).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/rdp-users", api.URIPathVariableObjectID), resources.ListADComputerRDPUsers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/dcom-users", api.URIPathVariableObjectID), resources.ListADComputerDCOMUsers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/ps-remote-users", api.URIPathVariableObjectID), resources.ListADComputerPSRemoteUsers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/sql-admins", api.URIPathVariableObjectID), resources.ListADComputerSQLAdmins).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/group-membership", api.URIPathVariableObjectID), resources.ListADGroupMembership).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/admin-rights", api.URIPathVariableObjectID), resources.ListADEntityAdminRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/rdp-rights", api.URIPathVariableObjectID), resources.ListADEntityRDPRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/dcom-rights", api.URIPathVariableObjectID), resources.ListADEntityDCOMRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/ps-remote-rights", api.URIPathVariableObjectID), resources.ListADEntityPSRemoteRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/constrained-users", api.URIPathVariableObjectID), resources.ListADComputerConstrainedDelegationUsers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/constrained-delegation-rights", api.URIPathVariableObjectID), resources.ListADEntityConstrainedDelegationRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(permissions.GraphDBRead),

		// Container Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/containers/{%s}", api.URIPathVariableObjectID), resources.GetContainerEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/containers/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),

		// Domain Entity API
		routerInst.PATCH(fmt.Sprintf("/api/v2/domains/{%s}", api.URIPathVariableObjectID), resources.PatchDomain).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}", api.URIPathVariableObjectID), resources.GetDomainEntityInfo).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/users", api.URIPathVariableObjectID), resources.ListADDomainContainedUsers).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/groups", api.URIPathVariableObjectID), resources.ListADDomainContainedGroups).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/computers", api.URIPathVariableObjectID), resources.ListADDomainContainedComputers).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/ous", api.URIPathVariableObjectID), resources.ListADDomainContainedOUs).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/gpos", api.URIPathVariableObjectID), resources.ListADDomainContainedGPOs).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-users", api.URIPathVariableObjectID), resources.ListADDomainForeignUsers).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-groups", api.URIPathVariableObjectID), resources.ListADDomainForeignGroups).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-admins", api.URIPathVariableObjectID), resources.ListADDomainForeignAdmins).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-gpo-controllers", api.URIPathVariableObjectID), resources.ListADDomainForeignGPOControllers).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/inbound-trusts", api.URIPathVariableObjectID), resources.ListADDomainInboundTrusts).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/outbound-trusts", api.URIPathVariableObjectID), resources.ListADDomainOutboundTrusts).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/dc-syncers", api.URIPathVariableObjectID), resources.ListADDomainDCSyncers).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/linked-gpos", api.URIPathVariableObjectID), resources.ListADEntityLinkedGPOs).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/adcs-escalations", api.URIPathVariableObjectID), resources.ListADCSEscalations).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),

		// GPO Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}", api.URIPathVariableObjectID), resources.GetGPOEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/computers", api.URIPathVariableObjectID), resources.ListADGPOAffectedComputers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/users", api.URIPathVariableObjectID), resources.ListADGPOAffectedUsers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/tier-zero", api.URIPathVariableObjectID), resources.ListADGPOAffectedTierZero).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/ous", api.URIPathVariableObjectID), resources.ListADGPOAffectedContainers).RequirePermissions(permissions.GraphDBRead),

		// AIACA Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/aiacas/{%s}", api.URIPathVariableObjectID), resources.GetAIACAEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/aiacas/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/aiacas/{%s}/pki-hierarchy", api.URIPathVariableObjectID), resources.ListCAPKIHierarchy).RequirePermissions(permissions.GraphDBRead),

		// RootCA Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/rootcas/{%s}", api.URIPathVariableObjectID), resources.GetRootCAEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/rootcas/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/rootcas/{%s}/pki-hierarchy", api.URIPathVariableObjectID), resources.ListRootCAPKIHierarchy).RequirePermissions(permissions.GraphDBRead),

		// EnterpriseCA Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/enterprisecas/{%s}", api.URIPathVariableObjectID), resources.GetEnterpriseCAEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/enterprisecas/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/enterprisecas/{%s}/pki-hierarchy", api.URIPathVariableObjectID), resources.ListCAPKIHierarchy).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/enterprisecas/{%s}/published-templates", api.URIPathVariableObjectID), resources.ListPublishedTemplates).RequirePermissions(permissions.GraphDBRead),

		// NTAuthStore Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/ntauthstores/{%s}", api.URIPathVariableObjectID), resources.GetNTAuthStoreEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ntauthstores/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ntauthstores/{%s}/trusted-cas", api.URIPathVariableObjectID), resources.ListTrustedCAs).RequirePermissions(permissions.GraphDBRead),

		// CertTemplate Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/certtemplates/{%s}", api.URIPathVariableObjectID), resources.GetCertTemplateEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/certtemplates/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/certtemplates/{%s}/published-to-cas", api.URIPathVariableObjectID), resources.ListPublishedToCAs).RequirePermissions(permissions.GraphDBRead),

		// OU Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}", api.URIPathVariableObjectID), resources.GetOUEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/gpos", api.URIPathVariableObjectID), resources.ListADEntityLinkedGPOs).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/users", api.URIPathVariableObjectID), resources.ListADOUContainedUsers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/groups", api.URIPathVariableObjectID), resources.ListADOUContainedGroups).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/computers", api.URIPathVariableObjectID), resources.ListADOUContainedComputers).RequirePermissions(permissions.GraphDBRead),

		// User Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}", api.URIPathVariableObjectID), resources.GetUserEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/sessions", api.URIPathVariableObjectID), resources.ListADUserSessions).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/memberships", api.URIPathVariableObjectID), resources.ListADGroupMembership).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/admin-rights", api.URIPathVariableObjectID), resources.ListADEntityAdminRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/rdp-rights", api.URIPathVariableObjectID), resources.ListADEntityRDPRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/dcom-rights", api.URIPathVariableObjectID), resources.ListADEntityDCOMRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/ps-remote-rights", api.URIPathVariableObjectID), resources.ListADEntityPSRemoteRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/sql-admin-rights", api.URIPathVariableObjectID), resources.ListADUserSQLAdminRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/constrained-delegation-rights", api.URIPathVariableObjectID), resources.ListADEntityConstrainedDelegationRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(permissions.GraphDBRead),

		// Group Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}", api.URIPathVariableObjectID), resources.GetGroupEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/sessions", api.URIPathVariableObjectID), resources.ListADGroupSessions).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/members", api.URIPathVariableObjectID), resources.ListADGroupMembers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/memberships", api.URIPathVariableObjectID), resources.ListADGroupMembership).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/admin-rights", api.URIPathVariableObjectID), resources.ListADEntityAdminRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/rdp-rights", api.URIPathVariableObjectID), resources.ListADEntityRDPRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/dcom-rights", api.URIPathVariableObjectID), resources.ListADEntityDCOMRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/ps-remote-rights", api.URIPathVariableObjectID), resources.ListADEntityPSRemoteRights).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),

		// IssuancePolicy Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/issuancepolicies/{%s}", api.URIPathVariableObjectID), resources.GetIssuancePolicyEntityInfo).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/issuancepolicies/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/issuancepolicies/{%s}/linkedtemplates", api.URIPathVariableObjectID), resources.ListADIssuancePolicyLinkedCertTemplates).RequirePermissions(permissions.GraphDBRead),

		// Data Quality Stats API
		routerInst.GET(fmt.Sprintf("/api/v2/ad-domains/{%s}/data-quality-stats", api.URIPathVariableDomainID), resources.GetADDataQualityStats).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/azure-tenants/{%s}/data-quality-stats", api.URIPathVariableTenantID), resources.GetAzureDataQualityStats).RequirePermissions(permissions.GraphDBRead).SupportsETAC(resources.DB, resources.DogTags),
		routerInst.GET(fmt.Sprintf("/api/v2/platform/{%s}/data-quality-stats", api.URIPathVariablePlatformID), resources.GetPlatformAggregateStats).RequirePermissions(permissions.GraphDBRead),

		// Datapipe API
		routerInst.GET("/api/v2/datapipe/status", resources.GetDatapipeStatus).RequireAuth(),
		// TODO: Update the permission on this once we get something more concrete
		routerInst.GET("/api/v2/analysis/status", resources.GetAnalysisRequest).RequirePermissions(permissions.GraphDBRead),
		routerInst.PUT("/api/v2/analysis", resources.RequestAnalysis).RequirePermissions(permissions.GraphDBWrite),

		// Custom Node Management
		routerInst.GET("/api/v2/custom-nodes", resources.GetCustomNodeKinds).RequireAuth(),
		routerInst.GET(fmt.Sprintf("/api/v2/custom-nodes/{%s}", v2.CustomNodeKindParameter), resources.GetCustomNodeKind).RequireAuth(),
		routerInst.POST("/api/v2/custom-nodes", resources.CreateCustomNodeKind).RequireAuth(),
		routerInst.PUT(fmt.Sprintf("/api/v2/custom-nodes/{%s}", v2.CustomNodeKindParameter), resources.UpdateCustomNodeKind).RequireAuth(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/custom-nodes/{%s}", v2.CustomNodeKindParameter), resources.DeleteCustomNodeKind).RequireAuth(),

		// Open Graph Schema
		routerInst.PUT("/api/v2/extensions", resources.OpenGraphSchemaIngest).CheckFeatureFlag(resources.DB, appcfg.FeatureOpenGraphExtensionManagement).RequireAuth(),
		routerInst.GET("/api/v2/extensions", resources.ListExtensions).CheckFeatureFlag(resources.DB, appcfg.FeatureOpenGraphExtensionManagement).RequireAuth(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/extensions/{%s}", api.URIPathVariableExtensionID), resources.DeleteExtension).CheckFeatureFlag(resources.DB, appcfg.FeatureOpenGraphExtensionManagement).RequireAuth(),

		// Graph Schema API
		routerInst.GET("/api/v2/graph-schema/edges", resources.ListEdgeTypes).CheckFeatureFlag(resources.DB, appcfg.FeatureOpenGraphExtensionManagement).RequirePermissions(permissions.GraphDBRead),
	)
}
