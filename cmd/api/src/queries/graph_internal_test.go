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

package queries

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_ApplyTimeoutReduction(t *testing.T) {
	// Query Weight			Reduction Factor 		  Runtime
	// 	0-4						1						x
	// 	5-9						2						x/2
	//	10-14					3						x/3
	//	15-19					4						x/4
	//	20-24					5						x/5
	//	25-29					6						x/6
	//	30-34					7						x/7
	//	35-39					8						x/8
	//	40-44					9						x/9
	//	45-49					10						x/10
	// 	50						11						x/11
	//	>50						Too complex

	var (
		inputRuntime      = 15 * time.Minute
		expectedReduction int64
	)

	// Start with weight of 2, increase by 5 in each iteration until reduction factor = 11
	// This will run the function and assess the results for each range of permissible query
	// weights, against their respective expected reduction factor and runtime.
	weight := int64(2)
	for expectedReduction = 1; expectedReduction < 12; expectedReduction++ {
		expectedRuntime := int64(inputRuntime.Seconds()) / expectedReduction
		reducedRuntime, reduction := applyTimeoutReduction(weight, inputRuntime)

		require.Equal(t, expectedReduction, reduction)
		require.Equal(t, expectedRuntime, int64(reducedRuntime.Seconds()))

		weight += 5
	}
}

const cacheKey = "ad-entity-query_queryName_objectID_1"

func Test_runMaybeCachedEntityQuery(t *testing.T) {
	var (
		mockCtrl         = gomock.NewController(t)
		mockDB           = graph_mocks.NewMockDatabase(mockCtrl)
		node             = graph.NewNode(0, graph.NewProperties(), graph.StringKind("kind"))
		ctx              = context.Background()
		happyQueryName   = "happyQuery"
		happyObjectID    = "happyObjectID"
		failureQueryName = "failureQuery"
		failureObjectID  = "failureObjectID"

		// Ensure we can match on the exact error from the delegate to ensure that none of the intermediate layers are
		// producing errors unexpectedly
		failureDelegateErr = errors.New("failed")
		failureDelegate    = func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
			return nil, failureDelegateErr
		}

		happyPathDelegateReturn = graph.NewNodeSet()
		happyPathDelegate       = func(tx graph.Transaction, node *graph.Node, skip, limit int) (graph.NodeSet, error) {
			return happyPathDelegateReturn, nil
		}
	)
	defer mockCtrl.Finish()

	cacheInstance, err := cache.NewCache(cache.Config{MaxSize: 100})
	require.Nil(t, err)
	graphQueryInst := &GraphQuery{
		Graph:              mockDB,
		Cache:              cacheInstance,
		SlowQueryThreshold: 0,
	}

	mockDB.EXPECT().ReadTransaction(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
		return txDelegate(nil)
	}).Times(2)

	t.Run("runMaybeCachedEntityQuery with delegate failure", func(t *testing.T) {
		_, err = graphQueryInst.runMaybeCachedEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     failureQueryName,
			ObjectID:      failureObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  failureDelegate,
		}, true)

		require.Equal(t, failureDelegateErr, err)
	})

	t.Run("runMaybeCachedEntityQuery happy path with no cached results", func(t *testing.T) {
		result, err := graphQueryInst.runMaybeCachedEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     happyQueryName,
			ObjectID:      happyObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  happyPathDelegate,
		}, true)

		require.Nil(t, err)
		// Result set is empty so assert on that
		require.Len(t, result, 0)
	})

	t.Run("runMaybeCachedEntityQuery happy path with cached results", func(t *testing.T) {
		key := fmt.Sprintf("ad-entity-query_%s_%s_%d", happyQueryName, happyObjectID, model.DataTypeList)
		cacheInstance.Set(key, graph.NodeSet{})

		result, err := graphQueryInst.runMaybeCachedEntityQuery(context.Background(), node, EntityQueryParameters{
			QueryName:     happyQueryName,
			ObjectID:      happyObjectID,
			RequestedType: model.DataTypeList,
			ListDelegate:  happyPathDelegate,
		}, true)

		require.Nil(t, err)
		// Result set is empty so assert on that
		require.Len(t, result, 0)
	})
}

