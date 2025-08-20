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
	DeleteSourceKindsByName(ctx context.Context, kinds graph.Kinds) error
	RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error
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
			INSERT INTO source_kinds (name)
			VALUES (?)
			ON CONFLICT (name) DO NOTHING;
		`

		result := s.db.WithContext(ctx).Exec(query, sourceKind)
		if err := result.Error; err != nil {
			return fmt.Errorf("failed to insert source kind %q: %w", sourceKind, err)
		}

		return nil
	}
}

type SourceKind struct {
	ID   int        `json:"id"`
	Name graph.Kind `json:"name"`
}

func (s *BloodhoundDB) GetSourceKinds(ctx context.Context) ([]SourceKind, error) {
	const query = `
		SELECT id, name
		FROM source_kinds;
	`

	type rawSourceKind struct {
		ID   int
		Name string
	}

	var kinds []rawSourceKind
	result := s.db.WithContext(ctx).Raw(query).Scan(&kinds)
	if err := result.Error; err != nil {
		return nil, fmt.Errorf("failed to fetch source kinds: %w", err)
	}

	out := make([]SourceKind, len(kinds))
	for i, k := range kinds {
		out[i] = SourceKind{
			ID:   k.ID,
			Name: graph.StringKind(k.Name),
		}
	}

	return out, nil
}

func (s *BloodhoundDB) DeleteSourceKindsByName(ctx context.Context, kinds graph.Kinds) error {
	if len(kinds) == 0 {
		return nil
	}

	// Convert to []string for the SQL query
	names := kinds.Strings()

	const query = `
		DELETE FROM source_kinds
		WHERE name = ANY (?);
	`

	result := s.db.WithContext(ctx).Exec(query, pq.Array(names))
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to delete source kinds by name: %w", err)
	}

	return nil
}
