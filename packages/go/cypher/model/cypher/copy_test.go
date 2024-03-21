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
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func validateCopy(t *testing.T, actual any) {
	require.Equal(t, actual, cypher.Copy(actual))
}

func int64Pointer(value int64) *int64 {
	return &value
}

func TestCopy(t *testing.T) {
	validateCopy(t, &cypher.RegularQuery{})
	validateCopy(t, &cypher.SingleQuery{})
	validateCopy(t, &cypher.SinglePartQuery{
		ReadingClauses: []*cypher.ReadingClause{{
			Match: &cypher.Match{
				Pattern: nil,
				Where:   nil,
			},
			Unwind: nil,
		}},
		UpdatingClauses: nil,
		Return:          nil,
	})

	validateCopy(t, &cypher.MultiPartQuery{})
	validateCopy(t, &cypher.MultiPartQuery{
		Parts: []*cypher.MultiPartQueryPart{{
			ReadingClauses: []*cypher.ReadingClause{{
				Match: &cypher.Match{
					Optional: true,
					Pattern: []*cypher.PatternPart{{
						Binding:                 cypher.NewVariableWithSymbol("p"),
						ShortestPathPattern:     true,
						AllShortestPathsPattern: true,
						PatternElements:         []*cypher.PatternElement{},
					}},
				},
			}},
		}},
		SinglePartQuery: &cypher.SinglePartQuery{
			ReadingClauses: []*cypher.ReadingClause{{
				Match: &cypher.Match{
					Pattern: nil,
					Where:   nil,
				},
				Unwind: nil,
			}},
			UpdatingClauses: nil,
			Return:          nil,
		},
	})

	validateCopy(t, &cypher.IDInCollection{})
	validateCopy(t, &cypher.FilterExpression{})
	validateCopy(t, &cypher.Quantifier{})

	validateCopy(t, &cypher.MultiPartQueryPart{})
	validateCopy(t, &cypher.Remove{})
	validateCopy(t, &cypher.ArithmeticExpression{})
	validateCopy(t, &cypher.PartialArithmeticExpression{
		Operator: cypher.OperatorAdd,
	})
	validateCopy(t, &cypher.Parenthetical{})
	validateCopy(t, &cypher.Comparison{})
	validateCopy(t, &cypher.PartialComparison{
		Operator: cypher.OperatorAdd,
	})
	validateCopy(t, &cypher.SetItem{
		Operator: cypher.OperatorAdditionAssignment,
	})
	validateCopy(t, &cypher.Order{})
	validateCopy(t, &cypher.Skip{})
	validateCopy(t, &cypher.Limit{})
	validateCopy(t, &cypher.RemoveItem{})
	validateCopy(t, &cypher.Comparison{})
	validateCopy(t, &cypher.FunctionInvocation{
		Distinct:  true,
		Namespace: []string{"a", "b", "c"},
		Name:      "d",
	})
	validateCopy(t, &cypher.Variable{
		Symbol: "A",
	})
	validateCopy(t, &cypher.Parameter{
		Symbol: "B",
	})
	validateCopy(t, &cypher.Literal{
		Value: "1234",
		Null:  false,
	})
	validateCopy(t, &cypher.Literal{
		Null: true,
	})
	validateCopy(t, &cypher.Projection{
		Distinct: true,
	})
	validateCopy(t, &cypher.ProjectionItem{})
	validateCopy(t, &cypher.PropertyLookup{
		Symbols: []string{"a", "b", "c"},
	})
	validateCopy(t, &cypher.Set{})
	validateCopy(t, &cypher.Delete{
		Detach: true,
	})
	validateCopy(t, &cypher.Create{
		Unique: true,
	})
	validateCopy(t, &cypher.KindMatcher{})
	validateCopy(t, &cypher.Conjunction{})
	validateCopy(t, &cypher.Disjunction{})
	validateCopy(t, &cypher.ExclusiveDisjunction{})
	validateCopy(t, &cypher.PatternPart{
		Binding:                 cypher.NewVariableWithSymbol("p"),
		ShortestPathPattern:     true,
		AllShortestPathsPattern: true,
	})
	validateCopy(t, &cypher.PatternElement{})
	validateCopy(t, &cypher.Negation{})
	validateCopy(t, &cypher.NodePattern{
		Binding: cypher.NewVariableWithSymbol("n"),
	})
	validateCopy(t, &cypher.RelationshipPattern{
		Binding:   cypher.NewVariableWithSymbol("r"),
		Direction: graph.DirectionOutbound,
	})
	validateCopy(t, &cypher.PatternRange{
		StartIndex: int64Pointer(1234),
	})
	validateCopy(t, &cypher.UpdatingClause{})
	validateCopy(t, &cypher.SortItem{
		Ascending: true,
	})
	validateCopy(t, []*cypher.PatternPart{})
	validateCopy(t, &cypher.PatternPredicate{
		PatternElements: []*cypher.PatternElement{{}},
	})

	// External types
	validateCopy(t, []string{})
	validateCopy(t, graph.Kinds{})
}
