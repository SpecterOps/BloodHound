// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
)

type Kind interface {
	GetKindByName(ctx context.Context, name string) (model.Kind, error)
	GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error)
}

func (s *BloodhoundDB) GetKindByName(ctx context.Context, name string) (model.Kind, error) {
	const query = `
		SELECT id, name
		FROM kind
		WHERE name = $1;
	`

	var kind model.Kind
	result := s.db.WithContext(ctx).Raw(query, name).Scan(&kind)

	if result.Error != nil {
		return model.Kind{}, result.Error
	}

	if result.RowsAffected == 0 || kind.ID == 0 {
		return model.Kind{}, ErrNotFound
	}

	return kind, nil
}

func (s *BloodhoundDB) GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error) {
	if len(ids) == 0 {
		return []model.Kind{}, nil
	}

	// Dedupe IDs so the length check against query results doesn't produce a
	// false-positive ErrNotFound when callers pass duplicate values.
	uniqueIDs := utils.Dedupe(ids)

	query := `
		SELECT id, name
		FROM kind
		WHERE id IN (?)
		ORDER BY id;
	`

	var kinds []model.Kind
	result := s.db.WithContext(ctx).Raw(query, uniqueIDs).Scan(&kinds)

	if err := result.Error; err != nil {
		return nil, fmt.Errorf("failed to fetch kinds by IDs: %w", err)
	}

	if len(kinds) != len(uniqueIDs) {
		return nil, ErrNotFound
	}

	return kinds, nil
}
