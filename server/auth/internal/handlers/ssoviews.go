// Copyright 2026 Specter Ops, Inc.
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

package handlers

import (
	"database/sql"
	"time"
)

type ProviderConfigView struct {
	AutoProvision AutoProvisionConfigView `json:"auto_provision"`
}

type AutoProvisionConfigView struct {
	Enabled       bool  `json:"enabled"`
	DefaultRoleID int32 `json:"default_role_id"`
	RoleProvision bool  `json:"role_provision"`
}

type ProviderTimestampsView struct {
	ID        int32        `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

type OIDCProviderView struct {
	ClientID      string `json:"client_id"`
	Issuer        string `json:"issuer"`
	SSOProviderID int    `json:"sso_provider_id"`
	ProviderTimestampsView
}

type SAMLProviderDetailsView struct {
	Name                       string   `json:"name"`
	DisplayName                string   `json:"display_name"`
	IssuerURI                  string   `json:"idp_issuer_uri"`
	SingleSignOnURI            string   `json:"idp_sso_uri"`
	RootURIVersion             int      `json:"root_uri_version"`
	PrincipalAttributeMappings []string `json:"principal_attribute_mappings"`
	SSOProviderID              *int32   `json:"sso_provider_id"`
	ProviderTimestampsView
}
