// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package racfhound_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/racfhound"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestPathfindingRelationships(t *testing.T) {
	relationshipKinds := racfhound.PathfindingRelationships()

	require.True(t, relationshipKinds.ContainsOneOf(graph.StringKind("RACFMemberOf")))
	require.True(t, relationshipKinds.ContainsOneOf(graph.StringKind("RACFCanWrite")))
	require.False(t, relationshipKinds.ContainsOneOf(graph.StringKind("RACFHasSubgroup")))
	require.False(t, relationshipKinds.ContainsOneOf(graph.StringKind("RACFEvidencedBy")))
}

func TestIsNonPathfindingRelationship(t *testing.T) {
	require.True(t, racfhound.IsNonPathfindingRelationship(graph.StringKind("RACFHasSubgroup")))
	require.True(t, racfhound.IsNonPathfindingRelationship(graph.StringKind("RACFSubgroupOf")))
	require.False(t, racfhound.IsNonPathfindingRelationship(graph.StringKind("RACFMemberOf")))
}
