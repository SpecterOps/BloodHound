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
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/walk"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type State int

const (
	StateTranslatingStart State = iota
	StateTranslatingPatternPart
	StateTranslatingMatch
	StateTranslatingCreate
	StateTranslatingWhere
	StateTranslatingProjection
	StateTranslatingOrderBy
	StateTranslatingUpdateClause
	StateTranslatingPatternPredicate
	StateTranslatingNestedExpression
)

func (s State) String() string {
	switch s {
	case StateTranslatingStart:
		return "start"
	case StateTranslatingPatternPart:
		return "pattern part"
	case StateTranslatingMatch:
		return "match clause"
	case StateTranslatingWhere:
		return "where clause"
	case StateTranslatingProjection:
		return "projection"
	case StateTranslatingOrderBy:
		return "order by"
	case StateTranslatingPatternPredicate:
		return "pattern predicate"
	case StateTranslatingNestedExpression:
		return "nested expression"
	default:
		return ""
	}
}

type Translator struct {
	walk.HierarchicalVisitor[cypher.SyntaxNode]

	ctx            context.Context
	kindMapper     pgsql.KindMapper
	translation    Result
	state          []State
	treeTranslator *ExpressionTreeTranslator
	properties     map[string]pgsql.Expression
	pattern        *Pattern
	match          *Match
	projections    *ProjectionClause
	mutations      *Mutations
	query          *Query
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
		properties:     map[string]pgsql.Expression{},
		pattern:        &Pattern{},
	}
}

func (s *Translator) currentState() State {
	return s.state[len(s.state)-1]
}

func (s *Translator) pushState(state State) {
	s.state = append(s.state, state)
}

func (s *Translator) popState() {
	s.state = s.state[:len(s.state)-1]
}

func (s *Translator) exitState(expectedState State) {
	if currentState := s.currentState(); currentState != expectedState {
		s.SetErrorf("expected state %s but found %s", expectedState, currentState)
	} else {
		s.state = s.state[:len(s.state)-1]
	}
}

func (s *Translator) inState(expectedState State) bool {
	for _, state := range s.state {
		if state == expectedState {
			return true
		}
	}

	return false
}

