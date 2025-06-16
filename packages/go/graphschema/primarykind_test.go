// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/stretchr/testify/require"
)

func Test_PrimaryNodeKind(t *testing.T) {
	t.Parallel()

	t.Run("detects meta kinds", func(t *testing.T) {
		primaryKind := PrimaryNodeKind(graph.Kinds{meta})
		require.Equal(t, meta, primaryKind)
	})

	t.Run("ad local group overrides unknown", func(t *testing.T) {
		primaryKind := PrimaryNodeKind(graph.Kinds{ad.Entity, ad.LocalGroup})
		require.Equal(t, ad.LocalGroup, primaryKind)
	})

	t.Run("detects valid kind", func(t *testing.T) {
		primaryKind := PrimaryNodeKind(graph.Kinds{ad.Entity, ad.Computer})
		require.Equal(t, ad.Computer, primaryKind)
	})

	t.Run("falls back to base kind if no valid kinds", func(t *testing.T) {
		primaryKind := PrimaryNodeKind(graph.Kinds{ad.Entity, graph.StringKind("Villain")})
		require.Equal(t, ad.Entity, primaryKind)
	})

	t.Run("falls back to unknown if nothing detected", func(t *testing.T) {
		primaryKind := PrimaryNodeKind(graph.Kinds{graph.StringKind("Hero")})
		require.Equal(t, unknownKind, primaryKind)
	})

}
