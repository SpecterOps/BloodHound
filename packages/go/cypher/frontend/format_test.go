package frontend

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCypherEmitter_StripLiterals(t *testing.T) {
	var (
		buffer            = &bytes.Buffer{}
		regularQuery, err = ParseCypher(DefaultCypherContext(), "match (n {value: 'PII'}) where n.other = 'more pii' and n.number = 411 return n.name, n")
		emitter           = CypherEmitter{
			StripLiterals: true,
		}
	)

	require.Nil(t, err)
	require.Nil(t, emitter.Write(regularQuery, buffer))
	require.Equal(t, "match (n {value: $STRIPPED}) where n.other = $STRIPPED and n.number = $STRIPPED return n.name, n", buffer.String())
}
