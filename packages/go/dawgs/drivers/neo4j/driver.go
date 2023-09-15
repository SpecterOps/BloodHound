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

package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

const (
	DriverName = "neo4j"
)

func readCfg() neo4j.SessionConfig {
	return neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	}
}

func writeCfg() neo4j.SessionConfig {
	return neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	}
}

type driver struct {
	driver                    neo4j.Driver
	limiter                   channels.ConcurrencyLimiter
	defaultTransactionTimeout time.Duration
	batchWriteSize            int
	writeFlushSize            int
	traversalMemoryLimit      size.Size
}

func (s *driver) SetBatchWriteSize(size int) {
	s.batchWriteSize = size
}

func (s *driver) SetWriteFlushSize(size int) {
	s.writeFlushSize = size
}

func (s *driver) BatchOperation(ctx context.Context, batchDelegate graph.BatchDelegate) error {
	// Attempt to acquire a connection slot or wait for a bit until one becomes available
	if !s.limiter.Acquire(ctx) {
		return graph.ErrContextTimedOut
	} else {
		defer s.limiter.Release()
	}

	var (
		cfg = graph.TransactionConfig{
			Timeout: s.defaultTransactionTimeout,
		}

		session = s.driver.NewSession(writeCfg())
		batch   = newBatchOperation(ctx, session, cfg, s.writeFlushSize, s.batchWriteSize, s.traversalMemoryLimit)
	)

	defer session.Close()
	defer batch.Close()

	if err := batchDelegate(batch); err != nil {
		return err
	}

	return batch.Commit()
}

func (s *driver) Close() error {
	return s.driver.Close()
}

func (s *driver) transaction(ctx context.Context, txDelegate graph.TransactionDelegate, session neo4j.Session, options []graph.TransactionOption) error {
	// Attempt to acquire a connection slot or wait for a bit until one becomes available
	if !s.limiter.Acquire(ctx) {
		return graph.ErrContextTimedOut
	} else {
		defer s.limiter.Release()
	}

	cfg := graph.TransactionConfig{
		Timeout: s.defaultTransactionTimeout,
	}

	// Apply the transaction options
	for _, option := range options {
		option(&cfg)
	}

	tx := newTransaction(ctx, session, cfg, s.writeFlushSize, s.batchWriteSize, s.traversalMemoryLimit)
	defer tx.Close()

	if err := txDelegate(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *driver) ReadTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	session := s.driver.NewSession(readCfg())
	defer session.Close()

	return s.transaction(ctx, txDelegate, session, options)
}

func (s *driver) WriteTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	session := s.driver.NewSession(writeCfg())
	defer session.Close()

	return s.transaction(ctx, txDelegate, session, options)
}

func (s *driver) FetchSchema(ctx context.Context) (*graph.Schema, error) {
	schema := graph.NewSchema()

	return schema, s.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if result := tx.Run("call db.indexes() yield name, uniqueness, provider, labelsOrTypes, properties;", nil); result.Error() != nil {
			return result.Error()
		} else {
			defer result.Close()

			var (
				name       string
				uniqueness string
				provider   string
				labels     []string
				properties []string
			)

			for result.Next() {
				if err := result.Scan(&name, &uniqueness, &provider, &labels, &properties); err != nil {
					return err
				}

				// Need this for neo4j 4.4+ which creates a weird index by default
				if len(labels) == 0 {
					continue
				}

				if len(labels) > 1 || len(properties) > 1 {
					return fmt.Errorf("composite index types are currently not supported")
				}

				label := labels[0]
				property := properties[0]

				if uniqueness == "UNIQUE" {
					schema.EnsureKind(graph.StringKind(label)).Constraint(property, name, parseProviderType(provider))
				} else {
					schema.EnsureKind(graph.StringKind(label)).Index(property, name, parseProviderType(provider))
				}
			}

			return result.Error()
		}
	})
}

func (s *driver) AssertSchema(ctx context.Context, schema *graph.Schema) error {
	if existingSchema, err := s.FetchSchema(ctx); err != nil {
		return fmt.Errorf("could not load schema: %w", err)
	} else {
		return assertAgainst(ctx, schema, existingSchema, s)
	}
}

func (s *driver) Run(ctx context.Context, query string, parameters map[string]any) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Run(query, parameters).Error()
	})
}
