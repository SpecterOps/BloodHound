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
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) buildInlineProjection(part *QueryPart) (pgsql.Select, error) {
	sqlSelect := pgsql.Select{
		Where: part.projections.Constraints,
	}

	// If there's a projection frame set, some additional negotiation is required to identify which frame the
	// from-statement should be written to. Some of this would be better figured out during the translation
	// of the projection where query scope and other components are not yet fully translated.
	if part.projections.Frame != nil {
		// Look up to see if there are CTE expressions registered. If there are then it is likely
		// there was a projection between this CTE and the previous multipart query part
		hasCTEs := part.Model.CommonTableExpressions != nil && len(part.Model.CommonTableExpressions.Expressions) > 0

		if part.Frame.Previous == nil || hasCTEs {
			sqlSelect.From = []pgsql.FromClause{{
				Source: part.projections.Frame.Binding.Identifier,
			}}
		} else {
			sqlSelect.From = []pgsql.FromClause{{
				Source: part.Frame.Previous.Binding.Identifier,
			}}
		}
	}

	for _, projection := range part.projections.Items {
		builtProjection := projection.SelectItem

		if projection.Alias.Set {
			builtProjection = &pgsql.AliasedExpression{
				Expression: builtProjection,
				Alias:      projection.Alias,
			}
		}

		sqlSelect.Projection = append(sqlSelect.Projection, builtProjection)
	}

	if len(part.projections.GroupBy) > 0 {
		for _, groupBy := range part.projections.GroupBy {
			sqlSelect.GroupBy = append(sqlSelect.GroupBy, groupBy)
		}
	}

	return sqlSelect, nil
}

func (s *Translator) buildTailProjection() error {
	var (
		currentPart           = s.query.CurrentPart()
		currentFrame          = s.query.Scope.CurrentFrame()
		singlePartQuerySelect = pgsql.Select{}
	)

	singlePartQuerySelect.From = []pgsql.FromClause{{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{currentFrame.Binding.Identifier},
		},
	}}

	if projectionConstraint, err := s.treeTranslator.ConsumeAll(); err != nil {
		return err
	} else if projection, err := buildExternalProjection(s.query.Scope, currentPart.projections.Items); err != nil {
		return err
	} else if err := RewriteFrameBindings(s.query.Scope, projectionConstraint.Expression); err != nil {
		return err
	} else {
		singlePartQuerySelect.Projection = projection
		singlePartQuerySelect.Where = projectionConstraint.Expression
	}

	currentPart.Model.Body = singlePartQuerySelect

	if currentPart.Skip.Set {
		currentPart.Model.Offset = currentPart.Skip
	}

	if currentPart.Limit.Set {
		currentPart.Model.Limit = currentPart.Limit
	}

	if len(currentPart.OrderBy) > 0 {
		currentPart.Model.OrderBy = currentPart.OrderBy
	}

	return nil
}

func (s *Translator) translateMatch() error {
	currentQueryPart := s.query.CurrentPart()

	for _, part := range currentQueryPart.ConsumeCurrentPattern().Parts {
		if !part.IsTraversal {
			if err := s.translateNonTraversalPatternPart(part); err != nil {
				return err
			}
		} else {
			if err := s.translateTraversalPatternPart(part, false); err != nil {
				return err
			}
		}

		// Render this pattern part in the current query part
		if err := s.buildPatternPart(part); err != nil {
			return err
		}

		// Declare the pattern variable in scope if set
		if part.PatternBinding.Set {
			s.query.Scope.Declare(part.PatternBinding.Value.Identifier)
		}
	}

	return s.buildPatternPredicates()
}

func (s *Translator) translateTraversalPatternPart(part *PatternPart, isolatedProjection bool) error {
	var scopeSnapshot *Scope

	if isolatedProjection {
		scopeSnapshot = s.query.Scope.Snapshot()
	}

	for idx, traversalStep := range part.TraversalSteps {
		if traversalStepFrame, err := s.query.Scope.PushFrame(); err != nil {
			return err
		} else {
			// Assign the new scope frame to the traversal step
			traversalStep.Frame = traversalStepFrame
		}

		if traversalStep.Expansion.Set {
			if err := s.translateTraversalPatternPartWithExpansion(idx == 0, traversalStep); err != nil {
				return err
			}
		} else {
			if err := s.translateTraversalPatternPartWithoutExpansion(idx == 0, traversalStep); err != nil {
				return err
			}
		}
	}

	if isolatedProjection {
		s.query.Scope = scopeSnapshot
	}

	return nil
}

