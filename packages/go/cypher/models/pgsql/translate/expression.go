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
	"fmt"

	"github.com/specterops/bloodhound/cypher/models/cypher"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/walk"
)

func unwrapParenthetical(parenthetical pgsql.Expression) pgsql.Expression {
	next := parenthetical

	for next != nil {
		switch typedNext := next.(type) {
		case *pgsql.Parenthetical:
			next = typedNext.Expression

		default:
			return next
		}
	}

	return parenthetical
}

func (s *Translator) translatePropertyLookup(lookup *cypher.PropertyLookup) error {
	if translatedAtom, err := s.treeTranslator.PopOperand(); err != nil {
		return err
	} else {
		switch typedTranslatedAtom := translatedAtom.(type) {
		case pgsql.Identifier:
			if fieldIdentifierLiteral, err := pgsql.AsLiteral(lookup.Symbols[0]); err != nil {
				return err
			} else {
				s.treeTranslator.PushOperand(pgsql.CompoundIdentifier{typedTranslatedAtom, pgsql.ColumnProperties})
				s.treeTranslator.PushOperand(fieldIdentifierLiteral)

				if err := s.treeTranslator.CompleteBinaryExpression(s.scope, pgsql.OperatorPropertyLookup); err != nil {
					return err
				}
			}

		case pgsql.FunctionCall:
			if fieldIdentifierLiteral, err := pgsql.AsLiteral(lookup.Symbols[0]); err != nil {
				return err
			} else if componentName, typeOK := fieldIdentifierLiteral.Value.(string); !typeOK {
				return fmt.Errorf("expected a string component name in translated literal but received type: %T", fieldIdentifierLiteral.Value)
			} else {
				switch typedTranslatedAtom.Function {
				case pgsql.FunctionCurrentDate, pgsql.FunctionLocalTime, pgsql.FunctionCurrentTime, pgsql.FunctionLocalTimestamp, pgsql.FunctionNow:
					switch componentName {
					case cypher.ITTCEpochSeconds:
						s.treeTranslator.PushOperand(pgsql.FunctionCall{
							Function: pgsql.FunctionExtract,
							Parameters: []pgsql.Expression{pgsql.ProjectionFrom{
								Projection: []pgsql.SelectItem{
									pgsql.EpochIdentifier,
								},
								From: []pgsql.FromClause{{
									Source: translatedAtom,
								}},
							}},
							CastType: pgsql.Numeric,
						})

					case cypher.ITTCEpochMilliseconds:
						s.treeTranslator.PushOperand(pgsql.NewBinaryExpression(
							pgsql.FunctionCall{
								Function: pgsql.FunctionExtract,
								Parameters: []pgsql.Expression{pgsql.ProjectionFrom{
									Projection: []pgsql.SelectItem{
										pgsql.EpochIdentifier,
									},
									From: []pgsql.FromClause{{
										Source: translatedAtom,
									}},
								}},
								CastType: pgsql.Numeric,
							},
							pgsql.OperatorMultiply,
							pgsql.NewLiteral(1000, pgsql.Int4),
						))

					default:
						return fmt.Errorf("unsupported date time instant type component %s from function call %s", componentName, typedTranslatedAtom.Function)
					}

				default:
					return fmt.Errorf("unsupported instant type component %s from function call %s", componentName, typedTranslatedAtom.Function)
				}
			}
		}
	}

	return nil
}

func translateCypherAssignmentOperator(operator cypher.AssignmentOperator) (pgsql.Operator, error) {
	switch operator {
	case cypher.OperatorAssignment:
		return pgsql.OperatorAssignment, nil
	case cypher.OperatorLabelAssignment:
		return pgsql.OperatorKindAssignment, nil
	default:
		return pgsql.UnsetOperator, fmt.Errorf("unsupported assignment operator %s", operator)
	}
}

func ExtractSyntaxNodeReferences(root pgsql.SyntaxNode) (*pgsql.IdentifierSet, error) {
	dependencies := pgsql.NewIdentifierSet()

	return dependencies, walk.PgSQL(root, walk.NewSimpleVisitor[pgsql.SyntaxNode](
		func(node pgsql.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
			switch typedNode := node.(type) {
			case pgsql.Identifier:
				// Filter for reserved identifiers
				if !pgsql.IsReservedIdentifier(typedNode) {
					dependencies.Add(typedNode)
				}

			case pgsql.CompoundIdentifier:
				identifier := typedNode.Root()

				if !pgsql.IsReservedIdentifier(identifier) {
					dependencies.Add(identifier)
				}
			}
		},
	))
}

