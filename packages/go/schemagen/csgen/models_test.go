package csgen

import (
	"github.com/stretchr/testify/require"
	"testing"
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
