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

package appdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/server/assetgrouptags/internal/services"
)

const (
	// assetGroupTagTypeTier identifies tag rows that represent a tier.
	assetGroupTagTypeTier = 1
	// tierZeroPosition is the position of the tier zero tag among the tiers.
	tierZeroPosition = 1
)

// tagRow is the package-local DB row type for an asset group tag. db: tags drive
// pgx.RowToStructByName scanning.
type tagRow struct {
	ID int `db:"id"`
}

// toAssetGroupTag translates a raw tag row into the domain model.
func toAssetGroupTag(row tagRow) services.AssetGroupTag {
	return services.AssetGroupTag{ID: row.ID}
}

// GetAssetGroupTagByID fetches an asset group tag by its id, returning
// ErrAssetGroupTagNotFound when no matching, non-deleted tag exists.
func (s *Store) GetAssetGroupTagByID(ctx context.Context, id int) (services.AssetGroupTag, error) {
	var query = fmt.Sprintf("SELECT id FROM %s WHERE id = $1 AND deleted_at IS NULL", tableAssetGroupTags)

	rows, err := s.db.Query(ctx, query, id)
	if err != nil {
		return services.AssetGroupTag{}, err
	}
	defer rows.Close()

	if row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[tagRow]); errors.Is(err, pgx.ErrNoRows) {
		return services.AssetGroupTag{}, services.ErrAssetGroupTagNotFound
	} else if err != nil {
		return services.AssetGroupTag{}, fmt.Errorf("reading rows: %w", err)
	} else {
		return toAssetGroupTag(row), nil
	}
}

// GetTierZeroTag fetches the tier zero asset group tag, returning
// ErrTierZeroTagNotFound when none is configured.
func (s *Store) GetTierZeroTag(ctx context.Context) (services.AssetGroupTag, error) {
	var query = fmt.Sprintf("SELECT id FROM %s WHERE deleted_at IS NULL AND type = $1 AND position = $2 ORDER BY name ASC", tableAssetGroupTags)

	rows, err := s.db.Query(ctx, query, assetGroupTagTypeTier, tierZeroPosition)
	if err != nil {
		return services.AssetGroupTag{}, err
	}
	defer rows.Close()

	if tagRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[tagRow]); err != nil {
		return services.AssetGroupTag{}, fmt.Errorf("reading rows: %w", err)
	} else if len(tagRows) == 0 {
		return services.AssetGroupTag{}, services.ErrTierZeroTagNotFound
	} else {
		return toAssetGroupTag(tagRows[0]), nil
	}
}
