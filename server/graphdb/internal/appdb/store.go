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
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
)

// tableKind is the kind table joined when resolving entity kinds to their integer ids.
const tableKind = "kind"

// kindRow is the package-local DB row type for a resolved kind. The id is the
// schema_node_kinds or schema_relationship_kinds row id; the name is the kind name
// from the kind table. db: tags drive pgx.RowToStructByName scanning. ID is nullable
// to support LEFT JOIN queries where the kind exists in the kind table but not in the
// schema_*_kinds table.
type kindRow struct {
	ID   *int32 `db:"id"`
	Name string `db:"name"`
}

// toKind translates a raw kind row into the domain model.
func toKind(row kindRow) services.Kind {
	return services.Kind{ID: row.ID, Name: row.Name}
}

// graphReader is the minimal graph-database surface this store relies on to read graph
// entities. Only ReadTransaction is required, keeping the abstraction scoped to what is
// exercised here.
type graphReader interface {
	ReadTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error
}

// pgxQuerier lists only the pgx methods this package actually calls against PostgreSQL.
type pgxQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Store reads graph entity data from the graph database and resolves entity kinds against
// the schema kind tables. Callers receive services-layer sentinels rather than raw driver
// errors.
type Store struct {
	graph graphReader
	db    pgxQuerier
}

// NewStore returns a Store backed by the provided graph database and pgx connection pool.
func NewStore(graphDatabase graphReader, db pgxQuerier) *Store {
	return &Store{graph: graphDatabase, db: db}
}
