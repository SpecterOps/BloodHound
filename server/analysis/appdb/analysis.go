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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/analysis/service"
	"github.com/specterops/bloodhound/server/appdb/pgxutils"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// Store performs analysis-request persistence operations directly against a PostgreSQL
// connection. Callers receive appdb-level sentinels rather than raw driver errors.
type Store struct {
	db pgxutils.PgxQuerier
}

// NewStore returns a Store backed by the provided pgx connection pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{db: pool}
}

// GetAnalysisRequest returns the currently pending analysis request, or ErrNotFound when
// no request is present.
func (s *Store) GetAnalysisRequest(ctx context.Context) (service.RequestedAnalysis, error) {
	var (
		row      analysisRequest
		sqlQuery string
		args     []any
		err      error
	)

	sqlQuery, args, err = psql.Select(
		sm.Columns(
			"requested_by",
			"request_type",
			"requested_at",
			"delete_all_graph",
			"delete_sourceless_graph",
			"delete_source_kinds",
			"delete_relationships",
		),
		sm.From("analysis_request_switch"),
		sm.Limit(1),
	).Build(ctx)
	if err != nil {
		return service.RequestedAnalysis{}, err
	}

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
		return service.RequestedAnalysis{}, service.ErrNotFound
	}
	if err != nil {
		return service.RequestedAnalysis{}, err
	}
	return toRequestedAnalysis(row), nil
}

// CreateAnalysisRequest atomically inserts a new analysis request for the given user when none
// is currently pending. The analysis_request_switch table is a DB-level singleton (PRIMARY KEY
// (singleton) with CHECK (singleton)), so INSERT ... ON CONFLICT (singleton) DO NOTHING is
// race-free and idempotent. The currently-pending request is returned alongside a boolean
// indicating whether this call created it (true) or a request was already pending (false).
func (s *Store) CreateAnalysisRequest(ctx context.Context, requestedBy string) (service.RequestedAnalysis, bool, error) {
	var (
		now            = time.Now().UTC()
		sqlQuery       string
		args           []any
		err            error
		commandTag     pgconn.CommandTag
		currentRequest service.RequestedAnalysis
	)

	sqlQuery, args, err = psql.Insert(
		im.Into("analysis_request_switch",
			"requested_by",
			"request_type",
			"requested_at",
			"delete_all_graph",
			"delete_sourceless_graph",
			"delete_source_kinds",
			"delete_relationships",
		),
		im.Values(psql.Arg(
			requestedBy,
			string(service.RequestedAnalysisTypeAnalysis),
			now,
			false,
			false,
			[]string{},
			[]string{},
		)),
		im.OnConflict("singleton").DoNothing(),
	).Build(ctx)
	if err != nil {
		return service.RequestedAnalysis{}, false, err
	}

	commandTag, err = s.db.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return service.RequestedAnalysis{}, false, err
	}

	currentRequest, err = s.GetAnalysisRequest(ctx)
	if err != nil {
		return service.RequestedAnalysis{}, false, err
	}

	return currentRequest, commandTag.RowsAffected() == 1, nil
}
