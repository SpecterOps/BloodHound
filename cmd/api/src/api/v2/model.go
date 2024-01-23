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

package v2

import (
	"github.com/gorilla/schema"
	"github.com/specterops/bloodhound/cache"
	_ "github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"github.com/specterops/bloodhound/src/serde"
)

type ListPermissionsResponse struct {
	Permissions model.Permissions `json:"permissions"`
}

type ListRolesResponse struct {
	Roles model.Roles `json:"roles"`
}

type ListUsersResponse struct {
	Users model.Users `json:"users"`
}

type ListTokensResponse struct {
	Tokens model.AuthTokens `json:"tokens"`
}

type SAMLSignOnEndpoint struct {
	Name          string    `json:"name"`
	InitiationURL serde.URL `json:"initiation_url"`
}

type ListSAMLSignOnEndpointsResponse struct {
	Endpoints []SAMLSignOnEndpoint `json:"endpoints"`
}

type ListSAMLProvidersResponse struct {
	SAMLProviders model.SAMLProviders `json:"saml_providers"`
}

type UpdateUserRequest struct {
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	EmailAddress   string  `json:"email_address"`
	Principal      string  `json:"principal"`
	Roles          []int32 `json:"roles"`
	SAMLProviderID string  `json:"saml_provider_id"`
	IsDisabled     bool    `json:"is_disabled"`
}

type CreateUserRequest struct {
	UpdateUserRequest
	SetUserSecretRequest
}

type DeleteSAMLProviderResponse struct {
	AffectedUsers model.Users `json:"affected_users"`
}

type SetUserSecretRequest struct {
	Secret             string `json:"secret" validate:"password,length=12,lower=1,upper=1,special=1,numeric=1"`
	NeedsPasswordReset bool   `json:"needs_password_reset"`
}

type CreateUserToken struct {
	TokenName string `json:"token_name"`
	UserID    string `json:"user_id"`
}

type CreateSAMLAuthProviderRequest struct {
	Name                       string   `json:"name"`
	DisplayName                string   `json:"display_name"`
	SigningCertificate         string   `json:"signing_certificate"`
	IssuerURI                  string   `json:"issuer_uri"`
	SingleSignOnURI            string   `json:"single_signon_uri"`
	PrincipalAttributeMappings []string `json:"principal_attribute_mappings"`
}

type UpdateSAMLAuthProviderRequest struct {
	Name                       string   `json:"name"`
	DisplayName                string   `json:"display_name"`
	SigningCertificate         string   `json:"signing_certificate"`
	IssuerURI                  string   `json:"issuer_uri"`
	SingleSignOnURI            string   `json:"single_signon_uri"`
	PrincipalAttributeMappings []string `json:"principal_attribute_mappings"`
}

type SecretInitializationRequest struct {
	AdminEmailAddress string `json:"admin_email_address"`
	Secret            string `json:"secret"`
}

type IDPValidationResponse struct {
	ErrorMessage string `json:"error_message"`
	Successful   bool   `json:"successful"`
}

type PagedNodeListEntry struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	DistinguishedName string `json:"distinguished_name"`
	ObjectID          string `json:"object_id"`
}

type SAMLInitializationRequest struct {
	AdminEmailAddress            string `json:"admin_email_address"`
	IdentityProviderProviderName string `json:"idp_name"`
	IdentityProviderURL          string `json:"idp_url"`
	ServiceProviderCertificate   string `json:"sp_certificate"`
	ServiceProviderKey           string `json:"sp_private_key"`
}

// Resources holds the database and configuration dependencies to be passed around the API functions
type Resources struct {
	Decoder                    *schema.Decoder
	DB                         database.Database
	Graph                      graph.Database // TODO: to be phased out in favor of graph queries
	GraphQuery                 queries.Graph
	Config                     config.Configuration
	QueryParameterFilterParser model.QueryParameterFilterParser
	Cache                      cache.Cache
	CollectorManifests         config.CollectorManifests
	TaskNotifier               datapipe.Tasker
}

func NewResources(
	rdms database.Database,
	graphDB *graph.DatabaseSwitch,
	cfg config.Configuration,
	apiCache cache.Cache,
	graphQuery queries.Graph,
	collectorManifests config.CollectorManifests,
	taskNotifier datapipe.Tasker,
) Resources {
	return Resources{
		Decoder:                    schema.NewDecoder(),
		DB:                         rdms,
		Graph:                      graphDB, // TODO: to be phased out in favor of graph queries
		GraphQuery:                 graphQuery,
		Config:                     cfg,
		QueryParameterFilterParser: model.NewQueryParameterFilterParser(),
		Cache:                      apiCache,
		CollectorManifests:         collectorManifests,
		TaskNotifier:               taskNotifier,
	}
}
