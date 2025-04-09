// Copyright 2025 Specter Ops, Inc.
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
	"log/slog"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func GetTypeHint(expression pgsql.Expression) (pgsql.DataType, bool) {
	if typeHintedExpression, isTypeHinted := expression.(pgsql.TypeHinted); isTypeHinted {
		return typeHintedExpression.TypeHint(), true
	}

	return pgsql.UnsetDataType, false
}

func applyUnaryExpressionTypeHints(expression *pgsql.UnaryExpression) error {
	if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(expression.Operand); isPropertyLookup {
		expression.Operand = rewritePropertyLookupOperator(propertyLookup, pgsql.Boolean)
	}

	return nil
}

func inferBinaryExpressionType(expression *pgsql.BinaryExpression) (pgsql.DataType, error) {
	var (
		leftHint, isLeftHinted   = GetTypeHint(expression.LOperand)
		rightHint, isRightHinted = GetTypeHint(expression.ROperand)
	)

	if isLeftHinted {
		if isRightHinted {
			if higherLevelHint, matchesOrConverts := leftHint.OperatorResultType(rightHint, expression.Operator); !matchesOrConverts {
				return pgsql.UnsetDataType, fmt.Errorf("left and right operands for binary expression \"%s\" are not compatible: %s != %s", expression.Operator, leftHint, rightHint)
			} else {
				return higherLevelHint, nil
			}
		} else if inferredRightHint, err := InferExpressionType(expression.ROperand); err != nil {
			return pgsql.UnsetDataType, err
		} else if inferredRightHint == pgsql.UnknownDataType {
			// Assume the right side is convertable and return the left operand hint
			return leftHint, nil
		} else if upcastHint, matchesOrConverts := leftHint.OperatorResultType(inferredRightHint, expression.Operator); !matchesOrConverts {
			return pgsql.UnsetDataType, fmt.Errorf("left and right operands for binary expression \"%s\" are not compatible: %s != %s", expression.Operator, leftHint, inferredRightHint)
		} else {
			return upcastHint, nil
		}
	} else if isRightHinted {
		// There's no left type, attempt to infer it
		if inferredLeftHint, err := InferExpressionType(expression.LOperand); err != nil {
			return pgsql.UnsetDataType, err
		} else if inferredLeftHint == pgsql.UnknownDataType {
			// Assume the right side is convertable and return the left operand hint
			return rightHint, nil
		} else if upcastHint, matchesOrConverts := rightHint.OperatorResultType(inferredLeftHint, expression.Operator); !matchesOrConverts {
			return pgsql.UnsetDataType, fmt.Errorf("left and right operands for binary expression \"%s\" are not compatible: %s != %s", expression.Operator, rightHint, inferredLeftHint)
		} else {
			return upcastHint, nil
		}
	} else {
		// If neither side has specific type information then check the operator to see if it implies some type
		// hinting before resorting to inference
		switch expression.Operator {
		case pgsql.OperatorCypherStartsWith, pgsql.OperatorCypherContains, pgsql.OperatorCypherEndsWith:
			// String operations imply the operands must be text
			return pgsql.Text, nil

		case pgsql.OperatorAnd, pgsql.OperatorOr:
			// Boolean operators that the operands must be boolean
			return pgsql.Boolean, nil

		default:
			// The operator does not imply specific type information onto the operands. Attempt to infer any
			// information as a last ditch effort to type the AST nodes
			if inferredLeftHint, err := InferExpressionType(expression.LOperand); err != nil {
				return pgsql.UnsetDataType, err
			} else if inferredRightHint, err := InferExpressionType(expression.ROperand); err != nil {
				return pgsql.UnsetDataType, err
			} else if inferredLeftHint == pgsql.UnknownDataType && inferredRightHint == pgsql.UnknownDataType {
				// Unable to infer any type information, this may be resolved elsewhere so this is not explicitly
				// an error condition
				return pgsql.UnknownDataType, nil
			} else if higherLevelHint, matchesOrConverts := inferredLeftHint.OperatorResultType(inferredRightHint, expression.Operator); !matchesOrConverts {
				return pgsql.UnsetDataType, fmt.Errorf("left and right operands for binary expression \"%s\" are not compatible: %s != %s", expression.Operator, inferredLeftHint, inferredRightHint)
			} else {
				return higherLevelHint, nil
			}
		}
	}
}

func InferExpressionType(expression pgsql.Expression) (pgsql.DataType, error) {
	switch typedExpression := expression.(type) {
	case pgsql.Identifier, pgsql.RowColumnReference:
		return pgsql.UnknownDataType, nil

	case pgsql.CompoundIdentifier:
		if len(typedExpression) != 2 {
			return pgsql.UnsetDataType, fmt.Errorf("expected a compound identifier to have only 2 components but found: %d", len(typedExpression))
		}

		// Infer type information for well known column names
		switch typedExpression[1] {
		// TODO: Graph ID should be int2
		case pgsql.ColumnGraphID, pgsql.ColumnID, pgsql.ColumnStartID, pgsql.ColumnEndID:
			return pgsql.Int8, nil

		case pgsql.ColumnKindID:
			return pgsql.Int2, nil

		case pgsql.ColumnKindIDs:
			return pgsql.Int2Array, nil

		case pgsql.ColumnProperties:
			return pgsql.JSONB, nil

		default:
			return pgsql.UnknownDataType, nil
		}

	case pgsql.TypeHinted:
		return typedExpression.TypeHint(), nil

	case *pgsql.BinaryExpression:
		switch typedExpression.Operator {
		case pgsql.OperatorJSONTextField:
			// Text field lookups could be text or an unknown lookup - reduce it to an unknown type
			return pgsql.UnknownDataType, nil

		case pgsql.OperatorPropertyLookup, pgsql.OperatorJSONField:
			// This is unknown, not unset meaning that it can be re-cast by future inference inspections
			return pgsql.UnknownDataType, nil

		case pgsql.OperatorAnd, pgsql.OperatorOr, pgsql.OperatorEquals, pgsql.OperatorGreaterThan, pgsql.OperatorGreaterThanOrEqualTo,
			pgsql.OperatorLessThan, pgsql.OperatorLessThanOrEqualTo, pgsql.OperatorIn, pgsql.OperatorJSONBFieldExists,
			pgsql.OperatorLike, pgsql.OperatorILike, pgsql.OperatorPGArrayOverlap:
			return pgsql.Boolean, nil

		default:
			return inferBinaryExpressionType(typedExpression)
		}

	case *pgsql.Parenthetical:
		return InferExpressionType(typedExpression.Expression)

	default:
		slog.Info(fmt.Sprintf("unable to infer type hint for expression type: %T", expression))
		return pgsql.UnknownDataType, nil
	}
}

