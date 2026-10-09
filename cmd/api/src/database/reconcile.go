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

package database

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// reconcileConfig holds the callbacks needed to reconcile a collection of existing items against a desired input set.
// TInput is the desired-state type, TExisting is the current-state type, and K is a comparable key used to
// match the two (e.g. string for name, int32 for ID). Callers construct a config via a factory that closes
// over any scope-specific values the callbacks require.
type reconcileConfig[TInput any, TExisting any, K comparable] struct {
	getInputKey    func(input TInput) K
	getExistingKey func(existing TExisting) K
	create         func(ctx context.Context, input TInput) (TExisting, error)
	update         func(ctx context.Context, existing TExisting, input TInput) (TExisting, error)
	delete         func(ctx context.Context, existing TExisting) error
}

// reconcile brings a collection of existing items in line with a desired input set using set-differencing.
// Items present in inputs but not in existingItems are created; items present in both are updated;
// items present in existingItems but not in inputs are deleted. Matching is done by key K.
// The caller is responsible for supplying the current existingItems slice upfront.
func reconcile[TInput any, TExisting any, K comparable](
	ctx context.Context,
	inputs []TInput,
	existingRows []TExisting,
	config reconcileConfig[TInput, TExisting, K],
) (model.ReconcileResult[TExisting], error) {
	var (
		existingByKey = make(map[K]TExisting, len(existingRows))
		inputKeys     = make(map[K]bool, len(inputs))
		result        model.ReconcileResult[TExisting]
	)

	for _, existing := range existingRows {
		existingByKey[config.getExistingKey(existing)] = existing
	}

	for _, input := range inputs {
		inputKeys[config.getInputKey(input)] = true
	}

	// Delete pass — remove items not present in the input
	for _, existing := range existingRows {
		if inputKeys[config.getExistingKey(existing)] {
			continue
		}
		if err := config.delete(ctx, existing); err != nil {
			return result, err
		}
		result.Deleted = append(result.Deleted, existing)
	}

	// Create/Update pass — upsert each desired input
	for _, input := range inputs {
		if existing, found := existingByKey[config.getInputKey(input)]; found {
			if updated, err := config.update(ctx, existing, input); err != nil {
				return result, err
			} else {
				result.Updated = append(result.Updated, updated)
			}
		} else if created, err := config.create(ctx, input); err != nil {
			return result, err
		} else {
			result.Created = append(result.Created, created)
		}
	}

	return result, nil
}

// mergeReconcileResults combines two reconciliation outcomes without modifying either input.
func mergeReconcileResults[T any](first, second model.ReconcileResult[T]) model.ReconcileResult[T] {
	return model.ReconcileResult[T]{
		Created: append(append([]T(nil), first.Created...), second.Created...),
		Updated: append(append([]T(nil), first.Updated...), second.Updated...),
		Deleted: append(append([]T(nil), first.Deleted...), second.Deleted...),
	}
}
