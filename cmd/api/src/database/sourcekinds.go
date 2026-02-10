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

	"github.com/lib/pq"
	"github.com/specterops/dawgs/graph"
)

type SourceKindsData interface {
	GetSourceKinds(ctx context.Context) ([]SourceKind, error)
	DeactivateSourceKindsByName(ctx context.Context, kinds graph.Kinds) error
	RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error
	GetSourceKindByName(ctx context.Context, name string) (SourceKind, error)
	GetSourceKindsByIds(ctx context.Context, ids ...int32) ([]SourceKind, error)
}

// RegisterSourceKind returns a function that inserts a source kind by name,
// using the provided context. The returned function can be called later with just the name.
// The function is curried in this way because it is primarily used in datapipe during ingest decoding when
// there is no ctx in scope.
func (s *BloodhoundDB) RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error {
	return func(sourceKind graph.Kind) error {
		if sourceKind == nil || sourceKind == graph.EmptyKind {
			return nil
		}

		const query = `
			INSERT INTO source_kinds (name, active)
			VALUES (?, true)
			ON CONFLICT (name) DO UPDATE SET active = true;
		`

		result := s.db.WithContext(ctx).Exec(query, sourceKind)
		if err := result.Error; err != nil {
			return fmt.Errorf("failed to insert source kind %q: %w", sourceKind, err)
		}

		return nil
	}
}

type SourceKind struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func (s *SourceKind) ToKind() graph.Kind {
	return graph.StringKind(s.Name)
}

func (s *BloodhoundDB) GetSourceKinds(ctx context.Context) ([]SourceKind, error) {
	const query = `
		SELECT id, name, active
		FROM source_kinds
		WHERE active = true
		ORDER BY name ASC;
	`

	var kinds []SourceKind
	if err := s.db.WithContext(ctx).Raw(query).Scan(&kinds).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch source kinds: %w", err)
	}

	return kinds, nil
}

func (s *BloodhoundDB) GetSourceKindByName(ctx context.Context, name string) (SourceKind, error) {
	const query = `
		SELECT id, name, active
		FROM source_kinds
		WHERE name = $1 AND active = true;
	`

	var sourceKind SourceKind
	result := s.db.WithContext(ctx).Raw(query, name).Scan(&sourceKind)

	if result.Error != nil {
		return SourceKind{}, result.Error
	}

	if result.RowsAffected == 0 || sourceKind.ID == 0 {
		return SourceKind{}, ErrNotFound
	}

	return sourceKind, nil
}

func (s *BloodhoundDB) GetSourceKindsByIds(ctx context.Context, ids ...int32) ([]SourceKind, error) {
	if len(ids) == 0 {
		return []SourceKind{}, nil
	}

	const query = `
		SELECT id, name, active
		FROM source_kinds
		WHERE id = ANY(?) AND active = true;
	`

	var sourceKinds []SourceKind
	if err := s.db.WithContext(ctx).Raw(query, pq.Array(ids)).Scan(&sourceKinds).Error; err != nil {
		return nil, err
	}

	return sourceKinds, nil
}

func (s *BloodhoundDB) GetSourceKindByID(ctx context.Context, id int) (SourceKind, error) {
	const query = `
		SELECT id, name, active
		FROM source_kinds
		WHERE id = $1 AND active = true;
	`

	var sourceKind SourceKind
	result := s.db.WithContext(ctx).Raw(query, id).Scan(&sourceKind)

	if result.Error != nil {
		return SourceKind{}, result.Error
	}

	if result.RowsAffected == 0 || sourceKind.ID == 0 {
		return SourceKind{}, ErrNotFound
	}

	return sourceKind, nil
}

func (s *BloodhoundDB) DeactivateSourceKindsByName(ctx context.Context, kinds graph.Kinds) error {
	if len(kinds) == 0 {
		return nil
	}

	// Convert to []string for the SQL query
	names := kinds.Strings()

	// Source Kinds Base & AZBase are excluded from being deactivated.
	const query = `
		UPDATE source_kinds
		SET active = false
		WHERE name = ANY (?)
		AND name NOT IN ('Base', 'AZBase');
	`

	result := s.db.WithContext(ctx).Exec(query, pq.Array(names))
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to deactivate source kinds by name: %w", err)
	}

	return nil
}
