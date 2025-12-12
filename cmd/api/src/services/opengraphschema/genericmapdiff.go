// Copyright 2025 Specter Ops, Inc.
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
package opengraphschema

// MapDiffActions - Actions required to sync two maps
//
//  1. ItemsToUpsert (Upserts): Represents the full SourceMap. In set theory terms, this is
//     (SourceMap ∩ DestinationMap) ∪ (SourceMap - DestinationMap). This is the left circle and
//     middle intersection of a Venn diagram, containing all elements that should be present in the final state.
//
//  2. ItemsToDelete (Deletes): Represents the items present exclusively in the DestinationMap
//     that are absent in the `SourceMap`. In set theory terms, this is the asymmetric set difference
//     (DestinationMap - SourceMap). This represents items in the right circle of a Venn diagram
//     that are *not* in the intersection, and thus must be deleted.
type MapDiffActions[V any] struct {
	ItemsToDelete []V
	ItemsToUpsert []V
}

// GenerateMapSynchronizationDiffActions compares two maps (`SourceMap` and `DestinationMap`) using their
// keys (`K`) to compute the required synchronization actions based on set theory.
//
// The function generates two lists of *values* (`V`) representing the operations needed to make
// `DestinationMap` an exact replica of `SourceMap`:
//
//  1. ItemsToUpsert (Upserts): Represents the full SourceMap. In set theory terms, this is
//     (SourceMap ∩ DestinationMap) ∪ (SourceMap - DestinationMap). This is the left circle and
//     middle intersection of a Venn diagram, containing all elements that should be present in the final state.
//
//  2. ItemsToDelete (Deletes): Represents the items present exclusively in the DestinationMap
//     that are absent in the `SourceMap`. In set theory terms, this is the asymmetric set difference
//     (DestinationMap - SourceMap). This represents items in the right circle of a Venn diagram
//     that are *not* in the intersection, and thus must be deleted.
//
// The OnUpsert func can be used
func GenerateMapSynchronizationDiffActions[K comparable, V any](src, dst map[K]V, onUpsert func(*V, *V)) MapDiffActions[V] {
	actions := MapDiffActions[V]{
		ItemsToDelete: make([]V, 0),
		ItemsToUpsert: make([]V, 0),
	}

	srcKeys := make(map[K]bool)
	for k := range src {
		srcKeys[k] = true
	}

	// 1. Identify keys to delete from the dst
	for k := range dst {
		if !srcKeys[k] {
			actions.ItemsToDelete = append(actions.ItemsToDelete, dst[k])
		}
	}

	// 2. Identify keys to upsert (all keys in src)
	for k, v := range src {

		if onUpsert != nil {

			// Retrieve the existing value from dst map, if it exists
			dstVal, existsInDst := dst[k]

			// Pass the key, the src value pointer, and the dst value pointer
			if existsInDst {
				onUpsert(&v, &dstVal)
			} else {
				// If it's a new key, pass nil for the dst value pointer
				onUpsert(&v, nil)
			}
		}

		actions.ItemsToUpsert = append(actions.ItemsToUpsert, v)
	}

	return actions
}
