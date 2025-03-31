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

	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

type PropertyLookup struct {
	Reference pgsql.CompoundIdentifier
	Field     string
}

func expressionToPropertyLookupBinaryExpression(expression pgsql.Expression) (*pgsql.BinaryExpression, bool) {
	if nextExpression := expression; nextExpression != nil {
		for {
			switch typedExpression := nextExpression.(type) {
			case pgsql.AnyExpression:
				// This is here to unwrap Any expressions that have been passed in as a property lookup. This is
				// common when dealing with array operators. In the future this check should be handled by the
				// caller to simplify the logic here.
				nextExpression = typedExpression.Expression
				continue

			case *pgsql.BinaryExpression:
				return typedExpression, pgsql.OperatorIsPropertyLookup(typedExpression.Operator)
			}

			// Break this loop on all other matches so that flow hits the outer return
			break
		}
	}

	return nil, false
}

func binaryExpressionToPropertyLookup(expression *pgsql.BinaryExpression) (PropertyLookup, error) {
	if reference, typeOK := expression.LOperand.(pgsql.CompoundIdentifier); !typeOK {
		return PropertyLookup{}, fmt.Errorf("expected left operand for property lookup to be a compound identifier but found type: %T", expression.LOperand)
	} else if field, typeOK := expression.ROperand.(pgsql.Literal); !typeOK {
		return PropertyLookup{}, fmt.Errorf("expected right operand for property lookup to be a literal but found type: %T", expression.ROperand)
	} else if field.CastType != pgsql.Text {
		return PropertyLookup{}, fmt.Errorf("expected property lookup field a string literal but found data type: %s", field.CastType)
	} else if stringField, typeOK := field.Value.(string); !typeOK {
		return PropertyLookup{}, fmt.Errorf("expected property lookup field a string literal but found data type: %T", field)
	} else {
		return PropertyLookup{
			Reference: reference,
			Field:     stringField,
		}, nil
	}
}

func decomposePropertyLookup(expression pgsql.Expression) (PropertyLookup, error) {
	if propertyLookupBinExp, isPropertyLookup := expressionToPropertyLookupBinaryExpression(expression); !isPropertyLookup {
		return PropertyLookup{}, fmt.Errorf("expected binary expression for property lookup decomposition but found type: %T", expression)
	} else {
		return binaryExpressionToPropertyLookup(propertyLookupBinExp)
	}
}