func rewritePropertyLookupOperator(propertyLookup *pgsql.BinaryExpression, dataType pgsql.DataType) pgsql.Expression {
	if dataType.IsArrayType() {
		// Ensure that array conversions use JSONB
		propertyLookup.Operator = pgsql.OperatorJSONField

		return pgsql.FunctionCall{
			Function:   pgsql.FunctionJSONBToTextArray,
			Parameters: []pgsql.Expression{propertyLookup},
			CastType:   dataType,
		}
	}

	switch dataType {
	case pgsql.Text:
		propertyLookup.Operator = pgsql.OperatorJSONTextField
		return propertyLookup

	case pgsql.Date, pgsql.TimestampWithoutTimeZone, pgsql.TimestampWithTimeZone, pgsql.TimeWithoutTimeZone, pgsql.TimeWithTimeZone:
		propertyLookup.Operator = pgsql.OperatorJSONTextField
		return pgsql.NewTypeCast(propertyLookup, dataType)

	case pgsql.UnknownDataType:
		propertyLookup.Operator = pgsql.OperatorJSONTextField
		return propertyLookup

	default:
		propertyLookup.Operator = pgsql.OperatorJSONTextField
		return pgsql.NewTypeCast(propertyLookup, dataType)
	}
}

func lookupRequiresElementType(typeHint pgsql.DataType, operator pgsql.Operator, otherOperand pgsql.SyntaxNode) bool {
	if typeHint.IsArrayType() {
		switch operator {
		case pgsql.OperatorIn:
			return true
		}

		switch otherOperand.(type) {
		case pgsql.AnyExpression:
			return true
		}
	}

	return false
}

func TypeCastExpression(expression pgsql.Expression, dataType pgsql.DataType) (pgsql.Expression, error) {
	if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(expression); isPropertyLookup {
		var lookupTypeHint = dataType

		if lookupRequiresElementType(dataType, propertyLookup.Operator, propertyLookup.ROperand) {
			// Take the base type of the array type hint: <unit> in <collection>
			lookupTypeHint = dataType.ArrayBaseType()
		}

		return rewritePropertyLookupOperator(propertyLookup, lookupTypeHint), nil
	}

	return pgsql.NewTypeCast(expression, dataType), nil
}

func rewritePropertyLookupOperands(expression *pgsql.BinaryExpression) error {
	var (
		leftPropertyLookup, hasLeftPropertyLookup   = expressionToPropertyLookupBinaryExpression(expression.LOperand)
		rightPropertyLookup, hasRightPropertyLookup = expressionToPropertyLookupBinaryExpression(expression.ROperand)
	)

	// Ensure that direct property comparisons prefer JSONB - JSONB
	if hasLeftPropertyLookup && hasRightPropertyLookup {
		leftPropertyLookup.Operator = pgsql.OperatorJSONField
		rightPropertyLookup.Operator = pgsql.OperatorJSONField

		return nil
	}

	if hasLeftPropertyLookup {
		// This check exists here to prevent from overwriting a property lookup that's part of a <value> in <list>
		// binary expression. This may want for better ergonomics in the future
		if anyExpression, isAnyExpression := expression.ROperand.(*pgsql.AnyExpression); isAnyExpression {
			expression.LOperand = rewritePropertyLookupOperator(leftPropertyLookup, anyExpression.CastType.ArrayBaseType())
		} else if rOperandTypeHint, err := InferExpressionType(expression.ROperand); err != nil {
			return err
		} else {
			switch expression.Operator {
			case pgsql.OperatorIn:
				expression.LOperand = rewritePropertyLookupOperator(leftPropertyLookup, rOperandTypeHint.ArrayBaseType())

			case pgsql.OperatorCypherStartsWith, pgsql.OperatorCypherEndsWith, pgsql.OperatorCypherContains, pgsql.OperatorRegexMatch:
				expression.LOperand = rewritePropertyLookupOperator(leftPropertyLookup, pgsql.Text)

			default:
				expression.LOperand = rewritePropertyLookupOperator(leftPropertyLookup, rOperandTypeHint)
			}
		}
	}

	if hasRightPropertyLookup {
		if lOperandTypeHint, err := InferExpressionType(expression.LOperand); err != nil {
			return err
		} else {
			switch expression.Operator {
			case pgsql.OperatorIn:
				if arrayType, err := lOperandTypeHint.ToArrayType(); err != nil {
					return err
				} else {
					expression.ROperand = rewritePropertyLookupOperator(rightPropertyLookup, arrayType)
				}

			case pgsql.OperatorCypherStartsWith, pgsql.OperatorCypherEndsWith, pgsql.OperatorCypherContains, pgsql.OperatorRegexMatch:
				expression.ROperand = rewritePropertyLookupOperator(rightPropertyLookup, pgsql.Text)

			default:
				expression.ROperand = rewritePropertyLookupOperator(rightPropertyLookup, lOperandTypeHint)
			}
		}
	}

	return nil
}

