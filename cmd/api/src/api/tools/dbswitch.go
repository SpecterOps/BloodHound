// Copyright 2023 Specter Ops, Inc.
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

package tools

import (
	"context"
	"errors"

	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/cmd/api/src/config"
)

const (
	pgErrorUniqueViolationCode           = "23505"
	pgErrorUniqueViolationConstraintName = "pg_type_typname_nsp_index"
)

type postgresqlConnection interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

var newGraphDriverConnection = func(ctx context.Context, cfg config.Configuration) (postgresqlConnection, error) {
	return newPostgresqlConnection(ctx, cfg)
}

func newPostgresqlConnection(ctx context.Context, cfg config.Configuration) (*pgx.Conn, error) {
	if pgCfg, err := pgx.ParseConfig(cfg.Database.PostgreSQLConnectionString()); err != nil {
		return nil, err
	} else {
		return pgx.ConnectConfig(ctx, pgCfg)
	}
}

func HasGraphDriverSet(ctx context.Context, pgxConn *pgx.Conn) (bool, error) {
	var (
		exists bool
		row    = pgxConn.QueryRow(ctx, `select exists(select * from database_switch limit 1);`)
	)

	return exists, row.Scan(&exists)
}

func getGraphDriver(ctx context.Context, pgxConn postgresqlConnection) (string, error) {
	var (
		driverName string
		row        = pgxConn.QueryRow(ctx, `select driver from database_switch limit 1;`)
	)

	return driverName, row.Scan(&driverName)
}

func SetGraphDriver(ctx context.Context, cfg config.Configuration, driverName string) error {
	if pgxConn, err := newPostgresqlConnection(ctx, cfg); err != nil {
		return err
	} else {
		defer pgxConn.Close(ctx)

		if hasDriver, err := HasGraphDriverSet(ctx, pgxConn); err != nil {
			return err
		} else if hasDriver {
			_, err := pgxConn.Exec(ctx, `update database_switch set driver = $1;`, driverName)
			return err
		} else {
			_, err := pgxConn.Exec(ctx, `insert into database_switch (driver) values ($1);`, driverName)
			return err
		}
	}
}

// ResolveGraphDriver initializes a graph driver connection and creates the `database_switch` table if it does not exist. It returns the graph driver name, or an error.
func ResolveGraphDriver(ctx context.Context, cfg config.Configuration) (string, error) {
	driverName := cfg.GraphDriver

	if pgxConn, err := newGraphDriverConnection(ctx, cfg); err != nil {
		return "", err
	} else {
		defer pgxConn.Close(ctx)
		if _, err := pgxConn.Exec(ctx, `create table if not exists database_switch (driver text not null, primary key(driver));`); err != nil {
			var pgError *pgconn.PgError
			if errors.As(err, &pgError) && pgError.Code == pgErrorUniqueViolationCode && pgError.ConstraintName == pgErrorUniqueViolationConstraintName {
				slog.InfoContext(ctx, "Concurrent database_switch table CREATE; falling back to primary")
			} else {
				return "", err
			}
		}

		if setDriverName, err := getGraphDriver(ctx, pgxConn); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.InfoContext(
					ctx,
					"No database driver has been set for migration, using configured driver",
					slog.String("configured_driver", driverName))
			} else {
				return "", err
			}
		} else {
			driverName = setDriverName
		}
	}

	return driverName, nil
}
