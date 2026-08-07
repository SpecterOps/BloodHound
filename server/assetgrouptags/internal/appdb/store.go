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

// Package appdb stores asset group data in PostgreSQL. It owns the SQL for this
// package and reads asset group tags directly from the pgx pool, translating
// driver not-found errors into the services-layer sentinels.
package appdb

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// tableAssetGroupTags is the table backing asset group tag rows.
const tableAssetGroupTags = "asset_group_tags"

// pgxQuerier lists only the pgx methods this package actually calls against
// PostgreSQL, so the same Store code can run against either a pool or a
// transaction.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Store reads asset group data from PostgreSQL.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}
