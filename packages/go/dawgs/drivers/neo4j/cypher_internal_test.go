package neo4j

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StripCypher(t *testing.T) {
	var (
		query = "match (u1:User {domain: \"DOMAIN1\"}), (u2:User {domain: \"DOMAIN2\"}) where u1.samaccountname <> \"krbtgt\" and u1.samaccountname = u2.samaccountname with u2 match p1 = (u2)-[*1..]->(g:Group) with p1 match p2 = (u2)-[*1..]->(g:Group) return p1, p2"
	)

	result := stripCypherQuery(query)

	require.Equalf(t, false, strings.Contains(result, "DOMAIN1"), "Cypher query not sanitized. Contains sensitive value: %s", result)
	require.Equalf(t, false, strings.Contains(result, "DOMAIN2"), "Cypher query not sanitized. Contains sensitive value: %s", result)
}
