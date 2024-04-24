package pgsql

import "fmt"

type SyntaxNode interface {
	NodeType() string
}

type Statement interface {
	SyntaxNode
	AsStatement() Statement
}

type Expression interface {
	SyntaxNode
	AsExpression() Expression
}

type Projection interface {
	Expression
	AsProjection() Projection
}

func ExpressionAs[T any](expression Expression) (T, error) {
	var emptyT T

	if expression == nil {
		return emptyT, nil
	}

	if projection, isT := expression.(T); isT {
		return projection, nil
	}

	return emptyT, fmt.Errorf("type %T is not a projection", expression)
}

type MergeAction interface {
	Expression
	AsMergeAction() MergeAction
}

type SetExpression interface {
	Expression
	AsSetExpression() SetExpression
}

type ConflictAction interface {
	Expression
	AsConflictAction() ConflictAction
}
