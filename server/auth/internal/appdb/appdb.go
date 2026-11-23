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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	sqlCurrentTimestamp = "current_timestamp"
)

type pgxQuerier interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Store struct {
	db pgxQuerier
}

func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

func qualify(tableName, columnName string) string {
	return tableName + "." + columnName
}

func qualifyAll(tableName string, columnNames ...string) []string {
	var columns = make([]string, 0, len(columnNames))

	for _, columnName := range columnNames {
		columns = append(columns, qualify(tableName, columnName))
	}

	return columns
}
