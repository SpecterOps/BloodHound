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
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) translateRelationshipPattern(relationshipPattern *cypher.RelationshipPattern) error {
	var (
		currentQueryPart = s.query.CurrentPart()
		patternPart      = currentQueryPart.currentPattern.CurrentPart()
	)

	if bindingResult, err := s.bindPatternExpression(relationshipPattern, pgsql.EdgeComposite); err != nil {
		return err
	} else {
		if err := s.translateRelationshipPatternToStep(bindingResult, patternPart, relationshipPattern); err != nil {
			return err
		}

		if currentQueryPart.HasProperties() {
			var propertyConstraints pgsql.Expression

			for key, value := range currentQueryPart.ConsumeProperties() {
				s.treeTranslator.Push(pgsql.NewPropertyLookup(pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnProperties}, pgsql.NewLiteral(key, pgsql.Text)))
				s.treeTranslator.Push(value)

				if newConstraint, err := s.treeTranslator.PopBinaryExpression(pgsql.OperatorEquals); err != nil {
					return err
				} else {
					propertyConstraints = pgsql.OptionalAnd(propertyConstraints, newConstraint)
				}
			}

			if err := s.treeTranslator.Constrain(pgsql.NewIdentifierSet().Add(bindingResult.Binding.Identifier), propertyConstraints); err != nil {
				return err
			}
		}

		// Capture the kind matchers for this relationship pattern
		if len(relationshipPattern.Kinds) > 0 {
			if kindIDs, err := s.kindMapper.MapKinds(s.ctx, relationshipPattern.Kinds); err != nil {
				return fmt.Errorf("failed to translate kinds: %w", err)
			} else if kindIDsLiteral, err := pgsql.AsLiteral(kindIDs); err != nil {
				return err
			} else if err := s.treeTranslator.Constrain(pgsql.NewIdentifierSet().Add(bindingResult.Binding.Identifier), pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnKindID},
				pgsql.OperatorEquals,
				pgsql.NewAnyExpression(kindIDsLiteral),
			)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Translator) translateRelationshipPatternToStep(bindingResult BindingResult, part *PatternPart, relationshipPattern *cypher.RelationshipPattern) error {
	var (
		expansion      models.Optional[Expansion]
		numSteps       = len(part.TraversalSteps)
		currentStep    = part.TraversalSteps[numSteps-1]
		isContinuation = currentStep.Edge != nil
	)

	if bindingResult.AlreadyBound {
		if isContinuation {
			// This is a traversal continuation so copy the right node identifier of the preceding step and then
			// add the new step
			nextStep := &PatternSegment{
				Edge:      bindingResult.Binding,
				Direction: relationshipPattern.Direction,
			}

			// Mark the left node as already bound as it's part of the previous step's continuation
			nextStep.LeftNode = currentStep.RightNode
			nextStep.LeftNodeBound = true

			part.TraversalSteps = append(part.TraversalSteps, nextStep)
		} else {
			// Carry over the left node identifier if the edge identifier for the preceding step isn't set
			currentStep.Edge = bindingResult.Binding
			currentStep.Direction = relationshipPattern.Direction
		}

		return nil
	}

	// Look for any relationship pattern ranges. These indicate some kind of variable expansion of the path pattern.
	if relationshipPattern.Range != nil {
		// Set the edge type to an expansion of edges
		bindingResult.Binding.DataType = pgsql.ExpansionEdge

		if !isContinuation {
			// If this isn't a continuation then the left node was defined in isolation from the preceding node
			// pattern. Retype the left node to an expansion root node and link it to the expansion
			currentStep.LeftNode.DataType = pgsql.ExpansionRootNode
		}

		expansion = models.ValueOptional(Expansion{
			MinDepth: models.PointerOptional(relationshipPattern.Range.StartIndex),
			MaxDepth: models.PointerOptional(relationshipPattern.Range.EndIndex),
		})

		if expansionPathBinding, err := s.scope.DefineNew(pgsql.ExpansionPath); err != nil {
			return err
		} else {
			// Set the path binding in the expansion struct for easier referencing upstream
			expansion.Value.PathBinding = expansionPathBinding

			if part.PatternBinding.Set {
				// If there's a bound pattern track this expansion's path as a dependency of the
				// pattern identifier
				part.PatternBinding.Value.DependOn(expansionPathBinding)
			}
		}
	} else if part.PatternBinding.Set {
		// If there's a bound pattern track this edge as a dependency of the pattern identifier
		part.PatternBinding.Value.DependOn(bindingResult.Binding)
	}

	if isContinuation {
		// This is a traversal continuation so copy the right node identifier of the preceding step and then
		// add the new step
		part.TraversalSteps = append(part.TraversalSteps, &PatternSegment{
			Edge:          bindingResult.Binding,
			Direction:     relationshipPattern.Direction,
			LeftNode:      currentStep.RightNode,
			LeftNodeBound: true,
			Expansion:     expansion,
		})
	} else {
		// Carry over the left node identifier if the edge identifier for the preceding step isn't set
		currentStep.Edge = bindingResult.Binding
		currentStep.Direction = relationshipPattern.Direction
		currentStep.Expansion = expansion
	}

	return nil
}
