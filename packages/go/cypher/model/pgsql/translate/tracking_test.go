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

// todo: wtf is constraint tracker
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

func TestTrackedIdentifier(t *testing.T) {

	var (
		trackedNodeIdent = translate.TrackedIdentifier{
			Identifier: pgsql.Identifier("whatsup"),
			DataType:   pgsql.NodeComposite,
		}

		expectedNodeComposite = pgsql.CompositeValue{
			Values: []pgsql.Expression{
				pgsql.CompoundIdentifier{trackedNodeIdent.Identifier, pgsql.ColumnID},
				pgsql.CompoundIdentifier{trackedNodeIdent.Identifier, pgsql.ColumnKindIDs},
				pgsql.CompoundIdentifier{trackedNodeIdent.Identifier, pgsql.ColumnProperties},
			},
			DataType: pgsql.NodeComposite,
		}

		trackedEdgeIdent = translate.TrackedIdentifier{
			Identifier: pgsql.Identifier("whatsup"),
			DataType:   pgsql.EdgeComposite,
		}

		expectedEdgeComposite = pgsql.CompositeValue{
			Values: []pgsql.Expression{
				pgsql.CompoundIdentifier{trackedEdgeIdent.Identifier, pgsql.ColumnID},
				pgsql.CompoundIdentifier{trackedEdgeIdent.Identifier, pgsql.ColumnStartID},
				pgsql.CompoundIdentifier{trackedEdgeIdent.Identifier, pgsql.ColumnEndID},
				pgsql.CompoundIdentifier{trackedEdgeIdent.Identifier, pgsql.ColumnKindID},
				pgsql.CompoundIdentifier{trackedEdgeIdent.Identifier, pgsql.ColumnProperties},
			},
			DataType: pgsql.EdgeComposite,
		}
	)

	// >>> build FromClause
	fromClauses, err := trackedNodeIdent.BuildFromClauses()
	require.Nil(t, err)
	require.Equal(t,
		[]pgsql.FromClause{{
			Relation: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{
					trackedNodeIdent.Identifier,
				},
			},
		}},
		fromClauses,
	)

	// >>> build CompositeValue
	// then result of this func is typically used in a top-level select like: `select (<trackedIdent>.id, <trackedIdent>.kind_ids, <trackedIdent>.properties)::nodecomposite as <alias>` where <trackedIdent> is likely to be of the form n0, n1, ... nX`
	compositeNodeValue, err := trackedNodeIdent.BuildCompositeValue()
	require.Nil(t, err)
	require.Equal(t, expectedNodeComposite, compositeNodeValue)

	compositeEdgeValue, err := trackedEdgeIdent.BuildCompositeValue()
	require.Nil(t, err)
	require.Equal(t, expectedEdgeComposite, compositeEdgeValue)

	// >>> build Projection
	proj, err := trackedNodeIdent.BuildProjection()
	require.Nil(t, err)
	require.Equal(t, expectedNodeComposite, proj)

}
