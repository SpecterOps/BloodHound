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
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

var (
	// Originates from BHE but copied here
	meta         = graph.StringKind("Meta")
	metaDetail   = graph.StringKind("MetaDetail")
	metaIncludes = graph.StringKind("MetaIncludes")
	metaKinds    = []graph.Kind{meta, metaDetail, metaIncludes}

	UnknownKind = graph.StringKind("Unknown")

	// Used for quick O(1) kind lookups
	ValidKinds = buildValidKinds()
)

func buildValidKinds() ValidPrimaryKinds {
	var (
		validKinds = make(ValidPrimaryKinds)
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

type ValidPrimaryKinds map[graph.Kind]DisplayKind

func (s ValidPrimaryKinds) Add(kindName, iconName, iconColor string, iconType DisplayNodeType) {
	s[graph.StringKind(kindName)] = DisplayKind{
		Name: kindName,
		Icon: DisplayNodeIcon{
			Type:  iconType,
			Name:  iconName,
			Color: iconColor,
		},
	}
}

// PrimaryNodeKind - tests if the provided kinds contain a primary or meta kind.
//
// It accepts a validPrimaryKinds map[graph.Kind]bool that contains valid primary kinds.
// This allows devs to validate kinds against an OpenGraph extension's kinds.
// It will return the first meta kind or the first primary kind it finds. During processing, if
// a source kind is found it will set the base kind to the source kind. If a primary/meta kind is not
// found, it will return the base kind which will be the "unknown" kind if no known base kinds are
// present.
func PrimaryNodeKind(validPrimaryKinds ValidPrimaryKinds, kinds graph.Kinds) graph.Kind {
	var (
		resultKind = UnknownKind
		baseKind   = resultKind
	)

	if validPrimaryKinds == nil {
		slog.Warn("PrimaryNodeKind: validPrimaryKinds is nil")
		validPrimaryKinds = ValidKinds
	}

	for _, kind := range kinds {
		// If this is a BHE meta kind, return early
		if kind.Is(metaKinds...) {
			return meta
		} else if kind.Is(ad.Entity, azure.Entity) {
			baseKind = kind
		} else if kind.Is(ad.LocalGroup) {
			// Allow ad.LocalGroup to overwrite NodeKindUnknown, but nothing else
			if resultKind == UnknownKind {
				resultKind = kind
			}
		} else if _, ok := validPrimaryKinds[kind]; ok {
			return kind
		}
	}

	if resultKind.Is(UnknownKind) {
		return baseKind
	} else {
		return resultKind
	}
}

func GetNodeKindDisplayLabel(validPrimaryKinds ValidPrimaryKinds, node *graph.Node) string {
	return GetNodeKind(validPrimaryKinds, node).String()
}

// GetNodeKind - returns the primary kind of the node.
func GetNodeKind(validPrimaryKinds ValidPrimaryKinds, node *graph.Node) graph.Kind {
	return PrimaryNodeKind(validPrimaryKinds, node.Kinds)
}
