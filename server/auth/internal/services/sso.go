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

package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/specterops/bloodhound/packages/go/params"
)

var ErrInvalidProviderQuery = errors.New("invalid SSO provider query")

// ProviderType identifies the authentication protocol used by an SSO provider
type ProviderType int

const (
	ProviderSAML ProviderType = 1
	ProviderOIDC ProviderType = 2
)

func (s ProviderType) String() string {
	switch s {
	case ProviderSAML:
		return "SAML"
	case ProviderOIDC:
		return "OIDC"
	default:
		return "Unknown"
	}
}

type SSOProvider struct {
	ID           int32
	Name         string
	Type         ProviderType
	Slug         string
	Config       ProviderConfig
	OIDCProvider *OIDCProvider
	SAMLProvider *SAMLProvider
}

type ProviderConfig struct {
	AutoProvision AutoProvisionConfig
}

type AutoProvisionConfig struct {
	Enabled       bool
	DefaultRoleID int32
	RoleProvision bool
}

type ProviderTimestamps struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}

type OIDCProvider struct {
	ClientID      string
	Issuer        string
	SSOProviderID int
	ProviderTimestamps
}

type SAMLProvider struct {
	Name                       string
	DisplayName                string
	IssuerURI                  string
	SingleSignOnURI            string
	RootURIVersion             int
	PrincipalAttributeMappings []string
	SSOProviderID              *int32
	ProviderTimestamps
}

func (s *Service) ListAuthProviders(ctx context.Context, filters params.Filters, sortItems params.SortItems) ([]SSOProvider, error) {
	return s.db.ListAuthProviders(ctx, filters, sortItems)
}