func applyTypeFunctionLikeTypeHints(expression *pgsql.BinaryExpression) error {
	switch typedLOperand := expression.LOperand.(type) {
	case pgsql.AnyExpression:
		if rOperandTypeHint, err := InferExpressionType(expression.ROperand); err != nil {
			return err
		} else {
			// In an any-expression where the type of the any-expression is unknown, attempt to infer it
			if !typedLOperand.CastType.IsKnown() {
				if rOperandArrayTypeHint, err := rOperandTypeHint.ToArrayType(); err != nil {
					return err
				} else {
					typedLOperand.CastType = rOperandArrayTypeHint
					expression.LOperand = typedLOperand
				}
			} else if !rOperandTypeHint.IsKnown() {
				expression.ROperand = pgsql.NewTypeCast(expression.ROperand, typedLOperand.CastType.ArrayBaseType())
			} else {
				// Validate against the array base type of the any-expression
				lOperandBaseType := typedLOperand.CastType.ArrayBaseType()

				if !lOperandBaseType.IsComparable(rOperandTypeHint, expression.Operator) {
					return fmt.Errorf("function call has return signature of type %s but is being compared using operator %s against type %s", typedLOperand.CastType, expression.Operator, rOperandTypeHint)
				}
			}
		}

	case pgsql.FunctionCall:
		if rOperandTypeHint, err := InferExpressionType(expression.ROperand); err != nil {
			return err
		} else {
			if !typedLOperand.CastType.IsKnown() {
				typedLOperand.CastType = rOperandTypeHint
				expression.LOperand = typedLOperand
			}

			if pgsql.OperatorIsComparator(expression.Operator) && !typedLOperand.CastType.IsComparable(rOperandTypeHint, expression.Operator) {
				return newFunctionCallComparatorError(typedLOperand, expression.Operator, rOperandTypeHint)
			}
		}
	}

	switch typedROperand := expression.ROperand.(type) {
	case pgsql.AnyExpression:
		if lOperandTypeHint, err := InferExpressionType(expression.LOperand); err != nil {
			return err
		} else {
			// In an any-expression where the type of the any-expression is unknown, attempt to infer it
			if !typedROperand.CastType.IsKnown() {
				if !lOperandTypeHint.IsKnown() {
					// If the left operand has no type information then assume this is a castable any array
					typedROperand.CastType = pgsql.AnyArray
				} else if rOperandArrayTypeHint, err := lOperandTypeHint.ToArrayType(); err != nil {
					return err
				} else {
					typedROperand.CastType = rOperandArrayTypeHint
					expression.ROperand = typedROperand
				}
			} else if !lOperandTypeHint.IsKnown() {
				expression.LOperand = pgsql.NewTypeCast(expression.LOperand, typedROperand.CastType.ArrayBaseType())
			} else {
				// Validate against the array base type of the any-expression
				rOperandBaseType := typedROperand.CastType.ArrayBaseType()

				if !typedROperand.CastType.IsComparable(lOperandTypeHint, expression.Operator) && !rOperandBaseType.IsComparable(lOperandTypeHint, expression.Operator) {
					return fmt.Errorf("function call has return signature of type %s but is being compared using operator %s against type %s", typedROperand.CastType, expression.Operator, lOperandTypeHint)
				}
			}
		}

	case pgsql.FunctionCall:
		if lOperandTypeHint, err := InferExpressionType(expression.LOperand); err != nil {
			return err
		} else {
			if !typedROperand.CastType.IsKnown() {
				typedROperand.CastType = lOperandTypeHint
				expression.ROperand = typedROperand
			} else if !lOperandTypeHint.IsKnown() {
				expression.LOperand = pgsql.NewTypeCast(expression.LOperand, typedROperand.CastType.ArrayBaseType())
			} else if pgsql.OperatorIsComparator(expression.Operator) && !typedROperand.CastType.IsComparable(lOperandTypeHint, expression.Operator) {
				return newFunctionCallComparatorError(typedROperand, expression.Operator, lOperandTypeHint)
			}
		}
	}

	return nil
}

func applyBinaryExpressionTypeHints(expression *pgsql.BinaryExpression) error {
	switch expression.Operator {
	case pgsql.OperatorPropertyLookup:
		// Don't directly hint property lookups but replace the operator with the JSON operator
		expression.Operator = pgsql.OperatorJSONTextField
		return nil
	}

	if err := rewritePropertyLookupOperands(expression); err != nil {
		return err
	}

	return applyTypeFunctionLikeTypeHints(expression)
}
