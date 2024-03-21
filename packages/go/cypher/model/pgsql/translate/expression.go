package translate

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
)

func operandsToBinaryExpression(operator pgsql.Operator, rightOperand, leftOperand pgsql.Expression) *pgsql.BinaryExpression {
	binaryExpression := &pgsql.BinaryExpression{
		Operator: operator,
	}

	// Inspect left operand
	switch typedOperand := leftOperand.(type) {
	case *pgsql.BinaryExpression:
		binaryExpression.LOperand = typedOperand
		binaryExpression.LDependencies = typedOperand.CombinedDependencies()

	case pgsql.Identifier:
		binaryExpression.LOperand = typedOperand
		binaryExpression.LDependencies = map[pgsql.Identifier]struct{}{
			typedOperand: {},
		}

	case pgsql.CompoundIdentifier:
		binaryExpression.LOperand = typedOperand
		binaryExpression.LDependencies = map[pgsql.Identifier]struct{}{
			typedOperand.Root(): {},
		}

	default:
		binaryExpression.LOperand = typedOperand
		binaryExpression.LDependencies = map[pgsql.Identifier]struct{}{}
	}

	// Inspect right operand
	switch operand := rightOperand.(type) {
	case *pgsql.BinaryExpression:
		binaryExpression.ROperand = operand
		binaryExpression.RDependencies = operand.CombinedDependencies()

	case pgsql.Identifier:
		binaryExpression.ROperand = operand
		binaryExpression.RDependencies = operandIdentifiers(operand)

	case pgsql.CompoundIdentifier:
		binaryExpression.ROperand = operand
		binaryExpression.RDependencies = map[pgsql.Identifier]struct{}{
			operand.Root(): {},
		}

	default:
		binaryExpression.ROperand = operand
		binaryExpression.RDependencies = map[pgsql.Identifier]struct{}{}
	}

	return binaryExpression
}

func operandIdentifiers(operand pgsql.Expression) map[pgsql.Identifier]struct{} {
	switch typedOperand := operand.(type) {
	case *pgsql.BinaryExpression:
		return typedOperand.CombinedDependencies()

	case pgsql.Identifier:
		return map[pgsql.Identifier]struct{}{
			typedOperand: {},
		}

	case pgsql.CompoundIdentifier:
		return map[pgsql.Identifier]struct{}{
			typedOperand.Root(): {},
		}

	default:
		return map[pgsql.Identifier]struct{}{}
	}
}

type ExpressionTreeTranslator struct {
	constraintTracker *ConstraintTracker
	stack             []pgsql.Expression
	disjunctionDepth  int
	conjunctionDepth  int
}

func NewExpressionTreeTranslator(constraintTracker *ConstraintTracker) *ExpressionTreeTranslator {
	return &ExpressionTreeTranslator{
		constraintTracker: constraintTracker,
		disjunctionDepth:  0,
		conjunctionDepth:  0,
	}
}

func (s *ExpressionTreeTranslator) ConstrainOperand() error {
	switch operand := s.Pop().(type) {
	case *pgsql.BinaryExpression:
		s.constraintTracker.Constrain(operand.CombinedDependencies(), operand)

	case pgsql.Identifier:
		s.constraintTracker.Constrain(pgsql.AsIdentifierSet(operand), operand)

	case pgsql.CompoundIdentifier:
		s.constraintTracker.Constrain(pgsql.AsIdentifierSet(operand.Root()), operand)

	case *pgsql.UnaryExpression:
		switch unaryOperand := operand.Operand.(type) {
		case *pgsql.BinaryExpression:
			s.constraintTracker.Constrain(unaryOperand.CombinedDependencies(), operand)

		case pgsql.Identifier:
			s.constraintTracker.Constrain(pgsql.AsIdentifierSet(unaryOperand), operand)

		case pgsql.CompoundIdentifier:
			s.constraintTracker.Constrain(pgsql.AsIdentifierSet(unaryOperand.Root()), operand)

		default:
			return fmt.Errorf("unable to extract unary expression operand type: %T", operand.Operand)
		}

	default:
		return fmt.Errorf("unable to extract operand type: %T", operand)
	}

	return nil
}

func (s *ExpressionTreeTranslator) ConstraintRemainingOperands() error {
	// Pull the right operand only if one exists
	for !s.IsEmpty() {
		if err := s.ConstrainOperand(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ExpressionTreeTranslator) ConstrainOperandPair() error {
	// Always expect a left operand
	if s.IsEmpty() {
		return fmt.Errorf("expected at least one operand for constraint extraction")
	}

	if err := s.ConstrainOperand(); err != nil {
		return err
	}

	// Pull the right operand only if one exists
	if !s.IsEmpty() {
		return s.ConstrainOperand()
	}

	return nil
}

func (s *ExpressionTreeTranslator) IsEmpty() bool {
	return len(s.stack) == 0
}

func (s *ExpressionTreeTranslator) Pop() pgsql.Expression {
	var (
		popIdx = len(s.stack) - 1
		next   = s.stack[popIdx]
	)

	switch typedNext := next.(type) {
	case *pgsql.BinaryExpression:
		// Track this operator for expression tree extraction
		switch typedNext.Operator {
		case pgsql.OperatorAnd:
			s.conjunctionDepth -= 1

		case pgsql.OperatorOr:
			s.disjunctionDepth -= 1
		}
	}

	s.stack = s.stack[:popIdx]
	return next
}

func (s *ExpressionTreeTranslator) Peek() pgsql.Expression {
	return s.stack[len(s.stack)-1]
}

func (s *ExpressionTreeTranslator) Push(expression pgsql.Expression) {
	s.stack = append(s.stack, expression)
}

func (s *ExpressionTreeTranslator) EnterOperator(operator pgsql.Operator) {
	// Track this operator for expression tree extraction
	switch operator {
	case pgsql.OperatorAnd:
		s.conjunctionDepth += 1

	case pgsql.OperatorOr:
		s.disjunctionDepth += 1
	}
}

func (s *ExpressionTreeTranslator) ExitOperator(operator pgsql.Operator) error {
	// Track if this is an extraction or not
	extracting := false

	// Track this operator for expression tree extraction and look to see if it's a candidate for rewriting
	switch operator {
	case pgsql.OperatorAnd:
		// If this is a conjunction, it may be a candidate for extraction. Conjunctions are not candidates for
		// extraction if they are contained by a disjunction.
		extracting = s.disjunctionDepth == 0 && s.conjunctionDepth > 0

		if !extracting {
			// If not extracting, add this conjunction to the depth count
			s.conjunctionDepth += 1
		}

	case pgsql.OperatorOr:
		s.disjunctionDepth += 1
	}

	if extracting {
		return s.ConstrainOperandPair()
	}

	// Create and push a new binary expression with the given operator
	s.Push(operandsToBinaryExpression(operator, s.Pop(), s.Pop()))
	return nil
}

