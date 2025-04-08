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
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type KindMapper interface {
	MapKindID(ctx context.Context, kindID int16) (graph.Kind, error)
	MapKindIDs(ctx context.Context, kindIDs ...int16) (graph.Kinds, error)
	MapKind(ctx context.Context, kind graph.Kind) (int16, error)
	MapKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error)
	AssertKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error)
}

func KindMapperFromGraphDatabase(graphDB graph.Database) (KindMapper, error) {
	switch typedGraphDB := graphDB.(type) {
	case *Driver:
		return typedGraphDB.SchemaManager, nil
	default:
		return nil, fmt.Errorf("unsupported graph database type: %T", typedGraphDB)
	}
}

type SchemaManager struct {
	defaultGraph    model.Graph
	pool            *pgxpool.Pool
	hasDefaultGraph bool
	graphs          map[string]model.Graph
	kindsByID       map[graph.Kind]int16
	kindIDsByKind   map[int16]graph.Kind
	lock            *sync.RWMutex
}

func NewSchemaManager(pool *pgxpool.Pool) *SchemaManager {
	return &SchemaManager{
		pool:            pool,
		hasDefaultGraph: false,
		graphs:          map[string]model.Graph{},
		kindsByID:       map[graph.Kind]int16{},
		kindIDsByKind:   map[int16]graph.Kind{},
		lock:            &sync.RWMutex{},
	}
}

func (s *SchemaManager) WriteTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	if cfg, err := renderConfig(batchWriteSize, readWriteTxOptions, options); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if tx, err := newTransactionWrapper(ctx, conn, s, cfg, true); err != nil {
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

func (s *SchemaManager) fetch(tx graph.Transaction) error {
	if kinds, err := query.On(tx).SelectKinds(); err != nil {
		return err
	} else {
		s.kindsByID = kinds

		for kind, kindID := range s.kindsByID {
			s.kindIDsByKind[kindID] = kind
		}
	}

	return nil
}

func (s *SchemaManager) GetKindIDsByKind() map[int16]graph.Kind {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.kindIDsByKind
}

func (s *SchemaManager) Fetch(ctx context.Context) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return s.fetch(tx)
	}, OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol))
}

func (s *SchemaManager) defineKinds(tx graph.Transaction, kinds graph.Kinds) error {
	for _, kind := range kinds {
		if kindID, err := query.On(tx).InsertOrGetKind(kind); err != nil {
			return err
		} else {
			s.kindsByID[kind] = kindID
			s.kindIDsByKind[kindID] = kind
		}
	}

	return nil
}

func (s *SchemaManager) mapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	var (
		missingKinds = make(graph.Kinds, 0, len(kinds))
		ids          = make([]int16, 0, len(kinds))
	)

	for _, kind := range kinds {
		if id, hasID := s.kindsByID[kind]; hasID {
			ids = append(ids, id)
		} else {
			missingKinds = append(missingKinds, kind)
		}
	}

	return ids, missingKinds
}

func (s *SchemaManager) MapKind(ctx context.Context, kind graph.Kind) (int16, error) {
	s.lock.RLock()

	if id, hasID := s.kindsByID[kind]; hasID {
		s.lock.RUnlock()
		return id, nil
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.Fetch(ctx); err != nil {
		return -1, err
	}

	if id, hasID := s.kindsByID[kind]; hasID {
		return id, nil
	} else {
		return -1, fmt.Errorf("unable to map kind: %s", kind.String())
	}
}

func (s *SchemaManager) MapKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error) {
	s.lock.RLock()

	if mappedKinds, missingKinds := s.mapKinds(kinds); len(missingKinds) == 0 {
		s.lock.RUnlock()
		return mappedKinds, nil
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.Fetch(ctx); err != nil {
		return nil, err
	}

	if mappedKinds, missingKinds := s.mapKinds(kinds); len(missingKinds) == 0 {
		return mappedKinds, nil
	} else {
		return nil, fmt.Errorf("unable to map kinds: %s", strings.Join(missingKinds.Strings(), ", "))
	}
}
func (s *SchemaManager) ReadTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	if cfg, err := renderConfig(batchWriteSize, readOnlyTxOptions, options); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		return txDelegate(&transaction{
			schemaManager:   s,
			queryExecMode:   cfg.QueryExecMode,
			ctx:             ctx,
			conn:            conn,
			targetSchemaSet: false,
		})
	}
}

func (s *SchemaManager) mapKindIDs(kindIDs []int16) (graph.Kinds, []int16) {
	var (
		missingIDs = make([]int16, 0, len(kindIDs))
		kinds      = make(graph.Kinds, 0, len(kindIDs))
	)

	for _, kindID := range kindIDs {
		if kind, hasKind := s.kindIDsByKind[kindID]; hasKind {
			kinds = append(kinds, kind)
		} else {
			missingIDs = append(missingIDs, kindID)
		}
	}

	return kinds, missingIDs
}

func (s *SchemaManager) MapKindID(ctx context.Context, kindID int16) (graph.Kind, error) {
	if kindIDs, err := s.MapKindIDs(ctx, kindID); err != nil {
		return nil, err
	} else {
		return kindIDs[0], nil
	}
}