func Test_cacheQueryResult(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		mockDB   = graph_mocks.NewMockDatabase(mockCtrl)
		result   = graph.NodeSet{}
	)

	cacheInstance, err := cache.NewCache(cache.Config{MaxSize: 100})
	require.Nil(t, err)

	graphQuery := &GraphQuery{
		Graph:              mockDB,
		SlowQueryThreshold: time.Minute.Milliseconds(),
		Cache:              cacheInstance,
	}

	// Happy path rejection for queries that run quick enough
	graphQuery.cacheQueryResult(time.Now().Add(-time.Second), cacheKey, result)

	// Happy path setting for queries that are slow enough
	graphQuery.cacheQueryResult(time.Now().Add(-time.Hour), cacheKey, result)

	// Force test for the error case
	graphQuery.cacheQueryResult(time.Now().Add(-time.Hour), cacheKey, result)

	// Force test for when the cache key is already set
	graphQuery.cacheQueryResult(time.Now().Add(-time.Hour), cacheKey, result)
}

func Test_sortAndSliceResults_sorting(t *testing.T) {
	var (
		matches = NodeSearchResults{
			ExactResults: []*graph.Node{
				graph.NewNode(1, graph.NewProperties().Set(common.Name.String(), "b@c.com"), ad.Entity)},
			FuzzyResults: []*graph.Node{
				graph.NewNode(2, graph.NewProperties().Set(common.Name.String(), "bab@c.com"), ad.Entity),
				graph.NewNode(3, graph.NewProperties().Set(common.Name.String(), "ab@c.com"), ad.Entity),
			},
		}
		skip     = 0
		limit    = 10
		expected = []*graph.Node{
			matches.ExactResults[0], matches.FuzzyResults[1], matches.FuzzyResults[0], // manually put fuzzyMatches' elements in alphabetical order for assertion
		}
	)

	actual := sortAndSliceResults(matches, limit, skip)

	require.Equal(t, 3, len(actual))
	require.Equal(t, actual, expected)
}

func Test_sortAndSliceResults_limit(t *testing.T) {
	var (
		matches = NodeSearchResults{
			ExactResults: []*graph.Node{
				graph.NewNode(1, graph.NewProperties().Set(common.Name.String(), "b@c.com"), ad.Entity),
				graph.NewNode(2, graph.NewProperties().Set(common.Name.String(), "b@c.com"), ad.Entity),
				graph.NewNode(3, graph.NewProperties().Set(common.Name.String(), "b@c.com"), ad.Entity),
			},
			FuzzyResults: []*graph.Node{
				graph.NewNode(4, graph.NewProperties().Set(common.Name.String(), "ab@c.com"), ad.Entity),
			},
		}
		skip     = 0
		limit    = 3
		expected = matches.ExactResults
	)

	actual := sortAndSliceResults(matches, limit, skip)

	require.Equal(t, 3, len(actual))
	require.Equal(t, actual, expected)
}

// extractOrComparisons unwraps a `query.Or(...)` criteria (the first entry returned by
// createFuzzyNodeSearchGraphCriteria/createNodeStartsWithSearchGraphCriteria) into its two
// underlying comparisons: the Name comparison and the ObjectID comparison, respectively.
func extractOrComparisons(t *testing.T, criteria graph.Criteria) (*cypher.Comparison, *cypher.Comparison) {
	t.Helper()

	parenthetical, ok := criteria.(*cypher.Parenthetical)
	require.True(t, ok, "expected *cypher.Parenthetical, got %T", criteria)

	disjunction, ok := parenthetical.Expression.(*cypher.Disjunction)
	require.True(t, ok, "expected *cypher.Disjunction, got %T", parenthetical.Expression)
	require.Len(t, disjunction.Expressions, 2)

	nameComparison, ok := disjunction.Expressions[0].(*cypher.Comparison)
	require.True(t, ok, "expected *cypher.Comparison, got %T", disjunction.Expressions[0])

	objectIDComparison, ok := disjunction.Expressions[1].(*cypher.Comparison)
	require.True(t, ok, "expected *cypher.Comparison, got %T", disjunction.Expressions[1])

	return nameComparison, objectIDComparison
}

