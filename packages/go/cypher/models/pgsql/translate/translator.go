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

package translate

import (
	"context"

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/walk"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type Translator struct {
	walk.HierarchicalVisitor[cypher.SyntaxNode]

	ctx            context.Context
	kindMapper     pgsql.KindMapper
	translation    Result
	treeTranslator *ExpressionTreeTranslator
	query          *Query
	scope          *Scope
}

func NewTranslator(ctx context.Context, kindMapper pgsql.KindMapper, parameters map[string]any) *Translator {
	if parameters == nil {
		parameters = map[string]any{}
	}

	return &Translator{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[cypher.SyntaxNode](),
		translation: Result{
			Parameters: parameters,
		},
		ctx:            ctx,
		kindMapper:     kindMapper,
		treeTranslator: NewExpressionTreeTranslator(),
		query:          &Query{},
		scope:          NewScope(),
	}
}

func (s *Translator) Enter(expression cypher.SyntaxNode) {
	switch typedExpression := expression.(type) {
	case *cypher.RegularQuery, *cypher.SingleQuery, *cypher.PatternElement,
		*cypher.Comparison, *cypher.Skip, *cypher.Limit, cypher.Operator, *cypher.ArithmeticExpression,
		*cypher.NodePattern, *cypher.RelationshipPattern, *cypher.Remove, *cypher.Set,
		*cypher.ReadingClause, *cypher.UnaryAddOrSubtractExpression, *cypher.PropertyLookup,
		*cypher.Negation, *cypher.Create, *cypher.Where, *cypher.ListLiteral,
		*cypher.FunctionInvocation, *cypher.Order, *cypher.RemoveItem, *cypher.SetItem,
		*cypher.MapItem, *cypher.UpdatingClause, *cypher.Delete, *cypher.With,
		*cypher.Return, *cypher.MultiPartQuery, *cypher.Properties:

	case *cypher.MultiPartQueryPart:
		if err := s.prepareMultiPartQueryPart(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.SinglePartQuery:
		if err := s.prepareSinglePartQueryPart(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Match:
		s.query.CurrentPart().currentPattern = &Pattern{}

	case graph.Kinds:
		s.treeTranslator.PushOperand(pgsql.KindListLiteral{
			Values: typedExpression,
		})

	case *cypher.KindMatcher:
		if err := s.translateKindMatcher(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Parameter:
		var (
			cypherIdentifier = pgsql.Identifier(typedExpression.Symbol)
			binding, bound   = s.scope.AliasedLookup(cypherIdentifier)
		)

		if !bound {
			if parameterBinding, err := s.scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
				s.SetError(err)
			} else {
				// Alias the old parameter identifier to the synthetic one
				if cypherIdentifier != "" {
					s.scope.Alias(cypherIdentifier, parameterBinding)
				}

				// Create a new container for the parameter and its value
				if newParameter, err := pgsql.AsParameter(parameterBinding.Identifier, typedExpression.Value); err != nil {
					s.SetError(err)
				} else if negotiatedValue, err := pgsql.NegotiateValue(typedExpression.Value); err != nil {
					s.SetError(err)
				} else {
					// Lift the parameter value into the parameters map
					s.translation.Parameters[parameterBinding.Identifier.String()] = negotiatedValue
					parameterBinding.Parameter = models.ValueOptional(newParameter)
				}

				// Set the outer reference
				binding = parameterBinding
			}
		}

		s.treeTranslator.PushOperand(binding.Parameter.Value)

	case *cypher.Variable:
		identifier := pgsql.Identifier(typedExpression.Symbol)

		if binding, resolved := s.scope.AliasedLookup(identifier); !resolved {
			s.SetErrorf("unable to resolve or otherwise lookup identifer %s", identifier)
		} else {
			s.treeTranslator.PushOperand(binding.Identifier)
		}

	case *cypher.Literal:
		literalValue := typedExpression.Value

		if stringValue, isString := typedExpression.Value.(string); isString {
			// Cypher parser wraps string literals with ' characters
			literalValue = stringValue[1 : len(stringValue)-1]
		}

		if newLiteral, err := pgsql.AsLiteral(literalValue); err != nil {
			s.SetError(err)
		} else {
			newLiteral.Null = typedExpression.Null
			s.treeTranslator.PushOperand(newLiteral)
		}

	case *cypher.Parenthetical:
		s.treeTranslator.PushParenthetical()

	case *cypher.SortItem:
		s.query.CurrentPart().OrderBy = append(s.query.CurrentPart().OrderBy, pgsql.OrderBy{
			Ascending: typedExpression.Ascending,
		})

	case *cypher.Projection:
		if err := s.prepareProjection(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.ProjectionItem:
		s.query.CurrentPart().PrepareProjection()

	case *cypher.PatternPredicate:
		if err := s.preparePatternPredicate(); err != nil {
			s.SetError(err)
		}

	case *cypher.PatternPart:
		if err := s.translatePatternPart(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.PartialComparison:
		s.treeTranslator.VisitOperator(pgsql.Operator(typedExpression.Operator))

	case *cypher.PartialArithmeticExpression:
		s.treeTranslator.VisitOperator(pgsql.Operator(typedExpression.Operator))

	case *cypher.Disjunction:
		for idx := 0; idx < typedExpression.Len()-1; idx++ {
			s.treeTranslator.VisitOperator(pgsql.OperatorOr)
		}

	case *cypher.Conjunction:
		for idx := 0; idx < typedExpression.Len()-1; idx++ {
			s.treeTranslator.VisitOperator(pgsql.OperatorAnd)
		}

	default:
		s.SetErrorf("unable to translate cypher type: %T", expression)
	}
}

func (s *Translator) Exit(expression cypher.SyntaxNode) {
	switch typedExpression := expression.(type) {
	case *cypher.NodePattern:
		if err := s.translateNodePattern(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.RelationshipPattern:
		if err := s.translateRelationshipPattern(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.MapItem:
		if value, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			s.query.CurrentPart().AddProperty(typedExpression.Key, value)
		}

	case *cypher.PatternPredicate:
		if err := s.translatePatternPredicate(); err != nil {
			s.SetError(err)
		}

	case *cypher.RemoveItem:
		if err := s.translateRemoveItem(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Delete:
		if err := s.translateDelete(s.scope, typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.SetItem:
		if err := s.translateSetItem(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.UpdatingClause:
		if err := s.translateUpdates(); err != nil {
			s.SetError(err)
		}

	case *cypher.ListLiteral:
		var (
			numExpressions = len(typedExpression.Expressions())
			literal        = pgsql.ArrayLiteral{
				Values:   make([]pgsql.Expression, numExpressions),
				CastType: pgsql.UnsetDataType,
			}
		)

		for idx := numExpressions - 1; idx >= 0; idx-- {
			if nextExpression, err := s.treeTranslator.PopOperand(); err != nil {
				s.SetError(err)
			} else {
				if typeHint, isTypeHinted := nextExpression.(pgsql.TypeHinted); isTypeHinted {
					if arrayCastType, err := typeHint.TypeHint().ToArrayType(); err != nil {
						s.SetError(err)
					} else if literal.CastType != pgsql.UnsetDataType && literal.CastType != arrayCastType {
						s.SetErrorf("expected array literal value type %s at index %d but found type %s", literal.CastType, idx, arrayCastType)
					} else {
						literal.CastType = arrayCastType
					}
				}

				literal.Values[idx] = nextExpression
			}
		}

		if literal.CastType == pgsql.UnsetDataType {
			s.SetErrorf("array literal has no available type hints")
		} else {
			s.treeTranslator.PushOperand(literal)
		}

	case *cypher.SortItem:
		// Rewrite the order by constraints
		if lookupExpression, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if err := RewriteFrameBindings(s.scope, lookupExpression); err != nil {
			s.SetError(err)
		} else {
			if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(lookupExpression); isPropertyLookup {
				// If sorting, use the raw type of the JSONB field
				propertyLookup.Operator = pgsql.OperatorJSONField
			}

			s.query.CurrentPart().CurrentOrderBy().Expression = lookupExpression
		}

	case *cypher.KindMatcher:
		if matcher, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			s.treeTranslator.PushOperand(matcher)
		}

	case *cypher.Parenthetical:
		// Pull the sub-expression we wrap
		if wrappedExpression, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if parenthetical, err := s.treeTranslator.PopParenthetical(); err != nil {
			s.SetError(err)
		} else {
			parenthetical.Expression = wrappedExpression
			s.treeTranslator.PushOperand(*parenthetical)
		}

	case *cypher.FunctionInvocation:
		s.translateFunction(typedExpression)

	case *cypher.UnaryAddOrSubtractExpression:
		if operand, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			s.treeTranslator.PushOperand(&pgsql.UnaryExpression{
				Operator: pgsql.Operator(typedExpression.Operator),
				Operand:  operand,
			})
		}

	case *cypher.Negation:
		if operand, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			for cursor := operand; cursor != nil; {
				switch typedCursor := cursor.(type) {
				case pgsql.Parenthetical:
					// Unwrap parentheticals
					cursor = typedCursor.Expression
					continue

				case *pgsql.BinaryExpression:
					switch typedCursor.Operator {
					case pgsql.OperatorLike, pgsql.OperatorILike:
						// If this is a string comparison operation then the negation requires wrapping the
						// operand references in coalesce functions. While this will kick out index acceleration
						// the negation will already damage the query planner's ability to utilize an index lookup.

						if leftPropertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(typedCursor.LOperand); isPropertyLookup {
							typedCursor.LOperand = pgsql.FunctionCall{
								Function: pgsql.FunctionCoalesce,
								Parameters: []pgsql.Expression{
									leftPropertyLookup,
									pgsql.NewLiteral("", pgsql.Text),
								},
								CastType: pgsql.Text,
							}
						}

						if rightPropertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(typedCursor.ROperand); isPropertyLookup {
							typedCursor.ROperand = pgsql.FunctionCall{
								Function: pgsql.FunctionCoalesce,
								Parameters: []pgsql.Expression{
									rightPropertyLookup,
									pgsql.NewLiteral("", pgsql.Text),
								},
								CastType: pgsql.Text,
							}
						}
					}
				}

				break
			}

			s.treeTranslator.PushOperand(&pgsql.UnaryExpression{
				Operator: pgsql.OperatorNot,
				Operand:  operand,
			})
		}

	case *cypher.Where:
		// Assign the last operands as identifier set constraints
		if err := s.treeTranslator.PopRemainingExpressionsAsConstraints(); err != nil {
			s.SetError(err)
		}

	case *cypher.PropertyLookup:
		if err := s.translatePropertyLookup(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.PartialComparison:
		if err := s.treeTranslator.CompleteBinaryExpression(s.scope, pgsql.Operator(typedExpression.Operator)); err != nil {
			s.SetError(err)
		}

	case *cypher.PartialArithmeticExpression:
		if err := s.treeTranslator.CompleteBinaryExpression(s.scope, pgsql.Operator(typedExpression.Operator)); err != nil {
			s.SetError(err)
		}

	case *cypher.Disjunction:
		for idx := 0; idx < typedExpression.Len()-1; idx++ {
			if err := s.treeTranslator.CompleteBinaryExpression(s.scope, pgsql.OperatorOr); err != nil {
				s.SetError(err)
			}
		}

	case *cypher.Conjunction:
		for idx := 0; idx < typedExpression.Len()-1; idx++ {
			if err := s.treeTranslator.CompleteBinaryExpression(s.scope, pgsql.OperatorAnd); err != nil {
				s.SetError(err)
			}
		}

	case *cypher.ProjectionItem:
		if err := s.translateProjectionItem(s.scope, typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Match:
		if err := s.translateMatch(); err != nil {
			s.SetError(err)
		}

	case *cypher.With:
		if err := s.translateWith(); err != nil {
			s.SetError(err)
		}

	case *cypher.MultiPartQueryPart:
		if err := s.translateMultiPartQueryPart(); err != nil {
			s.SetError(err)
		}

	case *cypher.SinglePartQuery:
		if err := s.buildSinglePartQuery(typedExpression); err != nil {
			s.SetError(err)
		}

		s.translation.Statement = *s.query.CurrentPart().Model

	case *cypher.MultiPartQuery:
		if err := s.buildMultiPartQuery(typedExpression.SinglePartQuery); err != nil {
			s.SetError(err)
		}
	}
}

type Result struct {
	Statement  pgsql.Statement
	Parameters map[string]any
}

func Translate(ctx context.Context, cypherQuery *cypher.RegularQuery, kindMapper pgsql.KindMapper, parameters map[string]any) (Result, error) {
	translator := NewTranslator(ctx, kindMapper, parameters)

	if err := walk.Cypher(cypherQuery, translator); err != nil {
		return Result{}, err
	}

	return translator.translation, nil
}
