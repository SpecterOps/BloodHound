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

//go:generate go run go.uber.org/mock/mockgen -source=result.go -copyright_file=../../../../LICENSE.header -destination=./mocks/result.go -package=graph_mocks .

package graph

import (
	"context"
	"errors"

	"github.com/specterops/bloodhound/dawgs/util/channels"
)

type KindsResult struct {
	ID    ID
	Kinds Kinds
}

type RelationshipTripleResult struct {
	ID      ID
	StartID ID
	EndID   ID
}

type RelationshipKindsResult struct {
	RelationshipTripleResult
	Kind Kind
}

type DirectionalResult struct {
	Direction    Direction
	Relationship *Relationship
	Node         *Node
}

func NewDirectionalResult(direction Direction, relationship *Relationship, node *Node) DirectionalResult {
	return DirectionalResult{
		Direction:    direction,
		Relationship: relationship,
		Node:         node,
	}
}

// Cursor is an interface that represents an active database operation. Cursors must be closed to prevent resource
// leaks.
type Cursor[T any] interface {
	// Error returns the error reference captured by this Result. When DAWGS database calls fail this value must be
	// populated by the underlying driver.
	Error() error

	// Close releases any active resources bound to this cursor
	Close()

	// Chan returns the type channel backed by this database cursor
	Chan() chan T
}

type ResultMarshaller[T any] func(scanner Scanner) (T, error)

type ResultIterator[T any] struct {
	ctx        context.Context
	result     Result
	cancelFunc func()
	valueC     chan T
	marshaller ResultMarshaller[T]
	error      error
}

func NewResultIterator[T any](ctx context.Context, result Result, marshaller ResultMarshaller[T]) Cursor[T] {
	var (
		cursorCtx, cancelFunc = context.WithCancel(ctx)
		resultIterator        = &ResultIterator[T]{
			ctx:        cursorCtx,
			cancelFunc: cancelFunc,
			result:     result,
			valueC:     make(chan T),
			marshaller: marshaller,
		}
	)

	resultIterator.start()
	return resultIterator
}

func (s *ResultIterator[T]) start() {
	go func() {
		defer close(s.valueC)

		for s.result.Next() {
			if nextValue, err := s.marshaller(s.result); err != nil {
				s.error = err
				break
			} else if !channels.Submit(s.ctx, s.valueC, nextValue) {
				s.error = ErrContextTimedOut
				break
			}
		}
	}()
}

func (s *ResultIterator[T]) Error() error {
	if resultErr := s.result.Error(); resultErr != nil {
		if s.error != nil {
			return errors.Join(resultErr, s.error)
		}

		return resultErr
	}

	return s.error
}

func (s *ResultIterator[T]) Close() {
	s.cancelFunc()
	s.result.Close()
}

func (s *ResultIterator[T]) Chan() chan T {
	return s.valueC
}
