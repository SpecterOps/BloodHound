// Copyright 2024 Specter Ops, Inc.
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

package walk

import (
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

type CancelableErrorHandler interface {
	Done() bool
	Error() error
	SetDone()
	SetError(err error)
	SetErrorf(format string, args ...any)
}

type Visitor[N any] interface {
	CancelableErrorHandler

	Consume()
	WasConsumed() bool
	Enter(node N)
	Visit(node N)
	Exit(node N)
}

type cancelableErrorHandler struct {
	done bool
	errs []error
}

func (s *cancelableErrorHandler) Done() bool {
	return s.done
}

func (s *cancelableErrorHandler) SetDone() {
	s.done = true
}

func (s *cancelableErrorHandler) SetError(err error) {
	if err != nil {
		s.errs = append(s.errs, err)
		s.done = true
	}
}

func (s *cancelableErrorHandler) SetErrorf(format string, args ...any) {
	s.SetError(fmt.Errorf(format, args...))
}

func (s *cancelableErrorHandler) Error() error {
	return errors.Join(s.errs...)
}

func NewCancelableErrorHandler() CancelableErrorHandler {
	return &cancelableErrorHandler{}
}

type composableVisitor[N any] struct {
	CancelableErrorHandler

	currentSyntaxNodeConsumed bool
}

func (s *composableVisitor[N]) Consume() {
	s.currentSyntaxNodeConsumed = true
}

func (s *composableVisitor[N]) WasConsumed() bool {
	consumed := s.currentSyntaxNodeConsumed
	s.currentSyntaxNodeConsumed = false

	return consumed
}

func (s *composableVisitor[N]) Enter(node N) {
}

func (s *composableVisitor[N]) Visit(node N) {
}

func (s *composableVisitor[N]) Exit(node N) {
}

func NewVisitor[E any]() Visitor[E] {
	return &composableVisitor[E]{
		CancelableErrorHandler: NewCancelableErrorHandler(),
	}
}

type Order int

const (
	OrderPrefix Order = iota
	OrderInfix
	OrderPostfix
)

type SimpleVisitorFunc[N any] func(node N, errorHandler CancelableErrorHandler)

type simpleVisitor[N any] struct {
	Visitor[N]

	order       Order
	visitorFunc SimpleVisitorFunc[N]
}

func NewSimpleVisitor[N any](visitorFunc SimpleVisitorFunc[N]) Visitor[N] {
	return &simpleVisitor[N]{
		Visitor:     NewVisitor[N](),
		visitorFunc: visitorFunc,
	}
}

func (s *simpleVisitor[N]) Enter(node N) {
	if s.order == OrderPrefix {
		s.visitorFunc(node, s)
	}
}

func (s *simpleVisitor[N]) Visit(node N) {
	if s.order == OrderInfix {
		s.visitorFunc(node, s)
	}
}

func (s *simpleVisitor[N]) Exit(node N) {
	if s.order == OrderPostfix {
		s.visitorFunc(node, s)
	}
}

type Cursor[N any] struct {
	Node        N
	Branches    []N
	BranchIndex int
}

func (s *Cursor[N]) AddBranches(branches ...N) {
	s.Branches = append(s.Branches, branches...)
}

func (s *Cursor[N]) NumBranchesRemaining() int {
	return len(s.Branches) - s.BranchIndex
}

func (s *Cursor[N]) IsFirstVisit() bool {
	return s.BranchIndex == 0
}

func (s *Cursor[N]) HasNext() bool {
	return s.BranchIndex < len(s.Branches)
}

func (s *Cursor[N]) NextBranch() N {
	nextBranch := s.Branches[s.BranchIndex]
	s.BranchIndex += 1

	return nextBranch
}

func Generic[E any](node E, visitor Visitor[E], cursorConstructor func(node E) (*Cursor[E], error)) error {
	var stack []*Cursor[E]

	if cursor, err := cursorConstructor(node); err != nil {
		return err
	} else {
		stack = append(stack, cursor)
	}

	for len(stack) > 0 && !visitor.Done() {
		var (
			nextNode     = stack[len(stack)-1]
			isFirstVisit = nextNode.IsFirstVisit()
		)

		if isFirstVisit {
			visitor.Enter(nextNode.Node)

			if err := visitor.Error(); err != nil {
				return err
			}
		}

		if nextNode.HasNext() && !visitor.WasConsumed() {
			if !isFirstVisit {
				visitor.Visit(nextNode.Node)

				if err := visitor.Error(); err != nil {
					return err
				}
			}

			if cursor, err := cursorConstructor(nextNode.NextBranch()); err != nil {
				return err
			} else {
				stack = append(stack, cursor)
			}
		} else {
			visitor.Exit(nextNode.Node)

			if err := visitor.Error(); err != nil {
				return err
			}

			stack = stack[0 : len(stack)-1]
		}
	}

	return nil
}

func PgSQL(node pgsql.SyntaxNode, visitor Visitor[pgsql.SyntaxNode]) error {
	return Generic(node, visitor, newSQLWalkCursor)
}

func Cypher(node cypher.SyntaxNode, visitor Visitor[cypher.SyntaxNode]) error {
	return Generic(node, visitor, newCypherWalkCursor)
}