func newFunctionCallComparatorError(functionCall pgsql.FunctionCall, operator pgsql.Operator, comparisonType pgsql.DataType) error {
	switch functionCall.Function {
	case pgsql.FunctionCoalesce:
		// This is a specific error statement for coalesce statements. These statements have ill-defined
		// type conversion semantics in Cypher. As such, exposing the type specificity of coalesce to the
		// user as a distinct error will help reduce the surprise of running on a non-Neo4j substrate.
		return fmt.Errorf("coalesce has type %s but is being compared against type %s - ensure that all arguments in the coalesce function match the type of the other side of the comparison", functionCall.CastType, comparisonType)
	}

	return nil
}

type Builder struct {
	stack []pgsql.Expression
}

func NewExpressionTreeBuilder() *Builder {
	return &Builder{}
}

func (s *Builder) Depth() int {
	return len(s.stack)
}

func (s *Builder) IsEmpty() bool {
	return len(s.stack) == 0
}

func (s *Builder) PopOperand(kindMapper *contextAwareKindMapper) (pgsql.Expression, error) {
	next := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]

	switch typedNext := next.(type) {
	case *pgsql.UnaryExpression:
		if err := applyUnaryExpressionTypeHints(typedNext); err != nil {
			return nil, err
		}

	case *pgsql.BinaryExpression:
		if err := applyBinaryExpressionTypeHints(kindMapper, typedNext); err != nil {
			return nil, err
		}
	}

	return next, nil
}

func (s *Builder) PeekOperand() pgsql.Expression {
	return s.stack[len(s.stack)-1]
}

func (s *Builder) PushOperand(operand pgsql.Expression) {
	s.stack = append(s.stack, operand)
}

func ConjoinExpressions(kindMapper *contextAwareKindMapper, expressions []pgsql.Expression) (pgsql.Expression, error) {
	var conjoined pgsql.Expression

	for _, expression := range expressions {
		if expression == nil {
			continue
		}

		if conjoined == nil {
			conjoined = expression
			continue
		}

		conjoinedBinaryExpression := pgsql.NewBinaryExpression(conjoined, pgsql.OperatorAnd, expression)

		if err := applyBinaryExpressionTypeHints(kindMapper, conjoinedBinaryExpression); err != nil {
			return nil, err
		}

		conjoined = conjoinedBinaryExpression
	}

	return conjoined, nil
}

type ExpressionTreeTranslator struct {
	UserConstraints        *ConstraintTracker
	TranslationConstraints *ConstraintTracker

	treeBuilder        *Builder
	kindMapper         *contextAwareKindMapper
	parentheticalDepth int
	disjunctionDepth   int
	conjunctionDepth   int
}

func NewExpressionTreeTranslator(kindMapper *contextAwareKindMapper) *ExpressionTreeTranslator {
	return &ExpressionTreeTranslator{
		UserConstraints:        NewConstraintTracker(),
		TranslationConstraints: NewConstraintTracker(),
		treeBuilder:            NewExpressionTreeBuilder(),
		kindMapper:             kindMapper,
	}
}

func mergeUserAndTranslationConstraints(userConstraints, translationConstraints *Constraint) *Constraint {
	if userConstraints.Expression != nil {
		// Fold the user constraints into the translation constraints wrapped in a parenthetical
		translationConstraints.Expression = pgsql.OptionalAnd(pgsql.NewParenthetical(userConstraints.Expression), translationConstraints.Expression)
	}

	return translationConstraints
}

func (s *ExpressionTreeTranslator) ConsumeConstraintsFromVisibleSet(visible *pgsql.IdentifierSet) (*Constraint, error) {
	if userConstraints, err := s.UserConstraints.ConsumeSet(s.kindMapper, visible); err != nil {
		return nil, err
	} else if translationConstraints, err := s.TranslationConstraints.ConsumeSet(s.kindMapper, visible); err != nil {
		return nil, err
	} else {
		return mergeUserAndTranslationConstraints(userConstraints, translationConstraints), nil
	}
}

