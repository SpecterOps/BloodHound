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
	"errors"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/analysis/services"
)

// pgxQuerier is the minimal pgx surface the analysis Store relies on. Each
// appdb package defines its own copy so the abstraction stays scoped to the
// methods actually exercised here.
type pgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Store performs analysis-request persistence operations directly against a PostgreSQL
// connection. Callers receive appdb-level sentinels rather than raw driver errors.
type Store struct {
	db pgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

// GetAnalysisRequest returns the currently pending analysis request, or ErrNotFound when
// no request is present.
func (s *Store) GetAnalysisRequest(ctx context.Context) (services.RequestedAnalysis, error) {
	var (
		row analysisRequest
		err error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"requested_by",
		"request_type",
		"requested_at",
		"delete_all_graph",
		"delete_sourceless_graph",
		"delete_source_kinds",
		"delete_relationships",
	)
	selectBuilder.From("analysis_request_switch")
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	err = s.db.QueryRow(ctx, sqlQuery, args...).Scan(
		&row.RequestedBy,
		&row.RequestType,
		&row.RequestedAt,
		&row.DeleteAllGraph,
		&row.DeleteSourcelessGraph,
		&row.DeleteSourceKinds,
		&row.DeleteRelationships,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return services.RequestedAnalysis{}, services.ErrNotFound
	}
	if err != nil {
		return services.RequestedAnalysis{}, err
	}
	return toRequestedAnalysis(row), nil
}

// CreateAnalysisRequest atomically inserts a new analysis request for the given user when none
// is currently pending. The analysis_request_switch table is a DB-level singleton (PRIMARY KEY
// (singleton) with CHECK (singleton)), so INSERT ... ON CONFLICT (singleton) DO NOTHING is
// race-free and idempotent. The currently-pending request is returned alongside a boolean
// indicating whether this call created it (true) or a request was already pending (false).
func (s *Store) CreateAnalysisRequest(ctx context.Context, requestedBy string) (services.RequestedAnalysis, bool, error) {
	var (
		now            = time.Now().UTC()
		err            error
		commandTag     pgconn.CommandTag
		currentRequest services.RequestedAnalysis
	)

	insertBuilder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	insertBuilder.InsertInto("analysis_request_switch")
	insertBuilder.Cols(
		"requested_by",
		"request_type",
		"requested_at",
		"delete_all_graph",
		"delete_sourceless_graph",
		"delete_source_kinds",
		"delete_relationships",
	)
	insertBuilder.Values(
		requestedBy,
		string(services.RequestedAnalysisTypeAnalysis),
		now,
		false,
		false,
		[]string{},
		[]string{},
	)
	insertBuilder.SQL("ON CONFLICT (singleton) DO NOTHING")

	sqlQuery, args := insertBuilder.Build()

	commandTag, err = s.db.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return services.RequestedAnalysis{}, false, err
	}

	currentRequest, err = s.GetAnalysisRequest(ctx)
	if err != nil {
		return services.RequestedAnalysis{}, false, err
	}

	return currentRequest, commandTag.RowsAffected() == 1, nil
}
