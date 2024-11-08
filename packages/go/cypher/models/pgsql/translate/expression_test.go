// Copyright 2024 Specter Ops, Inc.
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

package translate_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
	"github.com/specterops/bloodhound/cypher/models/pgsql/translate"
	"github.com/stretchr/testify/require"
)

func mustAsLiteral(value any) pgsql.Literal {
	if literal, err := pgsql.AsLiteral(value); err != nil {
		panic(fmt.Sprintf("%v", err))
	} else {
		return literal
	}
}

func TestInferExpressionType(t *testing.T) {
	type testCase struct {
		ExpectedType pgsql.DataType
		Expression   pgsql.Expression
		Exclusive    bool
	}

	testCases := []testCase{{
		ExpectedType: pgsql.Boolean,
		Expression: pgsql.NewBinaryExpression(
			pgsql.NewPropertyLookup(
				pgsql.CompoundIdentifier{"n", "properties"},
				mustAsLiteral("field_a"),
			),
			pgsql.OperatorAnd,
			pgsql.NewBinaryExpression(
				mustAsLiteral("123"),
				pgsql.OperatorIn,
				pgsql.ArrayLiteral{
					Values:   []pgsql.Expression{mustAsLiteral("a"), mustAsLiteral("b")},
					CastType: pgsql.TextArray,
				},
			),
		),
	}, {
		ExpectedType: pgsql.Boolean,
		Expression: pgsql.NewBinaryExpression(
			pgsql.NewPropertyLookup(
				pgsql.CompoundIdentifier{"n", "properties"},
				mustAsLiteral("field_a"),
			),
			pgsql.OperatorAnd,
			pgsql.NewPropertyLookup(
				pgsql.CompoundIdentifier{"n", "properties"},
				mustAsLiteral("field_b"),
			),
		),
	}, {
		ExpectedType: pgsql.Text,
		Expression: pgsql.NewBinaryExpression(
			mustAsLiteral("123"),
			pgsql.OperatorIn,
			pgsql.ArrayLiteral{
				Values:   []pgsql.Expression{mustAsLiteral("a"), mustAsLiteral("b")},
				CastType: pgsql.TextArray,
			},
		),
	}, {
		ExpectedType: pgsql.Text,
		Expression: pgsql.NewBinaryExpression(
			mustAsLiteral("123"),
			pgsql.OperatorConcatenate,
			mustAsLiteral("456"),
		),
	}, {
		ExpectedType: pgsql.Int8,
		Expression: pgsql.NewBinaryExpression(
			mustAsLiteral(123),
			pgsql.OperatorAdd,
			pgsql.NewBinaryExpression(
				mustAsLiteral(123),
				pgsql.OperatorMultiply,
				mustAsLiteral(1),
			),
		),
	}, {
		ExpectedType: pgsql.Int8,
		Expression: pgsql.NewBinaryExpression(
			mustAsLiteral(123),
			pgsql.OperatorAdd,
			pgsql.NewBinaryExpression(
				mustAsLiteral(int16(123)),
				pgsql.OperatorMultiply,
				mustAsLiteral(int16(1)),
			),
		),
	}, {
		ExpectedType: pgsql.Int4,
		Expression: pgsql.NewBinaryExpression(
			pgsql.NewPropertyLookup(
				pgsql.CompoundIdentifier{"n", "properties"},
				mustAsLiteral("field"),
			),
			pgsql.OperatorAdd,
			pgsql.NewBinaryExpression(
				mustAsLiteral(int16(123)),
				pgsql.OperatorMultiply,
				mustAsLiteral(int32(1)),
			),
		),
	}}

	var (
		exclusive    []testCase
		hasExclusive bool
	)

	for _, nextCase := range testCases {
		if hasExclusive {
			if nextCase.Exclusive {
				exclusive = append(exclusive, nextCase)
			}
		} else if nextCase.Exclusive {
			hasExclusive = true

			exclusive = exclusive[:0]
			exclusive = append(exclusive, nextCase)
		} else {
			exclusive = append(exclusive, nextCase)
		}
	}

	for _, nextCase := range exclusive {
		if testName, err := format.Expression(nextCase.Expression, format.NewOutputBuilder()); err != nil {
			t.Fatalf("unable to format test case expression: %v", err)
		} else {
			t.Run(testName, func(t *testing.T) {
				inferredType, err := translate.InferExpressionType(nextCase.Expression)

				require.Nil(t, err)
				require.Equal(t, nextCase.ExpectedType, inferredType)
			})
		}
	}
}

