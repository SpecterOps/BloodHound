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

package translate_test

import (
	"testing"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
	"github.com/specterops/bloodhound/cypher/models/pgsql/translate"
	"github.com/specterops/bloodhound/src/test"
	"github.com/stretchr/testify/require"
)

func mustDefineNew(t *testing.T, scope *translate.Scope, dataType pgsql.DataType) *translate.BoundIdentifier {
	binding, err := scope.DefineNew(dataType)
	require.Nil(t, err)

	return binding
}

func mustPushFrame(t *testing.T, scope *translate.Scope) *translate.Frame {
	frame, err := scope.PushFrame()
	require.Nil(t, err)

	return frame
}

func TestRewriteFrameBindings(t *testing.T) {
	type testCase struct {
		Case     pgsql.Expression
		Expected pgsql.Expression
	}

	var (
		scope = translate.NewScope()
		frame = mustPushFrame(t, scope)
		a     = mustDefineNew(t, scope, pgsql.Int)
	)

	frame.Reveal(a.Identifier)
	frame.Export(a.Identifier)

	a.LastProjection = frame

	// Cases
	testCases := []testCase{{
		Case: &pgsql.Parenthetical{
			Expression: a.Identifier,
		},
		Expected: &pgsql.Parenthetical{
			Expression: pgsql.CompoundIdentifier{frame.Binding.Identifier, a.Identifier},
		},
	}, {
		Case:     pgsql.NewBinaryExpression(a.Identifier, pgsql.OperatorEquals, a.Identifier),
		Expected: pgsql.NewBinaryExpression(pgsql.CompoundIdentifier{frame.Binding.Identifier, a.Identifier}, pgsql.OperatorEquals, pgsql.CompoundIdentifier{frame.Binding.Identifier, a.Identifier}),
	}, {
		Case: &pgsql.AliasedExpression{
			Expression: a.Identifier,
			Alias:      pgsql.AsOptionalIdentifier("name"),
		},
		Expected: &pgsql.AliasedExpression{
			Expression: pgsql.CompoundIdentifier{frame.Binding.Identifier, a.Identifier},
			Alias:      pgsql.AsOptionalIdentifier("name"),
		},
	}}

	for _, nextTestCase := range testCases {
		t.Run("", func(t *testing.T) {
			test.RequireNilErr(t, translate.RewriteFrameBindings(scope, nextTestCase.Case))

			var (
				formattedCase, formattedCaseErr         = format.SyntaxNode(nextTestCase.Case)
				formattedExpected, formattedExpectedErr = format.SyntaxNode(nextTestCase.Expected)
			)

			test.RequireNilErr(t, formattedCaseErr)
			test.RequireNilErr(t, formattedExpectedErr)

			require.Equal(t, formattedExpected, formattedCase)
		})
	}
}
