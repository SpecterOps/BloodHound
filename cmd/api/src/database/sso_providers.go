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
	"strings"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"gorm.io/gorm"
)

const (
	ssoProviderTableName = "sso_providers"
)

// SSOProviderData defines the methods required to interact with the sso_providers table
type SSOProviderData interface {
	CreateSSOProvider(ctx context.Context, name string, authProvider model.SessionAuthProvider) (model.SSOProvider, error)
	DeleteSSOProvider(ctx context.Context, id int) error
	GetAllSSOProviders(ctx context.Context, order string, sqlFilter model.SQLFilter) ([]model.SSOProvider, error)
	GetSSOProviderById(ctx context.Context, id int32) (model.SSOProvider, error)
	GetSSOProviderBySlug(ctx context.Context, slug string) (model.SSOProvider, error)
	GetSSOProviderUsers(ctx context.Context, id int) (model.Users, error)
}

// CreateSSOProvider creates an entry in the sso_providers table
// A slug will be created for the SSO Provider using the name argument as a base. The name will be lower cased and all spaces are replaced with `-`
func (s *BloodhoundDB) CreateSSOProvider(ctx context.Context, name string, authProvider model.SessionAuthProvider) (model.SSOProvider, error) {
	var (
		provider = model.SSOProvider{
			Name: name,
			Slug: strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			Type: authProvider,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateSSOIdentityProvider,
			Model:  &provider,
		}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Table(ssoProviderTableName).Create(&provider))
	})

	return provider, err
}

// DeleteSSOProvider deletes a sso_provider entry with a matching id
func (s *BloodhoundDB) DeleteSSOProvider(ctx context.Context, id int) error {
	var (
		ssoProvider = model.SSOProvider{}
		auditEntry  = model.AuditEntry{
			Action: model.AuditLogActionDeleteSSOIdentityProvider,
			Model:  &ssoProvider,
		}
	)

	// Populate the OIDCProvider and SAMLProvider fields, used in AuditData to log the details of the provider based on its type
	if result := s.db.Preload("OIDCProvider").Preload("SAMLProvider").
		Table(ssoProviderTableName).
		Where("id = ?", id).
		First(&ssoProvider); result.Error != nil {
		return CheckError(result)
	}

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		if result := tx.Table(ssoProviderTableName).Delete(&ssoProvider); result.RowsAffected == 0 {
			return ErrNotFound
		} else {
			return CheckError(result)
		}
	})

	return err
}

func (s *BloodhoundDB) GetAllSSOProviders(ctx context.Context, order string, sqlFilter model.SQLFilter) ([]model.SSOProvider, error) {
	var providers []model.SSOProvider

	query := s.db.WithContext(ctx).Model(&model.SSOProvider{})

	// Apply SQL filter if provided
	if sqlFilter.SQLString != "" {
		query = query.Where(sqlFilter.SQLString, sqlFilter.Params...)
	}

	// Backwards compatibility when FF is disabled
	if oidcFeatureFlag, err := s.GetFlagByKey(ctx, appcfg.FeatureOIDCSupport); err != nil {
		return providers, err
	} else if !oidcFeatureFlag.Enabled {
		query = query.Where("type = ?", model.SessionAuthProviderSAML)
	}

	// Apply sorting order if provided
	if order != "" {
		query = query.Order(order)
	} else {
		// Default ordering by created_at if no order is specified
		query = query.Order("created_at")
	}

	// Preload the associated OIDC and SAML providers
	query = query.Preload("OIDCProvider").Preload("SAMLProvider")

	result := query.Find(&providers)
	return providers, CheckError(result)
}

func (s *BloodhoundDB) GetSSOProviderBySlug(ctx context.Context, slug string) (model.SSOProvider, error) {
	var provider model.SSOProvider
	result := s.db.WithContext(ctx).Preload("OIDCProvider").Preload("SAMLProvider").Where("slug = ?", slug).Find(&provider)

	return provider, CheckError(result)
}

// GetSSOProviderUsers returns all the users associated with a given sso provider
func (s *BloodhoundDB) GetSSOProviderUsers(ctx context.Context, id int) (model.Users, error) {
	var (
		users model.Users
	)

	return users, CheckError(s.db.WithContext(ctx).Table("users").Where("sso_provider_id = ?", id).Find(&users))
}

func (s *BloodhoundDB) GetSSOProviderById(ctx context.Context, id int32) (model.SSOProvider, error) {
	var provider model.SSOProvider
	result := s.db.WithContext(ctx).Preload("OIDCProvider").Preload("SAMLProvider").Table(ssoProviderTableName).Where("id = ?", id).First(&provider)

	return provider, CheckError(result)
}
