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
package appdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// pgxQuerier is the read/write surface shared by both the connection pool and
// an open transaction (pgx.Tx satisfies it), letting a single helper run
// against either without a tx-vs-pool variant.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// pgxDriver is pgxQuerier plus the ability to open a transaction; only the
// connection pool satisfies it.
type pgxDriver interface {
	pgxQuerier
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// Store reads and writes OpenGraph extension schema data to/from PostgreSQL. Callers receive
// services-layer sentinels rather than raw driver errors.
type Store struct {
	db pgxDriver
}

const sqlCurrentTimestamp = "current_timestamp"

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(db pgxDriver) *Store {
	return &Store{db: db}
}
