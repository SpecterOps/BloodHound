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

package pgd

import "github.com/specterops/bloodhound/cypher/models/pgsql"

func Not(operand pgsql.Expression) *pgsql.UnaryExpression {
	return pgsql.NewUnaryExpression(pgsql.OperatorNot, operand)
}

func And(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorAnd, rOperand)
}

func LessThan(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorLessThan, rOperand)
}

func LessThanOrEqualTo(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorLessThanOrEqualTo, rOperand)
}

func GreaterThan(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorGreaterThan, rOperand)
}

func GreaterThanOrEqualTo(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorGreaterThanOrEqualTo, rOperand)
}

func Equals(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorEquals, rOperand)
}

func Concatenate(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorConcatenate, rOperand)
}

func EdgeHasKind(edge pgsql.Identifier, kindID int16) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(
		Column(edge, pgsql.ColumnKindID),
		pgsql.OperatorEquals,
		IntLiteral(kindID),
	)
}

func PropertyLookup(owner pgsql.Identifier, propertyName string) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(
		Properties(owner),
		pgsql.OperatorJSONTextField,
		TextLiteral(propertyName),
	)
}

func Add(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorAdd, rOperand)
}
