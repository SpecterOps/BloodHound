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
	"strings"

	"github.com/specterops/bloodhound/src/model"
)

const (
	ssoProviderTableName = "sso_providers"
)

var (
	ssoProviderTypeMapping = map[model.SessionAuthProvider]model.SSOProviderType{
		model.SessionAuthProviderOIDC: model.SSOProviderTypeOIDC,
		model.SessionAuthProviderSAML: model.SSOProviderTypeSAML,
	}
)

// SSOProviderData defines the methods required to interact with the sso_providers table
type SSOProviderData interface {
	CreateSSOProvider(ctx context.Context, name string, authType model.SessionAuthProvider) (model.SSOProvider, error)
	GetAllSSOProviders(ctx context.Context, order string, sqlFilter model.SQLFilter) ([]model.SSOProvider, error)
}

// CreateSSOProvider creates an entry in the sso_providers table
// A slug will be created for the SSO Provider using the name argument as a base. The name will be lower cased and all spaces are replaced with `-`
func (s *BloodhoundDB) CreateSSOProvider(ctx context.Context, name string, authType model.SessionAuthProvider) (model.SSOProvider, error) {
	if ssoProviderType, ok := ssoProviderTypeMapping[authType]; !ok {
		return model.SSOProvider{}, fmt.Errorf("error could not find a valid mapping from SessionAuthProvider to SSOProviderType: %d", authType)
	} else {

		provider := model.SSOProvider{
			Name: name,
			Slug: strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			Type: ssoProviderType,
		}

		return provider, CheckError(s.db.WithContext(ctx).Table(ssoProviderTableName).Create(&provider))
	}
}

func (s *BloodhoundDB) GetAllSSOProviders(ctx context.Context, order string, sqlFilter model.SQLFilter) ([]model.SSOProvider, error) {
	var (
		providers []model.SSOProvider
		query     = s.db.WithContext(ctx).Table(ssoProviderTableName)
	)
	// Apply SQL filter if provided
	if sqlFilter.SQLString != "" {
		query = query.Where(sqlFilter.SQLString, sqlFilter.Params...)
	}

	// Apply sorting order if provided
	if order != "" {
		query = query.Order(order)
	} else {
		// Default ordering by created_at if no order is specified
		query = query.Order("created_at")
	}

	result := query.Find(&providers)
	return providers, CheckError(result)
}
