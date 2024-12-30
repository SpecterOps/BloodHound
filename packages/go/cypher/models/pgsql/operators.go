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

package pgsql

type Operator string

func (s Operator) IsIn(others ...Operator) bool {
	for _, other := range others {
		if s == other {
			return true
		}
	}

	return false
}

func (s Operator) AsExpression() Expression {
	return s
}

func (s Operator) String() string {
	return string(s)
}

func (s Operator) NodeType() string {
	return "operator"
}

func OperatorIsIn(operator Expression, matchers ...Expression) bool {
	for _, matcher := range matchers {
		if operator == matcher {
			return true
		}
	}

	return false
}

func OperatorIsBoolean(operator Expression) bool {
	return OperatorIsIn(operator,
		OperatorAnd,
		OperatorOr,
		OperatorNot,
		OperatorEquals,
		OperatorNotEquals,
		OperatorGreaterThan,
		OperatorGreaterThanOrEqualTo,
		OperatorLessThan,
		OperatorLessThanOrEqualTo)
}

func OperatorIsPropertyLookup(operator Expression) bool {
	return OperatorIsIn(operator,
		OperatorPropertyLookup,
		OperatorJSONField,
		OperatorJSONTextField,
	)
}

const (
	UnsetOperator                Operator = ""
	OperatorUnion                Operator = "union"
	OperatorConcatenate          Operator = "||"
	OperatorArrayOverlap         Operator = "&&"
	OperatorEquals               Operator = "="
	OperatorNotEquals            Operator = "!="
	OperatorGreaterThan          Operator = ">"
	OperatorGreaterThanOrEqualTo Operator = ">="
	OperatorLessThan             Operator = "<"
	OperatorLessThanOrEqualTo    Operator = "<="
	OperatorLike                 Operator = "like"
	OperatorILike                Operator = "ilike"
	OperatorPGArrayOverlap       Operator = "operator (pg_catalog.&&)"
	OperatorAnd                  Operator = "and"
	OperatorOr                   Operator = "or"
	OperatorNot                  Operator = "not"
	OperatorJSONBFieldExists     Operator = "?"
	OperatorJSONField            Operator = "->"
	OperatorJSONTextField        Operator = "->>"
	OperatorAdd                  Operator = "+"
	OperatorSubtract             Operator = "-"
	OperatorMultiply             Operator = "*"
	OperatorDivide               Operator = "/"
	OperatorIn                   Operator = "in"
	OperatorIs                   Operator = "is"
	OperatorIsNot                Operator = "is not"
	OperatorSimilarTo            Operator = "similar to"
	OperatorRegexMatch           Operator = "~"
	OperatorAssignment           Operator = "="
	OperatorAdditionAssignment   Operator = "+="

	OperatorCypherRegexMatch Operator = "=~"
	OperatorCypherStartsWith Operator = "starts with"
	OperatorCypherContains   Operator = "contains"
	OperatorCypherEndsWith   Operator = "ends with"

	OperatorPropertyLookup Operator = "property_lookup"
	OperatorKindAssignment Operator = "kind_assignment"
)
