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
}

type PrimaryDisplayKinds map[graph.Kind]DisplayKind

func (s PrimaryDisplayKinds) Add(kindName, iconName, iconColor string, iconType DisplayNodeType) {
	s[graph.StringKind(kindName)] = DisplayKind{
		Name: kindName,
		Icon: DisplayNodeIcon{
			Type:  iconType,
			Name:  iconName,
			Color: iconColor,
		},
	}
}

// PrimaryDisplayKind - tests if the provided kinds contain a primary or meta kind.
//
// It accepts a primaryDisplayKinds map[graph.Kind]DisplayKind that contains primary display kinds.
// This allows devs to validate kinds against an OpenGraph extension's kinds.
// It will return the first meta kind or the first primary kind it finds. During processing, if
// a source kind is found it will set the base kind to the source kind. If a primary/meta kind is not
// found, it will return the base kind which will be the "unknown" kind if no known base kinds are
// present.
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
		} else if kind.Is(ad.Entity, azure.Entity) {
			baseKind = kind
		} else if kind.Is(ad.LocalGroup) {
			// Allow ad.LocalGroup to overwrite NodeKindUnknown, but nothing else
			if resultKind == UnknownKind {
				resultKind = kind
			}
		} else if _, ok := primaryDisplayKinds[kind]; ok {
			return kind
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