// requireCaseSensitiveComparison asserts that comparison is a plain (case-sensitive) comparison
// of the form `n.<propertyName> <operator> <term>`.
func requireCaseSensitiveComparison(t *testing.T, comparison *cypher.Comparison, propertyName string, operator cypher.Operator, term string) {
	t.Helper()

	propertyLookup, ok := comparison.Left.(*cypher.PropertyLookup)
	require.True(t, ok, "expected *cypher.PropertyLookup, got %T", comparison.Left)
	require.Equal(t, propertyName, propertyLookup.Symbol)

	require.Len(t, comparison.Partials, 1)
	require.Equal(t, operator, comparison.Partials[0].Operator)

	parameter, ok := comparison.Partials[0].Right.(*cypher.Parameter)
	require.True(t, ok, "expected *cypher.Parameter, got %T", comparison.Partials[0].Right)
	require.Equal(t, term, parameter.Value)
}

// requireCaseInsensitiveComparison asserts that comparison is a case-insensitive comparison of
// the form `toLower(n.<propertyName>) <operator> toLower(<term>)`.
func requireCaseInsensitiveComparison(t *testing.T, comparison *cypher.Comparison, propertyName string, operator cypher.Operator, term string) {
	t.Helper()

	functionInvocation, ok := comparison.Left.(*cypher.FunctionInvocation)
	require.True(t, ok, "expected *cypher.FunctionInvocation, got %T", comparison.Left)
	require.Equal(t, "toLower", functionInvocation.Name)
	require.Len(t, functionInvocation.Arguments, 1)

	propertyLookup, ok := functionInvocation.Arguments[0].(*cypher.PropertyLookup)
	require.True(t, ok, "expected *cypher.PropertyLookup, got %T", functionInvocation.Arguments[0])
	require.Equal(t, propertyName, propertyLookup.Symbol)

	require.Len(t, comparison.Partials, 1)
	require.Equal(t, operator, comparison.Partials[0].Operator)

	parameter, ok := comparison.Partials[0].Right.(*cypher.Parameter)
	require.True(t, ok, "expected *cypher.Parameter, got %T", comparison.Partials[0].Right)
	require.Equal(t, strings.ToLower(term), parameter.Value)
}

// requireNotEqualsExclusion asserts that criteria is a `query.Not(query.Equals(n.<propertyName>, <term>))`
// clause, and that it remains case-sensitive and untouched by the term's original casing.
func requireNotEqualsExclusion(t *testing.T, criteria graph.Criteria, propertyName string, term string) {
	t.Helper()

	negation, ok := criteria.(*cypher.Negation)
	require.True(t, ok, "expected *cypher.Negation, got %T", criteria)

	parenthetical, ok := negation.Expression.(*cypher.Parenthetical)
	require.True(t, ok, "expected *cypher.Parenthetical, got %T", negation.Expression)

	comparison, ok := parenthetical.Expression.(*cypher.Comparison)
	require.True(t, ok, "expected *cypher.Comparison, got %T", parenthetical.Expression)

	requireCaseSensitiveComparison(t, comparison, propertyName, cypher.OperatorEquals, term)
}

