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

package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
)

const (
	DriverName = "pg"

	poolInitConnectionTimeout = time.Second * 10
	defaultTransactionTimeout = time.Minute * 15

	// defaultBatchWriteSize is currently set to 2k. This is meant to strike a balance between the cost of thousands
	// of round-trips against the cost of locking tables for too long.
	defaultBatchWriteSize = 2_000
)

func afterPooledConnectionEstablished(ctx context.Context, conn *pgx.Conn) error {
	for _, dataType := range pgsql.CompositeTypes {
		if definition, err := conn.LoadType(ctx, dataType.String()); err != nil {
			if !StateObjectDoesNotExist.ErrorMatches(err) {
				return fmt.Errorf("failed to match composite type %s to database: %w", dataType, err)
			}
		} else {
			conn.TypeMap().RegisterType(definition)
		}
	}

	return nil
}

func afterPooledConnectionRelease(conn *pgx.Conn) bool {
	for _, dataType := range pgsql.CompositeTypes {
		if _, hasType := conn.TypeMap().TypeForName(dataType.String()); !hasType {
			// This connection should be destroyed since it does not contain information regarding the schema's
			// composite types
			log.Warnf(fmt.Sprintf("Unable to find expected data type: %s. This database connection will not be pooled.", dataType))
			return false
		}
	}

	return true
}

func newDatabase(connectionString string) (*Driver, error) {
	poolCtx, done := context.WithTimeout(context.Background(), poolInitConnectionTimeout)
	defer done()

	if poolCfg, err := pgxpool.ParseConfig(connectionString); err != nil {
		return nil, err
	} else {
		// TODO: Min and Max connections for the pool should be configurable
		poolCfg.MinConns = 5
		poolCfg.MaxConns = 50

		// Bind functions to the AfterConnect and AfterRelease hooks to ensure that composite type registration occurs.
		// Without composite type registration, the pgx connection type will not be able to marshal PG OIDs to their
		// respective Golang structs.
		poolCfg.AfterConnect = afterPooledConnectionEstablished
		poolCfg.AfterRelease = afterPooledConnectionRelease

		if pool, err := pgxpool.NewWithConfig(poolCtx, poolCfg); err != nil {
			return nil, err
		} else {
			driverInst := &Driver{
				pool:                      pool,
				defaultTransactionTimeout: defaultTransactionTimeout,
				batchWriteSize:            defaultBatchWriteSize,
			}

			// Because the schema manager will act on the database on its own it needs a reference to the driver
			// TODO: This cyclical dependency might want to be unwound
			driverInst.schemaManager = NewSchemaManager(driverInst)
			return driverInst, nil
		}
	}
}

func init() {
	dawgs.Register(DriverName, func(ctx context.Context, cfg dawgs.Config) (graph.Database, error) {
		if connectionString, typeOK := cfg.DriverCfg.(string); !typeOK {
			return nil, fmt.Errorf("expected string for configuration type but got %T", cfg)
		} else if graphDB, err := newDatabase(connectionString); err != nil {
			return nil, err
		} else if err := graphDB.AssertSchema(ctx, graph.Schema{}); err != nil {
			return nil, err
		} else {
			return graphDB, nil
		}
	})
}
