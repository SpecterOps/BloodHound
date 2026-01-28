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

package model

import (
	"time"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

// UnifiedGraphWPropertyKeys
type UnifiedGraphWPropertyKeys struct {
	NodeKeys []string               `json:"node_keys,omitempty"`
	EdgeKeys []string               `json:"edge_keys,omitempty"`
	Edges    []UnifiedEdge          `json:"edges"`
	Nodes    map[string]UnifiedNode `json:"nodes"`
	Literals graph.Literals         `json:"literals"`
}

// UnifiedGraph represents a single, generic and minimalistic graph
type UnifiedGraph struct {
	Nodes    map[string]UnifiedNode `json:"nodes"`
	Edges    []UnifiedEdge          `json:"edges"`
	Literals graph.Literals         `json:"literals"`
}

// NewUnifiedGraph returns a new UnifiedGraph struct with the Nodes field initialized to an empty map
func NewUnifiedGraph() UnifiedGraph {
	return UnifiedGraph{
		Nodes:    map[string]UnifiedNode{},
		Edges:    []UnifiedEdge{},
		Literals: graph.Literals{},
	}
}

// UnifiedNode represents a single node in a graph containing a minimal set of attributes for graph rendering
type UnifiedNode struct {
	Label         string         `json:"label"`
	Kind          string         `json:"kind"`
	Kinds         []string       `json:"kinds"`
	ObjectId      string         `json:"objectId"`
	IsTierZero    bool           `json:"isTierZero"`
	IsOwnedObject bool           `json:"isOwnedObject"`
	LastSeen      time.Time      `json:"lastSeen"`
	Properties    map[string]any `json:"properties,omitempty"`
	Hidden        bool           `json:"hidden,omitempty"`
}

// UnifiedEdge represents a single path segment in a graph containing a minimal set of attributes for graph rendering
type UnifiedEdge struct {
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Label      string         `json:"label"`
	Kind       string         `json:"kind"`
	LastSeen   time.Time      `json:"lastSeen"`
	Properties map[string]any `json:"properties,omitempty"`
}

func FromDAWGSNode(node *graph.Node, includeProperties bool) UnifiedNode {
	var (
		props       = node.Properties
		objectId    = getTypedPropertyOrDefault(props, common.ObjectID.String(), "")
		label       = getTypedPropertyOrDefault(props, common.Name.String(), objectId)
		lastSeen    = getTypedPropertyOrDefault(props, common.LastSeen.String(), time.Now())
		primaryKind = getTypedPropertyOrDefault(props, common.PrimaryKind.String(), "")
	)

	// only generic-ingested nodes have the PrimaryKind property set to control what icon the UI displays.
	kind := primaryKind
	if kind == "" {
		kind = analysis.GetNodeKind(node).String()
	}

	var properties map[string]any
	if includeProperties {
		properties = props.Map
	}

	return UnifiedNode{
		Label:         label,
		Kind:          kind,
		Kinds:         node.Kinds.Strings(),
		ObjectId:      objectId,
		IsTierZero:    tiering.IsTierZero(node),
		IsOwnedObject: tiering.IsOwned(node),
		LastSeen:      lastSeen,
		Properties:    properties,
	}
}

// This is being used with slices.Map so it is necessary to return a closure
func FromDAWGSRelationship(includeProperties bool) func(*graph.Relationship) UnifiedEdge {
	return func(rel *graph.Relationship) UnifiedEdge {
		var properties map[string]any

		if includeProperties {
			properties = rel.Properties.Map
		}

		return UnifiedEdge{
			Source:     rel.StartID.String(),
			Target:     rel.EndID.String(),
			Kind:       rel.Kind.String(),
			Label:      rel.Kind.String(),
			LastSeen:   getTypedPropertyOrDefault(rel.Properties, common.LastSeen.String(), time.Now()),
			Properties: properties,
		}
	}
}

func (s *UnifiedGraph) AddRelationship(rel *graph.Relationship, includeProperties bool) {
	formattedRelationship := FromDAWGSRelationship(includeProperties)(rel)
	s.Edges = append(s.Edges, formattedRelationship)
}

func (s *UnifiedGraph) AddNode(node *graph.Node, includeProperties bool) {
	formattedNode := FromDAWGSNode(node, includeProperties)
	s.Nodes[node.ID.String()] = formattedNode
}

func (s *UnifiedGraph) AddPathSet(paths graph.PathSet, includeProperties bool) {
	for _, path := range paths.Paths() {
		for _, node := range path.Nodes {
			s.AddNode(node, includeProperties)
		}

		for _, edge := range path.Edges {
			s.AddRelationship(edge, includeProperties)
		}
	}
}

type propType interface {
	string | time.Time
}

func getTypedPropertyOrDefault[T propType](props *graph.Properties, propName string, defaultValue T) T {
	var (
		prop  = props.GetOrDefault(propName, defaultValue)
		value any
	)
	switch any(defaultValue).(type) {
	case string:
		value, _ = prop.String()
	case time.Time:
		value, _ = prop.Time()
	}
	return value.(T)
}
