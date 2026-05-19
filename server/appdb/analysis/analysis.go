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

package analysis

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/appdb/pgxutils"
	"github.com/specterops/bloodhound/server/models"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// ErrNotFound is returned by Store operations when no matching analysis request row exists.
var ErrNotFound = errors.New("analysis request not found")

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
func (s *Store) GetAnalysisRequest(ctx context.Context) (models.RequestedAnalysis, error) {
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
		return models.RequestedAnalysis{}, err
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
		return models.RequestedAnalysis{}, ErrNotFound
	}
	if err != nil {
		return models.RequestedAnalysis{}, err
	}
	return toRequestedAnalysis(row), nil
}

// CreateAnalysisRequest inserts a new analysis request for the given user.
// If a request of any type already exists the call is a no-op (matching the
// legacy setAnalysisRequest behaviour: an incoming analysis request never
// overwrites an existing one).
func (s *Store) CreateAnalysisRequest(ctx context.Context, requestedBy string) error {
	var (
		err      error
		now      = time.Now().UTC()
		sqlQuery string
		args     []any
	)

	_, err = s.GetAnalysisRequest(ctx)
	if err == nil {
		// A request already exists — no-op.
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}

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
			string(models.RequestedAnalysisTypeAnalysis),
			now,
			false,
			false,
			[]string{},
			[]string{},
		)),
	).Build(ctx)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, sqlQuery, args...)
	return err
}