func (s *ExpressionTreeTranslator) ConsumeAllConstraints() (*Constraint, error) {
	if userConstraints, err := s.UserConstraints.ConsumeAll(s.kindMapper); err != nil {
		return nil, err
	} else if translationConstraints, err := s.TranslationConstraints.ConsumeAll(s.kindMapper); err != nil {
		return nil, err
	} else {
		return mergeUserAndTranslationConstraints(userConstraints, translationConstraints), nil
	}
}

func (s *ExpressionTreeTranslator) AddTranslationConstraint(requiredIdentifiers *pgsql.IdentifierSet, expression pgsql.Expression) error {
	return s.TranslationConstraints.Constrain(s.kindMapper, requiredIdentifiers, expression)
}

func (s *ExpressionTreeTranslator) AddUserConstraint(requiredIdentifiers *pgsql.IdentifierSet, expression pgsql.Expression) error {
	return s.UserConstraints.Constrain(s.kindMapper, requiredIdentifiers, expression)
}

func (s *ExpressionTreeTranslator) PushOperand(expression pgsql.Expression) {
	s.treeBuilder.PushOperand(expression)
}

func (s *ExpressionTreeTranslator) PeekOperand() pgsql.Expression {
	return s.treeBuilder.PeekOperand()
}

func (s *ExpressionTreeTranslator) PopOperand() (pgsql.Expression, error) {
	return s.treeBuilder.PopOperand(s.kindMapper)
}

func (s *ExpressionTreeTranslator) popOperandAsUserConstraint() error {
	if nextExpression, err := s.PopOperand(); err != nil {
		return err
	} else if identifierDeps, err := ExtractSyntaxNodeReferences(nextExpression); err != nil {
		return err
	} else {
		if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(nextExpression); isPropertyLookup {
			// If this is a bare property lookup rewrite it with the intended type of boolean
			nextExpression = rewritePropertyLookupOperator(propertyLookup, pgsql.Boolean)
		}

		return s.AddUserConstraint(identifierDeps, nextExpression)
	}
}