func TestCreateFuzzyNodeSearchGraphCriteria_Search(t *testing.T) {
	const (
		nameTerm     = "abc"
		objectIDTerm = "xyz"
	)

	t.Run("flag off: contains criteria stays case-sensitive", func(t *testing.T) {
		filters := createFuzzyNodeSearchGraphCriteria(nil, nameTerm, objectIDTerm, false, false)
		require.Len(t, filters, 3)

		nameComparison, objectIDComparison := extractOrComparisons(t, filters[0])
		requireCaseSensitiveComparison(t, nameComparison, common.Name.String(), cypher.OperatorContains, nameTerm)
		requireCaseSensitiveComparison(t, objectIDComparison, common.ObjectID.String(), cypher.OperatorContains, objectIDTerm)

		requireNotEqualsExclusion(t, filters[1], common.Name.String(), nameTerm)
		requireNotEqualsExclusion(t, filters[2], common.ObjectID.String(), objectIDTerm)
	})

	t.Run("flag on: contains criteria becomes case-insensitive", func(t *testing.T) {
		filters := createFuzzyNodeSearchGraphCriteria(nil, nameTerm, objectIDTerm, false, true)
		require.Len(t, filters, 3)

		nameComparison, objectIDComparison := extractOrComparisons(t, filters[0])
		requireCaseInsensitiveComparison(t, nameComparison, common.Name.String(), cypher.OperatorContains, nameTerm)
		requireCaseInsensitiveComparison(t, objectIDComparison, common.ObjectID.String(), cypher.OperatorContains, objectIDTerm)

		// Exact-match exclusion clauses remain case-sensitive even when the flag is on.
		requireNotEqualsExclusion(t, filters[1], common.Name.String(), nameTerm)
		requireNotEqualsExclusion(t, filters[2], common.ObjectID.String(), objectIDTerm)
	})

	t.Run("includeGroupFilter and kinds are appended regardless of the flag", func(t *testing.T) {
		kinds := graph.Kinds{ad.Entity}

		for _, useRawObjectID := range []bool{false, true} {
			filters := createFuzzyNodeSearchGraphCriteria(kinds, nameTerm, objectIDTerm, true, useRawObjectID)
			require.Len(t, filters, 5)
			require.Equal(t, groupFilter, filters[3])

			kindIn, ok := filters[4].(*cypher.KindMatcher)
			require.True(t, ok, "expected *cypher.KindMatcher, got %T", filters[4])
			require.Equal(t, kinds, kindIn.Kinds)
		}
	})
}

func TestCreateNodeStartsWithSearchGraphCriteria_Search(t *testing.T) {
	const (
		nameTerm     = "abc"
		objectIDTerm = "xyz"
	)

	t.Run("flag off: starts-with criteria stays case-sensitive", func(t *testing.T) {
		filters := createNodeStartsWithSearchGraphCriteria(nil, nameTerm, objectIDTerm, false)
		require.Len(t, filters, 3)

		nameComparison, objectIDComparison := extractOrComparisons(t, filters[0])
		requireCaseSensitiveComparison(t, nameComparison, common.Name.String(), cypher.OperatorStartsWith, nameTerm)
		requireCaseSensitiveComparison(t, objectIDComparison, common.ObjectID.String(), cypher.OperatorStartsWith, objectIDTerm)

		requireNotEqualsExclusion(t, filters[1], common.Name.String(), nameTerm)
		requireNotEqualsExclusion(t, filters[2], common.ObjectID.String(), objectIDTerm)
	})

	t.Run("flag on: starts-with criteria becomes case-insensitive", func(t *testing.T) {
		filters := createNodeStartsWithSearchGraphCriteria(nil, nameTerm, objectIDTerm, true)
		require.Len(t, filters, 3)

		nameComparison, objectIDComparison := extractOrComparisons(t, filters[0])
		requireCaseInsensitiveComparison(t, nameComparison, common.Name.String(), cypher.OperatorStartsWith, nameTerm)
		requireCaseInsensitiveComparison(t, objectIDComparison, common.ObjectID.String(), cypher.OperatorStartsWith, objectIDTerm)

		// Exact-match exclusion clauses remain case-sensitive even when the flag is on.
		requireNotEqualsExclusion(t, filters[1], common.Name.String(), nameTerm)
		requireNotEqualsExclusion(t, filters[2], common.ObjectID.String(), objectIDTerm)
	})

	t.Run("kinds are appended regardless of the flag", func(t *testing.T) {
		kinds := graph.Kinds{ad.Entity}

		for _, useRawObjectID := range []bool{false, true} {
			filters := createNodeStartsWithSearchGraphCriteria(kinds, nameTerm, objectIDTerm, useRawObjectID)
			require.Len(t, filters, 4)

			kindIn, ok := filters[3].(*cypher.KindMatcher)
			require.True(t, ok, "expected *cypher.KindMatcher, got %T", filters[3])
			require.Equal(t, kinds, kindIn.Kinds)
		}
	})
}
