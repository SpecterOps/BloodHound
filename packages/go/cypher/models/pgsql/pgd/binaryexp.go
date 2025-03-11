package pgd

import "github.com/specterops/bloodhound/cypher/models/pgsql"

func Not(operand pgsql.Expression) *pgsql.UnaryExpression {
	return pgsql.NewUnaryExpression(pgsql.OperatorNot, operand)
}

func And(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorAnd, rOperand)
}

func LessThan(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorLessThan, rOperand)
}

func Equals(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorEquals, rOperand)
}

func Concatenate(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorConcatenate, rOperand)
}

func EdgeHasKind(edge pgsql.Identifier, kindID int16) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(
		Column(edge, pgsql.ColumnKindID),
		pgsql.OperatorEquals,
		IntLiteral(kindID),
	)
}

func PropertyLookup(owner pgsql.Identifier, propertyName string) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(
		Properties(owner),
		pgsql.OperatorJSONTextField,
		TextLiteral(propertyName),
	)
}

func Add(lOperand, rOperand pgsql.Expression) *pgsql.BinaryExpression {
	return pgsql.NewBinaryExpression(lOperand, pgsql.OperatorAdd, rOperand)
}