func (s *Translator) translateTraversalPatternPartWithoutExpansion(isFirstTraversalStep bool, traversalStep *PatternSegment) error {
	if constraints, err := consumePatternConstraints(isFirstTraversalStep, nonRecursivePattern, traversalStep, s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else {
		if isFirstTraversalStep {
			hasPreviousFrame := traversalStep.Frame.Previous != nil

			if hasPreviousFrame {
				// Pull the implicitly joined result set's visibility to avoid violating SQL expectation on explicit vs
				// implicit join order
				for _, knownIdentifier := range traversalStep.Frame.Known().Slice() {
					if binding, bound := s.query.Scope.Lookup(knownIdentifier); !bound {
						return errors.New("unknown traversal step identifier: " + knownIdentifier.String())
					} else if binding.LastProjection == traversalStep.Frame.Previous {
						traversalStep.Frame.Stash(binding.Identifier)
					}
				}
			}

			//
			if err := RewriteFrameBindings(s.query.Scope, constraints.LeftNode.Expression); err != nil {
				return err
			} else {
				traversalStep.LeftNodeConstraints = constraints.LeftNode.Expression
			}

			if leftNodeJoinCondition, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
				return err
			} else if err := RewriteFrameBindings(s.query.Scope, leftNodeJoinCondition); err != nil {
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
		} else if err := RewriteFrameBindings(s.query.Scope, edgeJoinCondition); err != nil {
			return err
		} else {
			traversalStep.EdgeJoinCondition = edgeJoinCondition
		}

		if err := RewriteFrameBindings(s.query.Scope, constraints.Edge.Expression); err != nil {
			return err
		} else {
			traversalStep.EdgeConstraints = constraints.Edge
		}

		traversalStep.Frame.Export(traversalStep.RightNode.Identifier)

		if err := RewriteFrameBindings(s.query.Scope, constraints.RightNode.Expression); err != nil {
			return err
		} else {
			traversalStep.RightNodeConstraints = constraints.RightNode.Expression
		}

		if rightNodeJoinCondition, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.query.Scope, rightNodeJoinCondition); err != nil {
			return err
		} else {
			traversalStep.RightNodeJoinCondition = rightNodeJoinCondition
		}
	}

	if boundProjections, err := buildVisibleProjections(s.query.Scope); err != nil {
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

func (s *Translator) translateTraversalPatternPartWithExpansion(isFirstTraversalStep bool, traversalStep *PatternSegment) error {
	if constraints, err := consumePatternConstraints(isFirstTraversalStep, recursivePattern, traversalStep, s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else {
		// If one side of the expansion has constraints but the other does not this may be an opportunity to reorder the traversal
		// to start with tighter search bounds
		constraints.OptimizePatternConstraintBalance(traversalStep)

		if isFirstTraversalStep {
			if err := RewriteFrameBindings(s.query.Scope, constraints.LeftNode.Expression); err != nil {
				return err
			}

			traversalStep.Expansion.Value.PrimerConstraints = pgsql.OptionalAnd(traversalStep.Expansion.Value.PrimerConstraints, constraints.LeftNode.Expression)

			if leftNodeJoinCondition, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
				return err
			} else if err := RewriteFrameBindings(s.query.Scope, leftNodeJoinCondition); err != nil {
				return err
			} else {
				traversalStep.Expansion.Value.LeftNodeJoinCondition = leftNodeJoinCondition
			}
		}

		if expansionNodeConstraints, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
			return err
		} else {
			traversalStep.Expansion.Value.ExpansionNodeConstraints = expansionNodeConstraints
		}

		if err := RewriteFrameBindings(s.query.Scope, constraints.Edge.Expression); err != nil {
			return err
		}

		if isFirstTraversalStep {
			traversalStep.Expansion.Value.PrimerConstraints = pgsql.OptionalAnd(traversalStep.Expansion.Value.PrimerConstraints, constraints.Edge.Expression)
		}

		traversalStep.Expansion.Value.ExpansionEdgeConstraints = constraints.Edge.Expression

		if constraints.RightNode.Expression != nil {
			if err := RewriteFrameBindings(s.query.Scope, constraints.RightNode.Expression); err != nil {
				return err
			} else {
				traversalStep.Expansion.Value.TerminalNodeConstraints = constraints.RightNode.Expression
			}
		}
	}

	// Export the path from the traversal's scope
	traversalStep.Frame.Export(traversalStep.Expansion.Value.PathBinding.Identifier)

	// Push a new frame that contains currently projected scope from the expansion recursive CTE
	if expansionFrame, err := s.query.Scope.PushFrame(); err != nil {
		return err
	} else {
		traversalStep.Expansion.Value.Frame = expansionFrame
		traversalStep.Expansion.Value.RecursiveConstraints = pgsql.OptionalAnd(traversalStep.Expansion.Value.ExpansionEdgeConstraints, expansionConstraints(expansionFrame.Binding.Identifier, traversalStep.Expansion.Value.MinDepth, traversalStep.Expansion.Value.MaxDepth))

		// Remove the previous projections of the root and terminal node to reproject them after expansion
		traversalStep.LeftNode.Dematerialize()
		traversalStep.RightNode.Dematerialize()

		if boundProjections, err := buildVisibleProjections(s.query.Scope); err != nil {
			return err
		} else {
			// Zip through all projected identifiers and update their last projected frame
			for _, binding := range boundProjections.Bindings {
				binding.MaterializedBy(expansionFrame)
			}

			traversalStep.Expansion.Value.Projection = boundProjections.Items
		}

		if err := s.query.Scope.PopFrame(); err != nil {
			return err
		}
	}

	if boundProjections, err := buildVisibleProjections(s.query.Scope); err != nil {
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

func (s *Translator) translateNonTraversalPatternPart(part *PatternPart) error {
	if nextFrame, err := s.query.Scope.PushFrame(); err != nil {
		return err
	} else {
		part.NodeSelect.Frame = nextFrame

		nextFrame.Export(part.NodeSelect.Binding.Identifier)

		if constraint, err := s.treeTranslator.IdentifierConstraints.ConsumeSet(nextFrame.Known()); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.query.Scope, constraint.Expression); err != nil {
			return err
		} else {
			part.NodeSelect.Constraint = constraint
		}

		if boundProjections, err := buildVisibleProjections(s.query.Scope); err != nil {
			return err
		} else {
			// Zip through all projected identifiers and update their last projected frame
			for _, binding := range boundProjections.Bindings {
				binding.MaterializedBy(nextFrame)
			}

			part.NodeSelect.Select.Projection = boundProjections.Items
		}
	}

	return nil
}
