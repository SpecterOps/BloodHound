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
	CreateOIDCProvider(ctx context.Context, name, authURL, tokenURL, clientID string) (model.OIDCProvider, error)
}

// CreateOIDCProvider creates a new entry for an OIDC provider
func (s *BloodhoundDB) CreateOIDCProvider(ctx context.Context, name, authURL, tokenURL, clientID string) (model.OIDCProvider, error) {
	provider := model.OIDCProvider{
		Name:     name,
		ClientID: clientID,
		AuthURL:  authURL,
		TokenURL: tokenURL,
	}

	return provider, CheckError(s.db.WithContext(ctx).Table("oidc_providers").Create(&provider))
}
