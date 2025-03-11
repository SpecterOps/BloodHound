package pgd

import (
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func Any(expr pgsql.Expression, castType pgsql.DataType) *pgsql.AnyExpression {
	return pgsql.NewAnyExpression(expr, castType)
}

func Column(root, column pgsql.Identifier) pgsql.CompoundIdentifier {
	return pgsql.CompoundIdentifier{root, column}
}

func EntityID(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnID)
}

func StartID(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnStartID)
}

func EndID(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnEndID)
}

func Properties(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnProperties)
}
