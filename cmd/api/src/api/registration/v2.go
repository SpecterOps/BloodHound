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

	"github.com/specterops/bloodhound/params"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/middleware"
	"github.com/specterops/bloodhound/src/api/router"
	"github.com/specterops/bloodhound/src/api/saml"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	authapi "github.com/specterops/bloodhound/src/api/v2/auth"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

func samlWriteAPIErrorResponse(request *http.Request, response http.ResponseWriter, statusCode int, message string) {
	api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(statusCode, message, request), response)
}

func registerV2Auth(cfg config.Configuration, db database.Database, permissions auth.PermissionSet, routerInst *router.Router, authenticator api.Authenticator) {
	var (
		loginResource      = authapi.NewLoginResource(cfg, authenticator, db)
		managementResource = authapi.NewManagementResource(cfg, db, auth.NewAuthorizer())
		samlResource       = saml.NewSAMLRootResource(cfg, db, samlWriteAPIErrorResponse)
	)

	router.With(middleware.DefaultRateLimitMiddleware,
		// Login resources
		routerInst.POST("/api/v2/login", loginResource.Login),
		routerInst.GET("/api/v2/self", managementResource.GetSelf),
		routerInst.POST("/api/v2/logout", loginResource.Logout),

		// Login path prefix matcher for SAML providers
		routerInst.PathPrefix("/api/v2/login/saml/{saml_provider_name}", middleware.ContextMiddleware(samlResource)),

		// SAML resources
		routerInst.GET("/api/v2/saml", managementResource.ListSAMLProviders).RequirePermissions(db, permissions.AuthManageProviders),
		routerInst.GET("/api/v2/saml/sso", managementResource.ListSAMLSignOnEndpoints),
		routerInst.POST("/api/v2/saml/providers", managementResource.CreateSAMLProviderMultipart).RequirePermissions(db, permissions.AuthManageProviders),
		routerInst.GET(fmt.Sprintf("/api/v2/saml/providers/{%s}", api.URIPathVariableSAMLProviderID), managementResource.GetSAMLProvider).RequirePermissions(db, permissions.AuthManageProviders),
		routerInst.DELETE(fmt.Sprintf("/api/v2/saml/providers/{%s}", api.URIPathVariableSAMLProviderID), managementResource.DeleteSAMLProvider).RequirePermissions(db, permissions.AuthManageProviders),

		// Permissions
		routerInst.GET("/api/v2/permissions", managementResource.ListPermissions).RequirePermissions(db, permissions.AuthManageSelf),
		routerInst.GET(fmt.Sprintf("/api/v2/permissions/{%s}", api.URIPathVariablePermissionID), managementResource.GetPermission).RequirePermissions(db, permissions.AuthManageSelf),

		// Roles
		routerInst.GET("/api/v2/roles", managementResource.ListRoles).RequirePermissions(db, permissions.AuthManageSelf),
		routerInst.GET(fmt.Sprintf("/api/v2/roles/{%s}", api.URIPathVariableRoleID), managementResource.GetRole).RequirePermissions(db, permissions.AuthManageSelf),

		// User management for all BloodHound users
		routerInst.GET("/api/v2/bloodhound-users", managementResource.ListUsers).RequirePermissions(db, permissions.AuthManageUsers),
		routerInst.POST("/api/v2/bloodhound-users", managementResource.CreateUser).RequirePermissions(db, permissions.AuthManageUsers),

		routerInst.GET(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), managementResource.GetUser).RequirePermissions(db, permissions.AuthManageUsers),
		routerInst.PATCH(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), managementResource.UpdateUser).RequirePermissions(db, permissions.AuthManageUsers),
		routerInst.DELETE(fmt.Sprintf("/api/v2/bloodhound-users/{%s}", api.URIPathVariableUserID), managementResource.DeleteUser).RequirePermissions(db, permissions.AuthManageUsers),

		routerInst.PUT(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/secret", api.URIPathVariableUserID), managementResource.PutUserAuthSecret).AuthorizeUserManagementAccess(db).RequireUserId(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/secret", api.URIPathVariableUserID), managementResource.ExpireUserAuthSecret).AuthorizeUserManagementAccess(db).RequireUserId(),

		routerInst.POST(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa", api.URIPathVariableUserID), managementResource.EnrollMFA).AuthorizeUserManagementAccess(db).RequireUserId(),
		routerInst.DELETE(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa", api.URIPathVariableUserID), managementResource.DisenrollMFA).AuthorizeUserManagementAccess(db).RequireUserId(),
		routerInst.GET(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa-activation", api.URIPathVariableUserID), managementResource.GetMFAActivationStatus).AuthorizeUserManagementAccess(db).RequireUserId(),
		routerInst.POST(fmt.Sprintf("/api/v2/bloodhound-users/{%s}/mfa-activation", api.URIPathVariableUserID), managementResource.ActivateMFA).AuthorizeUserManagementAccess(db).RequireUserId(),

		routerInst.POST("/api/v2/tokens", managementResource.CreateAuthToken).RequirePermissions(db, permissions.AuthCreateToken).AuthorizeUserManagementAccess(db),
		routerInst.GET("/api/v2/tokens", managementResource.ListAuthTokens).RequirePermissions(db, permissions.AuthCreateToken).AuthorizeUserManagementAccess(db),
		routerInst.DELETE(fmt.Sprintf("/api/v2/tokens/{%s}", api.URIPathVariableTokenID), managementResource.DeleteAuthToken).RequirePermissions(db, permissions.AuthCreateToken).AuthorizeUserManagementAccess(db),
	)
}

// NewV2API sets up dependencies, authorization and a router, and then defines the BloodHound V2 API endpoints on said router
func NewV2API(cfg config.Configuration, resources v2.Resources, routerInst *router.Router, authenticator api.Authenticator) {
	var permissions = auth.Permissions()

	// Register the auth API endpoints
	registerV2Auth(cfg, resources.DB, permissions, routerInst, authenticator)

	// Collector APIs
	routerInst.GET(fmt.Sprintf("/api/v2/collectors/{%s}", v2.CollectorTypePathParameterName), resources.GetCollectorManifest).RequireAuth(resources.DB)
	routerInst.GET(fmt.Sprintf("/api/v2/collectors/{%s}/{%s:v[0-9]+.[0-9]+.[0-9]+|latest}", v2.CollectorTypePathParameterName, v2.CollectorReleaseTagPathParameterName), resources.DownloadCollectorByVersion).RequireAuth(resources.DB)
	routerInst.GET(fmt.Sprintf("/api/v2/collectors/{%s}/{%s:v[0-9]+.[0-9]+.[0-9]+|latest}/checksum", v2.CollectorTypePathParameterName, v2.CollectorReleaseTagPathParameterName), resources.DownloadCollectorChecksumByVersion).RequireAuth(resources.DB)

	// Ingest APIs
	//TODO: What permission should we use here? GraphDB Write
	routerInst.GET("/api/v2/file-upload", resources.ListFileUploadJobs).RequireAuth(resources.DB)
	routerInst.POST("/api/v2/file-upload/start", resources.StartFileUploadJob).RequirePermissions(resources.DB, permissions.GraphDBWrite)
	routerInst.POST(fmt.Sprintf("/api/v2/file-upload/{%s}", v2.FileUploadJobIdPathParameterName), resources.ProcessFileUpload).RequirePermissions(resources.DB, permissions.GraphDBWrite)
	routerInst.POST(fmt.Sprintf("/api/v2/file-upload/{%s}/end", v2.FileUploadJobIdPathParameterName), resources.EndFileUploadJob).RequirePermissions(resources.DB, permissions.GraphDBWrite)

	router.With(middleware.DefaultRateLimitMiddleware,
		// Version API
		routerInst.GET("/api/version", v2.GetVersion).RequireAuth(resources.DB),

		// Swagger API
		routerInst.PathPrefix("/api/v2/swagger", v2.SwaggerHandler()),

		// Search API
		routerInst.GET("/api/v2/search", resources.SearchHandler).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET("/api/v2/available-domains", resources.GetAvailableDomains).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Audit API
		// TODO: This might actually need its own permission that's assigned to the Administrator user by default
		routerInst.GET("/api/v2/audit", resources.ListAuditLogs).RequirePermissions(resources.DB, permissions.AuthManageUsers),

		// App Config API
		routerInst.GET("/api/v2/config", resources.GetApplicationConfigurations).RequirePermissions(resources.DB, permissions.AppReadApplicationConfiguration),
		routerInst.PUT("/api/v2/config", resources.SetApplicationConfiguration).RequirePermissions(resources.DB, permissions.AppWriteApplicationConfiguration),

		routerInst.GET("/api/v2/features", resources.GetFlags).RequirePermissions(resources.DB, permissions.AppReadApplicationConfiguration),
		routerInst.PUT("/api/v2/features/{feature_id}/toggle", resources.ToggleFlag).RequirePermissions(resources.DB, permissions.AppWriteApplicationConfiguration),

		// Asset Groups API
		routerInst.GET("/api/v2/asset-groups", resources.ListAssetGroups).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.POST("/api/v2/asset-groups", resources.CreateAssetGroup).RequirePermissions(resources.DB, permissions.GraphDBWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}", api.URIPathVariableAssetGroupID), resources.GetAssetGroup).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/custom-selectors", api.URIPathVariableAssetGroupID), resources.GetAssetGroupCustomMemberCount).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.DELETE(fmt.Sprintf("/api/v2/asset-groups/{%s}", api.URIPathVariableAssetGroupID), resources.DeleteAssetGroup).RequirePermissions(resources.DB, permissions.GraphDBWrite),
		routerInst.PUT(fmt.Sprintf("/api/v2/asset-groups/{%s}", api.URIPathVariableAssetGroupID), resources.UpdateAssetGroup).RequirePermissions(resources.DB, permissions.GraphDBWrite),
		routerInst.DELETE(fmt.Sprintf("/api/v2/asset-groups/{%s}/selectors/{%s}", api.URIPathVariableAssetGroupID, api.URIPathVariableAssetGroupSelectorID), resources.DeleteAssetGroupSelector).RequirePermissions(resources.DB, permissions.GraphDBWrite),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/collections", api.URIPathVariableAssetGroupID), resources.ListAssetGroupCollections).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/asset-groups/{%s}/members", api.URIPathVariableAssetGroupID), resources.ListAssetGroupMembers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.PUT(fmt.Sprintf("/api/v2/asset-groups/{%s}/selectors", api.URIPathVariableAssetGroupID), resources.UpdateAssetGroupSelectors).RequirePermissions(resources.DB, permissions.GraphDBWrite),
		// DEPRECATED: this has been changed to a PUT endpoint above, and must be removed for API V3
		routerInst.POST(fmt.Sprintf("/api/v2/asset-groups/{%s}/selectors", api.URIPathVariableAssetGroupID), resources.UpdateAssetGroupSelectors).RequirePermissions(resources.DB, permissions.GraphDBWrite),

		//QA API
		routerInst.GET("/api/v2/completeness", resources.GetDatabaseCompleteness).RequirePermissions(resources.DB, permissions.GraphDBRead),

		routerInst.GET("/api/v2/pathfinding", resources.GetPathfindingResult).Queries("start_node", "{start_node}", "end_node", "{end_node}").RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET("/api/v2/graphs/shortest-path", resources.GetShortestPath).Queries(params.StartNode.String(), params.StartNode.RouteMatcher(), params.EndNode.String(), params.EndNode.RouteMatcher()).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET("/api/v2/graphs/edge-composition", resources.GetEdgeComposition).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs), // TODO: Cleanup #ADCSFeatureFlag after full launch.

		// TODO discuss if this should be a post endpoint
		routerInst.GET("/api/v2/graph-search", resources.GetSearchResult).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Cypher Queries API
		routerInst.POST("/api/v2/graphs/cypher", resources.CypherSearch).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET("/api/v2/saved-queries", resources.ListSavedQueries).RequirePermissions(resources.DB, permissions.SavedQueriesRead),
		routerInst.POST("/api/v2/saved-queries", resources.CreateSavedQuery).RequirePermissions(resources.DB, permissions.SavedQueriesWrite),
		routerInst.DELETE(fmt.Sprintf("/api/v2/saved-queries/{%s}", api.URIPathVariableSavedQueryID), resources.DeleteSavedQuery).RequirePermissions(resources.DB, permissions.SavedQueriesWrite),

		// Azure Entity API
		routerInst.GET("/api/v2/azure/{entity_type}", resources.GetAZEntity).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Base Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/base/{%s}", api.URIPathVariableObjectID), resources.GetBaseEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/base/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/base/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Computer Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}", api.URIPathVariableObjectID), resources.GetComputerEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/sessions", api.URIPathVariableObjectID), resources.ListADComputerSessions).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/admin-users", api.URIPathVariableObjectID), resources.ListADComputerAdmins).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/rdp-users", api.URIPathVariableObjectID), resources.ListADComputerRDPUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/dcom-users", api.URIPathVariableObjectID), resources.ListADComputerDCOMUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/ps-remote-users", api.URIPathVariableObjectID), resources.ListADComputerPSRemoteUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/sql-admins", api.URIPathVariableObjectID), resources.ListADComputerSQLAdmins).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/group-membership", api.URIPathVariableObjectID), resources.ListADGroupMembership).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/admin-rights", api.URIPathVariableObjectID), resources.ListADEntityAdminRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/rdp-rights", api.URIPathVariableObjectID), resources.ListADEntityRDPRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/dcom-rights", api.URIPathVariableObjectID), resources.ListADEntityDCOMRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/ps-remote-rights", api.URIPathVariableObjectID), resources.ListADEntityPSRemoteRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/constrained-users", api.URIPathVariableObjectID), resources.ListADComputerConstrainedDelegationUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/constrained-delegation-rights", api.URIPathVariableObjectID), resources.ListADEntityConstrainedDelegationRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/computers/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Container Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/containers/{%s}", api.URIPathVariableObjectID), resources.GetContainerEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/containers/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Domain Entity API
		routerInst.PATCH(fmt.Sprintf("/api/v2/domains/{%s}", api.URIPathVariableObjectID), resources.PatchDomain).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}", api.URIPathVariableObjectID), resources.GetDomainEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/users", api.URIPathVariableObjectID), resources.ListADDomainContainedUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/groups", api.URIPathVariableObjectID), resources.ListADDomainContainedGroups).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/computers", api.URIPathVariableObjectID), resources.ListADDomainContainedComputers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/ous", api.URIPathVariableObjectID), resources.ListADDomainContainedOUs).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/gpos", api.URIPathVariableObjectID), resources.ListADDomainContainedGPOs).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-users", api.URIPathVariableObjectID), resources.ListADDomainForeignUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-groups", api.URIPathVariableObjectID), resources.ListADDomainForeignGroups).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-admins", api.URIPathVariableObjectID), resources.ListADDomainForeignAdmins).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/foreign-gpo-controllers", api.URIPathVariableObjectID), resources.ListADDomainForeignGPOControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/inbound-trusts", api.URIPathVariableObjectID), resources.ListADDomainInboundTrusts).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/outbound-trusts", api.URIPathVariableObjectID), resources.ListADDomainOutboundTrusts).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/dc-syncers", api.URIPathVariableObjectID), resources.ListADDomainDCSyncers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/domains/{%s}/linked-gpos", api.URIPathVariableObjectID), resources.ListADEntityLinkedGPOs).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// GPO Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}", api.URIPathVariableObjectID), resources.GetGPOEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/computers", api.URIPathVariableObjectID), resources.ListADGPOAffectedComputers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/users", api.URIPathVariableObjectID), resources.ListADGPOAffectedUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/tier-zero", api.URIPathVariableObjectID), resources.ListADGPOAffectedTierZero).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/gpos/{%s}/ous", api.URIPathVariableObjectID), resources.ListADGPOAffectedContainers).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// AIACA Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/aiacas/{%s}", api.URIPathVariableObjectID), resources.GetAIACAEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs),                  // TODO: Cleanup #ADCSFeatureFlag after full launch.
		routerInst.GET(fmt.Sprintf("/api/v2/aiacas/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs), // TODO: Cleanup #ADCSFeatureFlag after full launch.

		// RootCA Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/rootcas/{%s}", api.URIPathVariableObjectID), resources.GetRootCAEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs),                 // TODO: Cleanup #ADCSFeatureFlag after full launch.
		routerInst.GET(fmt.Sprintf("/api/v2/rootcas/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs), // TODO: Cleanup #ADCSFeatureFlag after full launch.

		// EnterpriseCA Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/enterprisecas/{%s}", api.URIPathVariableObjectID), resources.GetEnterpriseCAEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs),           // TODO: Cleanup #ADCSFeatureFlag after full launch.
		routerInst.GET(fmt.Sprintf("/api/v2/enterprisecas/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs), // TODO: Cleanup #ADCSFeatureFlag after full launch.

		// NTAuthStore Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/ntauthstores/{%s}", api.URIPathVariableObjectID), resources.GetNTAuthStoreEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs),            // TODO: Cleanup #ADCSFeatureFlag after full launch.
		routerInst.GET(fmt.Sprintf("/api/v2/ntauthstores/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs), // TODO: Cleanup #ADCSFeatureFlag after full launch.

		// CertTemplate Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/certtemplates/{%s}", api.URIPathVariableObjectID), resources.GetCertTemplateEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs),           // TODO: Cleanup #ADCSFeatureFlag after full launch.
		routerInst.GET(fmt.Sprintf("/api/v2/certtemplates/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead).CheckFeatureFlag(resources.DB, appcfg.FeatureAdcs), // TODO: Cleanup #ADCSFeatureFlag after full launch.

		// OU Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}", api.URIPathVariableObjectID), resources.GetOUEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/gpos", api.URIPathVariableObjectID), resources.ListADEntityLinkedGPOs).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/users", api.URIPathVariableObjectID), resources.ListADOUContainedUsers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/groups", api.URIPathVariableObjectID), resources.ListADOUContainedGroups).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/ous/{%s}/computers", api.URIPathVariableObjectID), resources.ListADOUContainedComputers).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// User Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}", api.URIPathVariableObjectID), resources.GetUserEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/sessions", api.URIPathVariableObjectID), resources.ListADUserSessions).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/memberships", api.URIPathVariableObjectID), resources.ListADGroupMembership).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/admin-rights", api.URIPathVariableObjectID), resources.ListADEntityAdminRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/rdp-rights", api.URIPathVariableObjectID), resources.ListADEntityRDPRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/dcom-rights", api.URIPathVariableObjectID), resources.ListADEntityDCOMRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/ps-remote-rights", api.URIPathVariableObjectID), resources.ListADEntityPSRemoteRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/sql-admin-rights", api.URIPathVariableObjectID), resources.ListADUserSQLAdminRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/constrained-delegation-rights", api.URIPathVariableObjectID), resources.ListADEntityConstrainedDelegationRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/users/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Group Entity API
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}", api.URIPathVariableObjectID), resources.GetGroupEntityInfo).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/sessions", api.URIPathVariableObjectID), resources.ListADGroupSessions).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/members", api.URIPathVariableObjectID), resources.ListADGroupMembers).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/memberships", api.URIPathVariableObjectID), resources.ListADGroupMembership).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/admin-rights", api.URIPathVariableObjectID), resources.ListADEntityAdminRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/rdp-rights", api.URIPathVariableObjectID), resources.ListADEntityRDPRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/dcom-rights", api.URIPathVariableObjectID), resources.ListADEntityDCOMRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/ps-remote-rights", api.URIPathVariableObjectID), resources.ListADEntityPSRemoteRights).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/controllables", api.URIPathVariableObjectID), resources.ListADEntityControllables).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/groups/{%s}/controllers", api.URIPathVariableObjectID), resources.ListADEntityControllers).RequirePermissions(resources.DB, permissions.GraphDBRead),

		//Data Quality Stats API
		routerInst.GET(fmt.Sprintf("/api/v2/ad-domains/{%s}/data-quality-stats", api.URIPathVariableDomainID), resources.GetADDataQualityStats).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/azure-tenants/{%s}/data-quality-stats", api.URIPathVariableTenantID), resources.GetAzureDataQualityStats).RequirePermissions(resources.DB, permissions.GraphDBRead),
		routerInst.GET(fmt.Sprintf("/api/v2/platform/{%s}/data-quality-stats", api.URIPathVariablePlatformID), resources.GetPlatformAggregateStats).RequirePermissions(resources.DB, permissions.GraphDBRead),

		// Datapipe API
		routerInst.GET("/api/v2/datapipe/status", resources.GetDatapipeStatus).RequireAuth(resources.DB),
		//TODO: Update the permission on this once we get something more concrete
		routerInst.PUT("/api/v2/analysis", resources.RequestAnalysis).RequirePermissions(resources.DB, permissions.GraphDBWrite),
	)
}
