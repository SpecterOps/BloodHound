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

func PrimaryNodeKind(kinds graph.Kinds) graph.Kind {
	var (
		resultKind = unknownKind
		baseKind   = resultKind
	)

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
		} else if ValidKinds[kind] {
			resultKind = kind
		}
	}

	if resultKind.Is(unknownKind) {
		return baseKind
	} else {
		return resultKind
	}
}
