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

// Package assetgroups is the wireup module for the asset groups feature. It composes
// the store and service so other feature slices can obtain a ready-to-use adapter
// without reaching into the persistence layer.
package assetgroups

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/assetgroups/internal/appdb"
	"github.com/specterops/bloodhound/server/assetgroups/internal/services"
)

// AssetGroupTag is the domain representation of an asset group tag exposed to other
// feature slices. It aliases the internal services type so callers need not import the
// internal package.
type AssetGroupTag = services.AssetGroupTag

// AssetGroupsRequestAdapter is the exported asset groups capability other feature slices
// depend on. It is satisfied by the internal service constructed by
// NewAssetGroupsRequestAdapter.
type AssetGroupsRequestAdapter interface {
	ResolveTagIDsWithFallback(ctx context.Context, maybeAssetGroupTagID string) ([]int, error)
	GetTierZeroTag(ctx context.Context) (AssetGroupTag, error)
}

// NewAssetGroupsRequestAdapter builds a ready-to-use asset groups adapter backed by the
// pgx pool, wiring the store and service together so callers obtain the service without
// reaching into the persistence layer.
func NewAssetGroupsRequestAdapter(pool *pgxpool.Pool) AssetGroupsRequestAdapter {
	return services.NewService(appdb.NewStore(pool))
}
