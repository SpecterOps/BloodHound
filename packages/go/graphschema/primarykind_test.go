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

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func Test_PrimaryNodeKind(t *testing.T) {

	t.Run("detects meta kinds", func(t *testing.T) {
		primaryKind := PrimaryDisplayKind(nil, graph.Kinds{Meta})
		require.Equal(t, Meta, primaryKind)
	})

	t.Run("ad local group overrides unknown", func(t *testing.T) {
		primaryKind := PrimaryDisplayKind(nil, graph.Kinds{ad.Entity, ad.LocalGroup})
		require.Equal(t, ad.LocalGroup, primaryKind)
	})

	t.Run("detects valid kind", func(t *testing.T) {
		primaryKind := PrimaryDisplayKind(nil, graph.Kinds{ad.Entity, ad.Computer})
		require.Equal(t, ad.Computer, primaryKind)
	})

	t.Run("falls back to base kind if no valid kinds", func(t *testing.T) {
		primaryKind := PrimaryDisplayKind(nil, graph.Kinds{ad.Entity, graph.StringKind("Villain")})
		require.Equal(t, ad.Entity, primaryKind)
	})

	t.Run("falls back to unknown if nothing detected", func(t *testing.T) {
		primaryKind := PrimaryDisplayKind(nil, graph.Kinds{graph.StringKind("Hero")})
		require.Equal(t, UnknownKind, primaryKind)
	})

	t.Run("returns the first valid kind in the kinds array", func(t *testing.T) {
		primaryKind := PrimaryDisplayKind(PrimaryDisplayKinds{
			graph.StringKind("Hero"):    DisplayKind{},
			graph.StringKind("Villain"): DisplayKind{},
		}, graph.Kinds{graph.StringKind("Hero"), graph.StringKind("Villain")})
		require.Equal(t, graph.StringKind("Hero"), primaryKind)
	})

	t.Run("returns open graph display kind over its source kind", func(t *testing.T) {
		var (
			dogparkEntity       = graph.StringKind("dogpark_Entity")
			dogparkLargeDogArea = graph.StringKind("dogpark_LargeDogArea")
		)

		primaryKind := PrimaryDisplayKind(PrimaryDisplayKinds{
			dogparkEntity:       DisplayKind{Name: "dogpark_Entity", IsSourceKind: true},
			dogparkLargeDogArea: DisplayKind{Name: "dogpark_LargeDogArea"},
		}, graph.Kinds{dogparkEntity, dogparkLargeDogArea})
		require.Equal(t, dogparkLargeDogArea, primaryKind)
	})

	t.Run("falls back to open graph source kind when no display kind is present", func(t *testing.T) {
		dogparkEntity := graph.StringKind("dogpark_Entity")

		primaryKind := PrimaryDisplayKind(PrimaryDisplayKinds{
			dogparkEntity: DisplayKind{Name: "dogpark_Entity", IsSourceKind: true},
		}, graph.Kinds{dogparkEntity, graph.StringKind("dogpark_Squirrel")})
		require.Equal(t, dogparkEntity, primaryKind)
	})
}

func Test_PrimaryDisplayKinds_Add(t *testing.T) {
	t.Run("adds a new display kind with its icon", func(t *testing.T) {
		primaryDisplayKinds := make(PrimaryDisplayKinds)

		primaryDisplayKinds.Add("dogpark_LargeDogArea", "dog", "blue", DisplayNodeTypeFontAwesome, false)

		require.Equal(t, DisplayKind{
			Name: "dogpark_LargeDogArea",
			Icon: DisplayNodeIcon{
				Type:  DisplayNodeTypeFontAwesome,
				Name:  "dog",
				Color: "blue",
			},
			IsSourceKind: false,
		}, primaryDisplayKinds[graph.StringKind("dogpark_LargeDogArea")])
	})

	t.Run("adds a new source kind flagged with IsSourceKind", func(t *testing.T) {
		primaryDisplayKinds := make(PrimaryDisplayKinds)

		primaryDisplayKinds.Add("dogpark_Entity", "", "", "", true)

		require.Equal(t, DisplayKind{
			Name:         "dogpark_Entity",
			IsSourceKind: true,
		}, primaryDisplayKinds[graph.StringKind("dogpark_Entity")])
	})

	t.Run("adding an overlapping source kind after a custom kind preserves existing icon and updates isSourceKind", func(t *testing.T) {
		primaryDisplayKinds := make(PrimaryDisplayKinds)

		// A custom node kind is added first with its icon attributes.
		primaryDisplayKinds.Add("dogpark_Entity", "dog", "blue", DisplayNodeTypeFontAwesome, false)
		// A source kind with the same name is then added with empty icon attributes.
		primaryDisplayKinds.Add("dogpark_Entity", "", "", "", true)

		require.Equal(t, DisplayKind{
			Name: "dogpark_Entity",
			Icon: DisplayNodeIcon{
				Type:  DisplayNodeTypeFontAwesome,
				Name:  "dog",
				Color: "blue",
			},
			IsSourceKind: true,
		}, primaryDisplayKinds[graph.StringKind("dogpark_Entity")])
	})

	t.Run("adding an overlapping custom kind after a source kind preserves existing icon and maintains the original isSourceKind value", func(t *testing.T) {
		primaryDisplayKinds := make(PrimaryDisplayKinds)

		// A source kind is added first with an icon.
		primaryDisplayKinds.Add("dogpark_Entity", "dog", "blue", DisplayNodeTypeFontAwesome, true)
		// A custom node kind with the same name is then added with empty icon attributes.
		primaryDisplayKinds.Add("dogpark_Entity", "", "", "", false)

		require.Equal(t, DisplayKind{
			Name: "dogpark_Entity",
			Icon: DisplayNodeIcon{
				Type:  DisplayNodeTypeFontAwesome,
				Name:  "dog",
				Color: "blue",
			},
			IsSourceKind: true,
		}, primaryDisplayKinds[graph.StringKind("dogpark_Entity")])
	})
}
