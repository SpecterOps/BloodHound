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

package model_test

import (
	"testing"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func validateCopy(t *testing.T, actual any) {
	require.Equal(t, actual, model.Copy(actual))
}

func int64Pointer(value int64) *int64 {
	return &value
}

func TestCopy(t *testing.T) {
	validateCopy(t, &model.RegularQuery{})
	validateCopy(t, &model.SingleQuery{})
	validateCopy(t, &model.SinglePartQuery{
		ReadingClauses: []*model.ReadingClause{{
			Match: &model.Match{
				Pattern: nil,
				Where:   nil,
			},
			Unwind: nil,
		}},
		UpdatingClauses: nil,
		Return:          nil,
	})

	validateCopy(t, &model.MultiPartQuery{})
	validateCopy(t, &model.MultiPartQuery{
		Parts: []*model.MultiPartQueryPart{{
			ReadingClauses: []*model.ReadingClause{{
				Match: &model.Match{
					Optional: true,
					Pattern: []*model.PatternPart{{
						Binding:                 model.NewVariableWithSymbol("p"),
						ShortestPathPattern:     true,
						AllShortestPathsPattern: true,
						PatternElements:         []*model.PatternElement{},
					}},
				},
			}},
		}},
		SinglePartQuery: &model.SinglePartQuery{
			ReadingClauses: []*model.ReadingClause{{
				Match: &model.Match{
					Pattern: nil,
					Where:   nil,
				},
				Unwind: nil,
			}},
			UpdatingClauses: nil,
			Return:          nil,
		},
	})

	validateCopy(t, &model.IDInCollection{})
	validateCopy(t, &model.FilterExpression{})
	validateCopy(t, &model.Quantifier{})

	validateCopy(t, &model.MultiPartQueryPart{})
	validateCopy(t, &model.Remove{})
	validateCopy(t, &model.ArithmeticExpression{})
	validateCopy(t, &model.PartialArithmeticExpression{
		Operator: model.OperatorAdd,
	})
	validateCopy(t, &model.Parenthetical{})
	validateCopy(t, &model.Comparison{})
	validateCopy(t, &model.PartialComparison{
		Operator: model.OperatorAdd,
	})
	validateCopy(t, &model.SetItem{
		Operator: model.OperatorAdditionAssignment,
	})
	validateCopy(t, &model.Order{})
	validateCopy(t, &model.Skip{})
	validateCopy(t, &model.Limit{})
	validateCopy(t, &model.RemoveItem{})
	validateCopy(t, &model.Comparison{})
	validateCopy(t, &model.FunctionInvocation{
		Distinct:  true,
		Namespace: []string{"a", "b", "c"},
		Name:      "d",
	})
	validateCopy(t, &model.Variable{
		Symbol: "A",
	})
	validateCopy(t, &model.Parameter{
		Symbol: "B",
	})
	validateCopy(t, &model.Literal{
		Value: "1234",
		Null:  false,
	})
	validateCopy(t, &model.Literal{
		Null: true,
	})
	validateCopy(t, &model.Projection{
		Distinct: true,
		All:      true,
	})
	validateCopy(t, &model.ProjectionItem{})
	validateCopy(t, &model.PropertyLookup{
		Symbols: []string{"a", "b", "c"},
	})
	validateCopy(t, &model.Set{})
	validateCopy(t, &model.Delete{
		Detach: true,
	})
	validateCopy(t, &model.Create{
		Unique: true,
	})
	validateCopy(t, &model.KindMatcher{})
	validateCopy(t, &model.Conjunction{})
	validateCopy(t, &model.Disjunction{})
	validateCopy(t, &model.ExclusiveDisjunction{})
	validateCopy(t, &model.PatternPart{
		Binding:                 model.NewVariableWithSymbol("p"),
		ShortestPathPattern:     true,
		AllShortestPathsPattern: true,
	})
	validateCopy(t, &model.PatternElement{})
	validateCopy(t, &model.Negation{})
	validateCopy(t, &model.NodePattern{
		Binding: model.NewVariableWithSymbol("n"),
	})
	validateCopy(t, &model.RelationshipPattern{
		Binding:   model.NewVariableWithSymbol("r"),
		Direction: graph.DirectionOutbound,
	})
	validateCopy(t, &model.PatternRange{
		StartIndex: int64Pointer(1234),
	})
	validateCopy(t, &model.UpdatingClause{})
	validateCopy(t, &model.SortItem{
		Ascending: true,
	})
	validateCopy(t, []*model.PatternPart{})
	validateCopy(t, &model.PatternPredicate{
		PatternElements: []*model.PatternElement{{}},
	})

	// External types
	validateCopy(t, []string{})
	validateCopy(t, graph.Kinds{})
}
