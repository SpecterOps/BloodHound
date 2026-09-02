// Copyright 2025 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/dawgs/graph"
)

type SourceKindsData interface {
	GetSourceKinds(ctx context.Context) ([]model.SourceKind, error)
	DeleteSourceKindsByName(ctx context.Context, name ...string) error
	RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error
	GetSourceKindByName(ctx context.Context, name string) (model.SourceKind, error)
	GetSourceKindsByIDs(ctx context.Context, ids ...int) ([]model.SourceKind, error)
	GetSourceKindByID(ctx context.Context, id int) (model.SourceKind, error)
}

// RegisterSourceKind returns a function that inserts a source kind by name,
// using the provided context. The returned function can be called later with just the name.
// The function is curried in this way because it is primarily used in datapipe during ingest decoding when
// there is no ctx in scope.
//
// Since this function inserts into the kinds table, the business logic calling this func
// must also call the DAWGS RefreshKinds function to ensure the kinds are reloaded into the in memory kind map.
func (s *BloodhoundDB) RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error {
	return func(sourceKind graph.Kind) error {
		if sourceKind == nil || sourceKind == graph.EmptyKind {
			return nil
		}

		const query = "SELECT id, kind_id FROM upsert_source_kind(?)"

		result := s.db.WithContext(ctx).Exec(query, sourceKind)
		if err := result.Error; err != nil {
			return fmt.Errorf("failed to insert source kind %q: %w", sourceKind, err)
		}

		return nil
	}
}

func (s *BloodhoundDB) GetSourceKinds(ctx context.Context) ([]model.SourceKind, error) {
	var (
		sourceKinds []model.SourceKind
		query       = fmt.Sprintf(`
		SELECT sk.id, k.name
		FROM %s sk
		JOIN %s k ON k.id = sk.kind_id
		ORDER BY name ASC;
	`, model.SourceKind{}.TableName(), model.Kind{}.TableName())
	)

	result := s.db.WithContext(ctx).Raw(query).Scan(&sourceKinds)
	if err := result.Error; err != nil {
		return nil, fmt.Errorf("failed to fetch source kinds: %w", err)
	}

	return sourceKinds, nil
}

func (s *BloodhoundDB) GetSourceKindByName(ctx context.Context, name string) (model.SourceKind, error) {
	var (
		sourceKind model.SourceKind
		query      = fmt.Sprintf(`
		SELECT sk.id, k.name
		FROM %s sk
		JOIN %s k ON k.id = sk.kind_id
		WHERE k.name = $1;
	`, model.SourceKind{}.TableName(), model.Kind{}.TableName())
	)

	result := s.db.WithContext(ctx).Raw(query, name).Scan(&sourceKind)
	if result.Error != nil {
		return model.SourceKind{}, result.Error
	}

	if result.RowsAffected == 0 || sourceKind.ID == 0 {
		return model.SourceKind{}, ErrNotFound
	}

	return sourceKind, nil
}

// GetSourceKindByID - retrieves source_kind by source_kind table id
func (s *BloodhoundDB) GetSourceKindByID(ctx context.Context, id int) (model.SourceKind, error) {
	if sourceKinds, err := s.GetSourceKindsByIDs(ctx, id); err != nil {
		return model.SourceKind{}, err
	} else {
		return sourceKinds[0], nil
	}
}

func (s *BloodhoundDB) GetSourceKindsByIDs(ctx context.Context, ids ...int) ([]model.SourceKind, error) {
	var sourceKinds []model.SourceKind
	if len(ids) == 0 {
		return sourceKinds, nil
	}

	// Dedupe IDs so the length check against query results doesn't produce a
	// false-positive ErrNotFound when callers pass duplicate values.
	uniqueIDs := utils.Dedupe(ids)

	query := fmt.Sprintf(`
		SELECT sk.id, k.name
		FROM %s sk
		JOIN %s k ON k.id = sk.kind_id
		WHERE sk.id IN (?)
		ORDER BY sk.id;
	`, model.SourceKind{}.TableName(), model.Kind{}.TableName())
	result := s.db.WithContext(ctx).Raw(query, uniqueIDs).Scan(&sourceKinds)
	if err := result.Error; err != nil {
		return nil, fmt.Errorf("failed to fetch source kinds by IDs: %w", err)
	}

	if len(sourceKinds) != len(uniqueIDs) {
		return nil, ErrNotFound
	}

	return sourceKinds, nil
}

func (s *BloodhoundDB) DeleteSourceKindsByName(ctx context.Context, names ...string) error {
	if len(names) == 0 {
		return nil
	}

	// Source Kinds Base & AZBase are excluded from being deleted.
	var query = fmt.Sprintf(`
		DELETE FROM %s sk
		WHERE sk.kind_id IN
		 (SELECT id FROM %s k WHERE k.name IN (?) and k.name NOT IN ('Base', 'AZBase'))
	`, model.SourceKind{}.TableName(), model.Kind{}.TableName())

	result := s.db.WithContext(ctx).Exec(query, names)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to delete source kinds by name: %w", err)
	}

	return nil
}
