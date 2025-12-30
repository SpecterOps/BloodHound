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
	"reflect"

	"gorm.io/gorm"
)

type Transactable[T any] interface {
	Transaction(ctx context.Context, fn func(tx T) error) error
}

type TransactionExecutor[D any] interface {
	RunTransaction(ctx context.Context, db D, fn func(txDb D) error) error
}

type TxFactory[T any, D any] func(db D) T

type handle[D any] struct {
	db D
}

type TransactableDB[T Transactable[T], D any] struct {
	*handle[D]
	executor  TransactionExecutor[D]
	txFactory TxFactory[T, D]
}

func (t *TransactableDB[T, D]) Transaction(ctx context.Context, fn func(tx T) error) error {
	return t.executor.RunTransaction(ctx, t.db, func(txDb D) error {
		return fn(t.txFactory(txDb))
	})
}

func (t *TransactableDB[T, D]) DB() D {
	return t.db
}

type TransactableOption[T Transactable[T], D any] func(*TransactableDB[T, D])

func WithExecutor[T Transactable[T], D any](executor TransactionExecutor[D]) TransactableOption[T, D] {
	return func(t *TransactableDB[T, D]) {
		t.executor = executor
	}
}

func WithTxFactory[T Transactable[T], D any](factory TxFactory[T, D]) TransactableOption[T, D] {
	return func(t *TransactableDB[T, D]) {
		t.txFactory = factory
	}
}

func (t *TransactableDB[T, D]) ConfigureTransactable(db D, opts ...TransactableOption[T, D]) {
	t.handle = &handle[D]{db: db}
	for _, opt := range opts {
		opt(t)
	}
}

func GormExecutor[T Transactable[T]]() TransactableOption[T, *gorm.DB] {
	return WithExecutor[T, *gorm.DB](gormExecutor{})
}

type gormExecutor struct{}

func (gormExecutor) RunTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}

type Wirable[D any] interface {
	Wire(db D)
}

func AutowireEmbedded[D any](parent any, db D) {
	parentVal := reflect.ValueOf(parent)
	if parentVal.Kind() == reflect.Ptr {
		parentVal = parentVal.Elem()
	}
	if parentVal.Kind() != reflect.Struct {
		return
	}

	wirableType := reflect.TypeOf((*Wirable[D])(nil)).Elem()

	for i := 0; i < parentVal.NumField(); i++ {
		field := parentVal.Field(i)
		fieldType := parentVal.Type().Field(i)

		if !fieldType.Anonymous || !fieldType.IsExported() {
			continue
		}

		var fieldPtr reflect.Value
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			fieldPtr = field
		} else if field.CanAddr() {
			fieldPtr = field.Addr()
		} else {
			continue
		}

		if fieldPtr.Type().Implements(wirableType) {
			wireMethod := fieldPtr.MethodByName("Wire")
			if wireMethod.IsValid() {
				wireMethod.Call([]reflect.Value{reflect.ValueOf(db)})
			}
		}
	}
}

func WithAutowire[T Transactable[T], D any](parent any) TransactableOption[T, D] {
	return func(t *TransactableDB[T, D]) {
		AutowireEmbedded(parent, t.db)
	}
}

func WireDB[D any](db D, fields ...*D) {
	for _, f := range fields {
		*f = db
	}
}
