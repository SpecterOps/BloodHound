package translate

import (
	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/cypher/model/walk"
)

type IdentifierRewriter struct {
	walk.HierarchicalVisitor[pgsql.SyntaxNode]

	old pgsql.Identifier
	new pgsql.Identifier
}

func (s *IdentifierRewriter) Enter(node pgsql.SyntaxNode) {
	switch typedExpression := node.(type) {
	case *pgsql.BinaryExpression:
		switch typedLOperand := typedExpression.LOperand.(type) {
		case pgsql.CompoundIdentifier:
			typedLOperand.Replace(s.old, s.new)

		case pgsql.Identifier:
			if s.old == typedLOperand {
				typedExpression.LOperand = s.new
			}
		}

		switch typedROperand := typedExpression.ROperand.(type) {
		case pgsql.CompoundIdentifier:
			typedROperand.Replace(s.old, s.new)

		case pgsql.Identifier:
			if s.old == typedROperand {
				typedExpression.LOperand = s.new
			}
		}
	}
}

func NewIdentifierRewriter(old, new pgsql.Identifier) walk.HierarchicalVisitor[pgsql.SyntaxNode] {
	return &IdentifierRewriter{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[pgsql.SyntaxNode](),
		old:                 old,
		new:                 new,
	}
}

func RewriteExpressionIdentifiers(expression pgsql.Expression, old, new pgsql.Identifier) error {
	return walk.PgSQL(expression, NewIdentifierRewriter(old, new))
}
