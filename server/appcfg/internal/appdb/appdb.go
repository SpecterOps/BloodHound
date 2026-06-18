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
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

const (
	tableDatapipeStatus = "datapipe_status"
)

// pgxQuerier lists only the pgx methods this package actually calls.
type pgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// datapipeStatusRow is the package-local representation of a row in the datapipe_status table.
// It exists only to hold raw scanned values; callers receive the application-level services.DatapipeStatus.
// The db struct tags map column names to struct fields and enable automatic scanning via pgx.RowToStructByName.
type datapipeStatusRow struct {
	Status                  string    `db:"status"`
	UpdatedAt               time.Time `db:"updated_at"`
	LastCompleteAnalysisAt  null.Time `db:"last_complete_analysis_at"`
	LastAnalysisRunAt       null.Time `db:"last_analysis_run_at"`
	NextScheduledAnalysisAt null.Time `db:"next_scheduled_analysis_at"`
}

// toDatapipeStatus translates a raw DB row into the domain model.
func toDatapipeStatus(row datapipeStatusRow) services.DatapipeStatus {
	return services.DatapipeStatus{
		Status:                  services.DatapipeStatusType(row.Status),
		UpdatedAt:               row.UpdatedAt,
		LastCompleteAnalysisAt:  row.LastCompleteAnalysisAt,
		LastAnalysisRunAt:       row.LastAnalysisRunAt,
		NextScheduledAnalysisAt: row.NextScheduledAnalysisAt,
	}
}

// Store implements persistence for application configuration features.
type Store struct {
	db pgxQuerier
}

// NewStore returns a new Store backed by the given pgx connection pool.
func NewStore(db pgxQuerier) *Store {
	return &Store{db: db}
}

func (s *Store) GetDatapipeStatus(ctx context.Context) (services.DatapipeStatus, error) {
	var (
		row  datapipeStatusRow
		rows pgx.Rows
		err  error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"status",
		"updated_at",
		"last_complete_analysis_at",
		"last_analysis_run_at",
		"next_scheduled_analysis_at",
	)
	selectBuilder.From(tableDatapipeStatus)
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return services.DatapipeStatus{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[datapipeStatusRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.DatapipeStatus{}, services.ErrNotFound
	}
	if err != nil {
		return services.DatapipeStatus{}, fmt.Errorf("reading rows: %w", err)
	}

	return toDatapipeStatus(row), nil
}
