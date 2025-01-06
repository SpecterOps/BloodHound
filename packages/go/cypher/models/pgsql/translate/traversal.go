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

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func (s *Translator) buildDirectionalTraversalPatternRoot(leftNodeConstraints, edgeConstraints, rightNodeConstraints *Constraint, traversalStep *PatternSegment, direction graph.Direction) (pgsql.Select, error) {
	directionalSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	var (
		leftNodeJoinConstraint  *pgsql.BinaryExpression
		rightNodeJoinConstraint *pgsql.BinaryExpression
	)

	switch direction {
	case graph.DirectionOutbound:
		leftNodeJoinConstraint = &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
		}

		rightNodeJoinConstraint = &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.RightNode.Identifier, pgsql.ColumnID},
		}

	case graph.DirectionInbound:
		leftNodeJoinConstraint = &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
		}

		rightNodeJoinConstraint = &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.RightNode.Identifier, pgsql.ColumnID},
		}

	default:
		return directionalSelect, fmt.Errorf("direction must be set for directional traversal")
	}

	if leftNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{leftNodeConstraints.Expression, leftNodeJoinConstraint}); err != nil {
		return directionalSelect, err
	} else if rightNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rightNodeConstraints.Expression, rightNodeJoinConstraint}); err != nil {
		return directionalSelect, err
	} else {
		if err := rewriteIdentifierReferences(traversalStep.Frame, []pgsql.Expression{leftNodeJoinCondition, edgeConstraints.Expression, rightNodeJoinCondition}); err != nil {
			return directionalSelect, err
		}

		directionalSelect.Where = edgeConstraints.Expression
		directionalSelect.From = append(directionalSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			Joins: []pgsql.Join{{
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: leftNodeJoinCondition,
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: rightNodeJoinCondition,
				},
			}},
		})
	}

	return directionalSelect, nil
}

func (s *Translator) buildDirectionlessTraversalPatternRoot(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	if leftNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(traversalStep.LeftNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if rightNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(traversalStep.RightNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if edgeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Visible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else {
		directionalSelect := pgsql.Select{
			Projection: traversalStep.Projection,
		}

		if traversalStep.Frame.Previous != nil {
			directionalSelect.From = append(directionalSelect.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
				},
			})
		}

		switch traversalStep.Direction {
		case graph.DirectionInbound, graph.DirectionOutbound:
			nextSelect, err := s.buildDirectionalTraversalPatternRoot(leftNodeConstraints, edgeConstraints, rightNodeConstraints, traversalStep, traversalStep.Direction)

			return pgsql.Query{
				Body: nextSelect,
			}, err

		case graph.DirectionBoth:
			if inboundSelect, err := s.buildDirectionalTraversalPatternRoot(leftNodeConstraints, edgeConstraints, rightNodeConstraints, traversalStep, graph.DirectionInbound); err != nil {
				return pgsql.Query{}, err
			} else if outboundSelect, err := s.buildDirectionalTraversalPatternRoot(leftNodeConstraints, edgeConstraints, rightNodeConstraints, traversalStep, graph.DirectionOutbound); err != nil {
				return pgsql.Query{}, err
			} else {
				return pgsql.Query{
					Body: pgsql.SetOperation{
						Distinct: true,
						LOperand: inboundSelect,
						ROperand: outboundSelect,
						Operator: pgsql.OperatorUnion,
					},
				}, nil

			}

		}
	}

	return pgsql.Query{}, fmt.Errorf("unsupported")
}

func (s *Translator) buildTraversalPatternRoot(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	if traversalStep.Direction == graph.DirectionBoth {
		return s.buildDirectionlessTraversalPatternRoot(part, traversalStep)
	}

	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	if leftNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(traversalStep.LeftNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if rightNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(traversalStep.RightNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if edgeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Visible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else {
		if traversalStep.Frame.Previous != nil {
			nextSelect.From = append(nextSelect.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
				},
			})
		}

		if leftNodeJoinConstraint, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else if leftNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{leftNodeConstraints.Expression, leftNodeJoinConstraint}); err != nil {
			return pgsql.Query{}, err
		} else if rightNodeJoinConstraint, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else if rightNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rightNodeConstraints.Expression, rightNodeJoinConstraint}); err != nil {
			return pgsql.Query{}, err
		} else {
			if err := rewriteIdentifierReferences(traversalStep.Frame, []pgsql.Expression{leftNodeJoinCondition, edgeConstraints.Expression, rightNodeJoinCondition}); err != nil {
				return pgsql.Query{}, err
			}

			nextSelect.Where = edgeConstraints.Expression
			nextSelect.From = append(nextSelect.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
					Binding: models.ValueOptional(traversalStep.Edge.Identifier),
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: leftNodeJoinCondition,
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})
		}
	}

	return pgsql.Query{
		Body: nextSelect,
	}, nil
}

func (s *Translator) buildTraversalPatternStep(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	if rightNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(traversalStep.RightNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if edgeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Visible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if rightNodeJoinConstraint, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
		return pgsql.Query{}, err
	} else if rightNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rightNodeConstraints.Expression, rightNodeJoinConstraint}); err != nil {
		return pgsql.Query{}, err
	} else {
		if err := rewriteIdentifierReferences(traversalStep.Frame, []pgsql.Expression{edgeConstraints.Expression, rightNodeJoinCondition}); err != nil {
			return pgsql.Query{}, err
		}

		nextSelect.Where = edgeConstraints.Expression

		if traversalStep.Frame.Previous != nil {
			nextSelect.From = append(nextSelect.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
				},
			})
		}
		nextSelect.From = []pgsql.FromClause{{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
			},
		}, {
			Source: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			Joins: []pgsql.Join{{
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: rightNodeJoinCondition,
				},
			}},
		}}
	}

	return pgsql.Query{
		Body: nextSelect,
	}, nil
}
