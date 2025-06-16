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

package cypher_test

import (
	"testing"

	model2 "github.com/specterops/bloodhound/cypher/models/cypher"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func validateCopy(t *testing.T, actual any) {
	require.Equal(t, actual, model2.Copy(actual))
}

func int64Pointer(value int64) *int64 {
	return &value
}

func TestCopy(t *testing.T) {
	t.Parallel()
	validateCopy(t, &model2.RegularQuery{})
	validateCopy(t, &model2.SingleQuery{})
	validateCopy(t, &model2.SinglePartQuery{
		ReadingClauses: []*model2.ReadingClause{{
			Match: &model2.Match{
				Pattern: nil,
				Where:   nil,
			},
			Unwind: nil,
		}},
		UpdatingClauses: nil,
		Return:          nil,
	})

	validateCopy(t, &model2.MultiPartQuery{})
	validateCopy(t, &model2.MultiPartQuery{
		Parts: []*model2.MultiPartQueryPart{{
			ReadingClauses: []*model2.ReadingClause{{
				Match: &model2.Match{
					Optional: true,
					Pattern: []*model2.PatternPart{{
						Variable:                model2.NewVariableWithSymbol("p"),
						ShortestPathPattern:     true,
						AllShortestPathsPattern: true,
						PatternElements:         []*model2.PatternElement{},
					}},
				},
			}},
		}},
		SinglePartQuery: &model2.SinglePartQuery{
			ReadingClauses: []*model2.ReadingClause{{
				Match: &model2.Match{
					Pattern: nil,
					Where:   nil,
				},
				Unwind: nil,
			}},
			UpdatingClauses: nil,
			Return:          nil,
		},
	})

	validateCopy(t, &model2.IDInCollection{})
	validateCopy(t, &model2.FilterExpression{})
	validateCopy(t, &model2.Quantifier{})

	validateCopy(t, &model2.MultiPartQueryPart{})
	validateCopy(t, &model2.Remove{})
	validateCopy(t, &model2.ArithmeticExpression{})
	validateCopy(t, &model2.PartialArithmeticExpression{
		Operator: model2.OperatorAdd,
	})
	validateCopy(t, &model2.Parenthetical{})
	validateCopy(t, &model2.Comparison{})
	validateCopy(t, &model2.PartialComparison{
		Operator: model2.OperatorAdd,
	})
	validateCopy(t, &model2.SetItem{
		Operator: model2.OperatorAdditionAssignment,
	})
	validateCopy(t, &model2.Order{})
	validateCopy(t, &model2.Skip{})
	validateCopy(t, &model2.Limit{})
	validateCopy(t, &model2.RemoveItem{})
	validateCopy(t, &model2.Comparison{})
	validateCopy(t, &model2.FunctionInvocation{
		Distinct:  true,
		Namespace: []string{"a", "b", "c"},
		Name:      "d",
	})
	validateCopy(t, &model2.Variable{
		Symbol: "A",
	})
	validateCopy(t, &model2.Parameter{
		Symbol: "B",
	})
	validateCopy(t, &model2.Literal{
		Value: "1234",
		Null:  false,
	})
	validateCopy(t, &model2.Literal{
		Null: true,
	})
	validateCopy(t, &model2.Projection{
		Distinct: true,
		All:      true,
	})
	validateCopy(t, &model2.ProjectionItem{})
	validateCopy(t, &model2.PropertyLookup{
		Symbol: "a",
	})
	validateCopy(t, &model2.Set{})
	validateCopy(t, &model2.Delete{
		Detach: true,
	})
	validateCopy(t, &model2.Create{
		Unique: true,
	})
	validateCopy(t, &model2.KindMatcher{})
	validateCopy(t, &model2.Conjunction{})
	validateCopy(t, &model2.Disjunction{})
	validateCopy(t, &model2.ExclusiveDisjunction{})
	validateCopy(t, &model2.PatternPart{
		Variable:                model2.NewVariableWithSymbol("p"),
		ShortestPathPattern:     true,
		AllShortestPathsPattern: true,
	})
	validateCopy(t, &model2.PatternElement{})
	validateCopy(t, &model2.Negation{})
	validateCopy(t, &model2.NodePattern{
		Variable: model2.NewVariableWithSymbol("n"),
	})
	validateCopy(t, &model2.RelationshipPattern{
		Variable:  model2.NewVariableWithSymbol("r"),
		Direction: graph.DirectionOutbound,
	})
	validateCopy(t, &model2.PatternRange{
		StartIndex: int64Pointer(1234),
	})
	validateCopy(t, &model2.UpdatingClause{})
	validateCopy(t, &model2.SortItem{
		Ascending: true,
	})
	validateCopy(t, []*model2.PatternPart{})
	validateCopy(t, &model2.PatternPredicate{
		PatternElements: []*model2.PatternElement{{}},
	})

	// External types
	validateCopy(t, []string{})
	validateCopy(t, graph.Kinds{})
}
