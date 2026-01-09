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
package database

import (
	"context"
)

// TODO: Where should these live??

// MapDiffActions - Actions required to sync two maps
//
//  1. ItemsToUpdate (Updates): Represents the items present in both the SourceMap and DestinationMap.
//     In set theory terms, this is (SourceMap ∩ DestinationMap). This is the overlapping portion
//     of both circles of a Venn diagram.
//
//  2. ItemsToDelete (Deletes): Represents the items present exclusively in the DestinationMap
//     that are absent in the SourceMap. In set theory terms, this is (DestinationMap - SourceMap).
//     This represents items in the right circle of a Venn diagram that are *not* in the intersection, and thus must be deleted.
//
//  3. ItemsToInsert (Inserts): Represents the items present exclusively in the SourceMap that are absent in the DestinationMap.
//     In set theory terms this is (SourceMap - DestinationMap). This represents that items in the left circle of a Venn
//     diagram that are *not* in the intersection, and thus must be deleted.
type MapDiffActions[V any] struct {
	ItemsToDelete []V
	ItemsToUpdate []V
	ItemsToInsert []V
}

// GenerateMapSynchronizationDiffActions compares two maps (`SourceMap` and `DestinationMap`) using their
// keys (`K`) to compute the required synchronization actions based on set theory.
//
// The function generates three lists of *values* (`V`) representing the operations needed to make
// `DestinationMap` an exact replica of `SourceMap`:
//
//  1. ItemsToUpdate (Updates): Represents the items present in both the SourceMap and DestinationMap.
//     In set theory terms, this is (SourceMap ∩ DestinationMap). This is the overlapping portion
//     of both circles of a Venn diagram.
//
//  2. ItemsToDelete (Deletes): Represents the items present exclusively in the DestinationMap
//     that are absent in the SourceMap. In set theory terms, this is (DestinationMap - SourceMap).
//     This represents items in the right circle of a Venn diagram that are *not* in the intersection, and thus must be deleted.
//
//  3. ItemsToInsert (Inserts): Represents the items present exclusively in the SourceMap that are absent in the DestinationMap.
//     In set theory terms this is (SourceMap - DestinationMap). This represents that items in the left circle of a Venn
//     diagram that are *not* in the intersection, and thus must be deleted.
//
// An optional onMatch function can be provided to perform struct updates.
func GenerateMapSynchronizationDiffActions[K comparable, V any](src, dst map[K]V, onMatch func(*V, *V)) MapDiffActions[V] {

	var (
		actions = MapDiffActions[V]{
			ItemsToDelete: make([]V, 0),
			ItemsToUpdate: make([]V, 0),
			ItemsToInsert: make([]V, 0),
		}
	)

	if src == nil {
		src = make(map[K]V)
	} else if dst == nil {
		dst = make(map[K]V)
	}

	// 1. Identify keys to delete from the dst
	// These will be keys that exist in dst but not in src
	for k, v := range dst {
		if _, exists := src[k]; !exists {
			actions.ItemsToDelete = append(actions.ItemsToDelete, v)
		}
	}

	// 2. Identify keys to upsert (all keys in src)
	for k, v := range src {

		// Retrieve the existing value from dst map, if it exists
		dstVal, existsInDst := dst[k]

		// Pass the key, the src value pointer, and the dst value pointer
		if existsInDst {
			if onMatch != nil {
				onMatch(&v, &dstVal)
			}
			actions.ItemsToUpdate = append(actions.ItemsToUpdate, v)
		} else {
			// If it's a new key, pass nil for the dst value pointer
			actions.ItemsToInsert = append(actions.ItemsToInsert, v)
		}
	}

	return actions
}

// HandleMapDiffAction iterates through a set of MapDiffActions and executes the
// provided callback functions for items marked for deletion, update, or insertion.
//
// The function processes actions in the following order:
// 1. Deletions (using deleteFunc)
// 2. Updates (using updateFunc)
// 3. Insertions (using insertFunc)
//
// It returns the first error encountered during any of the operations, halting
// further processing. If all operations succeed, it returns nil.
func HandleMapDiffAction[V any](ctx context.Context, actions MapDiffActions[V], deleteFunc, updateFunc, insertFunc func(context.Context, V) error) error {
	var err error
	if len(actions.ItemsToDelete) > 0 {
		for _, deletedGraphSchemaKind := range actions.ItemsToDelete {
			if err = deleteFunc(ctx, deletedGraphSchemaKind); err != nil {
				return err
			}
		}
	}

	if len(actions.ItemsToUpdate) > 0 {
		for _, updatedGraphSchemaKind := range actions.ItemsToUpdate {
			if err = updateFunc(ctx, updatedGraphSchemaKind); err != nil {
				return err
			}
		}
	}
	if len(actions.ItemsToInsert) > 0 {
		for _, newGraphSchemaEdgeKind := range actions.ItemsToInsert {
			if err = insertFunc(ctx, newGraphSchemaEdgeKind); err != nil {
				return err
			}
		}
	}
	return nil
}
