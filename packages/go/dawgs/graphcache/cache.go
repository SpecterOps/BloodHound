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

package graphcache

import (
	"sync"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type Cache struct {
	nodes         *graph.IndexedSlice[graph.ID, *graph.Node]
	relationships *graph.IndexedSlice[graph.ID, *graph.Relationship]
	lock          *sync.RWMutex
}

func New() Cache {
	return Cache{
		nodes:         graph.NewIndexedSlice[graph.ID, *graph.Node](),
		relationships: graph.NewIndexedSlice[graph.ID, *graph.Relationship](),
		lock:          &sync.RWMutex{},
	}
}

func (s Cache) SizeOf() size.Size {
	// We're not including the total size of the lock here on purpose. Copying a locker can be dangerous and to get
	// the size of the backing struct for the pointer we would have to dereference and copy it.
	return size.Of(s) + s.nodes.SizeOf() + s.relationships.SizeOf()
}

func (s Cache) NodesCached() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.nodes.Len()
}

func (s Cache) RelationshipsCached() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.nodes.Len()
}

func (s Cache) GetNode(id graph.ID) *graph.Node {
	s.lock.RLock()
	defer s.lock.RUnlock()

	node, _ := s.nodes.CheckedGet(id)
	return node
}

func (s Cache) CheckedGetNode(id graph.ID) (*graph.Node, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.nodes.CheckedGet(id)
}

func (s Cache) GetNodes(ids []graph.ID) ([]*graph.Node, []graph.ID) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.nodes.GetAll(ids)
}

func (s Cache) PutNodes(nodes []*graph.Node) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, node := range nodes {
		s.nodes.Put(node.ID, node)
	}
}

func (s Cache) PutNodeSet(nodes graph.NodeSet) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, node := range nodes {
		s.nodes.Put(node.ID, node)
	}
}

func (s Cache) GetRelationship(id graph.ID) *graph.Relationship {
	s.lock.RLock()
	defer s.lock.RUnlock()

	relationship, _ := s.relationships.CheckedGet(id)
	return relationship
}

func (s Cache) CheckedGetRelationship(id graph.ID) (*graph.Relationship, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.relationships.CheckedGet(id)
}

func (s Cache) GetRelationships(ids []graph.ID) ([]*graph.Relationship, []graph.ID) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.relationships.GetAll(ids)
}

func (s Cache) PutRelationships(relationships []*graph.Relationship) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, relationship := range relationships {
		s.relationships.Put(relationship.ID, relationship)
	}
}

func (s Cache) PutRelationshipSet(relationships graph.RelationshipSet) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, relationship := range relationships {
		s.relationships.Put(relationship.ID, relationship)
	}
}

func fetchNodesByIDQuery(tx graph.Transaction, ids []graph.ID) graph.NodeQuery {
	return tx.Nodes().Filterf(func() graph.Criteria {
		return query.InIDs(query.NodeID(), ids...)
	})
}

func fetchRelationshipsByIDQuery(tx graph.Transaction, ids []graph.ID) graph.RelationshipQuery {
	return tx.Relationships().Filterf(func() graph.Criteria {
		return query.InIDs(query.RelationshipID(), ids...)
	})
}

func FetchNodesByID(tx graph.Transaction, cache Cache, ids []graph.ID) ([]*graph.Node, error) {
	var (
		cachedNodes, missingNodeIDs = cache.GetNodes(ids)
		toBeCachedCount             = 0
	)

	if len(missingNodeIDs) > 0 {
		if err := fetchNodesByIDQuery(tx, missingNodeIDs).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for next := range cursor.Chan() {
				cachedNodes = append(cachedNodes, next)
				toBeCachedCount++
			}

			return cursor.Error()
		}); err != nil {
			return nil, err
		}

		if toBeCachedCount > 0 {
			cache.PutNodes(cachedNodes[len(cachedNodes)-toBeCachedCount:])
		}

		if len(missingNodeIDs) != toBeCachedCount {
			return nil, graph.ErrMissingResultExpectation
		}
	}

	return cachedNodes, nil
}

func FetchRelationshipsByID(tx graph.Transaction, cache Cache, ids []graph.ID) (graph.RelationshipSet, error) {
	cachedRelationships, missingRelationshipIDs := cache.GetRelationships(ids)

	if len(missingRelationshipIDs) > 0 {
		if err := fetchRelationshipsByIDQuery(tx, ids).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for next := range cursor.Chan() {
				cachedRelationships = append(cachedRelationships, next)
			}

			return cursor.Error()
		}); err != nil {
			return nil, err
		}

		cache.PutRelationships(cachedRelationships[len(cachedRelationships)-len(missingRelationshipIDs):])
	}

	return graph.NewRelationshipSet(cachedRelationships...), nil
}

func ShallowFetchRelationships(cache Cache, graphQuery graph.RelationshipQuery) ([]*graph.Relationship, error) {
	var relationships []*graph.Relationship

	if err := graphQuery.FetchKinds(func(cursor graph.Cursor[graph.RelationshipKindsResult]) error {
		for next := range cursor.Chan() {
			relationships = append(relationships, graph.NewRelationship(next.ID, next.StartID, next.EndID, nil, next.Kind))
		}

		return cursor.Error()
	}); err != nil {
		return nil, err
	}

	cache.PutRelationships(relationships)
	return relationships, nil
}

func ShallowFetchNodesByID(tx graph.Transaction, cache Cache, ids []graph.ID) ([]*graph.Node, error) {
	cachedNodes, missingNodeIDs := cache.GetNodes(ids)

	if len(missingNodeIDs) > 0 {
		if err := fetchNodesByIDQuery(tx, missingNodeIDs).FetchKinds(func(cursor graph.Cursor[graph.KindsResult]) error {
			for next := range cursor.Chan() {
				cachedNodes = append(cachedNodes, graph.NewNode(next.ID, nil, next.Kinds...))
			}

			return cursor.Error()
		}); err != nil {
			return nil, err
		}

		cache.PutNodes(cachedNodes[len(cachedNodes)-len(missingNodeIDs):])
	}

	return cachedNodes, nil
}
