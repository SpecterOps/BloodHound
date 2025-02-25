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
	var sqlSelect pgsql.Select

	if part.projections.Frame != nil {
		sqlSelect.From = []pgsql.FromClause{{
			Source: part.projections.Frame.Binding.Identifier,
		}}
	}

	if projectionConstraint, err := s.treeTranslator.ConsumeAll(); err != nil {
		return sqlSelect, err
	} else {
		sqlSelect.Where = projectionConstraint.Expression
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
	if isFirstTraversalStep {
		// If this is the first traversal step, visit the left-hand node
		traversalStep.Frame.Export(traversalStep.LeftNode.Identifier)

		hasPreviousFrame := traversalStep.Frame.Previous != nil

		if hasPreviousFrame {
			// Pull the implicitly joined result set's visibility to avoid violating SQL expectation on explicit vs
			// implicit join order
			for _, knownIdentifier := range traversalStep.Frame.Known().Slice() {
				if binding, bound := s.query.Scope.Lookup(knownIdentifier); !bound {
					return errors.New("unknown traversal step identifier: " + knownIdentifier.String())
				} else if binding.LastProjection == traversalStep.Frame.Previous {
					traversalStep.Frame.Veil(binding.Identifier)
				}
			}
		}

		if leftNodeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Known(), s.treeTranslator.IdentifierConstraints); err != nil {
			return err
		} else if leftNodeJoinConstraint, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
			return err
		} else if leftNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{leftNodeConstraints.Expression, leftNodeJoinConstraint}); err != nil {
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

	if edgeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Known(), s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else if err := RewriteFrameBindings(s.query.Scope, edgeConstraints.Expression); err != nil {
		return err
	} else {
		traversalStep.EdgeConstraint = edgeConstraints
	}

	traversalStep.Frame.Export(traversalStep.RightNode.Identifier)

	if rightNodeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Known(), s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else if rightNodeJoinConstraint, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
		return err
	} else if rightNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rightNodeConstraints.Expression, rightNodeJoinConstraint}); err != nil {
		return err
	} else if err := RewriteFrameBindings(s.query.Scope, rightNodeJoinCondition); err != nil {
		return err
	} else {
		traversalStep.RightNodeJoinCondition = rightNodeJoinCondition
	}

	if boundProjections, err := buildVisibleScopeProjections(s.query.Scope); err != nil {
		return err
	} else {
		// Zip through all projected identifiers and update their last projected frame
		for _, binding := range boundProjections.Bindings {
			binding.LastProjection = traversalStep.Frame
		}

		traversalStep.Projection = boundProjections.Items
	}

	return nil
}

func (s *Translator) translateTraversalPatternPartWithExpansion(isFirstTraversalStep bool, traversalStep *PatternSegment) error {
	if isFirstTraversalStep {
		traversalStep.Frame.Export(traversalStep.LeftNode.Identifier)

		if rootNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(traversalStep.LeftNode.Identifier), s.treeTranslator.IdentifierConstraints); err != nil {
			return err
		} else if rootNodeJoinConstraints, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
			return err
		} else if rootNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rootNodeConstraints.Expression, rootNodeJoinConstraints}); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.query.Scope, rootNodeJoinCondition); err != nil {
			return err
		} else {
			traversalStep.Expansion.Value.PrimerRootNodeConstraints = rootNodeJoinCondition
		}
	}

	if expansionNodeConstraints, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
		return err
	} else {
		traversalStep.Expansion.Value.ExpansionNodeConstraints = expansionNodeConstraints
	}

	// The exclusions below are done at this step in the process since the recursive descent portion of the query no longer has
	// a reference to the root node and any dependent interaction between the root and terminal nodes would require an
	// additional join. By not consuming the remaining constraints for the root and terminal nodes, they become visible up
	// in the outer select of the recursive CTE.
	traversalStep.Frame.Export(traversalStep.Edge.Identifier)

	if edgeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Known().Remove(traversalStep.LeftNode.Identifier), s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else if err := RewriteFrameBindings(s.query.Scope, edgeConstraints.Expression); err != nil {
		return err
	} else if isFirstTraversalStep {
		traversalStep.Expansion.Value.PrimerConstraints = edgeConstraints.Expression
	} else {
		traversalStep.Expansion.Value.ExpansionEdgeConstraints = edgeConstraints.Expression
	}

	traversalStep.Frame.Export(traversalStep.RightNode.Identifier)

	if terminalNodeConstraints, err := consumeConstraintsFrom(traversalStep.Frame.Known().Remove(traversalStep.LeftNode.Identifier), s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else if terminalNodeConstraints.Expression != nil {
		if err := RewriteFrameBindings(s.query.Scope, terminalNodeConstraints.Expression); err != nil {
			return err
		} else {
			traversalStep.Expansion.Value.TerminalNodeConstraints = terminalNodeConstraints.Expression
		}
	}

	// Export the path from the traversal's scope
	traversalStep.Frame.Export(traversalStep.Expansion.Value.PathBinding.Identifier)

	// Push a new frame that contains currently projected scope from the expansion recursive CTE
	if expansionFrame, err := s.query.Scope.PushFrame(); err != nil {
		return err
	} else {
		traversalStep.Expansion.Value.Frame = expansionFrame
		traversalStep.Expansion.Value.RecursiveConstraints = pgsql.OptionalAnd(
			traversalStep.Expansion.Value.PrimerConstraints,
			expansionConstraints(expansionFrame.Binding.Identifier),
		)

		if boundProjections, err := buildVisibleScopeProjections(s.query.Scope); err != nil {
			return err
		} else {
			// Zip through all projected identifiers and update their last projected frame
			for _, binding := range boundProjections.Bindings {
				binding.LastProjection = expansionFrame
			}

			traversalStep.Expansion.Value.Projection = boundProjections.Items
		}

		if err := s.query.Scope.PopFrame(); err != nil {
			return err
		}
	}

	if boundProjections, err := buildVisibleScopeProjections(s.query.Scope); err != nil {
		return err
	} else {
		// Zip through all projected identifiers and update their last projected frame
		for _, binding := range boundProjections.Bindings {
			binding.LastProjection = traversalStep.Frame
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

		if constraint, err := consumeConstraintsFrom(nextFrame.Known(), s.treeTranslator.IdentifierConstraints); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.query.Scope, constraint.Expression); err != nil {
			return err
		} else {
			part.NodeSelect.Constraint = constraint
		}

		if boundProjections, err := buildVisibleScopeProjections(s.query.Scope); err != nil {
			return err
		} else {
			// Zip through all projected identifiers and update their last projected frame
			for _, binding := range boundProjections.Bindings {
				binding.LastProjection = nextFrame
			}

			part.NodeSelect.Select.Projection = boundProjections.Items
		}
	}

	return nil
}
