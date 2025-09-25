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

package graphify

import (
	"testing"

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestDeduplicateKinds(t *testing.T) {
	kinds := []graph.Kind{graph.StringKind("Same"), graph.StringKind("Same"), graph.StringKind("Different")}

	deduped := deduplicateKinds(kinds)
	require.Len(t, deduped, 2)
	require.Equal(t, deduped[0].String(), "Same")
	require.Equal(t, deduped[1].String(), "Different")
}

func TestMergeNodeKinds(t *testing.T) {
	kinds := []graph.Kind{
		graph.StringKind("Same"),
		graph.StringKind("Same"),
		graph.StringKind("Different"),
		graph.EmptyKind,
	}

	merged := MergeNodeKinds(graph.StringKind("Base"), kinds...)
	require.Len(t, merged, 3)
	require.Equal(t, merged[0].String(), "Base")
	require.Equal(t, merged[1].String(), "Same")
	require.Equal(t, merged[2].String(), "Different")

	merged = MergeNodeKinds(graph.EmptyKind, kinds...)
	require.Len(t, merged, 2)
	require.Equal(t, merged[0].String(), "Same")
	require.Equal(t, merged[1].String(), "Different")
}
