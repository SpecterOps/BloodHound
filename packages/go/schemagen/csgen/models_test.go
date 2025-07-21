// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package csgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_TopLevel(t *testing.T) {
	topLevel := Namespace{
		Name: "SharpHoundCommon.Enums",
		Nodes: []SyntaxNode{
			Class{
				Modifiers:  []Modifier{ModifierStatic},
				Visibility: VisibilityPublic,
				Name:       "PropertyNames",
				Nodes: []SyntaxNode{
					BinaryExpression{
						LeftOperand: ClassMemberAssignment{
							Visibility: VisibilityPublic,
							Modifier:   Modifiers{ModifierConst},
							Type:       TypeString,
							Symbol:     "Enroll",
						},
						Operator: OperatorEquals,
						RightOperand: Literal{
							Value: "Enroll",
						},
					},
				},
			},
		},
	}

	expected := "namespace SharpHoundCommon.Enums {\npublic static class PropertyNames {\npublic const string Enroll = \"Enroll\";\n\n}\n}"

	builder := NewOutputBuilder()

	err := WalkSyntaxTree(topLevel, builder)
	require.NoError(t, err)
	output := builder.Build()
	require.True(t, len(output) > 0)
	require.Equal(t, output, expected)
}
