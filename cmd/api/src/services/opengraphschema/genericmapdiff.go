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

// MapSyncActions -
type MapSyncActions[V any] struct {
	ValuesToDelete []V
	ValuesToUpsert []V
}

// DiffMapsToSyncActions compares a DestinationMap with a SourceMap and generates
// the list of actions required to make the DestinationMap match the SourceMap exactly.
// It accepts a consolidation hook to modify values before they are added to the upsert list.
func DiffMapsToSyncActions[K comparable, V any](dst, src map[K]V, onUpsert func(*V, *V)) MapSyncActions[V] {
	actions := MapSyncActions[V]{
		ValuesToDelete: make([]V, 0),
		ValuesToUpsert: make([]V, 0),
	}

	srcKeys := make(map[K]bool)
	for k := range src {
		srcKeys[k] = true
	}

	// 1. Identify keys to delete from the destination
	for k := range dst {
		if !srcKeys[k] {
			actions.ValuesToDelete = append(actions.ValuesToDelete, dst[k])
		}
	}

	// 2. Identify keys to upsert (all keys in src)
	for k, v := range src {

		if onUpsert != nil {

			// Retrieve the existing value from dst map, if it exists
			dstVal, existsInDst := dst[k]

			// Pass the key, the source value pointer, and the destination value pointer
			if existsInDst {
				onUpsert(&v, &dstVal)
			} else {
				// If it's a new key, pass nil for the destination value pointer
				onUpsert(&v, nil)
			}
		}

		actions.ValuesToUpsert = append(actions.ValuesToUpsert, v)
	}

	return actions
}