func TestExpressionTreeTranslator(t *testing.T) {
	// Tree translator is a stack oriented expression tree builder
	var (
		treeTranslator = translate.NewExpressionTreeTranslator()
		scope          = translate.NewScope()
	)

	// Case: Translating the constraint: a.name = 'a' and a.num_a > 1 and b.name = 'b' and a.other = b.other

	// Perform a prefix visit of the parent expression and its operator. This is used for tracking
	// conjunctions and disjunctions.
	treeTranslator.PushOperator(pgsql.OperatorEquals)

	// Postfix visit and push the compound identifier first: a.name
	treeTranslator.Push(pgsql.CompoundIdentifier{"a", "name"})

	// Postfix visit and push the literal next: "a"
	treeTranslator.Push(mustAsLiteral("a"))

	// Perform a postfix visit of the parent expression and its operator.
	require.Nil(t, treeTranslator.PopPushOperator(scope, pgsql.OperatorEquals))

	// Expect one newly created binary expression to be the only thing left on the tree
	// translator's operand stack
	require.Equal(t, 1, treeTranslator.Depth())
	require.IsType(t, &pgsql.BinaryExpression{}, treeTranslator.Peek())

	// Continue with: and a.num_a > 1
	// Preform a prefix visit of the 'and' operator:
	treeTranslator.PushOperator(pgsql.OperatorAnd)

	// Preform a prefix visit of the '>' operator:
	treeTranslator.PushOperator(pgsql.OperatorGreaterThan)

	// Postfix visit and push the compound identifier first: a.num_a
	treeTranslator.Push(pgsql.CompoundIdentifier{"a", "num_a"})

	// Postfix visit and push the literal next: 1
	treeTranslator.Push(mustAsLiteral(1))

	// Perform a postfix visit of the parent expression and its operator.
	require.Nil(t, treeTranslator.PopPushOperator(scope, pgsql.OperatorGreaterThan))

	// Perform a postfix visit of the conjoining parent expression and its operator.
	require.Nil(t, treeTranslator.PopPushOperator(scope, pgsql.OperatorAnd))

	// Continue with: and b.name = "b"
	// Preform a prefix visit of the 'and' operator:
	treeTranslator.PushOperator(pgsql.OperatorAnd)

	// Preform a prefix visit of the '=' operator:
	treeTranslator.PushOperator(pgsql.OperatorEquals)

	// Postfix visit and push the compound identifier first: b.name
	treeTranslator.Push(pgsql.CompoundIdentifier{"b", "name"})

	// Postfix visit and push the literal next: "b"
	treeTranslator.Push(mustAsLiteral("b"))

	// Perform a postfix visit of the parent expression and its operator.
	require.Nil(t, treeTranslator.PopPushOperator(scope, pgsql.OperatorEquals))

	// Perform a postfix visit of the conjoining parent expression and its operator.
	require.Nil(t, treeTranslator.PopPushOperator(scope, pgsql.OperatorAnd))

	// Continue with: and a.other = b.other
	// enter Op(and), enter Op(=)
	treeTranslator.PushOperator(pgsql.OperatorAnd)
	treeTranslator.PushOperator(pgsql.OperatorEquals)

	// push LOperand, push ROperand
	treeTranslator.Push(pgsql.CompoundIdentifier{"a", "other"})
	treeTranslator.Push(pgsql.CompoundIdentifier{"b", "other"})

	// exit  exit Op(=), Op(and)
	treeTranslator.PopPushOperator(scope, pgsql.OperatorEquals)
	treeTranslator.PopPushOperator(scope, pgsql.OperatorAnd)

	// Assign remaining operands as constraints
	treeTranslator.ConstrainRemainingOperands()

	// Pull out the 'a' constraint
	aIdentifier := pgsql.AsIdentifierSet("a")
	expectedTranslation := "a.name = 'a' and a.num_a > 1"
	validateConstraints(t, treeTranslator, aIdentifier, expectedTranslation)

	// Pull out the 'b' constraint next
	bIdentifier := pgsql.AsIdentifierSet("b")
	expectedTranslation = "b.name = 'b'"
	validateConstraints(t, treeTranslator, bIdentifier, expectedTranslation)

	// Pull out the constraint that depends on both 'a' and 'b' identifiers
	idents := pgsql.AsIdentifierSet("a", "b")
	expectedTranslation = "a.other = b.other"
	validateConstraints(t, treeTranslator, idents, expectedTranslation)
}

func validateConstraints(t *testing.T, constraintTracker *translate.ExpressionTreeTranslator, idents *pgsql.IdentifierSet, expectedTranslation string) {
	constraint, err := constraintTracker.ConsumeSet(idents)

	require.NotNil(t, constraint)
	require.True(t, constraint.Dependencies.Matches(idents))
	require.Nil(t, err)

	formattedConstraint, err := format.Expression(constraint.Expression, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, expectedTranslation, formattedConstraint)
}
