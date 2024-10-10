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
	GetOIDCProvider(ctx context.Context, id int64) (model.OIDCProvider, error)
	CreateOIDCProvider(ctx context.Context, name, issuer, clientID string) (model.OIDCProvider, error)
	DeleteOIDCProvider(ctx context.Context, id int64) error
}

func (s *BloodhoundDB) GetOIDCProvider(ctx context.Context, id int64) (model.OIDCProvider, error) {
	var (
		oidcProvider model.OIDCProvider
		result       = s.db.WithContext(ctx).Table("oidc_providers").Where("id = ?", id).First(&oidcProvider)
	)
	return oidcProvider, CheckError(result)
}

// CreateOIDCProvider creates a new entry for an OIDC provider as well as the associated SSO provider
func (s *BloodhoundDB) CreateOIDCProvider(ctx context.Context, name, issuer, clientID string) (model.OIDCProvider, error) {
	var (
		oidcProvider = model.OIDCProvider{
			ClientID: clientID,
			Issuer:   issuer,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateOIDCIdentityProvider,
			Model:  &oidcProvider, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	// Create both the sso_providers and oidc_providers rows in a single transaction
	// If one of these requests errors, both changes will be rolled back
	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if ssoProvider, err := bhdb.CreateSSOProvider(ctx, name, model.SessionAuthProviderOIDC); err != nil {
			return err
		} else {
			oidcProvider.SSOProviderID = int(ssoProvider.ID)
			return CheckError(tx.WithContext(ctx).Table(oidcProvidersTableName).Create(&oidcProvider))
		}
	})

	return oidcProvider, err
}

// DeleteOIDCProvider deletes a oidc_provider matching the given id
func (s *BloodhoundDB) DeleteOIDCProvider(ctx context.Context, id int64) error {
	result := s.db.WithContext(ctx).Table("oidc_providers").Where("id = ?", id).Delete(&model.OIDCProvider{})
	if result.RowsAffected == 0 {
		return ErrNotFound
	} else {
		return CheckError(result)
	}
}
