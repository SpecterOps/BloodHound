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
	"encoding/json"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

const tableSchemaKindInfo = "schema_kind_info"

func (s *Store) CreateKindInfos(ctx context.Context, kindInfos []services.KindInfo) error {
	var (
		insertBuilder = sqlbuilder.PostgreSQL.NewInsertBuilder()
		sqlQuery      string
		args          []any
		err           error
	)

	if len(kindInfos) == 0 {
		return nil
	}

	insertBuilder.InsertInto(tableSchemaKindInfo)
	insertBuilder.Cols(
		"kind_id",
		"node_kind_id",
		"relationship_kind_id",
		"info_key",
		"title",
		"position",
		"content",
	)

	for _, kindInfo := range kindInfos {
		content := kindInfo.Content
		if len(content) == 0 {
			content = json.RawMessage("{}")
		}

		insertBuilder.Values(
			kindInfo.KindID,
			kindInfo.NodeKindID,
			kindInfo.RelationshipKindID,
			kindInfo.InfoKey,
			kindInfo.Title,
			kindInfo.Position,
			sqlbuilder.Build("$?::jsonb", string(content)),
		)
	}

	sqlQuery, args = insertBuilder.Build()

	if _, err = s.db.Exec(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("creating kind infos: %w", err)
	}

	return nil
}
