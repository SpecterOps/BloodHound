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
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type KindMapper interface {
	MapKindID(kindID int16) (graph.Kind, bool)
	MapKindIDs(kindIDs ...int16) (graph.Kinds, []int16)
	MapKind(kind graph.Kind) (int16, bool)
	MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds)
	AssertKinds(tx graph.Transaction, kinds graph.Kinds) ([]int16, error)
}

type SchemaManager struct {
	defaultGraph    model.Graph
	hasDefaultGraph bool
	graphs          map[string]model.Graph
	kindsByID       map[graph.Kind]int16
	kindIDsByKind   map[int16]graph.Kind
	lock            *sync.RWMutex
}

func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		hasDefaultGraph: false,
		graphs:          map[string]model.Graph{},
		kindsByID:       map[graph.Kind]int16{},
		kindIDsByKind:   map[int16]graph.Kind{},
		lock:            &sync.RWMutex{},
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

func (s *SchemaManager) MapKind(kind graph.Kind) (int16, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	id, hasID := s.kindsByID[kind]
	return id, hasID
}

func (s *SchemaManager) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.mapKinds(kinds)
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

func (s *SchemaManager) MapKindID(kindID int16) (graph.Kind, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	kind, hasKind := s.kindIDsByKind[kindID]
	return kind, hasKind
}

func (s *SchemaManager) MapKindIDs(kindIDs ...int16) (graph.Kinds, []int16) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.mapKindIDs(kindIDs)
}

func (s *SchemaManager) AssertKinds(tx graph.Transaction, kinds graph.Kinds) ([]int16, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing any kind definitions
	s.lock.RLock()

	if kindIDs, missingKinds := s.mapKinds(kinds); len(missingKinds) == 0 {
		// All kinds are defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return kindIDs, nil
	}

	// Release the read-lock here so that we can acquire a write-lock
	s.lock.RUnlock()

	// Acquire a write-lock and release on-exit
	s.lock.Lock()
	defer s.lock.Unlock()

	// We have to re-acquire the missing kinds since there's a potential for another writer to acquire the write-lock
	// in between release of the read-lock and acquisition of the write-lock for this operation
	_, missingKinds := s.mapKinds(kinds)

	if err := s.defineKinds(tx, missingKinds); err != nil {
		return nil, err
	}

	kindIDs, _ := s.mapKinds(kinds)
	return kindIDs, nil
}

func (s *SchemaManager) SetDefaultGraph(tx graph.Transaction, schema graph.Graph) error {
	// Validate the schema if the graph already exists in the database
	if definition, err := query.On(tx).SelectGraphByName(schema.Name); err != nil {
		return err
	} else {
		s.graphs[schema.Name] = definition
		
		s.defaultGraph = definition
		s.hasDefaultGraph = true
	}

	return nil
}

func (s *SchemaManager) AssertDefaultGraph(tx graph.Transaction, schema graph.Graph) error {
	if graphInstance, err := s.AssertGraph(tx, schema); err != nil {
		return err
	} else {
		s.lock.Lock()
		defer s.lock.Unlock()

		s.defaultGraph = graphInstance
		s.hasDefaultGraph = true
	}

	return nil
}

func (s *SchemaManager) DefaultGraph() (model.Graph, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.defaultGraph, s.hasDefaultGraph
}

func (s *SchemaManager) AssertGraph(tx graph.Transaction, schema graph.Graph) (model.Graph, error) {
	// Acquire a read-lock first to fast-pass validate if we're missing the graph definitions
	s.lock.RLock()

	if graphInstance, isDefined := s.graphs[schema.Name]; isDefined {
		// The graph is defined. Release the read-lock here before returning
		s.lock.RUnlock()
		return graphInstance, nil
	}

	// Release the read-lock here so that we can acquire a write-lock
	s.lock.RUnlock()

	// Acquire a write-lock and create the graph definition
	s.lock.Lock()
	defer s.lock.Unlock()

	if graphInstance, isDefined := s.graphs[schema.Name]; isDefined {
		// The graph was defined by a different actor between the read unlock and the write lock.
		return graphInstance, nil
	}

	// Validate the schema if the graph already exists in the database
	if definition, err := query.On(tx).SelectGraphByName(schema.Name); err != nil {
		// ErrNoRows signifies that this graph must be created
		if !errors.Is(err, pgx.ErrNoRows) {
			return model.Graph{}, err
		}
	} else if assertedDefinition, err := query.On(tx).AssertGraph(schema, definition); err != nil {
		return model.Graph{}, err
	} else {
		s.graphs[schema.Name] = assertedDefinition
		return assertedDefinition, nil
	}

	// Create the graph
	if definition, err := query.On(tx).CreateGraph(schema); err != nil {
		return model.Graph{}, err
	} else {
		s.graphs[schema.Name] = definition
		return definition, nil
	}
}

func (s *SchemaManager) AssertSchema(tx graph.Transaction, schema graph.Schema) error {
	s.lock.Lock()
	defer s.lock.Unlock()

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
