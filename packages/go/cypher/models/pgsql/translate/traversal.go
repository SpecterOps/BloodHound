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

func (s *Translator) buildDirectionalTraversalPatternRoot(leftNodeConstraints, edgeConstraints, rightNodeConstraints pgsql.Expression, traversalStep *PatternSegment, direction graph.Direction) (pgsql.Select, error) {
	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}
	nextSelect.Where = edgeConstraints
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
				Constraint: leftNodeConstraints,
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: rightNodeConstraints,
			},
		}},
	})
	return nextSelect, nil
}

func (s *Translator) buildDirectionlessTraversalPatternRoot(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
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
		nextSelect, err := s.buildDirectionalTraversalPatternRoot(traversalStep.LeftNodeJoinCondition, traversalStep.EdgeConstraints.Expression, traversalStep.RightNodeJoinCondition, traversalStep, traversalStep.Direction)

		return pgsql.Query{
			Body: nextSelect,
		}, err

	case graph.DirectionBoth:
		nextSelect, err := s.buildDirectionalTraversalPatternRoot(traversalStep.LeftNodeJoinCondition, traversalStep.EdgeConstraints.Expression, traversalStep.RightNodeJoinCondition, traversalStep, graph.DirectionOutbound)

		return pgsql.Query{
			Body: nextSelect,
		}, err
	}

	return pgsql.Query{}, fmt.Errorf("unsupported")
}

func (s *Translator) buildTraversalPatternRoot(partFrame *Frame, part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	if traversalStep.Direction == graph.DirectionBoth {
		return s.buildDirectionlessTraversalPatternRoot(part, traversalStep)
	}

	nextSelect := pgsql.Select{
		Projection: traversalStep.Projection,
	}

	if partFrame.Previous != nil {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{partFrame.Previous.Binding.Identifier},
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

func (s *Translator) buildTraversalPatternStep(partFrame *Frame, part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
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
