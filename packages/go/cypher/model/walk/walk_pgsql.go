package walk

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
)

func newSQLWalkCursor(expression pgsql.SyntaxNode) (*Cursor[pgsql.SyntaxNode], error) {
	switch typedExpression := expression.(type) {
	case pgsql.CompoundIdentifier, pgsql.Operator, pgsql.Literal:
		return &Cursor[pgsql.SyntaxNode]{
			Expression: expression,
		}, nil

	case *pgsql.UnaryExpression:
		return &Cursor[pgsql.SyntaxNode]{
			Expression: expression,
			Branches:   []pgsql.SyntaxNode{typedExpression.Operator, typedExpression.Operand},
		}, nil

	case *pgsql.BinaryExpression:
		return &Cursor[pgsql.SyntaxNode]{
			Expression: expression,
			Branches:   []pgsql.SyntaxNode{typedExpression.LOperand, typedExpression.Operator, typedExpression.ROperand},
		}, nil

	default:
		return nil, fmt.Errorf("unable to negotiate sql type %T into a translation cursor", expression)
	}
}