func (s *SchemaManager) MapKindIDs(ctx context.Context, kindIDs ...int16) (graph.Kinds, error) {
	s.lock.RLock()

	if kinds, missingKinds := s.mapKindIDs(kindIDs); len(missingKinds) == 0 {
		s.lock.RUnlock()
		return kinds, nil
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.Fetch(ctx); err != nil {
		return nil, err
	}

	if kinds, missingKinds := s.mapKindIDs(kindIDs); len(missingKinds) == 0 {
		return kinds, nil
	} else {
		return nil, fmt.Errorf("unable to map kind ids: %v", missingKinds)
	}
}

func (s *SchemaManager) assertKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error) {
	// Acquire a write-lock and release on-exit
	s.lock.Lock()
	defer s.lock.Unlock()

	// We have to re-acquire the missing kinds since there's a potential for another writer to acquire the write-lock
	// in between release of the read-lock and acquisition of the write-lock for this operation
	if _, missingKinds := s.mapKinds(kinds); len(missingKinds) > 0 {
		if err := s.WriteTransaction(ctx, func(tx graph.Transaction) error {
			return s.defineKinds(tx, missingKinds)
		}, OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol)); err != nil {
			return nil, err
		}
	}

	// Lookup the kinds again from memory as they should now be up to date
	kindIDs, _ := s.mapKinds(kinds)
	return kindIDs, nil
}

func (s *SchemaManager) AssertKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing any kind definitions
	s.lock.RLock()

	if kindIDs, missingKinds := s.mapKinds(kinds); len(missingKinds) == 0 {
		// All kinds are defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return kindIDs, nil
	}

	// Release the read-lock here so that we can acquire a write-lock
	s.lock.RUnlock()
	return s.assertKinds(ctx, kinds)
}

func (s *SchemaManager) setDefaultGraph(defaultGraph model.Graph, schema graph.Graph) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.hasDefaultGraph {
		// Another actor has already asserted or otherwise set a default graph
		return
	}

	s.graphs[schema.Name] = defaultGraph

	s.defaultGraph = defaultGraph
	s.hasDefaultGraph = true
}

func (s *SchemaManager) SetDefaultGraph(ctx context.Context, schema graph.Graph) error {
	return s.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Validate the schema if the graph already exists in the database
		if graphModel, err := query.On(tx).SelectGraphByName(schema.Name); err != nil {
			return err
		} else {
			s.setDefaultGraph(graphModel, schema)
			return nil
		}
	})
}

func (s *SchemaManager) AssertDefaultGraph(ctx context.Context, schema graph.Graph) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if graphModel, err := s.AssertGraph(tx, schema); err != nil {
			return err
		} else {
			s.setDefaultGraph(graphModel, schema)
		}

		return nil
	})
}

func (s *SchemaManager) DefaultGraph() (model.Graph, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.defaultGraph, s.hasDefaultGraph
}

func (s *SchemaManager) assertGraph(tx graph.Transaction, schema graph.Graph) (model.Graph, error) {
	var assertedGraph model.Graph

	// Validate the schema if the graph already exists in the database
	if definition, err := query.On(tx).SelectGraphByName(schema.Name); err != nil {
		// ErrNoRows is ignored as it signifies that this graph must be created
		if !errors.Is(err, pgx.ErrNoRows) {
			return model.Graph{}, err
		}

		if newDefinition, err := query.On(tx).CreateGraph(schema); err != nil {
			return model.Graph{}, err
		} else {
			assertedGraph = newDefinition
		}
	} else if assertedDefinition, err := query.On(tx).AssertGraph(schema, definition); err != nil {
		return model.Graph{}, err
	} else {
		// Graph existed and may have been updated
		assertedGraph = assertedDefinition
	}

	// Cache the graph definition and return it
	s.graphs[schema.Name] = assertedGraph
	return assertedGraph, nil
}

func (s *SchemaManager) AssertGraph(tx graph.Transaction, schema graph.Graph) (model.Graph, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing the graph definitions
	s.lock.RLock()

	if graphInstance, isDefined := s.graphs[schema.Name]; isDefined {
		// The graph is defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return graphInstance, nil
	}

	// Release the read-lock here so that we can acquire a write-lock next
	s.lock.RUnlock()

	s.lock.Lock()
	defer s.lock.Unlock()

	if graphInstance, isDefined := s.graphs[schema.Name]; isDefined {
		// The graph was defined by a different actor between the read unlock and the write lock, return it
		return graphInstance, nil
	}

	return s.assertGraph(tx, schema)
}

func (s *SchemaManager) assertSchema(tx graph.Transaction, schema graph.Schema) error {
	if err := query.On(tx).CreateSchema(); err != nil {
		return err
	}

	if err := s.fetch(tx); err != nil {
		return err
	}

	for _, graphSchema := range schema.Graphs {
		if _, missingKinds := s.mapKinds(graphSchema.Nodes); len(missingKinds) > 0 {
			if err := s.defineKinds(tx, missingKinds); err != nil {
				return err
			}
		}

		if _, missingKinds := s.mapKinds(graphSchema.Edges); len(missingKinds) > 0 {
			if err := s.defineKinds(tx, missingKinds); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SchemaManager) AssertSchema(ctx context.Context, schema graph.Schema) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return s.assertSchema(tx, schema)
	}, OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol))
}
