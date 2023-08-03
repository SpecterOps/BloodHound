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

package model

import (
	"fmt"
	"strings"
)

var (
	GreedyRangeQuantifier = NewRangeQuantifier("*")
)

const (
	TokenLiteralAsterisk = "*"

	QuantifierTypeAll     = QuantifierType("all")
	QuantifierTypeAny     = QuantifierType("any")
	QuantifierTypeNone    = QuantifierType("none")
	QuantifierTypeSingle  = QuantifierType("single")
	QuantifierTypeInvalid = QuantifierType("")

	SortAscending  = SortOrder("asc")
	SortDescending = SortOrder("desc")

	OperatorAssignment         = AssignmentOperator("=")
	OperatorAdditionAssignment = AssignmentOperator("+=")
	OperatorLabelAssignment    = AssignmentOperator("")

	OperatorAdd                  = Operator("+")
	OperatorSubtract             = Operator("-")
	OperatorMultiply             = Operator("*")
	OperatorDivide               = Operator("/")
	OperatorModulo               = Operator("%")
	OperatorPowerOf              = Operator("^")
	OperatorEquals               = Operator("=")
	OperatorRegexMatch           = Operator("=~")
	OperatorNotEquals            = Operator("<>")
	OperatorGreaterThan          = Operator(">")
	OperatorGreaterThanOrEqualTo = Operator(">=")
	OperatorLessThan             = Operator("<")
	OperatorLessThanOrEqualTo    = Operator("<=")
	OperatorStartsWith           = Operator("starts with")
	OperatorEndsWith             = Operator("ends with")
	OperatorContains             = Operator("contains")
	OperatorIn                   = Operator("in")
	OperatorIs                   = Operator("is")
	OperatorIsNot                = Operator("is not")
	OperatorNot                  = Operator("not")
	OperatorInvalid              = Operator("")
)

func ParseOperator(operator string) (Operator, error) {
	switch Operator(strings.ToLower(operator)) {
	case OperatorAdd:
		return OperatorAdd, nil
	case OperatorSubtract:
		return OperatorSubtract, nil
	case OperatorMultiply:
		return OperatorMultiply, nil
	case OperatorDivide:
		return OperatorDivide, nil
	case OperatorModulo:
		return OperatorModulo, nil
	case OperatorPowerOf:
		return OperatorPowerOf, nil
	case OperatorEquals:
		return OperatorEquals, nil
	case OperatorRegexMatch:
		return OperatorRegexMatch, nil
	case OperatorNotEquals:
		return OperatorNotEquals, nil
	case OperatorGreaterThan:
		return OperatorGreaterThan, nil
	case OperatorGreaterThanOrEqualTo:
		return OperatorGreaterThanOrEqualTo, nil
	case OperatorLessThan:
		return OperatorLessThan, nil
	case OperatorLessThanOrEqualTo:
		return OperatorLessThanOrEqualTo, nil
	case OperatorStartsWith:
		return OperatorStartsWith, nil
	case OperatorEndsWith:
		return OperatorEndsWith, nil
	case OperatorContains:
		return OperatorContains, nil
	case OperatorIn:
		return OperatorIn, nil
	case OperatorIs:
		return OperatorIs, nil
	case OperatorIsNot:
		return OperatorIsNot, nil
	case OperatorNot:
		return OperatorNot, nil
	case OperatorInvalid:
		fallthrough
	default:
		return OperatorInvalid, fmt.Errorf("invalid operator: \"%s\"", operator)
	}
}
