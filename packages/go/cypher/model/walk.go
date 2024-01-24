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

package model

import (
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type WalkCursor struct {
	Trunk         Expression
	Branches      []Expression
	currentBranch int
}

func (s *WalkCursor) CurrentBranch() Expression {
	return s.Branches[s.currentBranch]
}

func (s *WalkCursor) next() (Expression, bool) {
	if s.currentBranch < len(s.Branches) {
		next := s.Branches[s.currentBranch]
		s.currentBranch++

		return next, true
	}

	return nil, false
}

type WalkStack struct {
	stack []*WalkCursor
}

func newStack(root Expression) *WalkStack {
	return &WalkStack{
		stack: []*WalkCursor{{
			Branches:      []Expression{root},
			currentBranch: 0,
		}},
	}
}

func (s *WalkStack) Push(trunk Expression) *WalkCursor {
	cursor := &WalkCursor{
		Trunk:         trunk,
		currentBranch: 0,
	}

	s.stack = append(s.stack, cursor)
	return cursor
}

func (s *WalkStack) Trunk() Expression {
	if s.Empty() {
		return nil
	}

	return s.Peek().Trunk
}

func (s *WalkStack) Empty() bool {
	return len(s.stack) == 0
}

func (s *WalkStack) Peek() *WalkCursor {
	return s.stack[len(s.stack)-1]
}

func (s *WalkStack) PeekAt(depth int) *WalkCursor {
	if index := len(s.stack) - depth - 1; depth >= 0 {
		return s.stack[index]
	}

	return nil
}

func (s *WalkStack) Pop() {
	s.stack = s.stack[:len(s.stack)-1]
}

func CollectExpression(cursor *WalkCursor, expression Expression) {
	if expression != nil {
		cursor.Branches = append(cursor.Branches, expression)
	}
}

func CollectExpressions(cursor *WalkCursor, expressions []Expression) {
	for _, expression := range expressions {
		CollectExpression(cursor, expression)
	}
}

func Collect[T any](cursor *WalkCursor, expression *T) {
	if expression != nil {
		CollectExpression(cursor, expression)
	}
}

func CollectSlice[T any](cursor *WalkCursor, expressions []*T) {
	for _, expression := range expressions {
		Collect(cursor, expression)
	}
}

type Visitor interface {
	Enter(stack *WalkStack, expression Expression) error
	Exit(stack *WalkStack, expression Expression) error
}

type VisitorFunc func(stack *WalkStack, branch Expression) error

type visitor struct {
	enterVisitor VisitorFunc
	exitVisitor  VisitorFunc
}

func NewVisitor(enterVisitor VisitorFunc, exitVisitor VisitorFunc) Visitor {
	return visitor{
		enterVisitor: enterVisitor,
		exitVisitor:  exitVisitor,
	}
}

func (s visitor) Enter(stack *WalkStack, expression Expression) error {
	if s.enterVisitor != nil {
		return s.enterVisitor(stack, expression)
	}

	return nil
}

func (s visitor) Exit(stack *WalkStack, expression Expression) error {
	if s.exitVisitor != nil {
		return s.exitVisitor(stack, expression)
	}

	return nil
}

type CollectorFunc func(nextCursor *WalkCursor, expression Expression) bool

func cypherModelCollect(nextCursor *WalkCursor, expression Expression) bool {
	switch typedExpr := expression.(type) {
	case ExpressionList:
		CollectExpressions(nextCursor, typedExpr.GetAll())

	case *RegularQuery:
		Collect(nextCursor, typedExpr.SingleQuery)

	case *SingleQuery:
		Collect(nextCursor, typedExpr.SinglePartQuery)
		Collect(nextCursor, typedExpr.MultiPartQuery)

	case *MultiPartQuery:
		CollectSlice(nextCursor, typedExpr.Parts)
		Collect(nextCursor, typedExpr.SinglePartQuery)

	case *MultiPartQueryPart:
		CollectSlice(nextCursor, typedExpr.ReadingClauses)
		CollectSlice(nextCursor, typedExpr.UpdatingClauses)
		Collect(nextCursor, typedExpr.With)

	case *Quantifier:
		Collect(nextCursor, typedExpr.Filter)

	case *FilterExpression:
		Collect(nextCursor, typedExpr.Specifier)
		Collect(nextCursor, typedExpr.Where)

	case *IDInCollection:
		Collect(nextCursor, typedExpr.Variable)
		CollectExpression(nextCursor, typedExpr.Expression)

	case *With:
		Collect(nextCursor, typedExpr.Projection)
		Collect(nextCursor, typedExpr.Where)

	case *Unwind:
		CollectExpression(nextCursor, typedExpr.Expression)
		Collect(nextCursor, typedExpr.Binding)

	case *ReadingClause:
		Collect(nextCursor, typedExpr.Match)
		Collect(nextCursor, typedExpr.Unwind)

	case *SinglePartQuery:
		CollectSlice(nextCursor, typedExpr.ReadingClauses)
		CollectExpressions(nextCursor, typedExpr.UpdatingClauses)
		Collect(nextCursor, typedExpr.Return)

	case *Remove:
		CollectSlice(nextCursor, typedExpr.Items)

	case *Set:
		CollectSlice(nextCursor, typedExpr.Items)

	case *SetItem:
		CollectExpression(nextCursor, typedExpr.Left)
		CollectExpression(nextCursor, typedExpr.Right)

	case *Negation:
		CollectExpression(nextCursor, typedExpr.Expression)

	case *PartialComparison:
		CollectExpression(nextCursor, typedExpr.Right)

	case *Parenthetical:
		CollectExpression(nextCursor, typedExpr.Expression)

	case *PatternElement:
		CollectExpression(nextCursor, typedExpr.Element)

	case *Match:
		Collect(nextCursor, typedExpr.Where)
		CollectSlice(nextCursor, typedExpr.Pattern)

	case *Create:
		CollectSlice(nextCursor, typedExpr.Pattern)

	case *Return:
		Collect(nextCursor, typedExpr.Projection)

	case *FunctionInvocation:
		CollectExpressions(nextCursor, typedExpr.Arguments)

	case *Comparison:
		CollectExpression(nextCursor, typedExpr.Left)
		CollectSlice(nextCursor, typedExpr.Partials)

	case *PatternPredicate:
		CollectSlice(nextCursor, typedExpr.PatternElements)

	case *SortItem:
		CollectExpression(nextCursor, typedExpr.Expression)

	case *Order:
		CollectSlice(nextCursor, typedExpr.Items)

	case *Projection:
		CollectExpressions(nextCursor, typedExpr.Items)
		Collect(nextCursor, typedExpr.Order)

	case *ProjectionItem:
		CollectExpression(nextCursor, typedExpr.Expression)
		CollectExpression(nextCursor, typedExpr.Binding)

	case *ArithmeticExpression:
		CollectExpression(nextCursor, typedExpr.Left)
		CollectSlice(nextCursor, typedExpr.Partials)

	case *PartialArithmeticExpression:
		CollectExpression(nextCursor, typedExpr.Right)

	case *Delete:
		CollectExpressions(nextCursor, typedExpr.Expressions)

	case *KindMatcher:
		CollectExpression(nextCursor, typedExpr.Reference)

	case *RemoveItem:
		CollectExpression(nextCursor, typedExpr.KindMatcher)
		Collect(nextCursor, typedExpr.Property)

	case *PropertyLookup:
		CollectExpression(nextCursor, typedExpr.Atom)

	case *UpdatingClause:
		CollectExpression(nextCursor, typedExpr.Clause)

	case *NodePattern:
		CollectExpression(nextCursor, typedExpr.Properties)
		CollectExpression(nextCursor, typedExpr.Binding)

	case *PatternPart:
		CollectSlice(nextCursor, typedExpr.PatternElements)
		CollectExpression(nextCursor, typedExpr.Binding)

	case *RelationshipPattern:
		CollectExpression(nextCursor, typedExpr.Properties)
		CollectExpression(nextCursor, typedExpr.Binding)

	case *Properties:
		Collect(nextCursor, typedExpr.Parameter)

	case *Variable, *Literal, *Parameter, *RangeQuantifier, graph.Kinds:
	// Valid model elements but no further descent required

	case nil:
	default:
		return false
	}

	return true
}

func Walk(root Expression, visitor Visitor, extensions ...CollectorFunc) error {
	stack := newStack(root)

	for !stack.Empty() {
		currentCursor := stack.Peek()

		if nextExpr, hasNext := currentCursor.next(); hasNext {
			// On enter of new node
			if err := visitor.Enter(stack, nextExpr); err != nil {
				return err
			}

			if nextCursor := stack.Push(nextExpr); !cypherModelCollect(nextCursor, nextExpr) {
				collected := false

				for _, extension := range extensions {
					if extension(nextCursor, nextExpr) {
						collected = true
						break
					}
				}

				if !collected {
					return fmt.Errorf("unsupported type for model traversal %T", nextExpr)
				}
			}
		} else {
			stack.Pop()

			if err := visitor.Exit(stack, currentCursor.Trunk); err != nil {
				return err
			}
		}
	}

	return nil
}
