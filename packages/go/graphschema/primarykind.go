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
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

var (
	metaKinds = []graph.Kind{Meta, MetaDetail, MetaIncludes}

	UnknownKind = graph.StringKind("Unknown")

	// Used for quick O(1) kind lookups
	ValidKinds = buildValidKinds()
)

func buildValidKinds() PrimaryDisplayKinds {
	var (
		validKinds = make(PrimaryDisplayKinds)
		kindSlices = []graph.Kinds{
			ad.NodeKinds(),
			ad.Relationships(),
			azure.NodeKinds(),
			azure.Relationships(),
			common.NodeKinds(),
			common.Relationships(),
		}
	)

	for _, kindSlice := range kindSlices {
		for _, kind := range kindSlice {
			validKinds[kind] = DisplayKind{}
		}
	}

	// ad.Entity ("Base") and azure.Entity ("AZBase") are the built-in source kinds. They are seeded in the
	// source_kinds table and represent the origin/category of a node rather than a primary display kind.
	for _, sourceKind := range []graph.Kind{ad.Entity, azure.Entity} {
		validKinds[sourceKind] = DisplayKind{IsSourceKind: true}
	}

	return validKinds
}

type DisplayNodeType string

const (
	DisplayNodeTypeFontAwesome DisplayNodeType = "font-awesome"
)

type DisplayNodeIcon struct {
	Type  DisplayNodeType `json:"type"`
	Name  string          `json:"name"`
	Color string          `json:"color"`
}

type DisplayKind struct {
	Name string
	Icon DisplayNodeIcon
	// IsSourceKind marks a kind as a source/base kind (e.g. ad.Entity, azure.Entity, or an OpenGraph source
	// kind such as GithubBase). Source kinds represent the origin of a node and are only used as a fallback
	// display kind when no other primary display kind is present.
	IsSourceKind bool
}

type PrimaryDisplayKinds map[graph.Kind]DisplayKind

// Add - registers a display kind under the given kind name, keyed by its graph.StringKind form. It builds a
// DisplayKind from the provided name, icon attributes (name, color, and type), and the isSourceKind flag, which
// marks the kind as a source/base kind used only as a fallback display kind.
func (s PrimaryDisplayKinds) Add(kindName, iconName, iconColor string, iconType DisplayNodeType, isSourceKind bool) {
	var graphKind = graph.StringKind(kindName)

	if displayKind, exists := s[graphKind]; exists {
		iconName = displayKind.Icon.Name
		iconColor = displayKind.Icon.Color
		iconType = displayKind.Icon.Type

		if displayKind.IsSourceKind {
			isSourceKind = displayKind.IsSourceKind
		}
	}

	s[graphKind] = DisplayKind{
		Name: kindName,
		Icon: DisplayNodeIcon{
			Type:  iconType,
			Name:  iconName,
			Color: iconColor,
		},
		IsSourceKind: isSourceKind,
	}
}

// PrimaryDisplayKind - tests if the provided kinds contain a primary or meta kind.
//
// It accepts a primaryDisplayKinds map[graph.Kind]DisplayKind that contains primary display kinds.
// This allows devs to validate kinds against an OpenGraph extension's kinds.
// It will return the first meta kind or the first primary kind it finds. During processing, if
// a source kind is found (any kind flagged with IsSourceKind, such as ad.Entity, azure.Entity, or an
// OpenGraph source kind like GithubBase) it will set the base kind to that source kind. If a primary/meta
// kind is not found, it will return the base kind which will be the "unknown" kind if no known base kinds
// are present.
func PrimaryDisplayKind(primaryDisplayKinds PrimaryDisplayKinds, kinds graph.Kinds) graph.Kind {
	var (
		resultKind = UnknownKind
		baseKind   = resultKind
	)

	if primaryDisplayKinds == nil {
		primaryDisplayKinds = ValidKinds
	}

	for _, kind := range kinds {
		// If this is a BHE meta kind, return early
		if kind.Is(metaKinds...) {
			return Meta
		} else if kind.Is(ad.LocalGroup) {
			// Allow ad.LocalGroup to overwrite NodeKindUnknown, but nothing else. This is checked before the
			// generic display-kind lookup so that a higher-priority primary kind appearing later in the array
			// still wins over ad.LocalGroup.
			if resultKind == UnknownKind {
				resultKind = kind
			}
		} else if displayKind, ok := primaryDisplayKinds[kind]; ok {
			if displayKind.IsSourceKind {
				// Source/base kinds are only used as a fallback when no other primary kind is present
				baseKind = kind
			} else {
				return kind
			}
		}
	}

	if resultKind.Is(UnknownKind) {
		return baseKind
	} else {
		return resultKind
	}
}

func GetNodeKindDisplayLabel(primaryDisplayKinds PrimaryDisplayKinds, node *graph.Node) string {
	return GetNodeKind(primaryDisplayKinds, node).String()
}

// GetNodeKind - returns the primary kind of the node.
func GetNodeKind(primaryDisplayKinds PrimaryDisplayKinds, node *graph.Node) graph.Kind {
	return PrimaryDisplayKind(primaryDisplayKinds, node.Kinds)
}
