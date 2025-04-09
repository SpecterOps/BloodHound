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

package translate

import (
	"fmt"

	"github.com/specterops/bloodhound/cypher/models/walk"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func leftNodeConstraint(edgeIdentifier, nodeIdentifier pgsql.Identifier, direction graph.Direction) (pgsql.Expression, error) {
	switch direction {
	case graph.DirectionOutbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionInbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionBoth:
		return pgsql.NewBinaryExpression(
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			),
			pgsql.OperatorOr,
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			),
		), nil

	default:
		return nil, fmt.Errorf("unsupported direction: %d", direction)
	}
}

func leftNodeTraversalStepConstraint(traversalStep *TraversalStep) (pgsql.Expression, error) {
	return leftNodeConstraint(
		traversalStep.Edge.Identifier,
		traversalStep.LeftNode.Identifier,
		traversalStep.Direction)
}

func rightEdgeConstraint(traversalStep *TraversalStep) (pgsql.Expression, error) {
	switch traversalStep.Edge.DataType {
	case pgsql.EdgeComposite:
		switch traversalStep.Direction {
		case graph.DirectionOutbound:
			return &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
				ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
			}, nil

		case graph.DirectionInbound:
			return &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
				ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			}, nil

		case graph.DirectionBoth:
			return pgsql.NewBinaryExpression(
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
				),
				pgsql.OperatorOr,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				),
			), nil

		default:
			return nil, fmt.Errorf("unsupported direction: %d", traversalStep.Direction)
		}

	case pgsql.ExpansionEdge:
		switch traversalStep.Direction {
		case graph.DirectionOutbound:
			return pgsql.NewBinaryExpression(
				pgsql.RowColumnReference{
					Identifier: &pgsql.ArrayIndex{
						Expression: traversalStep.Edge.Identifier,
						Indexes: []pgsql.Expression{
							pgsql.FunctionCall{
								Function: pgsql.FunctionArrayLength,
								Parameters: []pgsql.Expression{
									traversalStep.Edge.Identifier,
									pgsql.NewLiteral(1, pgsql.Int),
								},
								CastType: pgsql.Int,
							},
						},
					},
					Column: pgsql.ColumnEndID,
				},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
			), nil

		case graph.DirectionInbound:
			return &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
				ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			}, nil

		default:
			return nil, fmt.Errorf("unsupported direction: %d", traversalStep.Direction)
		}

	default:
		return nil, fmt.Errorf("invalid root edge type: %s", traversalStep.Edge.DataType)
	}
}

func terminalNodeConstraint(edgeIdentifier, nodeIdentifier pgsql.Identifier, direction graph.Direction) (pgsql.Expression, error) {
	switch direction {
	case graph.DirectionOutbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionInbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionBoth:
		return pgsql.NewBinaryExpression(
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			),
			pgsql.OperatorOr,
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			),
		), nil

	default:
		return nil, fmt.Errorf("unsupported direction: %d", direction)
	}
}

func rightNodeTraversalStepJoinCondition(traversalStep *TraversalStep) (pgsql.Expression, error) {
	return terminalNodeConstraint(
		traversalStep.Edge.Identifier,
		traversalStep.RightNode.Identifier,
		traversalStep.Direction)
}

func isSyntaxNodeSatisfied(syntaxNode pgsql.SyntaxNode) (bool, error) {
	var (
		satisfied = true
		err       = walk.PgSQL(syntaxNode, walk.NewSimpleVisitor[pgsql.SyntaxNode](
			func(node pgsql.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
				switch typedNode := node.(type) {
				case pgsql.SyntaxNodeFuture:
					satisfied = typedNode.Satisfied()

					if !satisfied {
						errorHandler.SetDone()
					}
				}
			},
		))
	)

	return satisfied, err
}

// Constraint is an extracted expression that contains an identifier set of symbols required to be
// in scope for this constraint to be solvable.
type Constraint struct {
	Dependencies *pgsql.IdentifierSet
	Expression   pgsql.Expression
}

func (s *Constraint) Merge(other *Constraint) error {
	if other.Dependencies != nil && other.Expression != nil {
		newExpression := pgsql.OptionalAnd(s.Expression, other.Expression)

		switch typedNewExpression := newExpression.(type) {
		case *pgsql.UnaryExpression:
			if err := applyUnaryExpressionTypeHints(typedNewExpression); err != nil {
				return err
			}

		case *pgsql.BinaryExpression:
			if err := applyBinaryExpressionTypeHints(typedNewExpression); err != nil {
				return err
			}
		}

		s.Dependencies.MergeSet(other.Dependencies)
		s.Expression = newExpression
	}

	return nil
}

