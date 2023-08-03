// Copyright 2023 Specter Ops, Inc.
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

package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
)

type kindStr string

func (s kindStr) String() string {
	return string(s)
}

func (s kindStr) Is(others ...graph.Kind) bool {
	for _, other := range others {
		if s.String() == other.String() {
			return true
		}
	}

	return false
}

func TestGetNodeKindDisplayLabel(t *testing.T) {
	const unsupportedKind = kindStr("Unsupported Kind")

	require.Equal(t, ad.User.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.User)))
	require.Equal(t, ad.Group.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.Group, ad.LocalGroup)))
	require.Equal(t, ad.LocalGroup.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.LocalGroup)))
	require.Equal(t, azure.Group.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), azure.Entity, azure.Group)))
	require.Equal(t, unsupportedKind.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), unsupportedKind)))
	require.Equal(t, "Unknown", analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties())))
}
