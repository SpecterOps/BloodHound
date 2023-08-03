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
	"testing"

	"github.com/stretchr/testify/require"
)

func validateParsing(t *testing.T, expected Operator, operatorStr string) {
	parsed, err := ParseOperator(operatorStr)

	require.Nil(t, err)
	require.Equal(t, expected, parsed)
}

func TestParseOperator(t *testing.T) {
	validateParsing(t, OperatorAdd, "+")
	validateParsing(t, OperatorSubtract, "-")
	validateParsing(t, OperatorMultiply, "*")
	validateParsing(t, OperatorDivide, "/")
	validateParsing(t, OperatorModulo, "%")
	validateParsing(t, OperatorPowerOf, "^")
	validateParsing(t, OperatorEquals, "=")
	validateParsing(t, OperatorRegexMatch, "=~")
	validateParsing(t, OperatorNotEquals, "<>")
	validateParsing(t, OperatorGreaterThan, ">")
	validateParsing(t, OperatorGreaterThanOrEqualTo, ">=")
	validateParsing(t, OperatorLessThan, "<")
	validateParsing(t, OperatorLessThanOrEqualTo, "<=")
	validateParsing(t, OperatorStartsWith, "starts with")
	validateParsing(t, OperatorEndsWith, "ends with")
	validateParsing(t, OperatorContains, "contains")
	validateParsing(t, OperatorIn, "in")
	validateParsing(t, OperatorIs, "is")
	validateParsing(t, OperatorIsNot, "is not")
	validateParsing(t, OperatorNot, "not")

	parsedInvalid, err := ParseOperator("!@#(*)$(*!)(%")

	require.NotNil(t, err)
	require.Equal(t, OperatorInvalid, parsedInvalid)
}
