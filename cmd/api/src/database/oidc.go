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

package database

import (
	"context"

	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

const (
	oidcProvidersTableName = "oidc_providers"
)

// OIDCProviderData defines the interface required to interact with the oidc_providers table
type OIDCProviderData interface {
	CreateOIDCProvider(ctx context.Context, name, slug, issuer, clientID string) (model.OIDCProvider, error)
}

// CreateOIDCProvider creates a new entry for an OIDC provider
func (s *BloodhoundDB) CreateOIDCProvider(ctx context.Context, name, slug, issuer, clientID string) (model.OIDCProvider, error) {
	oidcProvider := model.OIDCProvider{
		Name:     name,
		ClientID: clientID,
		Issuer:   issuer,
	}

	// Create both the sso_providers and oidc_providers rows in a single transaction
	// If one of these requests errors, both changes will be rolled back
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if ssoProvider, err := s.CreateSSOProviderWithTransaction(ctx, tx, name, slug, model.SessionAuthProviderOIDC); err != nil {
			return err
		} else {
			oidcProvider.SSOProviderID = int(ssoProvider.ID)
			return CheckError(tx.WithContext(ctx).Table(oidcProvidersTableName).Create(&oidcProvider))
		}
	})

	return oidcProvider, err
}