// ConstraintTracker is a tool for associating constraints (e.g. binary or unary expressions
// that constrain a set of identifiers) with the identifier set they constrain.
//
// This is useful for rewriting a where-clause so that conjoined components can be isolated:
//
// Where Clause:
//
// where a.name = 'a' and b.name = 'b' and c.name = 'c' and a.num_a > 1 and a.ef = b.ef + c.ef
//
// Isolated Constraints:
//
//	"a":           a.name = 'a' and a.num_a > 1
//	"b":           b.name = 'b'
//	"c":           c.name = 'c'
//	"a", "b", "c": a.ef = b.ef + c.ef
type ConstraintTracker struct {
	Constraints []*Constraint
}

func NewConstraintTracker() *ConstraintTracker {
	return &ConstraintTracker{}
}

func (s *ConstraintTracker) HasConstraints(scope *pgsql.IdentifierSet) (bool, error) {
	for idx := 0; idx < len(s.Constraints); idx++ {
		nextConstraint := s.Constraints[idx]

		if syntaxNodeSatisfied, err := isSyntaxNodeSatisfied(nextConstraint.Expression); err != nil {
			return false, err
		} else if syntaxNodeSatisfied && scope.Satisfies(nextConstraint.Dependencies) {
			return true, nil
		}
	}

	return false, nil
}

func (s *ConstraintTracker) ConsumeAll() (*Constraint, error) {
	var (
		constraintExpressions = make([]pgsql.Expression, len(s.Constraints))
		matchedDependencies   = pgsql.NewIdentifierSet()
	)

	for idx, constraint := range s.Constraints {
		constraintExpressions[idx] = constraint.Expression
		matchedDependencies.MergeSet(constraint.Dependencies)
	}

	// Clear the internal constraint slice
	s.Constraints = s.Constraints[:0]

	if conjoined, err := ConjoinExpressions(constraintExpressions); err != nil {
		return nil, err
	} else {
		return &Constraint{
			Dependencies: matchedDependencies,
			Expression:   conjoined,
		}, nil
	}
}

/*
ConsumeSet takes a given scope (a set of identifiers considered in-scope) and locates all constraints that can
be satisfied by the scope's identifiers.

```

	visible := pgsql.IdentifierSet{
		"a": struct{}{},
		"b": struct{}{},
	}

	tracker := ConstraintTracker{
		Constraints: []*Constraint{{
			Dependencies: pgsql.IdentifierSet{
				"a": struct{}{},
			},
			Expression: &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{"a", "name"},
				ROperand: pgsql.Literal{
					Value: "a",
				},
			},
		}},
	}

	satisfiedConstraints, expression := tracker.ConsumeSet(visible)

```
*/
func (s *ConstraintTracker) ConsumeSet(visible *pgsql.IdentifierSet) (*Constraint, error) {
	var (
		matchedDependencies   = pgsql.NewIdentifierSet()
		constraintExpressions []pgsql.Expression
	)

	for idx := 0; idx < len(s.Constraints); {
		nextConstraint := s.Constraints[idx]

		// If this is a syntax node that has not been realized do not allow the constraint it represents
		// to be consumed even if the dependencies are satisfied
		if syntaxNodeSatisfied, err := isSyntaxNodeSatisfied(nextConstraint.Expression); err != nil {
			return nil, err
		} else if !syntaxNodeSatisfied || !visible.Satisfies(nextConstraint.Dependencies) {
			// This constraint isn't satisfied, move to the next one
			idx += 1
		} else {
			// Remove this constraint
			s.Constraints = append(s.Constraints[:idx], s.Constraints[idx+1:]...)

			// Append the constraint as a conjoined expression
			constraintExpressions = append(constraintExpressions, nextConstraint.Expression)

			// Track which identifiers were satisfied
			matchedDependencies.MergeSet(nextConstraint.Dependencies)
		}
	}

	if conjoined, err := ConjoinExpressions(constraintExpressions); err != nil {
		return nil, err
	} else {
		return &Constraint{
			Dependencies: matchedDependencies,
			Expression:   conjoined,
		}, nil
	}
}

