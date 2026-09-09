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

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

const (
	tableParameters = "parameters"
)

type ParameterRow struct {
	Key         string            `db:"key"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
	Value       types.JSONBObject `db:"value"`

	CreatedAt null.Time `db:"created_at"`
	UpdatedAt null.Time `db:"updated_at"`
	// DeletedAt is not actually stored in the DB
}

func toParameter(row ParameterRow) services.Parameter {
	return services.Parameter{
		Key:         services.ParameterKey(row.Key),
		Name:        row.Name,
		Description: row.Description,
		Value:       row.Value,
		CreatedAt:   row.CreatedAt.ValueOrZero(),
		UpdatedAt:   row.UpdatedAt.ValueOrZero(),
	}
}

func (s *Store) GetConfigurationParameter(ctx context.Context, parameterKey services.ParameterKey) (services.Parameter, error) {
	var (
		row  ParameterRow
		rows pgx.Rows
		err  error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"key",
		"name",
		"description",
		"value",
		"created_at",
		"updated_at",
	)
	selectBuilder.From(tableParameters)
	selectBuilder.Where(selectBuilder.Equal("key", string(parameterKey)))
	selectBuilder.Limit(1)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return services.Parameter{}, err
	}

	row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[ParameterRow])
	if errors.Is(err, pgx.ErrNoRows) {
		return services.Parameter{}, services.ErrNotFound
	}
	if err != nil {
		return services.Parameter{}, fmt.Errorf("reading rows: %w", err)
	}

	return toParameter(row), nil
}

func (s *Store) GetAllConfigurationParameters(ctx context.Context) (services.Parameters, error) {
	var (
		parameterRows []ParameterRow
		rows          pgx.Rows
		err           error
	)

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"key",
		"name",
		"description",
		"value",
		"created_at",
		"updated_at",
	)
	selectBuilder.From(tableParameters)

	sqlQuery, args := selectBuilder.Build()

	rows, err = s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	} else if parameterRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[ParameterRow]); err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	}

	returnParameters := []services.Parameter{}

	for _, row := range parameterRows {
		returnParameters = append(returnParameters, toParameter(row))
	}

	return services.Parameters(returnParameters), nil
}
