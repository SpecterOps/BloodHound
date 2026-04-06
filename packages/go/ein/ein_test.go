package ein

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validKinds() graph.Kinds {
	return slicesext.Concat(ad.NodeKinds(), ad.Relationships(), azure.NodeKinds(), azure.Relationships())
}

func TestParseKind(t *testing.T) {
	t.Run("all known strings map to their graph.Kind", func(t *testing.T) {
		for _, k := range validKinds() {
			res, err := ParseKind(k.String())
			require.Nil(t, err)
			assert.Equal(t, k, res, "expect string to map back to original kind")
		}
	})

	t.Run("unknown kind strings cause an error", func(t *testing.T) {
		_, err := ParseKind("unsupported")
		assert.Contains(t, err.Error(), "unsupported", "error contains unsupported kind string")
	})
}
