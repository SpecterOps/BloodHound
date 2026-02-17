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

		const query = `
			WITH dawgs_kind (id, name) AS ( SELECT id, name FROM upsert_kind(?))
			INSERT INTO source_kinds (kind_id, active)
			SELECT dk.id, true
			FROM dawgs_kind dk
			ON CONFLICT (kind_id) DO UPDATE SET active = true;
		`

		result := s.db.WithContext(ctx).Exec(query, sourceKind)
		if err := result.Error; err != nil {
			return fmt.Errorf("failed to insert source kind %q: %w", sourceKind, err)
		}

		return nil
	}
}

type SourceKind struct {
	ID     int        `json:"id"`
	Name   graph.Kind `json:"name"`
	Active bool       `json:"active"`
}

func (s *BloodhoundDB) GetSourceKinds(ctx context.Context) ([]SourceKind, error) {
	const query = `
		SELECT sk.id, k.name, sk.active
		FROM source_kinds sk
		JOIN kind k ON k.id = sk.kind_id
		WHERE sk.active = true
		ORDER BY name ASC;
	`

	type rawSourceKind struct {
		ID     int
		Name   string
		Active bool
	}

	var kinds []rawSourceKind
	result := s.db.WithContext(ctx).Raw(query).Scan(&kinds)
	if err := result.Error; err != nil {
		return nil, fmt.Errorf("failed to fetch source kinds: %w", err)
	}

	out := make([]SourceKind, len(kinds))
	for i, k := range kinds {
		out[i] = SourceKind{
			ID:     k.ID,
			Name:   graph.StringKind(k.Name),
			Active: k.Active,
		}
	}

	return out, nil
}

func (s *BloodhoundDB) GetSourceKindByName(ctx context.Context, name string) (SourceKind, error) {
	const query = `
		SELECT sk.id, k.name, sk.active
		FROM source_kinds sk
		JOIN kind k ON k.id = sk.kind_id
		WHERE k.name = $1 AND sk.active = true;
	`

	type rawSourceKind struct {
		ID     int
		Name   string
		Active bool
	}

	var raw rawSourceKind
	result := s.db.WithContext(ctx).Raw(query, name).Scan(&raw)

	if result.Error != nil {
		return SourceKind{}, result.Error
	}

	if result.RowsAffected == 0 || raw.ID == 0 {
		return SourceKind{}, ErrNotFound
	}

	kind := SourceKind{
		ID:     raw.ID,
		Name:   graph.StringKind(raw.Name),
		Active: raw.Active,
	}

	return kind, nil
}

// GetSourceKindByID - retrieves source_kind by source_kind table id
func (s *BloodhoundDB) GetSourceKindByID(ctx context.Context, id int) (SourceKind, error) {
	const query = `
		SELECT sk.id, k.name, sk.active
		FROM source_kinds sk
		JOIN kind k ON k.id = sk.kind_id
		WHERE sk.id = $1 AND sk.active = true;
	`
	type rawSourceKind struct {
		ID     int
		Name   string
		Active bool
	}

	var raw rawSourceKind
	result := s.db.WithContext(ctx).Raw(query, id).Scan(&raw)

	if result.Error != nil {
		return SourceKind{}, result.Error
	}

	if result.RowsAffected == 0 || raw.ID == 0 {
		return SourceKind{}, ErrNotFound
	}

	kind := SourceKind{
		ID:     raw.ID,
		Name:   graph.StringKind(raw.Name),
		Active: raw.Active,
	}

	return kind, nil
}

func (s *BloodhoundDB) DeactivateSourceKindsByName(ctx context.Context, kinds graph.Kinds) error {
	if len(kinds) == 0 {
		return nil
	}

	// Convert to []string for the SQL query
	names := kinds.Strings()

	// Source Kinds Base & AZBase are excluded from being deactivated.
	const query = `
		UPDATE source_kinds AS sk
		SET active = false
		FROM kind k
		WHERE sk.kind_id = k.id
		AND k.name = ANY (?)
		AND k.name NOT IN ('Base', 'AZBase');
	`

	result := s.db.WithContext(ctx).Exec(query, pq.Array(names))
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to deactivate source kinds by name: %w", err)
	}

	return nil
}