func (s *ConstraintTracker) Constrain(dependencies *pgsql.IdentifierSet, constraintExpression pgsql.Expression) error {
	for _, constraint := range s.Constraints {
		if constraint.Dependencies.Matches(dependencies) {
			joinedExpression := pgsql.NewBinaryExpression(
				constraintExpression,
				pgsql.OperatorAnd,
				constraint.Expression,
			)

			if err := applyBinaryExpressionTypeHints(joinedExpression); err != nil {
				return err
			}

			constraint.Expression = joinedExpression
			return nil
		}
	}

	s.Constraints = append(s.Constraints, &Constraint{
		Dependencies: dependencies,
		Expression:   constraintExpression,
	})

	return nil
}

// PatternConstraints is a struct that represents all constraints that can be solved during a traversal step
// pattern: `()-[]->()`.
type PatternConstraints struct {
	LeftNode  *Constraint
	Edge      *Constraint
	RightNode *Constraint
}

// OptimizePatternConstraintBalance considers the constraints that apply to a pattern segment's bound identifiers.
//
// If only the right side of the pattern segment is constrained, this could result in an imbalanced expansion where one side
// of the traversal has an extreme disparity in search space.
//
// In cases that match this heuristic, it's beneficial to begin the traversal with the most tightly constrained set
// of nodes. To accomplish this we flip the order of the traversal step.
func (s *PatternConstraints) OptimizePatternConstraintBalance(scope *Scope, traversalStep *TraversalStep) error {
	if leftNodeSelectivity, err := MeasureSelectivity(scope, traversalStep.LeftNodeBound, s.LeftNode.Expression); err != nil {
		return err
	} else if rightNodeSelectivity, err := MeasureSelectivity(scope, traversalStep.RightNodeBound, s.RightNode.Expression); err != nil {
		return err
	} else if rightNodeSelectivity > leftNodeSelectivity {
		// (a)-[*..]->(b:Constraint)
		// (a)<-[*..]-(b:Constraint)
		traversalStep.FlipNodes()
		s.FlipNodes()
	}

	return nil
}

func (s *PatternConstraints) FlipNodes() {
	oldLeftNode := s.LeftNode
	s.LeftNode = s.RightNode
	s.RightNode = oldLeftNode
}

const (
	recursivePattern    = true
	nonRecursivePattern = false
)

func consumePatternConstraints(isFirstTraversalStep, isRecursivePattern bool, traversalStep *TraversalStep, treeTranslator *ExpressionTreeTranslator) (PatternConstraints, error) {
	var (
		constraints PatternConstraints
		err         error
	)

	// Even if this isn't the first traversal and the node may be already bound, this should result in an empty
	// constraint instead of a nil value for `leftNode`
	if isFirstTraversalStep {
		// If this is the first traversal step then the left node is just coming into scope
		traversalStep.Frame.Export(traversalStep.LeftNode.Identifier)
	}

	// Track the identifiers visible at this frame to correctly assign the remaining constraints
	knownBindings := traversalStep.Frame.Known()

	if constraints.LeftNode, err = treeTranslator.ConsumeConstraintsFromVisibleSet(knownBindings); err != nil {
		return constraints, err
	}

	if isRecursivePattern {
		// The exclusion below is done at this step in the process since the recursive descent portion of an expansion
		// will no longer have a reference to the root node; any dependent interaction between the root and terminal
		// nodes would require an additional join. By not consuming the remaining constraints for the root and terminal
		// nodes, they become visible up in the outer select of the recursive CTE.
		knownBindings.Remove(traversalStep.LeftNode.Identifier)
	}

	// Export the edge identifier first
	traversalStep.Frame.Export(traversalStep.Edge.Identifier)
	knownBindings.Add(traversalStep.Edge.Identifier)

	if constraints.Edge, err = treeTranslator.ConsumeConstraintsFromVisibleSet(knownBindings); err != nil {
		return constraints, err
	}

	// Export the right node identifier last
	traversalStep.Frame.Export(traversalStep.RightNode.Identifier)
	knownBindings.Add(traversalStep.RightNode.Identifier)

	if constraints.RightNode, err = treeTranslator.ConsumeConstraintsFromVisibleSet(knownBindings); err != nil {
		return constraints, err
	}

	return constraints, nil
}
