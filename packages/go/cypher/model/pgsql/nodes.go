package pgsql

type StringLike interface {
	String() string
}

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
