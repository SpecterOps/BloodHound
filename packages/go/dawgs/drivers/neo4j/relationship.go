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

	neo4j_core "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/query/neo4j"
)

func directionToReturnCriteria(direction graph.Direction) (graph.Criteria, error) {
	switch direction {
	case graph.DirectionInbound:
		// Select the relationship and the end node
		return query.Returning(
			query.Relationship(),
			query.End(),
		), nil

	case graph.DirectionOutbound:
		// Select the relationship and the start node
		return query.Returning(
			query.Relationship(),
			query.Start(),
		), nil

	default:
		return nil, fmt.Errorf("bad direction: %d", direction)
	}
}

func newRelationship(internalRelationship neo4j_core.Relationship) *graph.Relationship {
	propertiesInst := internalRelationship.Props

	if propertiesInst == nil {
		propertiesInst = make(map[string]any)
	}

	return graph.NewRelationship(
		graph.ID(internalRelationship.Id),
		graph.ID(internalRelationship.StartId),
		graph.ID(internalRelationship.EndId),
		graph.AsProperties(propertiesInst),
		graph.StringKind(internalRelationship.Type),
	)
}

type RelationshipQuery struct {
	ctx          context.Context
	tx           innerTransaction
	queryBuilder *neo4j.QueryBuilder
}

func NewRelationshipQuery(ctx context.Context, tx innerTransaction) graph.RelationshipQuery {
	return &RelationshipQuery{
		ctx:          ctx,
		tx:           tx,
		queryBuilder: neo4j.NewEmptyQueryBuilder(),
	}
}

func (s *RelationshipQuery) run(statement string, parameters map[string]any) graph.Result {
	return s.tx.Raw(statement, parameters)
}

func (s *RelationshipQuery) Query(delegate func(results graph.Result) error, finalCriteria ...graph.Criteria) error {
	for _, criteria := range finalCriteria {
		s.queryBuilder.Apply(criteria)
	}

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if statement, err := s.queryBuilder.Render(); err != nil {
		return err
	} else if result := s.run(statement, s.queryBuilder.Parameters); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
		return delegate(result)
	}
}

func (s *RelationshipQuery) Debug() (string, map[string]any) {
	rendered, _ := s.queryBuilder.Render()
	return rendered, s.queryBuilder.Parameters
}

func (s *RelationshipQuery) Update(properties *graph.Properties) error {
	s.queryBuilder.Apply(query.Updatef(func() graph.Criteria {
		var updateStatements []graph.Criteria

		if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
			updateStatements = append(updateStatements, query.SetProperties(query.Relationship(), modifiedProperties))
		}

		if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
			updateStatements = append(updateStatements, query.DeleteProperties(query.Relationship(), deletedProperties...))
		}

		return updateStatements
	}))

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if cypherQuery, err := s.queryBuilder.Render(); err != nil {
		return graph.NewError(cypherQuery, err)
	} else {
		return s.run(cypherQuery, s.queryBuilder.Parameters).Error()
	}
}

func (s *RelationshipQuery) Delete() error {
	s.queryBuilder.Apply(query.Delete(
		query.Relationship(),
	))

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if statement, err := s.queryBuilder.Render(); err != nil {
		return err
	} else {
		return s.run(statement, s.queryBuilder.Parameters).Error()
	}
}

func (s *RelationshipQuery) OrderBy(criteria ...graph.Criteria) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.OrderBy(criteria...))
	return s
}

func (s *RelationshipQuery) Offset(offset int) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Offset(offset))
	return s
}

func (s *RelationshipQuery) Limit(limit int) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Limit(limit))
	return s
}

func (s *RelationshipQuery) Filter(criteria graph.Criteria) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Where(criteria))
	return s
}

func (s *RelationshipQuery) Filterf(criteriaDelegate graph.CriteriaProvider) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Where(criteriaDelegate()))
	return s
}

