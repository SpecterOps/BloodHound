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
	"github.com/specterops/bloodhound/cypher/models"
	cypher "github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"strings"
)

func (s *Translator) translateRelationshipPattern(scope *Scope, relationshipPattern *cypher.RelationshipPattern, part *PatternPart) error {
	if bindingResult, err := s.bindPatternExpression(scope, relationshipPattern, pgsql.EdgeComposite); err != nil {
		return err
	} else {
		if err := s.translateRelationshipPatternToStep(scope, bindingResult, part, relationshipPattern); err != nil {
			return err
		}

		if len(s.properties) > 0 {
			var propertyConstraints pgsql.Expression

			for key, value := range s.properties {
				propertyConstraints = pgsql.OptionalAnd(propertyConstraints, pgsql.NewBinaryExpression(
					pgsql.NewPropertyLookup(pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnProperties}, pgsql.NewLiteral(key, pgsql.Text)),
					pgsql.OperatorEquals,
					value,
				))
			}

			if err := part.Constraints.Constrain(pgsql.AsIdentifierSet(bindingResult.Binding.Identifier), propertyConstraints); err != nil {
				return err
			}
		}

		// Capture the kind matchers for this relationship pattern
		if len(relationshipPattern.Kinds) > 0 {
			if kindIDs, missingKinds := s.kindMapper.MapKinds(relationshipPattern.Kinds); len(missingKinds) > 0 {
				s.SetErrorf("unable to map kinds: %s", strings.Join(missingKinds.Strings(), ", "))
			} else if kindIDsLiteral, err := pgsql.AsLiteral(kindIDs); err != nil {
				s.SetError(err)
			} else {
				var (
					dependencies = pgsql.AsIdentifierSet(bindingResult.Binding.Identifier)
					expression   = pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnKindID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(kindIDsLiteral),
					)
				)

				if err := part.Constraints.Constrain(dependencies, expression); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Translator) translateRelationshipPatternToStep(scope *Scope, bindingResult BindingResult, part *PatternPart, relationshipPattern *cypher.RelationshipPattern) error {
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

			nextStep.LeftNode = currentStep.RightNode
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

		if expansionScopeBinding, err := scope.DefineNew(pgsql.ExpansionPattern); err != nil {
			return err
		} else {
			// Link the edge to the expansion
			expansionScopeBinding.Link(bindingResult.Binding)

			if !isContinuation {
				// If this isn't a continuation then the left node was defined in isolation from the preceding node
				// pattern. Retype the left node to an expansion root node and link it to the expansion
				currentStep.LeftNode.DataType = pgsql.ExpansionRootNode
				expansionScopeBinding.Link(currentStep.LeftNode)
			}

			expansion = models.ValueOptional(Expansion{
				Binding:  expansionScopeBinding,
				MinDepth: models.PointerOptional(relationshipPattern.Range.StartIndex),
				MaxDepth: models.PointerOptional(relationshipPattern.Range.EndIndex),
			})

			if expansionPathBinding, err := scope.DefineNew(pgsql.ExpansionPath); err != nil {
				return err
			} else {
				// Link the path array to the expansion that declares it
				expansionPathBinding.Link(expansion.Value.Binding)

				// Set the path binding in the expansion struct for easier referencing upstream
				expansion.Value.PathBinding = expansionPathBinding

				if part.PatternBinding.Set {
					// If there's a bound pattern track this expansion's path as a dependency of the
					// pattern identifier
					part.PatternBinding.Value.DependOn(expansionPathBinding)
				}
			}
		}
	} else if part.PatternBinding.Set {
		// If there's a bound pattern track this edge as a dependency of the pattern identifier
		part.PatternBinding.Value.DependOn(bindingResult.Binding)
	}

	if isContinuation {
		// This is a traversal continuation so copy the right node identifier of the preceding step and then
		// add the new step
		nextStep := &PatternSegment{
			Edge:        bindingResult.Binding,
			Direction:   relationshipPattern.Direction,
			LeftNode:    currentStep.RightNode,
			Definitions: []*BoundIdentifier{bindingResult.Binding},
			Expansion:   expansion,
		}

		if expansion.Set {
			nextStep.Definitions = append(nextStep.Definitions, expansion.Value.PathBinding)
		}

		part.TraversalSteps = append(part.TraversalSteps, nextStep)

		// The edge needs a constraint that ties it to the preceding edge
		if edgeConstraint, err := rightEdgeConstraint(currentStep, bindingResult.Binding.Identifier, nextStep.Direction); err != nil {
			return err
		} else if err := s.treeTranslator.ConstrainIdentifier(nextStep.Edge.Identifier, edgeConstraint); err != nil {
			return err
		}
	} else {
		// Carry over the left node identifier if the edge identifier for the preceding step isn't set
		currentStep.Edge = bindingResult.Binding
		currentStep.Direction = relationshipPattern.Direction
		currentStep.Expansion = expansion

		if expansion.Set {
			currentStep.Definitions = append(currentStep.Definitions, bindingResult.Binding, expansion.Value.PathBinding)
		} else {
			currentStep.Definitions = append(currentStep.Definitions, bindingResult.Binding)
		}
	}

	return nil
}
