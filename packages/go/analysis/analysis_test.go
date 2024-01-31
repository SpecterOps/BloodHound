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

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/slicesext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unsupportedKind = kindStr("Unsupported Kind")

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

func validKinds() graph.Kinds {
	return slicesext.Concat(ad.NodeKinds(), ad.Relationships(), azure.NodeKinds(), azure.Relationships())
}

func validKindStrings() []string {
	var (
		kindStrings = make([]string, 0, len(validKinds()))
	)

	for _, kind := range validKinds() {
		kindStrings = append(kindStrings, kind.String())
	}

	return kindStrings
}

func TestParseKind(t *testing.T) {
	t.Run("all known strings map to their graph.Kind", func(t *testing.T) {
		for _, k := range validKinds() {
			res, err := analysis.ParseKind(k.String())
			require.Nil(t, err)
			assert.Equal(t, k, res, "expect string to map back to original kind")
		}
	})

	t.Run("unknown kind strings cause an error", func(t *testing.T) {
		_, err := analysis.ParseKind(unsupportedKind.String())
		assert.Contains(t, err.Error(), unsupportedKind.String(), "error contains unsupported kind string")
	})
}

func TestParseKinds(t *testing.T) {
	t.Run("all known strings map to their graph.Kind", func(t *testing.T) {
		res, err := analysis.ParseKinds(validKindStrings()...)
		require.Nil(t, err)
		assert.Equal(t, validKinds(), res)
	})

	t.Run("unknown kind strings cause an error", func(t *testing.T) {
		_, err := analysis.ParseKinds(unsupportedKind.String())
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), unsupportedKind, "expect string to map back to original kind")
	})

	t.Run("no arguments provided should return the base kinds", func(t *testing.T) {
		res, err := analysis.ParseKinds()
		require.Nil(t, err)
		assert.Equal(t, graph.Kinds{ad.Entity, azure.Entity}, res)
	})
}

func TestGetNodeKindDisplayLabel(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(ad.Entity.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity)), "should return base kind if no other valid kinds are present")
	assert.Equal(ad.User.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.User)), "should return valid AD kind when base and kind are present")
	assert.Equal(ad.Group.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.Group, ad.LocalGroup)), "should return valid kind other than LocalGroup if one is present")
	assert.Equal(ad.LocalGroup.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, ad.LocalGroup)), "should return LocalGroup if no other valid kinds are present")
	assert.Equal(azure.Group.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), azure.Entity, azure.Group)), "should return valid Azure kind when base and kind are present")
	assert.Equal(analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), unsupportedKind)), "should return Unknown when only an unsupported kind is present")
	assert.Equal(ad.Entity.String(), analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties(), ad.Entity, unsupportedKind)), "should return valid kind if one is preseneven if an unsupported kind is also present")
	assert.Equal(analysis.NodeKindUnknown, analysis.GetNodeKindDisplayLabel(graph.PrepareNode(graph.NewProperties())), "should return Unknown if no node has no kinds on it")
}
