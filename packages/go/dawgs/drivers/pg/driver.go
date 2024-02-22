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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	readOnlyTxOptions = pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	}

	readWriteTxOptions = pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	}
)

type Config struct {
	Options            pgx.TxOptions
	QueryExecMode      pgx.QueryExecMode
	QueryResultFormats pgx.QueryResultFormats
	BatchWriteSize     int
}

func OptionSetQueryExecMode(queryExecMode pgx.QueryExecMode) graph.TransactionOption {
	return func(config *graph.TransactionConfig) {
		if pgCfg, typeOK := config.DriverConfig.(*Config); typeOK {
			pgCfg.QueryExecMode = queryExecMode
		}
	}
}

type Driver struct {
	pool                      *pgxpool.Pool
	schemaManager             *SchemaManager
	defaultTransactionTimeout time.Duration
	batchWriteSize            int
}

func (s *Driver) SetDefaultGraph(ctx context.Context, graphSchema graph.Graph) error {
	return s.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return s.schemaManager.SetDefaultGraph(tx, graphSchema)
	})
}

func (s *Driver) KindMapper() KindMapper {
	return s.schemaManager
}

func (s *Driver) SetBatchWriteSize(size int) {
	s.batchWriteSize = size
}

func (s *Driver) SetWriteFlushSize(size int) {
	// THis is a no-op function since PostgreSQL does not require transaction rotation like Neo4j does
}

func (s *Driver) BatchOperation(ctx context.Context, batchDelegate graph.BatchDelegate) error {
	if cfg, err := renderConfig(s.batchWriteSize, readWriteTxOptions, nil); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if batch, err := newBatch(ctx, conn, s.schemaManager, cfg); err != nil {
			return err
		} else {
			defer batch.Close()

			if err := batchDelegate(batch); err != nil {
				return err
			}

			return batch.Commit()
		}
	}
}

func (s *Driver) Close(ctx context.Context) error {
	s.pool.Close()
	return nil
}

func renderConfig(batchWriteSize int, pgxOptions pgx.TxOptions, userOptions []graph.TransactionOption) (*Config, error) {
	graphCfg := graph.TransactionConfig{
		DriverConfig: &Config{
			Options:            pgxOptions,
			QueryExecMode:      pgx.QueryExecModeCacheStatement,
			QueryResultFormats: pgx.QueryResultFormats{pgx.BinaryFormatCode},
			BatchWriteSize:     batchWriteSize,
		},
	}

	for _, option := range userOptions {
		option(&graphCfg)
	}

	if graphCfg.DriverConfig != nil {
		if pgCfg, typeOK := graphCfg.DriverConfig.(*Config); !typeOK {
			return nil, fmt.Errorf("invalid driver config type %T", graphCfg.DriverConfig)
		} else {
			return pgCfg, nil
		}
	}

	return nil, fmt.Errorf("driver config is nil")
}

func (s *Driver) ReadTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	if cfg, err := renderConfig(s.batchWriteSize, readOnlyTxOptions, options); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		return txDelegate(&transaction{
			schemaManager:   s.schemaManager,
			queryExecMode:   cfg.QueryExecMode,
			ctx:             ctx,
			conn:            conn,
			targetSchemaSet: false,
		})
	}
}

func (s *Driver) WriteTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	if cfg, err := renderConfig(s.batchWriteSize, readWriteTxOptions, options); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if tx, err := newTransaction(ctx, conn, s.schemaManager, cfg); err != nil {
			return err
		} else {
			defer tx.Close()

			if err := txDelegate(tx); err != nil {
				return err
			}

			return tx.Commit()
		}
	}
}

func (s *Driver) FetchSchema(ctx context.Context) (graph.Schema, error) {
	// TODO: This is not required for existing functionality as the SchemaManager type handles most of this negotiation
	//		 however, in the future this function would make it easier to make schema management generic and should be
	//		 implemented.
	return graph.Schema{}, fmt.Errorf("not implemented")
}

func (s *Driver) AssertSchema(ctx context.Context, schema graph.Schema) error {
	if err := s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if err := s.schemaManager.AssertSchema(tx, schema); err != nil {
			return err
		} else if schema.DefaultGraph.Name != "" {
			return s.schemaManager.AssertDefaultGraph(tx, schema.DefaultGraph)
		}

		return nil
	}, OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol)); err != nil {
		return err
	} else {
		// Resetting the pool must be done on every schema assertion as composite types may have changed OIDs
		s.pool.Reset()
	}

	return nil
}

func (s *Driver) Run(ctx context.Context, query string, parameters map[string]any) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		result := tx.Raw(query, parameters)
		defer result.Close()

		return result.Error()
	})
}