func (s *ExpressionTreeTranslator) PopRemainingExpressionsAsUserConstraints() error {
	// Pull the right operand only if one exists
	for !s.treeBuilder.IsEmpty() {
		if err := s.popOperandAsUserConstraint(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ExpressionTreeTranslator) ConstrainDisjointOperandPair() error {
	// Always expect a left operand
	if s.treeBuilder.IsEmpty() {
		return fmt.Errorf("expected at least one operand for constraint extraction")
	}

if rightOperand, err := s.treeBuilder.PopOperand(s.kindMapper); err != nil {
		return err
	} else if rightDependencies, err := ExtractSyntaxNodeReferences(rightOperand); err != nil {
		return err
	} else if s.treeBuilder.IsEmpty() {
		// If the tree builder is empty then this operand is at the top of the disjunction chain
		return s.AddUserConstraint(rightDependencies, rightOperand)
	} else if leftOperand, err := s.treeBuilder.PopOperand(s.kindMapper); err != nil {
		return err
	} else {
		newOrExpression := pgsql.NewBinaryExpression(
			leftOperand,
			pgsql.OperatorOr,
			rightOperand,
		)

		if err := applyBinaryExpressionTypeHints(s.kindMapper, newOrExpression); err != nil {
			return err
		}

		// This operation may not be complete; push it back on the stack
		s.PushOperand(newOrExpression)
		return nil
	}
}

func (s *ExpressionTreeTranslator) ConstrainConjoinedOperandPair() error {
	// Always expect a left operand
	if s.treeBuilder.IsEmpty() {
		return fmt.Errorf("expected at least one operand for constraint extraction")
	}

	if err := s.popOperandAsUserConstraint(); err != nil {
		return err
	}

	return nil
}

func (s *ExpressionTreeTranslator) PopBinaryExpression(operator pgsql.Operator) (*pgsql.BinaryExpression, error) {
	if rightOperand, err := s.PopOperand(); err != nil {
		return nil, err
	} else if leftOperand, err := s.PopOperand(); err != nil {
		return nil, err
	} else {
		newBinaryExpression := pgsql.NewBinaryExpression(leftOperand, operator, rightOperand)
		return newBinaryExpression, applyBinaryExpressionTypeHints(s.kindMapper, newBinaryExpression)
	}
}

func rewriteIdentityOperands(scope *Scope, newExpression *pgsql.BinaryExpression) error {
	switch typedLOperand := newExpression.LOperand.(type) {
	case pgsql.Identifier:
		// If the left side is an identifier we need to inspect the type of the identifier bound in our scope
		if boundLOperand, bound := scope.Lookup(typedLOperand); !bound {
			return fmt.Errorf("unknown identifier %s", typedLOperand)
		} else {
			switch typedROperand := newExpression.ROperand.(type) {
			case pgsql.Identifier:
				// If the right side is an identifier, inspect to see if the identifiers are an entity comparison.
				// For example: match (n1)-[]->(n2) where n1 <> n2 return n2
				if boundROperand, bound := scope.Lookup(typedROperand); !bound {
					return fmt.Errorf("unknown identifier %s", typedROperand)
				} else {
					switch boundLOperand.DataType {
					case pgsql.NodeCompositeArray:
						return fmt.Errorf("unsupported pgsql.NodeCompositeArray")

					case pgsql.NodeComposite, pgsql.ExpansionRootNode, pgsql.ExpansionTerminalNode:
						switch boundROperand.DataType {
						case pgsql.NodeComposite, pgsql.ExpansionRootNode, pgsql.ExpansionTerminalNode:
							// If this is a node entity comparison of some kind then the AST must be rewritten to use identity properties
							newExpression.LOperand = pgsql.CompoundIdentifier{typedLOperand, pgsql.ColumnID}
							newExpression.ROperand = pgsql.CompoundIdentifier{typedROperand, pgsql.ColumnID}

						case pgsql.NodeCompositeArray:
							newExpression.LOperand = pgsql.CompoundIdentifier{typedLOperand, pgsql.ColumnID}
							newExpression.ROperand = pgsql.CompoundIdentifier{typedROperand, pgsql.ColumnID}

						default:
							return fmt.Errorf("invalid comparison between types %s and %s", boundLOperand.DataType, boundROperand.DataType)
						}

					case pgsql.EdgeCompositeArray:
						return fmt.Errorf("unsupported pgsql.EdgeCompositeArray")

					case pgsql.EdgeComposite, pgsql.ExpansionEdge:
						switch boundROperand.DataType {
						case pgsql.EdgeComposite, pgsql.ExpansionEdge:
							// If this is an edge entity comparison of some kind then the AST must be rewritten to use identity properties
							newExpression.LOperand = pgsql.CompoundIdentifier{typedLOperand, pgsql.ColumnID}
							newExpression.ROperand = pgsql.CompoundIdentifier{typedROperand, pgsql.ColumnID}

						case pgsql.EdgeCompositeArray:
							newExpression.LOperand = pgsql.CompoundIdentifier{typedLOperand, pgsql.ColumnID}
							newExpression.ROperand = pgsql.CompoundIdentifier{typedROperand, pgsql.ColumnID}

						default:
							return fmt.Errorf("invalid comparison between types %s and %s", boundLOperand.DataType, boundROperand.DataType)
						}

					case pgsql.PathComposite:
						return fmt.Errorf("comparison for path identifiers is unsupported")
					}
				}
			}
		}
	}

	return nil
}

// isConcatenationOperation accepts two pgsql.DataType values and attempts to determine if the value are able to be
// concatenated.
//
// For further information regarding the conditional logic, please see the PgSQL upstream documentation:
// https://www.postgresql.org/docs/9.1/functions-string.html
func isConcatenationOperation(lOperandType, rOperandType pgsql.DataType) bool {
	// Any use of an array type automatically assumes concatenation
	if lOperandType.IsArrayType() || rOperandType.IsArrayType() {
		return true
	}

	// The case below must be able to infer operator intent from the following cases:
	// text + unknown
	// unknown + text
	// text + text

	// In the case below where both operands have no type information, no further
	// intent can be inferred for this operator
	switch lOperandType {
	case pgsql.Text:
		switch rOperandType {
		case pgsql.Text, pgsql.UnknownDataType:
			return true
		}

	case pgsql.UnknownDataType:
		switch rOperandType {
		case pgsql.Text:
			return true
		}
	}

	return false
}

func (s *ExpressionTreeTranslator) rewriteBinaryExpression(newExpression *pgsql.BinaryExpression) error {
	switch newExpression.Operator {
	case pgsql.OperatorAdd:
		// In the case of the use of the cypher `+` operator we must attempt to disambiguate if the intent
		// is to concatenate or to perform an addition
		if lOperandType, err := InferExpressionType(newExpression.LOperand); err != nil {
			return err
		} else if rOperandType, err := InferExpressionType(newExpression.ROperand); err != nil {
			return err
		} else if isConcatenationOperation(lOperandType, rOperandType) {
			newExpression.Operator = pgsql.OperatorConcatenate
		}

		s.PushOperand(newExpression)

	case pgsql.OperatorCypherContains:
		newExpression.Operator = pgsql.OperatorLike

		switch typedLOperand := newExpression.LOperand.(type) {
		case *pgsql.BinaryExpression:
			switch typedLOperand.Operator {
			case pgsql.OperatorPropertyLookup, pgsql.OperatorJSONField, pgsql.OperatorJSONTextField:
			default:
				return fmt.Errorf("unexpected operator %s for binary expression \"%s\" left operand", typedLOperand.Operator, newExpression.Operator)
			}
		}

		switch typedROperand := newExpression.ROperand.(type) {
		case pgsql.Literal:
			if rOperandDataType := typedROperand.TypeHint(); rOperandDataType != pgsql.Text {
				return fmt.Errorf("expected %s data type but found %s as right operand for operator %s", pgsql.Text, rOperandDataType, newExpression.Operator)
			} else if stringValue, isString := typedROperand.Value.(string); !isString {
				return fmt.Errorf("expected string but found %T as right operand for operator %s", typedROperand.Value, newExpression.Operator)
			} else {
				newExpression.ROperand = pgsql.NewLiteral("%"+stringValue+"%", rOperandDataType)
			}

		case *pgsql.Parenthetical:
			if typeCastedROperand, err := TypeCastExpression(typedROperand, pgsql.Text); err != nil {
				return err
			} else {
				newExpression.ROperand = pgsql.NewBinaryExpression(
					pgsql.NewLiteral("%", pgsql.Text),
					pgsql.OperatorConcatenate,
					pgsql.NewBinaryExpression(
						typeCastedROperand,
						pgsql.OperatorConcatenate,
						pgsql.NewLiteral("%", pgsql.Text),
					),
				)
			}

		case *pgsql.BinaryExpression:
			if stringLiteral, err := pgsql.AsLiteral("%"); err != nil {
				return err
			} else {
				if pgsql.OperatorIsPropertyLookup(typedROperand.Operator) {
					typedROperand.Operator = pgsql.OperatorJSONTextField
				}

				newExpression.ROperand = pgsql.NewTypeCast(pgsql.NewBinaryExpression(
					stringLiteral,
					pgsql.OperatorConcatenate,
					pgsql.NewBinaryExpression(
						&pgsql.Parenthetical{
							Expression: typedROperand,
						},
						pgsql.OperatorConcatenate,
						stringLiteral,
					),
				), pgsql.Text)
			}

		default:
			newExpression.ROperand = pgsql.NewBinaryExpression(
				pgsql.NewLiteral("%", pgsql.Text),
				pgsql.OperatorConcatenate,
				pgsql.NewBinaryExpression(
					typedROperand,
					pgsql.OperatorConcatenate,
					pgsql.NewLiteral("%", pgsql.Text),
				),
			)
		}

		s.PushOperand(newExpression)

	case pgsql.OperatorCypherRegexMatch:
		newExpression.Operator = pgsql.OperatorRegexMatch
		s.PushOperand(newExpression)

	case pgsql.OperatorCypherStartsWith:
		newExpression.Operator = pgsql.OperatorLike

		switch typedLOperand := newExpression.LOperand.(type) {
		case *pgsql.BinaryExpression:
			switch typedLOperand.Operator {
			case pgsql.OperatorPropertyLookup, pgsql.OperatorJSONField, pgsql.OperatorJSONTextField:
			default:
				return fmt.Errorf("unexpected operator %s for binary expression \"%s\" left operand", typedLOperand.Operator, newExpression.Operator)
			}
		}

		switch typedROperand := newExpression.ROperand.(type) {
		case pgsql.Literal:
			if rOperandDataType := typedROperand.TypeHint(); rOperandDataType != pgsql.Text {
				return fmt.Errorf("expected %s data type but found %s as right operand for operator %s", pgsql.Text, rOperandDataType, newExpression.Operator)
			} else if stringValue, isString := typedROperand.Value.(string); !isString {
				return fmt.Errorf("expected string but found %T as right operand for operator %s", typedROperand.Value, newExpression.Operator)
			} else {
				newExpression.ROperand = pgsql.NewLiteral(stringValue+"%", rOperandDataType)
			}

		case *pgsql.Parenthetical:
			if typeCastedROperand, err := TypeCastExpression(typedROperand, pgsql.Text); err != nil {
				return err
			} else {
				newExpression.ROperand = pgsql.NewBinaryExpression(
					typeCastedROperand,
					pgsql.OperatorConcatenate,
					pgsql.NewLiteral("%", pgsql.Text),
				)
			}

		case *pgsql.BinaryExpression:
			if stringLiteral, err := pgsql.AsLiteral("%"); err != nil {
				return err
			} else {
				if pgsql.OperatorIsPropertyLookup(typedROperand.Operator) {
					typedROperand.Operator = pgsql.OperatorJSONTextField
				}

				newExpression.ROperand = pgsql.NewTypeCast(pgsql.NewBinaryExpression(
					&pgsql.Parenthetical{
						Expression: typedROperand,
					},
					pgsql.OperatorConcatenate,
					stringLiteral,
				), pgsql.Text)
			}

		default:
			newExpression.ROperand = pgsql.NewBinaryExpression(
				typedROperand,
				pgsql.OperatorConcatenate,
				pgsql.NewLiteral("%", pgsql.Text),
			)
		}

		s.PushOperand(newExpression)

	case pgsql.OperatorCypherEndsWith:
		newExpression.Operator = pgsql.OperatorLike

		switch typedLOperand := newExpression.LOperand.(type) {
		case *pgsql.BinaryExpression:
			switch typedLOperand.Operator {
			case pgsql.OperatorPropertyLookup, pgsql.OperatorJSONField, pgsql.OperatorJSONTextField:
			default:
				return fmt.Errorf("unexpected operator %s for binary expression \"%s\" left operand", typedLOperand.Operator, newExpression.Operator)
			}
		}

		switch typedROperand := newExpression.ROperand.(type) {
		case pgsql.Literal:
			if rOperandDataType := typedROperand.TypeHint(); rOperandDataType != pgsql.Text {
				return fmt.Errorf("expected %s data type but found %s as right operand for operator %s", pgsql.Text, rOperandDataType, newExpression.Operator)
			} else if stringValue, isString := typedROperand.Value.(string); !isString {
				return fmt.Errorf("expected string but found %T as right operand for operator %s", typedROperand.Value, newExpression.Operator)
			} else {
				newExpression.ROperand = pgsql.NewLiteral("%"+stringValue, rOperandDataType)
			}

		case *pgsql.Parenthetical:
			if typeCastedROperand, err := TypeCastExpression(typedROperand, pgsql.Text); err != nil {
				return err
			} else {
				newExpression.ROperand = pgsql.NewBinaryExpression(
					pgsql.NewLiteral("%", pgsql.Text),
					pgsql.OperatorConcatenate,
					typeCastedROperand,
				)
			}

		case *pgsql.BinaryExpression:
			if pgsql.OperatorIsPropertyLookup(typedROperand.Operator) {
				typedROperand.Operator = pgsql.OperatorJSONTextField
			}

			newExpression.ROperand = pgsql.NewTypeCast(pgsql.NewBinaryExpression(
				pgsql.NewLiteral("%", pgsql.Text),
				pgsql.OperatorConcatenate,
				&pgsql.Parenthetical{
					Expression: typedROperand,
				},
			), pgsql.Text)

		default:
			newExpression.ROperand = pgsql.NewBinaryExpression(
				pgsql.NewLiteral("%", pgsql.Text),
				pgsql.OperatorConcatenate,
				typedROperand,
			)
		}

		s.PushOperand(newExpression)

	case pgsql.OperatorIs:
		switch typedLOperand := newExpression.LOperand.(type) {
		case *pgsql.BinaryExpression:
			switch typedLOperand.Operator {
			case pgsql.OperatorPropertyLookup, pgsql.OperatorJSONField, pgsql.OperatorJSONTextField:
				// This is a null-check against a property. This should be rewritten using the JSON field exists
				// operator instead. It can be
				switch typedROperand := newExpression.ROperand.(type) {
				case pgsql.Literal:
					if typedROperand.Null {
						newExpression.Operator = pgsql.OperatorJSONBFieldExists
						newExpression.LOperand = typedLOperand.LOperand
						newExpression.ROperand = typedLOperand.ROperand
					}

					s.PushOperand(pgsql.NewUnaryExpression(pgsql.OperatorNot, newExpression))
				}
			}
		}

	case pgsql.OperatorIsNot:
		switch typedLOperand := newExpression.LOperand.(type) {
		case *pgsql.BinaryExpression:
			switch typedLOperand.Operator {
			case pgsql.OperatorPropertyLookup, pgsql.OperatorJSONField, pgsql.OperatorJSONTextField:
				// This is a null-check against a property. This should be rewritten using the JSON field exists
				// operator instead. It can be
				switch typedROperand := newExpression.ROperand.(type) {
				case pgsql.Literal:
					if typedROperand.Null {
						newExpression.Operator = pgsql.OperatorJSONBFieldExists
						newExpression.LOperand = typedLOperand.LOperand
						newExpression.ROperand = typedLOperand.ROperand
					}

					s.PushOperand(newExpression)
				}
			}
		}

	case pgsql.OperatorIn:
		newExpression.Operator = pgsql.OperatorEquals

		switch typedROperand := newExpression.ROperand.(type) {
		case pgsql.TypeCast:
			switch typedInnerOperand := typedROperand.Expression.(type) {
			case *pgsql.BinaryExpression:
				if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(typedInnerOperand); isPropertyLookup {
					// Attempt to figure out the cast by looking at the left operand
					if leftHint, err := InferExpressionType(newExpression.LOperand); err != nil {
						return err
					} else if leftArrayHint, err := leftHint.ToArrayType(); err != nil {
						return err
					} else {
						// Ensure the lookup uses the JSONB type
						propertyLookup.Operator = pgsql.OperatorJSONField

						newExpression.ROperand = pgsql.NewAnyExpressionHinted(
							pgsql.FunctionCall{
								Function:   pgsql.FunctionJSONBToTextArray,
								Parameters: []pgsql.Expression{propertyLookup},
								CastType:   leftArrayHint,
							},
						)
					}
				}
			}

		case pgsql.TypeHinted:
			if lOperandTypeHint, err := InferExpressionType(newExpression.LOperand); err != nil {
				return err
			} else if lOperandTypeHint.IsArrayType() {
				newExpression.Operator = pgsql.OperatorPGArrayOverlap
			} else {
				newExpression.Operator = pgsql.OperatorEquals
				newExpression.ROperand = pgsql.NewAnyExpression(newExpression.ROperand, typedROperand.TypeHint())
			}

		default:
			// Attempt to figure out the cast by looking at the left operand
			if leftHint, err := InferExpressionType(newExpression.LOperand); err != nil {
				return err
			} else {
				newExpression.ROperand = pgsql.NewAnyExpression(newExpression.ROperand, leftHint)
			}
		}

		s.PushOperand(newExpression)

	default:
		s.PushOperand(newExpression)
	}

	return nil
}

func (s *ExpressionTreeTranslator) PushParenthetical() {
	s.PushOperand(&pgsql.Parenthetical{})
	s.parentheticalDepth += 1
}

func (s *ExpressionTreeTranslator) PopParenthetical() (*pgsql.Parenthetical, error) {
	s.parentheticalDepth -= 1

	if operand, err := s.treeBuilder.PopOperand(s.kindMapper); err != nil {
		return nil, err
	} else if parentheticalExpr, typeOK := operand.(*pgsql.Parenthetical); !typeOK {
		return nil, fmt.Errorf("expected type *pgsql.Parenthetical but received %T", operand)
	} else {
		return parentheticalExpr, nil
	}
}

func (s *ExpressionTreeTranslator) VisitOperator(operator pgsql.Operator) {
	// Track this operator for expression tree extraction
	switch operator {
	case pgsql.OperatorAnd:
		s.conjunctionDepth += 1

	case pgsql.OperatorOr:
		s.disjunctionDepth += 1
	}
}

func (s *ExpressionTreeTranslator) CompleteBinaryExpression(scope *Scope, operator pgsql.Operator) error {
	switch operator {
	case pgsql.OperatorAnd:
		if s.parentheticalDepth == 0 && s.disjunctionDepth == 0 {
			return s.ConstrainConjoinedOperandPair()
		}

		s.conjunctionDepth -= 1

	case pgsql.OperatorOr:
		if s.parentheticalDepth == 0 && s.conjunctionDepth == 0 {
			return s.ConstrainDisjointOperandPair()
		}

		s.disjunctionDepth -= 1
	}

	if newExpression, err := s.PopBinaryExpression(operator); err != nil {
		return err
	} else if err := rewriteIdentityOperands(scope, newExpression); err != nil {
		return err
	} else {
		return s.rewriteBinaryExpression(newExpression)
	}
}
