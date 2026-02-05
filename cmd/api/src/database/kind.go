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

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type Kind interface {
	GetKindByName(ctx context.Context, name string) (model.Kind, error)
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

func (s *BloodhoundDB) GetKindById(ctx context.Context, id int32) (model.Kind, error) {
	const query = `
		SELECT id, name
		FROM kind
		WHERE id = $1;
	`

	var kind model.Kind
	result := s.db.WithContext(ctx).Raw(query, id).Scan(&kind)

	if result.Error != nil {
		return model.Kind{}, result.Error
	}

	if result.RowsAffected == 0 || kind.ID == 0 {
		return model.Kind{}, ErrNotFound
	}

	return kind, nil
}
