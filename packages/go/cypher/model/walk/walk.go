package walk

import (
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
)

type CancelableErrorHandler interface {
	Done() bool
	Error() error
	SetDone()
	SetError(err error)
	SetErrorf(format string, args ...any)
}

type cancelableErrorHandler struct {
	done bool
	err  error
}

func (s *cancelableErrorHandler) Done() bool {
	return s.done
}

func (s *cancelableErrorHandler) SetDone() {
	s.done = true
}

func (s *cancelableErrorHandler) SetError(err error) {
	s.err = err
	s.done = true
}

func (s *cancelableErrorHandler) SetErrorf(format string, args ...any) {
	s.SetError(fmt.Errorf(format, args...))
}

func (s *cancelableErrorHandler) Error() error {
	return s.err
}

func NewCancelableErrorHandler() CancelableErrorHandler {
	return &cancelableErrorHandler{}
}

type composableHierarchicalVisitor[E any] struct {
	CancelableErrorHandler
}

func (s composableHierarchicalVisitor[E]) Enter(expression E) {
}

func (s composableHierarchicalVisitor[E]) Visit(expression E) {
}

func (s composableHierarchicalVisitor[E]) Exit(expression E) {
}

func NewComposableHierarchicalVisitor[E any]() HierarchicalVisitor[E] {
	return composableHierarchicalVisitor[E]{
		CancelableErrorHandler: NewCancelableErrorHandler(),
	}
}

type HierarchicalVisitor[N any] interface {
	CancelableErrorHandler

	Enter(node N)
	Visit(node N)
	Exit(node N)
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

func SetBranches[E any, T any](cursor *Cursor[E], branches ...T) error {
	for _, branch := range branches {
		if eTypedBranch, isEType := any(branch).(E); !isEType {
			var emptyE E
			return fmt.Errorf("branch type %T does not convert to %T", branch, emptyE)
		} else {
			cursor.Branches = append(cursor.Branches, eTypedBranch)
		}
	}

	return nil
}

type OrderedVisitors[E any] struct {
	visitors []HierarchicalVisitor[E]
	done     []bool
	err      error
}

func NewOrderedVisitors[E any](visitors ...HierarchicalVisitor[E]) HierarchicalVisitor[E] {
	return &OrderedVisitors[E]{
		visitors: visitors,
		done:     make([]bool, len(visitors)),
	}
}

func (s *OrderedVisitors[E]) Done() bool {
	if s.err != nil {
		return true
	}

	return len(s.visitors) == 0
}

func (s *OrderedVisitors[E]) Error() error {
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

func (s *OrderedVisitors[E]) SetDone() {
	for idx := 0; idx < len(s.done); idx++ {
		s.done[idx] = true
	}
}

func (s *OrderedVisitors[E]) SetVisitorDone(idx int) {
	s.done[idx] = true
}

func (s *OrderedVisitors[E]) SetError(err error) {
	s.err = err
}

func (s *OrderedVisitors[E]) SetErrorf(format string, args ...any) {
	s.SetError(fmt.Errorf(format, args...))
}

func (s *OrderedVisitors[E]) eachVisitor(eachFunc func(visitor HierarchicalVisitor[E])) {
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

func (s *OrderedVisitors[E]) Enter(expression E) {
	s.eachVisitor(func(visitor HierarchicalVisitor[E]) {
		visitor.Enter(expression)
	})
}

func (s *OrderedVisitors[E]) Visit(expression E) {
	s.eachVisitor(func(visitor HierarchicalVisitor[E]) {
		visitor.Visit(expression)
	})
}

func (s *OrderedVisitors[E]) Exit(expression E) {
	s.eachVisitor(func(visitor HierarchicalVisitor[E]) {
		visitor.Exit(expression)
	})
}

func Generic[E any](expression E, visitor HierarchicalVisitor[E], cursorConstructor func(expression E) (*Cursor[E], error)) error {
	var stack []*Cursor[E]

	if cursor, err := cursorConstructor(expression); err != nil {
		return err
	} else {
		stack = append(stack, cursor)
	}

	for len(stack) > 0 && !visitor.Done() {
		var (
			nextExpressionNode = stack[len(stack)-1]
			isFirstVisit       = nextExpressionNode.IsFirstVisit()
		)

		if isFirstVisit {
			visitor.Enter(nextExpressionNode.Node)

			if err := visitor.Error(); err != nil {
				return err
			}
		}

		if nextExpressionNode.HasNext() {
			if !isFirstVisit {
				visitor.Visit(nextExpressionNode.Node)

				if err := visitor.Error(); err != nil {
					return err
				}
			}

			if cursor, err := cursorConstructor(nextExpressionNode.NextBranch()); err != nil {
				return err
			} else {
				stack = append(stack, cursor)
			}
		} else {
			visitor.Exit(nextExpressionNode.Node)

			if err := visitor.Error(); err != nil {
				return err
			}

			stack = stack[0 : len(stack)-1]
		}
	}

	return nil
}

func PgSQL(node pgsql.SyntaxNode, visitors ...HierarchicalVisitor[pgsql.SyntaxNode]) error {
	if len(visitors) == 1 {
		return Generic(node, visitors[0], newSQLWalkCursor)
	}

	return Generic(node, NewOrderedVisitors[pgsql.SyntaxNode](visitors...), newSQLWalkCursor)
}

func Cypher(expression cypher.SyntaxNode, visitors ...HierarchicalVisitor[cypher.SyntaxNode]) error {
	if len(visitors) == 1 {
		return Generic(expression, visitors[0], newCypherWalkCursor)
	}

	return Generic(expression, NewOrderedVisitors[cypher.SyntaxNode](visitors...), newCypherWalkCursor)
}
