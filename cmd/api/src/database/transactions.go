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

package database

import (
	"context"

	"gorm.io/gorm"
)

type Transactable[T any] interface {
	Transaction(ctx context.Context, fn func(tx T) error) error
}

type TransactionExecutor[D any] interface {
	RunTransaction(ctx context.Context, db D, fn func(txDb D) error) error
}

type handle[D any] struct {
	db D
}

type TransactableDB[T Transactable[T], D any] struct {
	*handle[D]
	executor TransactionExecutor[D]
	withTxFn func(h *handle[D]) T
}

func (t *TransactableDB[T, D]) Transaction(ctx context.Context, fn func(tx T) error) error {
	return t.executor.RunTransaction(ctx, t.db, func(txDb D) error {
		return fn(t.withTxFn(&handle[D]{db: txDb}))
	})
}

type gormExecutor struct{}

func (gormExecutor) RunTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
