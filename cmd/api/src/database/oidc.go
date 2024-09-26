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
	)
	result := s.db.WithContext(ctx).Table("oidc_providers").Where("id = ?", id).Scan(&oidcProvider)
	if result.RowsAffected == 0 {
		return model.OIDCProvider{}, ErrNotFound
	}
	return oidcProvider, CheckError(result)
	//return oidcProvider, CheckError(s.db.WithContext(ctx).Raw("SELECT * FROM oidc_providers WHERE id = ?", id).Scan(&oidcProvider))
}

// CreateOIDCProvider creates a new entry for an OIDC provider
func (s *BloodhoundDB) CreateOIDCProvider(ctx context.Context, name, issuer, clientID string) (model.OIDCProvider, error) {
	provider := model.OIDCProvider{
		Name:     name,
		ClientID: clientID,
		Issuer:   issuer,
	}

	return provider, CheckError(s.db.WithContext(ctx).Table("oidc_providers").Create(&provider))
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