func (s *Translator) Enter(expression cypher.SyntaxNode) {
	switch typedExpression := expression.(type) {
	case *cypher.RegularQuery, *cypher.SingleQuery, *cypher.PatternElement, *cypher.Return,
		*cypher.Comparison, *cypher.Skip, *cypher.Limit, cypher.Operator, *cypher.ArithmeticExpression,
		*cypher.NodePattern, *cypher.RelationshipPattern, *cypher.Remove, *cypher.Set,
		*cypher.ReadingClause, *cypher.UnaryAddOrSubtractExpression, *cypher.PropertyLookup:
	// No operation for these syntax nodes

	case *cypher.Negation:
		s.pushState(StateTranslatingNestedExpression)

	case *cypher.SinglePartQuery:
		s.query = &Query{
			Scope: NewScope(),
			Model: &pgsql.Query{
				CommonTableExpressions: &pgsql.With{},
			},
		}

		s.mutations = NewMutations()

	case *cypher.Create:
		s.pushState(StateTranslatingCreate)

	case *cypher.Match:
		s.pushState(StateTranslatingMatch)

		// Start with a fresh match and where clause. Instantiation of the where clause here is necessary since
		// cypher will store identifier constraints in the query pattern which precedes the query where clause.
		s.pattern = &Pattern{}
		s.match = &Match{
			Scope:   s.query.Scope,
			Pattern: s.pattern,
		}

	case *cypher.Where:
		// Track that we're in a where clause first
		s.pushState(StateTranslatingWhere)

		// If there's a where AST node present in the cypher model we likely have an expression to translate
		s.pushState(StateTranslatingNestedExpression)

	case graph.Kinds:
		s.treeTranslator.Push(pgsql.KindListLiteral{
			Values: typedExpression,
		})

	case *cypher.KindMatcher:
		if err := s.translateKindMatcher(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Parameter:
		var (
			cypherIdentifier = pgsql.Identifier(typedExpression.Symbol)
			binding, bound   = s.query.Scope.AliasedLookup(cypherIdentifier)
		)

		if !bound {
			if parameterBinding, err := s.query.Scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
				s.SetError(err)
			} else {
				// Alias the old parameter identifier to the synthetic one
				if cypherIdentifier != "" {
					s.query.Scope.Alias(cypherIdentifier, parameterBinding)
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

		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			s.treeTranslator.Push(binding.Parameter.Value)

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.Variable:
		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			if binding, resolved := s.query.Scope.LookupString(typedExpression.Symbol); !resolved {
				s.SetErrorf("unable to find identifier %s", typedExpression.Symbol)
			} else {
				s.treeTranslator.Push(binding.Identifier)
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.ListLiteral:
		s.pushState(StateTranslatingNestedExpression)

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

			switch currentState := s.currentState(); currentState {
			case StateTranslatingNestedExpression:
				s.treeTranslator.Push(newLiteral)

			default:
				s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
			}
		}

	case *cypher.Parenthetical:
		s.pushState(StateTranslatingNestedExpression)
		s.treeTranslator.PushParenthetical()

	case *cypher.FunctionInvocation:
		s.pushState(StateTranslatingNestedExpression)

	case *cypher.Order:
		s.pushState(StateTranslatingOrderBy)

	case *cypher.SortItem:
		s.pushState(StateTranslatingNestedExpression)

		s.query.OrderBy = append(s.query.OrderBy, pgsql.OrderBy{
			Ascending: typedExpression.Ascending,
		})

	case *cypher.Projection:
		s.pushState(StateTranslatingProjection)

		if err := s.translateProjection(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.ProjectionItem:
		s.pushState(StateTranslatingNestedExpression)
		s.projections.PushProjection()

	case *cypher.PatternPredicate:
		s.pushState(StateTranslatingPatternPredicate)

		s.pattern = &Pattern{}

		if err := s.translatePatternPredicate(s.query.Scope); err != nil {
			s.SetError(err)
		}

	case *cypher.PatternPart:
		s.pushState(StateTranslatingPatternPart)

		if err := s.translatePatternPart(s.query.Scope, typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.PartialComparison:
		s.treeTranslator.PushOperator(pgsql.Operator(typedExpression.Operator))

	case *cypher.PartialArithmeticExpression:
		s.treeTranslator.PushOperator(pgsql.Operator(typedExpression.Operator))

	case *cypher.Disjunction:
		for idx := 0; idx < typedExpression.Len()-1; idx++ {
			s.treeTranslator.PushOperator(pgsql.OperatorOr)
		}

	case *cypher.Conjunction:
		for idx := 0; idx < typedExpression.Len()-1; idx++ {
			s.treeTranslator.PushOperator(pgsql.OperatorAnd)
		}

	case *cypher.UpdatingClause:
		s.pushState(StateTranslatingUpdateClause)

	case *cypher.RemoveItem:
		s.pushState(StateTranslatingNestedExpression)

	case *cypher.Delete:
		s.pushState(StateTranslatingNestedExpression)

	case *cypher.SetItem:
		s.pushState(StateTranslatingNestedExpression)

	case *cypher.Properties:
		clear(s.properties)

	case *cypher.MapItem:
		s.pushState(StateTranslatingNestedExpression)

	default:
		s.SetErrorf("unable to translate cypher type: %T", expression)
	}
}

func (s *Translator) Exit(expression cypher.SyntaxNode) {
	switch typedExpression := expression.(type) {
	case *cypher.NodePattern:
		if err := s.translateNodePattern(s.query.Scope, typedExpression, s.pattern.CurrentPart()); err != nil {
			s.SetError(err)
		}

	case *cypher.RelationshipPattern:
		if err := s.translateRelationshipPattern(s.query.Scope, typedExpression, s.pattern.CurrentPart()); err != nil {
			s.SetError(err)
		}

	case *cypher.MapItem:
		s.exitState(StateTranslatingNestedExpression)

		if value, err := s.treeTranslator.Pop(); err != nil {
			s.SetError(err)
		} else {
			s.properties[typedExpression.Key] = value
		}

	case *cypher.PatternPredicate:
		s.exitState(StateTranslatingPatternPredicate)

		// Retire the predicate scope frames and build the predicate
		for range s.pattern.CurrentPart().TraversalSteps {
			s.query.Scope.PopFrame()
		}

		if err := s.buildPatternPredicate(); err != nil {
			s.SetError(err)
		}

	case *cypher.RemoveItem:
		s.exitState(StateTranslatingNestedExpression)

		if err := s.translateRemoveItem(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Delete:
		s.exitState(StateTranslatingNestedExpression)

		if err := s.translateDelete(s.query.Scope, typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.SetItem:
		s.exitState(StateTranslatingNestedExpression)

		if err := s.translateSetItem(typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.UpdatingClause:
		s.exitState(StateTranslatingUpdateClause)

	case *cypher.ListLiteral:
		s.exitState(StateTranslatingNestedExpression)

		var (
			numExpressions = len(typedExpression.Expressions())
			literal        = pgsql.ArrayLiteral{
				Values:   make([]pgsql.Expression, numExpressions),
				CastType: pgsql.UnsetDataType,
			}
		)

		for idx := numExpressions - 1; idx >= 0; idx-- {
			if nextExpression, err := s.treeTranslator.Pop(); err != nil {
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
			s.treeTranslator.Push(literal)
		}

	case *cypher.Order:
		s.exitState(StateTranslatingOrderBy)

	case *cypher.SortItem:
		s.exitState(StateTranslatingNestedExpression)

		// Rewrite the order by constraints
		if lookupExpression, err := s.treeTranslator.Pop(); err != nil {
			s.SetError(err)
		} else if err := RewriteExpressionIdentifiers(lookupExpression, s.query.Scope.CurrentFrameBinding().Identifier, s.query.Scope.Visible()); err != nil {
			s.SetError(err)
		} else {
			s.query.CurrentOrderBy().Expression = lookupExpression
		}

	case *cypher.KindMatcher:
		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			if matcher, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				s.treeTranslator.Push(matcher)
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.Parenthetical:
		s.exitState(StateTranslatingNestedExpression)

		// Pull the sub-expression we wrap
		if wrappedExpression, err := s.treeTranslator.Pop(); err != nil {
			s.SetError(err)
		} else if parenthetical, err := s.treeTranslator.PopParenthetical(); err != nil {
			s.SetError(err)
		} else {
			parenthetical.Expression = wrappedExpression

			switch currentState := s.currentState(); currentState {
			case StateTranslatingNestedExpression:
				s.treeTranslator.Push(*parenthetical)

			default:
				s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
			}
		}

	case *cypher.FunctionInvocation:
		s.exitState(StateTranslatingNestedExpression)

		formattedName := strings.ToLower(typedExpression.Name)

		switch formattedName {
		case cypher.IdentityFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if referenceArgument, err := PopFromBuilderAs[pgsql.Identifier](s.treeTranslator); err != nil {
				s.SetError(err)
			} else {
				s.treeTranslator.Push(pgsql.CompoundIdentifier{referenceArgument, pgsql.ColumnID})
			}

		case cypher.LocalTimeFunction:
			if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.TimeWithoutTimeZone); err != nil {
				s.SetError(err)
			}

		case cypher.LocalDateTimeFunction:
			if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.TimestampWithoutTimeZone); err != nil {
				s.SetError(err)
			}

		case cypher.DateFunction:
			if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.Date); err != nil {
				s.SetError(err)
			}

		case cypher.DateTimeFunction:
			if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.TimestampWithTimeZone); err != nil {
				s.SetError(err)
			}

		case cypher.EdgeTypeFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else if identifier, isIdentifier := argument.(pgsql.Identifier); !isIdentifier {
				s.SetErrorf("expected an identifier for the cypher function: %s but received %T", typedExpression.Name, argument)
			} else {
				s.treeTranslator.Push(pgsql.CompoundIdentifier{identifier, pgsql.ColumnKindID})
			}

		case cypher.NodeLabelsFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else if identifier, isIdentifier := argument.(pgsql.Identifier); !isIdentifier {
				s.SetErrorf("expected an identifier for the cypher function: %s but received %T", typedExpression.Name, argument)
			} else {
				s.treeTranslator.Push(pgsql.CompoundIdentifier{identifier, pgsql.ColumnKindIDs})
			}

		case cypher.CountFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				s.treeTranslator.Push(pgsql.FunctionCall{
					Function:   pgsql.FunctionCount,
					Parameters: []pgsql.Expression{argument},
					CastType:   pgsql.Int8,
				})
			}

		case cypher.StringSplitToArrayFunction:
			if typedExpression.NumArguments() != 2 {
				s.SetError(fmt.Errorf("expected two arguments for cypher function %s", typedExpression.Name))
			} else if delimiter, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else if splitReference, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				if _, hasHint := GetTypeHint(splitReference); !hasHint {
					// Do our best to coerce the type into text
					if typedSplitRef, err := TypeCastExpression(splitReference, pgsql.Text); err != nil {
						s.SetError(err)
					} else {
						s.treeTranslator.Push(pgsql.FunctionCall{
							Function:   pgsql.FunctionStringToArray,
							Parameters: []pgsql.Expression{typedSplitRef, delimiter},
							CastType:   pgsql.TextArray,
						})
					}
				} else {
					s.treeTranslator.Push(pgsql.FunctionCall{
						Function:   pgsql.FunctionStringToArray,
						Parameters: []pgsql.Expression{splitReference, delimiter},
						CastType:   pgsql.TextArray,
					})
				}
			}

		case cypher.ToLowerFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				if propertyLookup, isPropertyLookup := asPropertyLookup(argument); isPropertyLookup {
					// Rewrite the property lookup operator with a JSON text field lookup
					propertyLookup.Operator = pgsql.OperatorJSONTextField
				}

				s.treeTranslator.Push(pgsql.FunctionCall{
					Function:   pgsql.FunctionToLower,
					Parameters: []pgsql.Expression{argument},
					CastType:   pgsql.Text,
				})
			}

		case cypher.ListSizeFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				var functionCall pgsql.FunctionCall

				if _, isPropertyLookup := asPropertyLookup(argument); isPropertyLookup {
					functionCall = pgsql.FunctionCall{
						Function:   pgsql.FunctionJSONBArrayLength,
						Parameters: []pgsql.Expression{argument},
						CastType:   pgsql.Int,
					}
				} else {
					functionCall = pgsql.FunctionCall{
						Function:   pgsql.FunctionArrayLength,
						Parameters: []pgsql.Expression{argument, pgsql.NewLiteral(1, pgsql.Int)},
						CastType:   pgsql.Int,
					}
				}

				s.treeTranslator.Push(functionCall)
			}

		case cypher.ToUpperFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				if propertyLookup, isPropertyLookup := asPropertyLookup(argument); isPropertyLookup {
					// Rewrite the property lookup operator with a JSON text field lookup
					propertyLookup.Operator = pgsql.OperatorJSONTextField
				}

				s.treeTranslator.Push(pgsql.FunctionCall{
					Function:   pgsql.FunctionToUpper,
					Parameters: []pgsql.Expression{argument},
					CastType:   pgsql.Text,
				})
			}

		case cypher.ToStringFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				s.treeTranslator.Push(pgsql.NewTypeCast(argument, pgsql.Text))
			}

		case cypher.ToIntegerFunction:
			if typedExpression.NumArguments() != 1 {
				s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
			} else if argument, err := s.treeTranslator.Pop(); err != nil {
				s.SetError(err)
			} else {
				s.treeTranslator.Push(pgsql.NewTypeCast(argument, pgsql.Int8))
			}

		case cypher.CoalesceFunction:
			if err := s.translateCoalesceFunction(typedExpression); err != nil {
				s.SetError(err)
			}

		default:
			s.SetErrorf("unknown cypher function: %s", typedExpression.Name)
		}

	case *cypher.UnaryAddOrSubtractExpression:
		if operand, err := s.treeTranslator.Pop(); err != nil {
			s.SetError(err)
		} else {
			s.treeTranslator.Push(&pgsql.UnaryExpression{
				Operator: pgsql.Operator(typedExpression.Operator),
				Operand:  operand,
			})
		}

	case *cypher.Negation:
		s.exitState(StateTranslatingNestedExpression)

		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			if operand, err := s.treeTranslator.Pop(); err != nil {
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

							if leftPropertyLookup, isPropertyLookup := asPropertyLookup(typedCursor.LOperand); isPropertyLookup {
								typedCursor.LOperand = pgsql.FunctionCall{
									Function: pgsql.FunctionCoalesce,
									Parameters: []pgsql.Expression{
										leftPropertyLookup,
										pgsql.NewLiteral("", pgsql.Text),
									},
									CastType: pgsql.Text,
								}
							}

							if rightPropertyLookup, isPropertyLookup := asPropertyLookup(typedCursor.ROperand); isPropertyLookup {
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

				s.treeTranslator.Push(&pgsql.UnaryExpression{
					Operator: pgsql.OperatorNot,
					Operand:  operand,
				})
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.PatternPart:
		s.exitState(StateTranslatingPatternPart)

	case *cypher.Where:
		// Validate state transitions
		s.exitState(StateTranslatingNestedExpression)
		s.exitState(StateTranslatingWhere)

		// Assign the last operands as identifier set constraints
		if err := s.treeTranslator.ConstrainRemainingOperands(); err != nil {
			s.SetError(err)
		}

	case *cypher.PropertyLookup:
		s.translatePropertyLookup(typedExpression)

	case *cypher.PartialComparison:
		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			if err := s.treeTranslator.PopPushOperator(s.query.Scope, pgsql.Operator(typedExpression.Operator)); err != nil {
				s.SetError(err)
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.PartialArithmeticExpression:
		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			if err := s.treeTranslator.PopPushOperator(s.query.Scope, pgsql.Operator(typedExpression.Operator)); err != nil {
				s.SetError(err)
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.Disjunction:
		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			for idx := 0; idx < typedExpression.Len()-1; idx++ {
				if err := s.treeTranslator.PopPushOperator(s.query.Scope, pgsql.OperatorOr); err != nil {
					s.SetError(err)
				}
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.Conjunction:
		switch currentState := s.currentState(); currentState {
		case StateTranslatingNestedExpression:
			for idx := 0; idx < typedExpression.Len()-1; idx++ {
				if err := s.treeTranslator.PopPushOperator(s.query.Scope, pgsql.OperatorAnd); err != nil {
					s.SetError(err)
				}
			}

		default:
			s.SetErrorf("invalid state \"%s\" for cypher AST node %T", s.currentState(), expression)
		}

	case *cypher.ProjectionItem:
		s.exitState(StateTranslatingNestedExpression)

		if err := s.translateProjectionItem(s.query.Scope, typedExpression); err != nil {
			s.SetError(err)
		}

	case *cypher.Projection:
		s.exitState(StateTranslatingProjection)

	case *cypher.Return:

	case *cypher.Create:
		s.exitState(StateTranslatingCreate)

	case *cypher.Match:
		s.exitState(StateTranslatingMatch)

		if err := s.buildMatch(s.match.Scope); err != nil {
			s.SetError(err)
		}

	case *cypher.SinglePartQuery:
		if s.mutations.Assignments.Len() > 0 {
			if err := s.translateUpdates(s.query.Scope); err != nil {
				s.SetError(err)
			}

			if err := s.buildUpdates(s.query.Scope); err != nil {
				s.SetError(err)
			}
		}

		if s.mutations.Deletions.Len() > 0 {
			if err := s.buildDeletions(s.query.Scope); err != nil {
				s.SetError(err)
			}
		}

		// If there was no return specified end the CTE chain with a bare select
		if typedExpression.Return == nil {
			if literalReturn, err := pgsql.AsLiteral(1); err != nil {
				s.SetError(err)
			} else {
				s.query.Model.Body = pgsql.Select{
					Projection: []pgsql.SelectItem{literalReturn},
				}
			}
		} else if err := s.buildProjection(s.query.Scope); err != nil {
			s.SetError(err)
		}

		s.translation.Statement = *s.query.Model
	}
}

type Result struct {
	Statement  pgsql.Statement
	Parameters map[string]any
}

func Translate(ctx context.Context, cypherQuery *cypher.RegularQuery, kindMapper pgsql.KindMapper, parameters map[string]any) (Result, error) {
	translator := NewTranslator(ctx, kindMapper, parameters)

	if err := walk.WalkCypher(cypherQuery, translator); err != nil {
		return Result{}, err
	}

	return translator.translation, nil
}
