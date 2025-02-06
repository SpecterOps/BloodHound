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

func (s *Translator) translatePatternPredicate(scope *Scope) error {
	// Set the pattern frame
	s.query.CurrentPart().pattern.Frame = scope.CurrentFrame()

	// All pattern predicates must be relationship patterns
	newPatternPart := s.query.CurrentPart().pattern.NewPart()
	newPatternPart.IsTraversal = true

	return nil
}

func (s *Translator) buildOptimizedRelationshipExistPredicate(part *PatternPart, traversalStep *PatternSegment) error {
	whereClause := pgsql.NewBinaryExpression(
		pgsql.NewBinaryExpression(
			pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
			pgsql.OperatorEquals,
			pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID}),
		pgsql.OperatorOr,
		pgsql.NewBinaryExpression(
			pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			pgsql.OperatorEquals,
			pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID}),
	)

	if err := rewriteIdentifierReferences(traversalStep.Frame, []pgsql.Expression{whereClause}); err != nil {
		return err
	}

	// explain analyze select * from node n0 where not exists(select 1 from edge e0 where e0.start_id = n0.id or e0.end_id = n0.id);
	s.treeTranslator.Push(pgsql.ExistsExpression{
		Subquery: pgsql.Subquery{
			Query: pgsql.Query{
				Body: pgsql.Select{
					Projection: []pgsql.SelectItem{
						pgsql.NewLiteral(1, pgsql.Int),
					},
					From: []pgsql.FromClause{{
						Source: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
							Binding: models.ValueOptional(traversalStep.Edge.Identifier),
						}},
					},
					Where: whereClause,
				},
			},
		},
	})

	return nil
}

func (s *Translator) buildPatternPredicate() error {
	if numPatternParts := len(s.query.CurrentPart().pattern.Parts); numPatternParts < 1 || numPatternParts > 1 {
		return fmt.Errorf("expected exactly one pattern part for pattern predicate but found: %d", numPatternParts)
	}

	var (
		lastFrame *Frame

		patternPart = s.query.CurrentPart().pattern.Parts[0]
		subQuery    = pgsql.Query{
			CommonTableExpressions: &pgsql.With{},
		}
	)

	if len(patternPart.TraversalSteps) == 1 {
		var (
			traversalStep        = patternPart.TraversalSteps[0]
			hasGlobalConstraints = s.treeTranslator.IdentifierConstraints.HasConstraints(
				pgsql.AsIdentifierSet(
					traversalStep.LeftNode.Identifier,
					traversalStep.Edge.Identifier,
					traversalStep.RightNode.Identifier,
				),
			)
			hasPredicateConstraints = patternPart.Constraints.HasConstraints(
				pgsql.AsIdentifierSet(
					traversalStep.LeftNode.Identifier,
					traversalStep.Edge.Identifier,
					traversalStep.RightNode.Identifier,
				),
			)
		)

		if traversalStep.Direction == graph.DirectionBoth && !hasPredicateConstraints && !hasGlobalConstraints {
			return s.buildOptimizedRelationshipExistPredicate(patternPart, traversalStep)
		}
	}

	for idx, traversalStep := range patternPart.TraversalSteps {
		if traversalStep.Expansion.Set {
			return fmt.Errorf("expansion in pattern predicate not supported")
		}

		if idx > 0 {
			if traversalStepQuery, err := s.buildTraversalPatternStep(patternPart, traversalStep); err != nil {
				return err
			} else {
				subQuery.AddCTE(pgsql.CommonTableExpression{
					Alias: pgsql.TableAlias{
						Name: traversalStep.Frame.Binding.Identifier,
					},
					Query: traversalStepQuery,
				})
			}
		} else {
			if traversalStepQuery, err := s.buildTraversalPatternRoot(patternPart, traversalStep); err != nil {
				return err
			} else {
				subQuery.AddCTE(pgsql.CommonTableExpression{
					Alias: pgsql.TableAlias{
						Name: traversalStep.Frame.Binding.Identifier,
					},
					Query: traversalStepQuery,
				})
			}
		}

		lastFrame = traversalStep.Frame
	}

	subQuery.Body = pgsql.Select{
		Projection: pgsql.Projection{
			pgsql.NewBinaryExpression(
				pgsql.FunctionCall{
					Function:   pgsql.FunctionCount,
					Parameters: []pgsql.Expression{pgsql.WildcardIdentifier},
				},
				pgsql.OperatorGreaterThan,
				pgsql.NewLiteral(0, pgsql.Int),
			),
		},

		From: []pgsql.FromClause{{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{lastFrame.Binding.Identifier},
			},
		}},
	}

	s.treeTranslator.Push(pgsql.Subquery{
		Query: subQuery,
	})

	return nil
}
