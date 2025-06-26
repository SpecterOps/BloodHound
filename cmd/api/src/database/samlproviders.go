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

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

const (
	samlProvidersTableName = "saml_providers"
)

// SAMLProviderData defines the interface required to interact with the oidc_providers table
type SAMLProviderData interface {
	CreateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider, config model.SSOProviderConfig) (model.SAMLProvider, error)
	GetAllSAMLProviders(ctx context.Context) (model.SAMLProviders, error)
	GetSAMLProvider(ctx context.Context, id int32) (model.SAMLProvider, error)
	GetSAMLProviderUsers(ctx context.Context, id int32) (model.Users, error)
	UpdateSAMLIdentityProvider(ctx context.Context, ssoProvider model.SSOProvider) (model.SAMLProvider, error)
}

// CreateSAMLIdentityProvider creates a new saml_providers row using the data in the input struct
// This also creates the corresponding sso_provider entry
// INSERT INTO saml_identity_providers (...) VALUES (...)
func (s *BloodhoundDB) CreateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider, config model.SSOProviderConfig) (model.SAMLProvider, error) {
	// Set the current version for root_uri_version
	samlProvider.RootURIVersion = model.SAMLRootURIVersion2

	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionCreateSAMLIdentityProvider,
		Model:  &samlProvider, // Pointer is required to ensure success log contains updated fields after transaction
	}

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		// Create the associated SSO provider
		if ssoProvider, err := bhdb.CreateSSOProvider(ctx, samlProvider.Name, model.SessionAuthProviderSAML, config); err != nil {
			return err
		} else {
			samlProvider.SSOProviderID = null.Int32From(ssoProvider.ID)
			return CheckError(tx.WithContext(ctx).Create(&samlProvider))
		}
	})

	return samlProvider, err
}

// GetAllSAMLProviders returns all SAML providers
// SELECT * FROM saml_providers
func (s *BloodhoundDB) GetAllSAMLProviders(ctx context.Context) (model.SAMLProviders, error) {
	var (
		samlProviders model.SAMLProviders
		result        = s.db.WithContext(ctx).Find(&samlProviders)
	)

	return samlProviders, CheckError(result)
}

// GetSAMLProvider returns a SAML provider corresponding to the ID provided
// SELECT * FOM saml_providers WHERE id = ..
func (s *BloodhoundDB) GetSAMLProvider(ctx context.Context, id int32) (model.SAMLProvider, error) {
	var (
		samlProvider model.SAMLProvider
		result       = s.db.WithContext(ctx).First(&samlProvider, id)
	)

	return samlProvider, CheckError(result)
}

// GetSAMLProviderUsers returns all users that are bound to the SAML provider ID provided
// SELECT * FROM users WHERE saml_provider_id = ..
func (s *BloodhoundDB) GetSAMLProviderUsers(ctx context.Context, id int32) (model.Users, error) {
	var users model.Users
	return users, CheckError(s.preload(model.UserAssociations()).WithContext(ctx).Where("sso_provider_id = ?", id).Find(&users))
}

// CreateSAMLProvider updates a saml_providers row using the data in the input struct
// UPDATE saml_identity_providers SET (...) VALUES (...) WHERE id = ...
func (s *BloodhoundDB) UpdateSAMLIdentityProvider(ctx context.Context, ssoProvider model.SSOProvider) (model.SAMLProvider, error) {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionUpdateSAMLIdentityProvider,
		Model:  ssoProvider.SAMLProvider, // Pointer is required to ensure success log contains updated fields after transaction
	}

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if _, err := bhdb.UpdateSSOProvider(ctx, ssoProvider); err != nil {
			return err
		} else if err := CheckError(tx.WithContext(ctx).Exec(
			fmt.Sprintf("UPDATE %s SET name = ?, display_name = ?, issuer_uri = ?, single_sign_on_uri = ?, metadata_xml = ?, updated_at = ? WHERE id = ?;", samlProvidersTableName),
			ssoProvider.SAMLProvider.Name, ssoProvider.SAMLProvider.DisplayName, ssoProvider.SAMLProvider.IssuerURI, ssoProvider.SAMLProvider.SingleSignOnURI, ssoProvider.SAMLProvider.MetadataXML, time.Now().UTC(), ssoProvider.SAMLProvider.ID),
		); err != nil {
			return err
		} else {
			// Ensure all existing sessions are invalidated within the tx
			return bhdb.TerminateUserSessionsBySSOProvider(ctx, ssoProvider)
		}
	})

	return *ssoProvider.SAMLProvider, err
}
