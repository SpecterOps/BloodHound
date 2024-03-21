package pgsql

import (
	"errors"
	"fmt"
)

var (
	ErrOperatorAlreadyAssigned = errors.New("expression operator already assigned")
	ErrOperandAlreadyAssigned  = errors.New("expression operand already assigned")
)

func PeekAs[T any](tree *Tree) (T, error) {
	expression := tree.Peek()

	if value, isT := expression.(T); isT {
		return value, nil
	}

	var emptyT T
	return emptyT, fmt.Errorf("unable to convert expression %T to type %T", expression, emptyT)
}

type Tree struct {
	stack []Expression
}

func NewTree(root Expression) *Tree {
	return &Tree{
		stack: []Expression{root},
	}
}

func (s *Tree) Depth() int {
	return len(s.stack)
}

func (s *Tree) Root() Expression {
	return s.stack[0]
}

func (s *Tree) RootExpression() Expression {
	return s.stack[0].(Expression)
}

func (s *Tree) Peek() Expression {
	return s.stack[len(s.stack)-1]
}

func (s *Tree) Ascend(depth int) {
	s.stack = s.stack[:len(s.stack)-depth]
}

func (s *Tree) Pop() Expression {
	nextNode := s.Peek()
	s.stack = s.stack[:len(s.stack)-1]

	return nextNode
}

func (s *Tree) ContinueBinaryExpression(operator Operator, operand Expression) error {
	if assignmentTarget, err := PeekAs[*BinaryExpression](s); err != nil {
		return err
	} else if assignmentTarget.LOperand == nil {
		assignmentTarget.LOperand = operand
		assignmentTarget.Operator = operator
	} else if assignmentTarget.ROperand == nil {
		assignmentTarget.ROperand = operand
		assignmentTarget.Operator = operator
	} else {
		assignmentTarget.ROperand = &BinaryExpression{
			Operator: operator,
			LOperand: assignmentTarget.ROperand,
			ROperand: operand,
		}

		s.Push(assignmentTarget.ROperand)
	}

	return nil
}

func (s *Tree) Push(expression Expression) {
	s.stack = append(s.stack, expression)
}
