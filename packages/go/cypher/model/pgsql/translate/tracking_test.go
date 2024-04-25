package translate_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/cypher/model/pgsql/translate"
	"github.com/stretchr/testify/require"
)

func TestIdentifierGenerator(t *testing.T) {

	identifierGenerator := translate.NewIdentifierGenerator()

	// default case
	identifier, err := identifierGenerator.NewIdentifier(pgsql.UnknownDataType)
	require.Equal(t, "identifier with data type UNKNOWN does not have a prefix case", err.Error())
	require.Equal(t, pgsql.Identifier(""), identifier)

	// node identifers
	identifier, err = identifierGenerator.NewIdentifier(pgsql.NodeComposite)
	require.Nil(t, err)
	require.Equal(t, pgsql.Identifier("n0"), identifier)

	identifier, err = identifierGenerator.NewIdentifier(pgsql.NodeComposite)
	require.Nil(t, err)
	require.Equal(t, pgsql.Identifier("n1"), identifier)

	// edge identifiers
	identifier, err = identifierGenerator.NewIdentifier(pgsql.EdgeComposite)
	require.Nil(t, err)
	require.Equal(t, pgsql.Identifier("e0"), identifier)

	identifier, err = identifierGenerator.NewIdentifier(pgsql.EdgeComposite)
	require.Nil(t, err)
	require.Equal(t, pgsql.Identifier("e1"), identifier)
}

func TestConstraintTracker(t *testing.T) {

	tracker := translate.NewConstraintTracker()

	var (
		deps pgsql.IdentifierSet = pgsql.AsIdentifierSet(
			pgsql.Identifier("hello"),
			pgsql.Identifier("world"),
		)
		expression pgsql.Expression = &pgsql.BinaryExpression{
			Operator: pgsql.OperatorPGArrayOverlap,
			LOperand: pgsql.CompoundIdentifier{"hello", pgsql.ColumnKindIDs},
			ROperand: pgsql.AsLiteral("some_kind"),
		}
	)

	/* BinaryExpr
	LOperand 	hello.kind_ids
	Operator 	operator (pg_catalog.&&)
	ROperand	array []::int2[]
	*/
	tracker.Constrain(deps, expression)

	result, expr := tracker.Consume(deps)

	fmt.Println(result, expr)
}
