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
	"bytes"
	"context"
	"github.com/specterops/bloodhound/cypher/backend/pgsql"
	"github.com/specterops/bloodhound/cypher/backend/pgsql/pgtransition"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

type liveQuery struct {
	ctx          context.Context
	tx           graph.Transaction
	kindMapper   KindMapper
	emitter      *pgsql.Emitter
	parameters   map[string]any
	queryBuilder *query.Builder
}

func newLiveQuery(ctx context.Context, tx graph.Transaction, kindMapper KindMapper) liveQuery {
	return liveQuery{
		ctx:          ctx,
		tx:           tx,
		kindMapper:   kindMapper,
		emitter:      pgsql.NewEmitter(false, kindMapper),
		parameters:   map[string]any{},
		queryBuilder: query.NewBuilder(nil),
	}
}

func (s *liveQuery) runAllShortestPathsQuery() graph.Result {
	if aspArguments, err := pgtransition.TranslateAllShortestPaths(s.queryBuilder.RegularQuery(), s.kindMapper); err != nil {
		return graph.NewErrorResult(err)
	} else {
		return s.tx.Raw(`select (t.nodes, t.edges)::pathComposite from all_shortest_paths(@p1, @p2, @p3, @p4) as t`, map[string]any{
			"p1": aspArguments.RootCriteria,
			"p2": aspArguments.TraversalCriteria,
			"p3": aspArguments.TerminalCriteria,
			"p4": aspArguments.MaxDepth,
		})
	}
}

func (s *liveQuery) runRegularQuery() graph.Result {
	buffer := &bytes.Buffer{}

	if regularQuery, err := s.queryBuilder.Build(); err != nil {
		return graph.NewErrorResult(err)
	} else if arguments, err := pgsql.Translate(regularQuery, s.kindMapper); err != nil {
		return graph.NewErrorResult(err)
	} else if err := s.emitter.Write(regularQuery, buffer); err != nil {
		return graph.NewErrorResult(err)
	} else {
		return s.tx.Raw(buffer.String(), arguments)
	}
}

func (s *liveQuery) Query(delegate func(results graph.Result) error, finalCriteria ...graph.Criteria) error {
	for _, criteria := range finalCriteria {
		s.queryBuilder.Apply(criteria)
	}

	if result := s.runRegularQuery(); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
		return delegate(result)
	}
}

func (s *liveQuery) exec(finalCriteria ...graph.Criteria) error {
	return s.Query(func(results graph.Result) error {
		return results.Error()
	}, finalCriteria...)
}

type nodeQuery struct {
	liveQuery
}

func (s *nodeQuery) Filter(criteria graph.Criteria) graph.NodeQuery {
	s.queryBuilder.Apply(query.Where(criteria))
	return s
}

func (s *nodeQuery) Filterf(criteriaDelegate graph.CriteriaProvider) graph.NodeQuery {
	return s.Filter(criteriaDelegate())
}

func (s *nodeQuery) Delete() error {
	return s.exec(query.Delete(
		query.Node(),
	))
}

func (s *nodeQuery) Update(properties *graph.Properties) error {
	return s.exec(query.Updatef(func() graph.Criteria {
		var updateStatements []graph.Criteria

		if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
			updateStatements = append(updateStatements, query.SetProperties(query.Node(), modifiedProperties))
		}

		if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
			updateStatements = append(updateStatements, query.DeleteProperties(query.Node(), deletedProperties...))
		}

		return updateStatements
	}))
}

func (s *nodeQuery) OrderBy(criteria ...graph.Criteria) graph.NodeQuery {
	s.queryBuilder.Apply(query.OrderBy(criteria...))
	return s
}

func (s *nodeQuery) Offset(offset int) graph.NodeQuery {
	s.queryBuilder.Apply(query.Offset(offset))
	return s
}

func (s *nodeQuery) Limit(limit int) graph.NodeQuery {
	s.queryBuilder.Apply(query.Limit(limit))
	return s
}

func (s *nodeQuery) Count() (int64, error) {
	var count int64

	return count, s.Query(func(results graph.Result) error {
		if !results.Next() {
			return graph.ErrNoResultsFound
		}

		return results.Scan(&count)
	}, query.Returning(
		query.Count(query.Node()),
	))
}

func (s *nodeQuery) First() (*graph.Node, error) {
	var node graph.Node

	return &node, s.Query(
		func(results graph.Result) error {
			if !results.Next() {
				return graph.ErrNoResultsFound
			}

			return results.Scan(&node)
		},
		query.Returning(
			query.Node(),
		),
		query.Limit(1),
	)
}

func (s *nodeQuery) Fetch(delegate func(cursor graph.Cursor[*graph.Node]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (*graph.Node, error) {
			var node graph.Node
			return &node, scanner.Scan(&node)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.Node(),
	))
}

func (s *nodeQuery) FetchIDs(delegate func(cursor graph.Cursor[graph.ID]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.ID, error) {
			var nodeID graph.ID
			return nodeID, scanner.Scan(&nodeID)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.NodeID(),
	))
}

func (s *nodeQuery) FetchKinds(delegate func(cursor graph.Cursor[graph.KindsResult]) error) error {
	return s.Query(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.KindsResult, error) {
			var (
				nodeID    graph.ID
				nodeKinds graph.Kinds
				err       = scanner.Scan(&nodeID, &nodeKinds)
			)

			return graph.KindsResult{
				ID:    nodeID,
				Kinds: nodeKinds,
			}, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.NodeID(),
		query.KindsOf(query.Node()),
	))
}
