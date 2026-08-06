// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/stretchr/testify/require"
)

const (
	createDatabaseSwitchTableSQL = `create table if not exists database_switch (driver text not null, primary key(driver));`
	getGraphDriverSQL            = `select driver from database_switch limit 1;`
)

func TestLookupGraphDriver(t *testing.T) {
	var (
		ctx              = context.Background()
		configuredDriver = "neo4j"
		storedDriver     = "pg"
		createError      = errors.New("create table failed")
		queryError       = errors.New("query driver failed")
		testCases        = []struct {
			name           string
			setup          func(mockConnection pgxmock.PgxConnIface)
			expectedDriver string
			expectedError  error
		}{
			{
				name: "returns configured driver when no driver is stored",
				setup: func(mockConnection pgxmock.PgxConnIface) {
					mockConnection.ExpectExec(createDatabaseSwitchTableSQL).
						WillReturnResult(pgxmock.NewResult("CREATE TABLE", 0))
					mockConnection.ExpectQuery(getGraphDriverSQL).
						WillReturnRows(pgxmock.NewRows([]string{"driver"}))
				},
				expectedDriver: configuredDriver,
			},
			{
				name: "returns stored driver",
				setup: func(mockConnection pgxmock.PgxConnIface) {
					mockConnection.ExpectExec(createDatabaseSwitchTableSQL).
						WillReturnResult(pgxmock.NewResult("CREATE TABLE", 0))
					mockConnection.ExpectQuery(getGraphDriverSQL).
						WillReturnRows(pgxmock.NewRows([]string{"driver"}).AddRow(storedDriver))
				},
				expectedDriver: storedDriver,
			},
			{
				name: "ignores concurrent table creation error",
				setup: func(mockConnection pgxmock.PgxConnIface) {
					mockConnection.ExpectExec(createDatabaseSwitchTableSQL).
						WillReturnError(&pgconn.PgError{
							Code:           pgErrorUniqueViolationCode,
							ConstraintName: pgErrorUniqueViolationConstraintName,
						})
					mockConnection.ExpectQuery(getGraphDriverSQL).
						WillReturnRows(pgxmock.NewRows([]string{"driver"}).AddRow(storedDriver))
				},
				expectedDriver: storedDriver,
			},
			{
				name: "returns table creation error",
				setup: func(mockConnection pgxmock.PgxConnIface) {
					mockConnection.ExpectExec(createDatabaseSwitchTableSQL).WillReturnError(createError)
				},
				expectedError: createError,
			},
			{
				name: "returns unique violation for a different constraint",
				setup: func(mockConnection pgxmock.PgxConnIface) {
					mockConnection.ExpectExec(createDatabaseSwitchTableSQL).
						WillReturnError(&pgconn.PgError{
							Code:           pgErrorUniqueViolationCode,
							ConstraintName: "other_constraint",
						})
				},
				expectedError: &pgconn.PgError{
					Code:           pgErrorUniqueViolationCode,
					ConstraintName: "other_constraint",
				},
			},
			{
				name: "returns driver query error",
				setup: func(mockConnection pgxmock.PgxConnIface) {
					mockConnection.ExpectExec(createDatabaseSwitchTableSQL).
						WillReturnResult(pgxmock.NewResult("CREATE TABLE", 0))
					mockConnection.ExpectQuery(getGraphDriverSQL).WillReturnError(queryError)
				},
				expectedError: queryError,
			},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockConnection, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)

			testCase.setup(mockConnection)
			mockConnection.ExpectClose()
			setPostgresqlConnectionFactory(t, func(context.Context, config.Configuration) (postgresqlConnection, error) {
				return mockConnection, nil
			})

			driverName, err := ResolveGraphDriver(ctx, config.Configuration{GraphDriver: configuredDriver})

			if testCase.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.expectedError.Error())
			}
			require.Equal(t, testCase.expectedDriver, driverName)
			require.NoError(t, mockConnection.ExpectationsWereMet())
		})
	}
}

func TestLookupGraphDriverReturnsConnectionError(t *testing.T) {
	var (
		connectionError = errors.New("connection failed")
		cfg             = config.Configuration{GraphDriver: "neo4j"}
	)

	setPostgresqlConnectionFactory(t, func(context.Context, config.Configuration) (postgresqlConnection, error) {
		return nil, connectionError
	})

	driverName, err := ResolveGraphDriver(context.Background(), cfg)

	require.ErrorIs(t, err, connectionError)
	require.Empty(t, driverName)
}

func setPostgresqlConnectionFactory(t *testing.T, factory func(context.Context, config.Configuration) (postgresqlConnection, error)) {
	t.Helper()

	originalFactory := newGraphDriverConnection
	newGraphDriverConnection = factory
	t.Cleanup(func() {
		newGraphDriverConnection = originalFactory
	})
}
