package pgtransition_test

import (
	"github.com/specterops/bloodhound/cypher/backend/pgsql/pgtransition"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test"
	"github.com/stretchr/testify/require"
	"testing"
)

type kindMapper struct{}

func (k kindMapper) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	return make([]int16, len(kinds)), nil
}

func TestTranslateAllShortestPaths(t *testing.T) {
	builder := query.NewBuilder(&query.Cache{})
	builder.Apply(query.Where(
		query.And(
			query.And(query.Equals(query.StartID(), graph.ID(1)), query.Equals(query.EndProperty("name"), "1")),
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor),
			query.Equals(query.EndID(), graph.ID(5)),
		),
	))

	aspArguments, err := pgtransition.TranslateAllShortestPaths(builder.RegularQuery(), kindMapper{})
	test.RequireNilErr(t, err)

	require.Equal(t, "s.id = 1", aspArguments.RootCriteria, "Root Criteria")
	require.Equal(t, "(r.kind_id = any(array[0]::int2[]) or r.kind_id = any(array[0]::int2[]) or r.kind_id = any(array[0]::int2[]) or r.kind_id = any(array[0]::int2[]))", aspArguments.TraversalCriteria, "Traversal Criteria")
	require.Equal(t, "e.properties->'name' = '1' and e.id = 5", aspArguments.TerminalCriteria, "Terminal Criteria")
}
