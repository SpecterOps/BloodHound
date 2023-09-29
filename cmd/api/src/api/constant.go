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

package api

const (
	// 32 Bytes or 256 Bits for HMAC-SHA2-256
	JWTSigningKeyByteLength = 32

	// Cookie Keys
	AuthTokenCookieName = "token"

	// UserInterfacePath is the static path to the UI landing page
	UserInterfacePath = "/ui"
	UserDisabledPath  = "/ui/user-disabled"

	// Authorization schemes
	AuthorizationSchemeBHESignature = "bhesignature"
	AuthorizationSchemeBearer       = "bearer"

	// Query parameters
	QueryParameterSortBy         = "sort_by"
	QueryParameterHydrateCounts  = "counts"
	QueryParameterHydrateDomains = "hydrate_domains"
	QueryParameterHydrateOUs     = "hydrate_ous"

	// URI path parameters
	URIPathVariableApplicationConfigurationParameter = "parameter"
	URIPathVariableAssetGroupID                      = "asset_group_id"
	URIPathVariableAssetGroupSelectorID              = "asset_group_selector_id"
	URIPathVariableAttackPathID                      = "attack_path_id"
	URIPathVariableClientID                          = "client_id"
	URIPathVariableDomainID                          = "domain_id"
	URIPathVariableEventID                           = "event_id"
	URIPathVariableFeatureID                         = "feature_id"
	URIPathVariableJobID                             = "job_id"
	URIPathVariableObjectID                          = "object_id"
	URIPathVariablePermissionID                      = "permission_id"
	URIPathVariablePlatformID                        = "platform_id"
	URIPathVariableRoleID                            = "role_id"
	URIPathVariableSAMLProviderID                    = "saml_provider_id"
	URIPathVariableServiceProviderName               = "saml_provider_name"
	URIPathVariableTaskID                            = "task_id"
	URIPathVariableTenantID                          = "tenant_id"
	URIPathVariableTokenID                           = "token_id"
	URIPathVariableUserID                            = "user_id"
	URIPathVariableSavedQueryID                      = "saved_query_id"
)
