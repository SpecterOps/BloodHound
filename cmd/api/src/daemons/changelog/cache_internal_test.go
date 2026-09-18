// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package changelog

import (
	"testing"

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestChangeCache(t *testing.T) {
	t.Run("node doesn't exist in cache. return true.", func(t *testing.T) {
		c := newCache(2)

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
			c      = newCache(2)
			change = &NodeChange{
				NodeID:     "123",
				Kinds:      nil,
				Properties: &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash        = change.IdentityKey()
			dataHash, err = change.Hash()
		)

		require.NoError(t, err)

		// simulate a full cache
		c.data[idHash] = dataHash

		shouldSubmit, err := c.shouldSubmit(change)
		require.NoError(t, err)
		require.False(t, shouldSubmit)
	})

	t.Run("node is cached. properties have a diff. return true", func(t *testing.T) {
		var (
			c         = newCache(2)
			oldChange = &NodeChange{
				NodeID:     "123",
				Kinds:      nil,
				Properties: &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash        = oldChange.IdentityKey()
			dataHash, err = oldChange.Hash()
		)

		require.NoError(t, err)

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
			c         = newCache(2)
			oldChange = &NodeChange{
				NodeID:     "123",
				Kinds:      nil,
				Properties: &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash        = oldChange.IdentityKey()
			dataHash, err = oldChange.Hash()
		)

		require.NoError(t, err)

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
		c := newCache(2)

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
			c         = newCache(2)
			oldChange = &EdgeChange{
				SourceNodeID: "123",
				TargetNodeID: "456",
				Kind:         graph.StringKind("a"),
				Properties:   &graph.Properties{Map: map[string]any{"a": 1}},
			}
			idHash        = oldChange.IdentityKey()
			dataHash, err = oldChange.Hash()
		)

		require.NoError(t, err)

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
