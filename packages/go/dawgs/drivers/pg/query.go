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

	"github.com/specterops/bloodhound/cypher/models/pgsql/translate"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

type liveQuery struct {
	ctx          context.Context
	tx           graph.Transaction
	kindMapper   KindMapper
	queryBuilder *query.Builder
}

func newLiveQuery(ctx context.Context, tx graph.Transaction, kindMapper KindMapper) liveQuery {
	return liveQuery{
		ctx:          ctx,
		tx:           tx,
		kindMapper:   kindMapper,
		queryBuilder: query.NewBuilder(nil),
	}
}

func (s *liveQuery) runRegularQuery(allShortestPaths bool) graph.Result {
	if regularQuery, err := s.queryBuilder.Build(allShortestPaths); err != nil {
		return graph.NewErrorResult(err)
	} else if translation, err := translate.FromCypher(s.ctx, regularQuery, s.kindMapper, false); err != nil {
		return graph.NewErrorResult(err)
	} else {
		return s.tx.Raw(translation.Statement, translation.Parameters)
	}
}

func (s *liveQuery) Query(delegate func(results graph.Result) error, finalCriteria ...graph.Criteria) error {
	for _, criteria := range finalCriteria {
		s.queryBuilder.Apply(criteria)
	}

	if result := s.runRegularQuery(false); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
		return delegate(result)
	}
}

func (s *liveQuery) QueryAllShortestPaths(delegate func(results graph.Result) error, finalCriteria ...graph.Criteria) error {
	for _, criteria := range finalCriteria {
		s.queryBuilder.Apply(criteria)
	}

	if result := s.runRegularQuery(true); result.Error() != nil {
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
