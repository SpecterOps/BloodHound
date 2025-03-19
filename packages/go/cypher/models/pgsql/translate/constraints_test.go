package translate

import (
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/pgd"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMeasureSelectivity(t *testing.T) {
	selectivity, err := MeasureSelectivity(pgd.Equals(
		pgsql.Identifier("123"),
		pgsql.Identifier("456"),
	))

	require.Nil(t, err)
	require.Equal(t, 30, selectivity)
}
