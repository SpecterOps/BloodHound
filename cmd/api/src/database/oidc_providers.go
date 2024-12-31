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
	"fmt"
	"time"

	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

const (
	oidcProvidersTableName = "oidc_providers"
)

// OIDCProviderData defines the interface required to interact with the oidc_providers table
type OIDCProviderData interface {
	CreateOIDCProvider(ctx context.Context, name, issuer, clientID string, config model.SSOProviderConfig) (model.OIDCProvider, error)
	UpdateOIDCProvider(ctx context.Context, ssoProvider model.SSOProvider) (model.OIDCProvider, error)
}

// CreateOIDCProvider creates a new entry for an OIDC provider as well as the associated SSO provider
func (s *BloodhoundDB) CreateOIDCProvider(ctx context.Context, name, issuer, clientID string, config model.SSOProviderConfig) (model.OIDCProvider, error) {
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

	// If we have a disabled autoprovision, wipe the auto provision config
	if !config.AutoProvision.Enabled {
		config.AutoProvision = model.SSOProviderAutoProvisionConfig{}
	}

	// Create both the sso_providers and oidc_providers rows in a single transaction
	// If one of these requests errors, both changes will be rolled back
	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if ssoProvider, err := bhdb.CreateSSOProvider(ctx, name, model.SessionAuthProviderOIDC, config); err != nil {
			return err
		} else {
			oidcProvider.SSOProviderID = int(ssoProvider.ID)
			return CheckError(tx.WithContext(ctx).Table(oidcProvidersTableName).Create(&oidcProvider))
		}
	})

	return oidcProvider, err
}

// UpdateOIDCProvider updates an OIDC provider as well as the associated SSO provider
func (s *BloodhoundDB) UpdateOIDCProvider(ctx context.Context, ssoProvider model.SSOProvider) (model.OIDCProvider, error) {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionUpdateOIDCIdentityProvider,
		Model:  ssoProvider.OIDCProvider, // Pointer is required to ensure success log contains updated fields after transaction
	}

	// update both the sso_providers, oidc_providers, and user_sessions rows in a single transaction
	// If one of these requests errors, all changes will be rolled back
	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if _, err := bhdb.UpdateSSOProvider(ctx, ssoProvider); err != nil {
			return err
		} else if err := CheckError(tx.WithContext(ctx).Exec(fmt.Sprintf("UPDATE %s SET client_id = ?, issuer = ?, updated_at = ? WHERE id = ?;", oidcProvidersTableName),
			ssoProvider.OIDCProvider.ClientID, ssoProvider.OIDCProvider.Issuer, time.Now().UTC(), ssoProvider.OIDCProvider.ID)); err != nil {
			return err
		} else {
			// Ensure all existing sessions are invalidated within the tx
			return bhdb.TerminateUserSessionsBySSOProvider(ctx, ssoProvider)
		}
	})

	return *ssoProvider.OIDCProvider, err
}
