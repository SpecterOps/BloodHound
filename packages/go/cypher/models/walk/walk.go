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

type HierarchicalVisitor[N any] interface {
	CancelableErrorHandler

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

type composableHierarchicalVisitor[N any] struct {
	CancelableErrorHandler
}

func (s composableHierarchicalVisitor[N]) Enter(node N) {
}

func (s composableHierarchicalVisitor[N]) Visit(node N) {
}

func (s composableHierarchicalVisitor[N]) Exit(node N) {
}

func NewComposableHierarchicalVisitor[E any]() HierarchicalVisitor[E] {
	return composableHierarchicalVisitor[E]{
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
	HierarchicalVisitor[N]

	order       Order
	visitorFunc SimpleVisitorFunc[N]
}

func NewSimpleVisitor[N any](visitorFunc SimpleVisitorFunc[N]) HierarchicalVisitor[N] {
	return &simpleVisitor[N]{
		HierarchicalVisitor: NewComposableHierarchicalVisitor[N](),
		visitorFunc:         visitorFunc,
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

type OrderedVisitors[N any] struct {
	visitors []HierarchicalVisitor[N]
	done     []bool
	err      error
}

func NewOrderedVisitors[N any](visitors ...HierarchicalVisitor[N]) HierarchicalVisitor[N] {
	return &OrderedVisitors[N]{
		visitors: visitors,
		done:     make([]bool, len(visitors)),
	}
}

func (s *OrderedVisitors[N]) Done() bool {
	if s.err != nil {
		return true
	}

	return len(s.visitors) == 0
}

func (s *OrderedVisitors[N]) Error() error {
	var errs []error

	if s.err != nil {
		errs = append(errs, s.err)
	}

	for _, visitor := range s.visitors {
		if err := visitor.Error(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (s *OrderedVisitors[N]) SetDone() {
	for idx := 0; idx < len(s.done); idx++ {
		s.done[idx] = true
	}
}

func (s *OrderedVisitors[N]) SetVisitorDone(idx int) {
	s.done[idx] = true
}

func (s *OrderedVisitors[N]) SetError(err error) {
	s.err = err
}

func (s *OrderedVisitors[N]) SetErrorf(format string, args ...any) {
	s.SetError(fmt.Errorf(format, args...))
}

func (s *OrderedVisitors[N]) eachVisitor(eachFunc func(visitor HierarchicalVisitor[N])) {
	for idx, visitor := range s.visitors {
		if s.done[idx] {
			continue
		}

		eachFunc(visitor)

		if visitor.Error() != nil {
			s.SetDone()
		} else if visitor.Done() {
			s.SetVisitorDone(idx)
		}
	}
}

func (s *OrderedVisitors[N]) Enter(node N) {
	s.eachVisitor(func(visitor HierarchicalVisitor[N]) {
		visitor.Enter(node)
	})
}

func (s *OrderedVisitors[N]) Visit(node N) {
	s.eachVisitor(func(visitor HierarchicalVisitor[N]) {
		visitor.Visit(node)
	})
}

func (s *OrderedVisitors[N]) Exit(node N) {
	s.eachVisitor(func(visitor HierarchicalVisitor[N]) {
		visitor.Exit(node)
	})
}

func Generic[E any](node E, visitor HierarchicalVisitor[E], cursorConstructor func(node E) (*Cursor[E], error)) error {
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

		if nextNode.HasNext() {
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

func PgSQL(node pgsql.SyntaxNode, visitors ...HierarchicalVisitor[pgsql.SyntaxNode]) error {
	if node == nil {
		return nil
	}

	if len(visitors) == 1 {
		// If there's only one visitor no need to wrap and add indirection to visit calls
		return Generic(node, visitors[0], newSQLWalkCursor)
	}

	return Generic(node, NewOrderedVisitors[pgsql.SyntaxNode](visitors...), newSQLWalkCursor)
}

func Cypher(node cypher.SyntaxNode, visitors ...HierarchicalVisitor[cypher.SyntaxNode]) error {
	if node == nil {
		return nil
	}

	if len(visitors) == 1 {
		// If there's only one visitor no need to wrap and add indirection to visit calls
		return Generic(node, visitors[0], newCypherWalkCursor)
	}

	return Generic(node, NewOrderedVisitors[cypher.SyntaxNode](visitors...), newCypherWalkCursor)
}
