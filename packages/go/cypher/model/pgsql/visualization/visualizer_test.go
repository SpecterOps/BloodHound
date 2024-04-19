package visualization

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/model/pgsql/translate"
	"github.com/stretchr/testify/require"
)

func TestGraphToPUMLDigraph(t *testing.T) {
	regularQuery, err := frontend.ParseCypher(frontend.NewContext(), "match (s), (e) where s.name = s.other + 1 / s.last and s.value = 1234 and not s.test and e.value = 1234 and e.comp = s.comp return s")
	require.Nil(t, err)

	sqlAST, err := translate.Translate(regularQuery)
	require.Nil(t, err)

	graph, err := SQLToDigraph(sqlAST)
	require.Nil(t, err)

	home, err := os.UserHomeDir()
	require.Nil(t, err)

	fout, err := os.OpenFile(filepath.Join(home, "graph.puml"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	defer fout.Close()

	require.Nil(t, err)
	require.Nil(t, GraphToPUMLDigraph(graph, fout))
}
