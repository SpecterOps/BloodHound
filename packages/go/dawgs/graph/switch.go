// Copyright 2023 Specter Ops, Inc.
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

package graph

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrAuthoritativeDatabaseSwitching = errors.New("switching authoritative database")
)

type DatabaseSwitch struct {
	activeContexts map[any]func()
	currentDB      Database
	inSwitch       bool
	ctxLock        *sync.Mutex
	currentDBLock  *sync.RWMutex
	writeFlushSize int
	batchWriteSize int
}

func NewDatabaseSwitch(ctx context.Context, initialDB Database) *DatabaseSwitch {
	return &DatabaseSwitch{
		activeContexts: map[any]func(){},
		currentDB:      initialDB,
		inSwitch:       false,
		ctxLock:        &sync.Mutex{},
		currentDBLock:  &sync.RWMutex{},
	}
}

func (s *DatabaseSwitch) SetDefaultGraph(ctx context.Context, graphSchema Graph) error {
	s.currentDBLock.RLock()
	defer s.currentDBLock.RUnlock()

	return s.currentDB.SetDefaultGraph(ctx, graphSchema)
}

func (s *DatabaseSwitch) Switch(db Database) {
	s.inSwitch = true

	defer func() {
		s.inSwitch = false
	}()

	s.cancelInternalContexts()

	s.currentDBLock.Lock()
	defer s.currentDBLock.Unlock()

	s.currentDB = db
}

func (s *DatabaseSwitch) SetWriteFlushSize(interval int) {
	s.writeFlushSize = interval
}

func (s *DatabaseSwitch) SetBatchWriteSize(interval int) {
	s.batchWriteSize = interval
}

func (s *DatabaseSwitch) newInternalContext(ctx context.Context) (context.Context, error) {
	s.ctxLock.Lock()
	defer s.ctxLock.Unlock()

	// Do not issue new contexts if we're in the process of switching authoritative databases
	if s.inSwitch {
		return nil, ErrAuthoritativeDatabaseSwitching
	}

	internalCtx, doneFunc := context.WithCancel(ctx)

	s.activeContexts[internalCtx] = doneFunc
	return internalCtx, nil
}

func (s *DatabaseSwitch) cancelInternalContexts() {
	s.ctxLock.Lock()
	defer s.ctxLock.Unlock()

	for _, doneFunc := range s.activeContexts {
		doneFunc()
	}
}

func (s *DatabaseSwitch) retireInternalContext(ctx context.Context) {
	s.ctxLock.Lock()
	defer s.ctxLock.Unlock()

	if doneFunc, exists := s.activeContexts[ctx]; exists {
		doneFunc()
		delete(s.activeContexts, ctx)
	}
}

func (s *DatabaseSwitch) ReadTransaction(ctx context.Context, txDelegate TransactionDelegate, options ...TransactionOption) error {
	if internalCtx, err := s.newInternalContext(ctx); err != nil {
		return err
	} else {
		defer s.retireInternalContext(internalCtx)

		s.currentDBLock.RLock()
		defer s.currentDBLock.RUnlock()

		return s.currentDB.ReadTransaction(internalCtx, txDelegate, options...)
	}
}

func (s *DatabaseSwitch) WriteTransaction(ctx context.Context, txDelegate TransactionDelegate, options ...TransactionOption) error {
	if internalCtx, err := s.newInternalContext(ctx); err != nil {
		return err
	} else {
		defer s.retireInternalContext(internalCtx)

		s.currentDBLock.RLock()
		defer s.currentDBLock.RUnlock()

		return s.currentDB.WriteTransaction(internalCtx, txDelegate, options...)
	}
}

func (s *DatabaseSwitch) BatchOperation(ctx context.Context, batchDelegate BatchDelegate) error {
	if internalCtx, err := s.newInternalContext(ctx); err != nil {
		return err
	} else {
		defer s.retireInternalContext(internalCtx)

		s.currentDBLock.RLock()
		defer s.currentDBLock.RUnlock()

		return s.currentDB.BatchOperation(internalCtx, batchDelegate)
	}
}

func (s *DatabaseSwitch) AssertSchema(ctx context.Context, dbSchema Schema) error {
	if internalCtx, err := s.newInternalContext(ctx); err != nil {
		return err
	} else {
		defer s.retireInternalContext(internalCtx)

		s.currentDBLock.RLock()
		defer s.currentDBLock.RUnlock()

		return s.currentDB.AssertSchema(ctx, dbSchema)
	}
}

func (s *DatabaseSwitch) Run(ctx context.Context, query string, parameters map[string]any) error {
	if internalCtx, err := s.newInternalContext(ctx); err != nil {
		return err
	} else {
		defer s.retireInternalContext(internalCtx)

		s.currentDBLock.RLock()
		defer s.currentDBLock.RUnlock()

		return s.currentDB.Run(internalCtx, query, parameters)
	}
}

func (s *DatabaseSwitch) Close(ctx context.Context) error {
	if internalCtx, err := s.newInternalContext(ctx); err != nil {
		return err
	} else {
		defer s.retireInternalContext(internalCtx)

		s.currentDBLock.RLock()
		defer s.currentDBLock.RUnlock()

		return s.currentDB.Close(ctx)
	}
}
