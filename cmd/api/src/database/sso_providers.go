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
	"gorm.io/gorm"
)

const (
	ssoProviderTableName = "sso_providers"
)

// SSOProviderData defines the methods required to interact with the sso_providers table
type SSOProviderData interface {
	GetSSOProvider(ctx context.Context, id int) (model.SSOProvider, error)
	CreateSSOProvider(ctx context.Context, name string, authProvider model.SessionAuthProvider) (model.SSOProvider, error)
	DeleteSSOProvider(ctx context.Context, id int) error
}

// GetSSOProvider gets an SSO Provider in the sso_providers table by the id given
func (s *BloodhoundDB) GetSSOProvider(ctx context.Context, id int) (model.SSOProvider, error) {
	var provider model.SSOProvider

	return provider, CheckError(s.db.WithContext(ctx).Table(ssoProviderTableName).Where("id = ?", id).First(&provider))
}

// CreateSSOProvider creates an entry in the sso_providers table
// A slug will be created for the SSO Provider using the name argument as a base. The name will be lower cased and all spaces are replaced with `-`
func (s *BloodhoundDB) CreateSSOProvider(ctx context.Context, name string, authProvider model.SessionAuthProvider) (model.SSOProvider, error) {
	provider := model.SSOProvider{
		Name: name,
		Slug: strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		Type: authProvider,
	}

	return provider, CheckError(s.db.WithContext(ctx).Table(ssoProviderTableName).Create(&provider))
}

// DeleteSSOProvider deletes a sso_provider entry with a matching id
func (s *BloodhoundDB) DeleteSSOProvider(ctx context.Context, id int) error {
	var (
		ssoProvider = model.SSOProvider{
			Serial: model.Serial{ID: int32(id)},
		}
		auditEntry = model.AuditEntry{
			Action: "DeleteSSOProvider",
			Model:  &ssoProvider}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Table(ssoProviderTableName).Where("id = ?", id).Delete(&model.SSOProvider{}))
	})

	return err
}
