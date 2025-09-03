package changelog

import (
	"testing"

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestChangeCache(t *testing.T) {
	t.Run("node doesn't exist in cache. return true.", func(t *testing.T) {
		c := newCache()

		node := &NodeChange{
			NodeID:     "123",
			Kinds:      nil,
			Properties: &graph.Properties{Map: map[string]any{"foo": "bar"}},
		}

		shouldSubmit, err := c.shouldSubmit(node)
		require.NoError(t, err)
		require.True(t, shouldSubmit)

	})

	t.Run("node is cached with same properties. return false", func(t *testing.T) {
		var (
			c      = newCache()
			change = &NodeChange{
				NodeID:     "123",
				Kinds:      nil,
				Properties: &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash      = change.IdentityKey()
			dataHash, _ = change.Hash()
		)

		// simulate a full cache
		c.data[idHash] = dataHash

		shouldSubmit, err := c.shouldSubmit(change)
		require.NoError(t, err)
		require.False(t, shouldSubmit)
	})

	t.Run("node is cached. properties have a diff. return true", func(t *testing.T) {
		var (
			c         = newCache()
			oldChange = &NodeChange{
				NodeID:     "123",
				Kinds:      nil,
				Properties: &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash      = oldChange.IdentityKey()
			dataHash, _ = oldChange.Hash()
		)

		// simulate a populated cache
		c.data[idHash] = dataHash

		newChange := &NodeChange{
			NodeID:     "123",
			Kinds:      nil,
			Properties: &graph.Properties{Map: map[string]any{"changed": 1}},
		}

		shouldSubmit, err := c.shouldSubmit(newChange)
		require.NoError(t, err)
		require.True(t, shouldSubmit)
	})

	t.Run("node is cached. kinds have a diff. return true", func(t *testing.T) {
		var (
			c         = newCache()
			oldChange = &NodeChange{
				NodeID:     "123",
				Kinds:      nil,
				Properties: &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash      = oldChange.IdentityKey()
			dataHash, _ = oldChange.Hash()
		)

		// simulate a populated cache
		c.data[idHash] = dataHash

		newChange := &NodeChange{
			NodeID:     "123",
			Kinds:      graph.StringsToKinds([]string{"kindA"}),
			Properties: &graph.Properties{Map: map[string]any{"changed": 1}},
		}

		shouldSubmit, err := c.shouldSubmit(newChange)
		require.NoError(t, err)
		require.True(t, shouldSubmit)
	})

	t.Run("edge doesn't exist in cache. return true.", func(t *testing.T) {
		c := newCache()

		node := &EdgeChange{
			SourceNodeID: "123",
			TargetNodeID: "456",
			Kind:         graph.StringKind("kindA"),
			Properties:   &graph.Properties{Map: map[string]any{"foo": "bar"}},
		}

		shouldSubmit, err := c.shouldSubmit(node)
		require.NoError(t, err)
		require.True(t, shouldSubmit)
	})

	t.Run("edge is cached. properties have a diff. return true", func(t *testing.T) {
		var (
			c         = newCache()
			oldChange = &EdgeChange{
				SourceNodeID: "123",
				TargetNodeID: "456",
				Kind:         graph.StringKind("a"),
				Properties:   &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash      = oldChange.IdentityKey()
			dataHash, _ = oldChange.Hash()
		)

		// simulate a populated cache
		c.data[idHash] = dataHash

		newChange := &EdgeChange{
			SourceNodeID: "123",
			TargetNodeID: "456",
			Kind:         graph.StringKind("a"),
			Properties:   &graph.Properties{Map: map[string]any{"changed": 1}},
		}

		shouldSubmit, err := c.shouldSubmit(newChange)
		require.NoError(t, err)
		require.True(t, shouldSubmit)
	})
}
