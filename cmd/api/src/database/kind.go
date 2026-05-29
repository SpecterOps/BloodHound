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
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
)

type Kind interface {
	GetKindsByNames(ctx context.Context, names ...string) ([]model.Kind, error)
	GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error)
	UpsertKind(ctx context.Context, name string) (model.Kind, error)
}

func (s *BloodhoundDB) GetKindsByNames(ctx context.Context, names ...string) ([]model.Kind, error) {
	var (
		uniqueNames = utils.Dedupe(names)
		query       = `SELECT id, name FROM kind WHERE name in (?);`
	)

	var kinds []model.Kind
	if result := s.db.WithContext(ctx).Raw(query, uniqueNames).Scan(&kinds); result.Error != nil {
		return kinds, result.Error
	} else if len(kinds) != len(uniqueNames) {
		return kinds, ErrNotFound
	}

	return kinds, nil
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
		return kinds, fmt.Errorf("failed to fetch kinds by IDs: %w", err)
	}

	if len(kinds) != len(uniqueIDs) {
		return kinds, ErrNotFound
	}

	return kinds, nil
}

// UpsertKind Since this function inserts into the kinds table, the business logic calling this func
// must also call the DAWGS RefreshKinds function to ensure the kinds are reloaded into the in memory kind map.
func (s *BloodhoundDB) UpsertKind(ctx context.Context, name string) (model.Kind, error) {
	var kind model.Kind
	if name == "" {
		return kind, errors.New("invalid kind name")
	}

	result := s.db.WithContext(ctx).Raw("SELECT id, name FROM upsert_kind(?)", name).Scan(&kind)
	return kind, CheckError(result)
}
