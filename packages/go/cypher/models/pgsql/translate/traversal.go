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
	"errors"

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func (s *Translator) buildDirectionlessTraversalPatternRoot(traversalStep *TraversalStep) (pgsql.Query, error) {
	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	if previousFrame, hasPrevious := s.previousValidFrame(traversalStep.Frame); hasPrevious {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{previousFrame.Binding.Identifier},
			},
		})
	}

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
				Constraint: traversalStep.LeftNodeJoinCondition,
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: traversalStep.RightNodeJoinCondition,
			},
		}},
	})

	// Append all constraints to the where clause
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.LeftNodeConstraints, nextSelect.Where)
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.EdgeConstraints.Expression, nextSelect.Where)
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.RightNodeConstraints, nextSelect.Where)

	return pgsql.Query{
		Body: nextSelect,
	}, nil
}

func (s *Translator) buildTraversalPatternRoot(partFrame *Frame, traversalStep *TraversalStep) (pgsql.Query, error) {
	if traversalStep.Direction == graph.DirectionBoth {
		return s.buildDirectionlessTraversalPatternRoot(traversalStep)
	}

	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	if traversalStep.LeftNodeBound {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{partFrame.Previous.Binding.Identifier},
			},
			Joins: []pgsql.Join{{
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
					Binding: models.ValueOptional(traversalStep.Edge.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: traversalStep.LeftNodeJoinCondition,
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: traversalStep.RightNodeJoinCondition,
				},
			}},
		})
	} else {
		if previousFrame, hasPrevious := s.previousValidFrame(traversalStep.Frame); hasPrevious {
			nextSelect.From = append(nextSelect.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{previousFrame.Binding.Identifier},
				},
			})
		}

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
					Constraint: traversalStep.LeftNodeJoinCondition,
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: traversalStep.RightNodeJoinCondition,
				},
			}},
		})
	}

	// Append all constraints to the where clause
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.LeftNodeConstraints, nextSelect.Where)
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.EdgeConstraints.Expression, nextSelect.Where)
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.RightNodeConstraints, nextSelect.Where)

	return pgsql.Query{
		Body: nextSelect,
	}, nil
}

func (s *Translator) buildTraversalPatternStep(partFrame *Frame, traversalStep *TraversalStep) (pgsql.Query, error) {
	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	if partFrame.Previous != nil {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{partFrame.Previous.Binding.Identifier},
			},
			Joins: []pgsql.Join{{
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
					Binding: models.ValueOptional(traversalStep.Edge.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: traversalStep.EdgeJoinCondition,
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: traversalStep.RightNodeJoinCondition,
				},
			}},
		})
	} else {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
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
					Constraint: traversalStep.RightNodeJoinCondition,
				},
			}},
		})
	}

	// Append all constraints to the where clause
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.LeftNodeConstraints, nextSelect.Where)
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.EdgeConstraints.Expression, nextSelect.Where)
	nextSelect.Where = pgsql.OptionalAnd(traversalStep.RightNodeConstraints, nextSelect.Where)

	return pgsql.Query{
		Body: nextSelect,
	}, nil
}

func (s *Translator) translateTraversalPatternPart(part *PatternPart, isolatedProjection bool) error {
	var scopeSnapshot *Scope

	if isolatedProjection {
		scopeSnapshot = s.scope.Snapshot()
	}

	for idx, traversalStep := range part.TraversalSteps {
		if traversalStepFrame, err := s.scope.PushFrame(); err != nil {
			return err
		} else {
			// Assign the new scope frame to the traversal step
			traversalStep.Frame = traversalStepFrame
		}

		if traversalStep.Expansion.Set {
			if err := s.translateTraversalPatternPartWithExpansion(idx == 0, part.AllShortestPaths || part.ShortestPath, traversalStep); err != nil {
				return err
			}
		} else {
			if err := s.translateTraversalPatternPartWithoutExpansion(idx == 0, traversalStep); err != nil {
				return err
			}
		}
	}

	if isolatedProjection {
		s.scope = scopeSnapshot
	}

	return nil
}

func (s *Translator) translateTraversalPatternPartWithoutExpansion(isFirstTraversalStep bool, traversalStep *TraversalStep) error {
	if constraints, err := consumePatternConstraints(isFirstTraversalStep, nonRecursivePattern, traversalStep, s.treeTranslator); err != nil {
		return err
	} else {
		if isFirstTraversalStep {
			hasPreviousFrame := traversalStep.Frame.Previous != nil

			if hasPreviousFrame {
				// Pull the implicitly joined result set's visibility to avoid violating SQL expectation on explicit vs
				// implicit join order
				for _, knownIdentifier := range traversalStep.Frame.Known().Slice() {
					if binding, bound := s.scope.Lookup(knownIdentifier); !bound {
						return errors.New("unknown traversal step identifier: " + knownIdentifier.String())
					} else if binding.LastProjection == traversalStep.Frame.Previous {
						traversalStep.Frame.Stash(binding.Identifier)
					}
				}
			}

			if err := RewriteFrameBindings(s.scope, constraints.LeftNode.Expression); err != nil {
				return err
			} else {
				traversalStep.LeftNodeConstraints = constraints.LeftNode.Expression
			}

			if leftNodeJoinCondition, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
				return err
			} else if err := RewriteFrameBindings(s.scope, leftNodeJoinCondition); err != nil {
				return err
			} else {
				traversalStep.LeftNodeJoinCondition = leftNodeJoinCondition
			}

			if hasPreviousFrame {
				traversalStep.Frame.RestoreStashed()
			}
		}

		traversalStep.Frame.Export(traversalStep.Edge.Identifier)

		if edgeJoinCondition, err := rightEdgeConstraint(traversalStep); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.scope, edgeJoinCondition); err != nil {
			return err
		} else {
			traversalStep.EdgeJoinCondition = edgeJoinCondition
		}

		if err := RewriteFrameBindings(s.scope, constraints.Edge.Expression); err != nil {
			return err
		} else {
			traversalStep.EdgeConstraints = constraints.Edge
		}

		traversalStep.Frame.Export(traversalStep.RightNode.Identifier)

		if err := RewriteFrameBindings(s.scope, constraints.RightNode.Expression); err != nil {
			return err
		} else {
			traversalStep.RightNodeConstraints = constraints.RightNode.Expression
		}

		if rightNodeJoinCondition, err := rightNodeTraversalStepJoinCondition(traversalStep); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.scope, rightNodeJoinCondition); err != nil {
			return err
		} else {
			traversalStep.RightNodeJoinCondition = rightNodeJoinCondition
		}
	}

	if boundProjections, err := buildVisibleProjections(s.scope); err != nil {
		return err
	} else {
		// Zip through all projected identifiers and update their last projected frame
		for _, binding := range boundProjections.Bindings {
			binding.MaterializedBy(traversalStep.Frame)
		}

		traversalStep.Projection = boundProjections.Items
	}

	return nil
}
