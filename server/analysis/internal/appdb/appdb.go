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
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/analysis/internal/services"
)

const (
	tableAnalysisRequestSwitch = "analysis_request_switch"
)

// queryExecer is the minimal surface satisfied by both *pgxpool.Pool and pgx.Tx.
// Helpers that must run inside a transaction accept this narrower interface.
type queryExecer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// pgxQuerier extends queryExecer with the ability to begin a transaction.
// Only *pgxpool.Pool satisfies this full interface; pgx.Tx satisfies only queryExecer.
type pgxQuerier interface {
	queryExecer
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// analysisRequest is the package-local representation of a row in the analysis_request_switch table.
// It exists only to hold raw scanned values; callers receive the application-level services.RequestedAnalysis.
// The db struct tags map column names to struct fields and enable automatic scanning via pgx.RowToStructByName.
type analysisRequest struct {
	RequestedBy           string    `db:"requested_by"`
	RequestType           string    `db:"request_type"`
	RequestedAt           time.Time `db:"requested_at"`
	DeleteAllGraph        bool      `db:"delete_all_graph"`
	DeleteSourcelessGraph bool      `db:"delete_sourceless_graph"`
	DeleteSourceKinds     []string  `db:"delete_source_kinds"`
	DeleteRelationships   []string  `db:"delete_relationships"`
}

// toRequestedAnalysis translates a raw DB row into the domain model.
func toRequestedAnalysis(row analysisRequest) services.RequestedAnalysis {
	return services.RequestedAnalysis{
		RequestedBy:           row.RequestedBy,
		RequestType:           services.RequestedAnalysisType(row.RequestType),
		RequestedAt:           row.RequestedAt,
		DeleteAllGraph:        row.DeleteAllGraph,
		DeleteSourcelessGraph: row.DeleteSourcelessGraph,
		DeleteSourceKinds:     row.DeleteSourceKinds,
		DeleteRelationships:   row.DeleteRelationships,
	}
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

// selectAnalysisRequest runs the SELECT query against the provided querier.
// It is used by both GetAnalysisRequest and the transactional DeleteAnalysisRequest
// so that the same logic executes whether or not an outer transaction is present.
func selectAnalysisRequest(ctx context.Context, querier queryExecer) (services.RequestedAnalysis, error) {
	var (
		row  analysisRequest
		rows pgx.Rows
		err  error
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
	selectBuilder.From(tableAnalysisRequestSwitch)
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	rows, err = querier.Query(ctx, sqlQuery, args...)
	if err != nil {
		return services.RequestedAnalysis{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[analysisRequest])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.RequestedAnalysis{}, services.ErrNoPendingRequest
	}
	if err != nil {
		return services.RequestedAnalysis{}, fmt.Errorf("reading rows: %w", err)
	}
	return toRequestedAnalysis(row), nil
}

// GetAnalysisRequest returns the currently pending analysis request, or ErrNoPendingRequest when
// no request is present.
func (s *Store) GetAnalysisRequest(ctx context.Context) (services.RequestedAnalysis, error) {
	return selectAnalysisRequest(ctx, s.db)
}

// CreateAnalysisRequest inserts a new analysis request when none is pending, using a transaction
// to keep the insert and read-back atomic. Returns the current request and whether this call
// created it.
func (s *Store) CreateAnalysisRequest(ctx context.Context, requestedBy string) (services.RequestedAnalysis, bool, error) {
	var (
		now            = time.Now().UTC()
		err            error
		tx             pgx.Tx
		commandTag     pgconn.CommandTag
		currentRequest services.RequestedAnalysis
	)

	insertBuilder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	insertBuilder.InsertInto(tableAnalysisRequestSwitch)
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

	tx, err = s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return services.RequestedAnalysis{}, false, fmt.Errorf("beginning transaction: %s", err)
	}
	defer func() {
		// Rollback is a no-op when the transaction has already been committed.
		_ = tx.Rollback(ctx)
	}()

	commandTag, err = tx.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return services.RequestedAnalysis{}, false, err
	}

	currentRequest, err = selectAnalysisRequest(ctx, tx)
	if err != nil {
		return services.RequestedAnalysis{}, false, err
	}

	if err = tx.Commit(ctx); err != nil {
		return services.RequestedAnalysis{}, false, fmt.Errorf("committing transaction: %s", err)
	}

	return currentRequest, commandTag.RowsAffected() == 1, nil
}

// DeleteAnalysisRequest removes the currently pending analysis request within a transaction.
// The row is first read under the transaction to guard against concurrent modifications.
// If the pending request is a deletion request, ErrDeletionRequestPending is returned and
// nothing is deleted; only an analysis request may be cancelled.
func (s *Store) DeleteAnalysisRequest(ctx context.Context) error {
	var (
		tx             pgx.Tx
		currentRequest services.RequestedAnalysis
		err            error
	)

	tx, err = s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction: %s", err)
	}
	defer func() {
		// Rollback is a no-op when the transaction has already been committed.
		_ = tx.Rollback(ctx)
	}()

	currentRequest, err = selectAnalysisRequest(ctx, tx)
	if err != nil {
		return err
	}

	if currentRequest.RequestType == services.RequestedAnalysisTypeDeletion {
		return services.ErrDeletionRequestPending
	}

	deleteBuilder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	deleteBuilder.DeleteFrom(tableAnalysisRequestSwitch)

	sqlQuery, args := deleteBuilder.Build()

	_, err = tx.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("deleting analysis request: %s", err)
	}

	return tx.Commit(ctx)
}
