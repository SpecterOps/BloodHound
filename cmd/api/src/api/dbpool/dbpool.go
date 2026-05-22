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
package dbpool

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/dawgs/drivers/pg"
)

const (
	poolInitConnectionTimeout = time.Second * 10
)

func newPoolCfg(cfg config.DatabaseConfiguration) (*pgxpool.Config, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgreSQLConnectionString())
	if err != nil {
		return nil, err
	}

	// TODO: Min and Max connections for the pool should be configurable
	poolCfg.MinConns = 5
	poolCfg.MaxConns = 50

	if cfg.EnableRDSIAMAuth {
		// Only enable the BeforeConnect handler if RDS IAM Auth is enabled
		poolCfg.BeforeConnect = func(ctx context.Context, connCfg *pgx.ConnConfig) error {
			if newPoolCfg, err := pgxpool.ParseConfig(cfg.RDSIAMAuthConnectionString()); err != nil {
				return err
			} else {
				connCfg.Host = newPoolCfg.ConnConfig.Host
				connCfg.Port = newPoolCfg.ConnConfig.Port

				connCfg.User = newPoolCfg.ConnConfig.User
				connCfg.Password = newPoolCfg.ConnConfig.Password
				connCfg.Database = newPoolCfg.ConnConfig.Database
			}

			return nil
		}
	}

	return poolCfg, nil
}

func NewDawgsPool(cfg config.DatabaseConfiguration) (*pgxpool.Pool, error) {
	if poolCfg, err := newPoolCfg(cfg); err != nil {
		return nil, err
	} else {
		return pg.NewPool(poolCfg)
	}
}

func NewAppPool(cfg config.DatabaseConfiguration) (*pgxpool.Pool, error) {
	poolCtx, done := context.WithTimeout(context.Background(), poolInitConnectionTimeout)
	defer done()

	if poolCfg, err := newPoolCfg(cfg); err != nil {
		return nil, err
	} else {
		return pgxpool.NewWithConfig(poolCtx, poolCfg)
	}
}
