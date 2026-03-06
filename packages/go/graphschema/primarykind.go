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
	// Originates from BHE but copied here
	meta         = graph.StringKind("Meta")
	metaDetail   = graph.StringKind("MetaDetail")
	metaIncludes = graph.StringKind("MetaIncludes")
	metaKinds    = []graph.Kind{meta, metaDetail, metaIncludes}

	unknownKind = graph.StringKind("Unknown")

	// Used for quick O(1) kind lookups
	ValidKinds = buildValidKinds()
)

func buildValidKinds() map[graph.Kind]bool {
	var (
		validKinds = make(map[graph.Kind]bool)
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
			validKinds[kind] = true
		}
	}

	return validKinds
}

type ValidPrimaryKinds map[graph.Kind]bool

// PrimaryNodeKind - tests if the provided kinds contain a primary or meta kind.
//
// It accepts a validPrimaryKinds map[graph.Kind]bool that contains valid primary kinds.
// This allows devs to validate kinds against an OpenGraph extension's kinds.
// It will return the first meta kind or the last primary kind it finds. During processing, if
// a source kind is found it will set the base kind to the source kind. If a primary/meta kind is not
// found, it will return the base kind which will be the "unknown" kind if no known base kinds are
// present.
func PrimaryNodeKind(validPrimaryKinds ValidPrimaryKinds, kinds graph.Kinds) graph.Kind {
	var (
		resultKind = unknownKind
		baseKind   = resultKind
	)

	if validPrimaryKinds == nil {
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
			if resultKind == unknownKind {
				resultKind = kind
			}
		} else if validPrimaryKinds[kind] {
			resultKind = kind
		}
	}

	if resultKind.Is(unknownKind) {
		return baseKind
	} else {
		return resultKind
	}
}