func (s *RelationshipQuery) Count() (int64, error) {
	var count int64

	return count, s.Query(func(results graph.Result) error {
		if !results.Next() {
			return graph.ErrNoResultsFound
		}

		return results.Scan(&count)
	}, query.Returning(
		query.Count(query.Relationship()),
	))
}

func (s *RelationshipQuery) FetchAllShortestPaths(delegate func(cursor graph.Cursor[graph.Path]) error) error {
	s.queryBuilder.Apply(query.Returning(
		query.Path(),
	))

	if err := s.queryBuilder.PrepareAllShortestPaths(); err != nil {
		return err
	} else if statement, err := s.queryBuilder.Render(); err != nil {
		return err
	} else if result := s.run(statement, s.queryBuilder.Parameters); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()

		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.Path, error) {
			var (
				nextPath graph.Path
				err      = scanner.Scan(&nextPath)
			)

			return nextPath, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}
}

func (s *RelationshipQuery) FetchTriples(delegate func(cursor graph.Cursor[graph.RelationshipTripleResult]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.RelationshipTripleResult, error) {
			var (
				startID        graph.ID
				relationshipID graph.ID
				endID          graph.ID
				err            = scanner.Scan(&startID, &relationshipID, &endID)
			)

			return graph.RelationshipTripleResult{
				ID:      relationshipID,
				StartID: startID,
				EndID:   endID,
			}, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.ReturningDistinct(
		query.StartID(),
		query.RelationshipID(),
		query.EndID(),
	))
}

func (s *RelationshipQuery) FetchKinds(delegate func(cursor graph.Cursor[graph.RelationshipKindsResult]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.RelationshipKindsResult, error) {
			var (
				startID          graph.ID
				relationshipID   graph.ID
				relationshipKind graph.Kind
				endID            graph.ID
				err              = scanner.Scan(&startID, &relationshipID, &relationshipKind, &endID)
			)

			return graph.RelationshipKindsResult{
				RelationshipTripleResult: graph.RelationshipTripleResult{
					ID:      relationshipID,
					StartID: startID,
					EndID:   endID,
				},
				Kind: relationshipKind,
			}, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.StartID(),
		query.RelationshipID(),
		query.KindsOf(query.Relationship()),
		query.EndID(),
	))
}

func (s *RelationshipQuery) First() (*graph.Relationship, error) {
	var relationship graph.Relationship

	return &relationship, s.Query(func(results graph.Result) error {
		if !results.Next() {
			return graph.ErrNoResultsFound
		}

		return results.Scan(&relationship)
	}, query.Returning(
		query.Relationship(),
	), query.Limit(1))
}

func (s *RelationshipQuery) Fetch(delegate func(cursor graph.Cursor[*graph.Relationship]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (*graph.Relationship, error) {
			var relationship graph.Relationship
			return &relationship, scanner.Scan(&relationship)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.Relationship(),
	))
}

func (s *RelationshipQuery) FetchDirection(direction graph.Direction, delegate func(cursor graph.Cursor[graph.DirectionalResult]) error) error {
	if returnCriteria, err := directionToReturnCriteria(direction); err != nil {
		return err
	} else {
		return s.Query(func(result graph.Result) error {
			cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.DirectionalResult, error) {
				var (
					relationship graph.Relationship
					node         graph.Node
				)

				if err := scanner.Scan(&relationship, &node); err != nil {
					return graph.DirectionalResult{}, err
				}

				return graph.DirectionalResult{
					Direction:    direction,
					Relationship: &relationship,
					Node:         &node,
				}, nil
			})

			defer cursor.Close()
			return delegate(cursor)
		}, returnCriteria)
	}
}

func (s *RelationshipQuery) FetchIDs(delegate func(cursor graph.Cursor[graph.ID]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.ID, error) {
			var relationshipID graph.ID
			return relationshipID, scanner.Scan(&relationshipID)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.RelationshipID(),
	))
}
