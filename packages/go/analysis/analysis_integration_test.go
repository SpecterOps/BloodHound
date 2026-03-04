// Copyright 2026 Specter Ops, Inc.
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

//go:build serial_integration

package analysis_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeKindDisplayLabel(t *testing.T) {
	dbInst := integration.SetupDB(t)

	primaryNodeKinds, err := dbInst.GetDisplayNodeGraphKinds(context.Background())
	require.NoError(t, err)

	assert.Equal(t, ad.Entity.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity)), "should return base kind if no other valid kinds are present")
	assert.Equal(t, ad.User.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.User)), "should return valid AD kind when base and kind are present")
	assert.Equal(t, ad.Group.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.Group, ad.LocalGroup)), "should return valid kind other than LocalGroup if one is present")
	assert.Equal(t, ad.LocalGroup.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.LocalGroup)), "should return LocalGroup if no other valid kinds are present")
	assert.Equal(t, azure.Group.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), azure.Entity, azure.Group)), "should return valid Azure kind when base and kind are present")
	assert.Equal(t, analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), unsupportedKind)), "should return Unknown when only an unsupported kind is present")
	assert.Equal(t, ad.Entity.String(), analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties(), ad.Entity, unsupportedKind)), "should return valid kind if one is preseneven if an unsupported kind is also present")
	assert.Equal(t, analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(primaryNodeKinds, graph.PrepareNode(graph.NewProperties())), "should return Unknown if no node has no kinds on it")
}
