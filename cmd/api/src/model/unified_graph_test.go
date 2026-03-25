// Copyright 2026 Specter Ops, Inc.
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
package model

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestUnifiedGraph_AddPathSet(t *testing.T) {
	testPathSetWithDupes := graph.NewPathSet(graph.Path{
		Nodes: []*graph.Node{
			{
				ID:    0,
				Kinds: graph.Kinds{ad.Entity, ad.Computer},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "12345",
					common.Name:  "computer one",
				}),
			},
			{
				ID:    1,
				Kinds: graph.Kinds{ad.Entity, ad.User},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "54321",
					common.Name:  "user one",
				}),
			},
			{
				ID:    3,
				Kinds: graph.Kinds{ad.Entity, ad.Computer},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "22322",
					common.Name:  "computer two",
				}),
			},
			{
				ID:    4,
				Kinds: graph.Kinds{ad.Entity, ad.User},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "41414",
					common.Name:  "user two",
				}),
			},
			{
				ID:    5,
				Kinds: graph.Kinds{ad.Entity, ad.Computer},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "64441",
					common.Name:  "computer three",
				}),
			},
		},
		Edges: []*graph.Relationship{
			{
				ID:         55,
				StartID:    0,
				EndID:      1,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         55,
				StartID:    0,
				EndID:      1,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         56,
				StartID:    2,
				EndID:      3,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         57,
				StartID:    3,
				EndID:      5,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         58,
				StartID:    2,
				EndID:      5,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
		},
	})
	testPathSetWithoutDupes := graph.NewPathSet(graph.Path{
		Nodes: []*graph.Node{
			{
				ID:    0,
				Kinds: graph.Kinds{ad.Entity, ad.Computer},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "12345",
					common.Name:  "computer one",
				}),
			},
			{
				ID:    1,
				Kinds: graph.Kinds{ad.Entity, ad.User},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "54321",
					common.Name:  "user one",
				}),
			},
			{
				ID:    3,
				Kinds: graph.Kinds{ad.Entity, ad.Computer},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "22322",
					common.Name:  "computer two",
				}),
			},
			{
				ID:    4,
				Kinds: graph.Kinds{ad.Entity, ad.User},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "41414",
					common.Name:  "user two",
				}),
			},
			{
				ID:    5,
				Kinds: graph.Kinds{ad.Entity, ad.Computer},
				Properties: graph.AsProperties(graph.PropertyMap{
					ad.DomainSID: "64441",
					common.Name:  "computer three",
				}),
			},
		},
		Edges: []*graph.Relationship{
			{
				ID:         54,
				StartID:    0,
				EndID:      1,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         55,
				StartID:    2,
				EndID:      5,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         56,
				StartID:    2,
				EndID:      3,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         57,
				StartID:    3,
				EndID:      5,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
			{
				ID:         58,
				StartID:    0,
				EndID:      5,
				Kind:       ad.GenericWrite,
				Properties: graph.NewProperties(),
			},
		},
	})

	t.Run("Should NOT return the exact amount of edges from the PathSet", func(t *testing.T) {
		testGraph := NewUnifiedGraph()
		testGraph.AddPathSet(nil, testPathSetWithDupes, true)

		require.Equal(t, 4, len(testGraph.Edges))
	})

	t.Run("Should return the exact amount of edges from the PathSet", func(t *testing.T) {
		testGraph := NewUnifiedGraph()
		testGraph.AddPathSet(nil, testPathSetWithoutDupes, true)

		require.Equal(t, len(testGraph.Edges), len(testGraph.Edges))
	})
}
