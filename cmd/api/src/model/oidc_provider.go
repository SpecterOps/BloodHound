// Copyright 2024 Specter Ops, Inc.
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

package model

// OIDCProvider contains the data needed to initiate an OIDC secure login flow
type OIDCProvider struct {
	ClientID      string `json:"client_id"`
	Issuer        string `json:"issuer"`
	SSOProviderID int    `json:"sso_provider_id"`

	Serial
}

func (OIDCProvider) TableName() string {
	return "oidc_providers"
}

func (s OIDCProvider) AuditData() AuditData {
	return AuditData{
		"id":              s.ID,
		"client_id":       s.ClientID,
		"issuer":          s.Issuer,
		"sso_provider_id": s.SSOProviderID,
	}
}
