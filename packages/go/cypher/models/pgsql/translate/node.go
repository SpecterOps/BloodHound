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
	"fmt"

	"github.com/specterops/bloodhound/cypher/models"
	cypher "github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) translateNodePattern(nodePattern *cypher.NodePattern) error {
	var (
		queryPart   = s.query.CurrentPart()
		patternPart = queryPart.pattern.CurrentPart()
	)

	if bindingResult, err := s.bindPatternExpression(nodePattern, pgsql.NodeComposite); err != nil {
		return err
	} else if err := s.translateNodePatternToStep(patternPart, bindingResult); err != nil {
		return err
	} else {
		if len(queryPart.properties) > 0 {
			var propertyConstraints pgsql.Expression

			for key, value := range queryPart.properties {
				s.treeTranslator.Push(pgsql.NewPropertyLookup(pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnProperties}, pgsql.NewLiteral(key, pgsql.Text)))
				s.treeTranslator.Push(value)

				if newConstraint, err := s.treeTranslator.PopBinaryExpression(pgsql.OperatorEquals); err != nil {
					return err
				} else {
					propertyConstraints = pgsql.OptionalAnd(propertyConstraints, newConstraint)
				}
			}

			if err := patternPart.Constraints.Constrain(pgsql.AsIdentifierSet(bindingResult.Binding.Identifier), propertyConstraints); err != nil {
				return err
			}
		}

		if len(nodePattern.Kinds) > 0 {
			if kindIDs, err := s.kindMapper.MapKinds(s.ctx, nodePattern.Kinds); err != nil {
				s.SetError(fmt.Errorf("failed to translate kinds: %w", err))
			} else if kindIDsLiteral, err := pgsql.AsLiteral(kindIDs); err != nil {
				s.SetError(err)
			} else {
				expression := pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnKindIDs},
					pgsql.OperatorPGArrayOverlap,
					kindIDsLiteral,
				)

				if err := patternPart.Constraints.Constrain(pgsql.AsIdentifierSet(bindingResult.Binding.Identifier), expression); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Translator) translateNodePatternSegment(part *PatternPart, bindingResult BindingResult) error {
	// If this isn't a traversal
	if nodeFrame, err := s.query.Scope.PushFrame(); err != nil {
		return err
	} else {
		part.NodeSelect.Frame = nodeFrame
	}

	// Track if the node select is introducing a new variable into scope
	if bindingResult.AlreadyBound {
		part.NodeSelect.IsDefinition = false

		if boundProjections, err := buildVisibleScopeProjections(s.query.Scope, nil); err != nil {
			return err
		} else {
			part.NodeSelect.Select.Projection = boundProjections.Items
		}
	} else {
		part.NodeSelect.IsDefinition = true

		if boundProjections, err := buildVisibleScopeProjections(s.query.Scope, []*BoundIdentifier{bindingResult.Binding}); err != nil {
			return err
		} else {
			part.NodeSelect.Select.Projection = boundProjections.Items
		}
	}

	// Make this the node select of the pattern part
	part.NodeSelect.Binding = bindingResult.Binding
	return nil
}

func (s *Translator) translateNodePatternSegmentWithTraversal(currentSegment *PatternSegment) error {
	// Note: The order below matters as it will change the order of projections in the resulting translation

	// If either of the nodes are not bound at this point then this traversal must materialize them
	if !currentSegment.LeftNodeBound {
		currentSegment.Definitions = append(currentSegment.Definitions, currentSegment.LeftNode)
	}

	// Add the edge symbol as part of the definitions that are being materialized by this expansion
	currentSegment.Definitions = append(currentSegment.Definitions, currentSegment.Edge)

	// If there's an expansion attached to this traversal, ensure that the symbol for the expansion's scope frame
	// is part of the definitions that are being materialized by this expansion
	if currentSegment.Expansion.Set {
		currentSegment.Definitions = append(currentSegment.Definitions, currentSegment.Expansion.Value.PathBinding)
	}

	// If the right node has not been bound before add it to the pattern part's list of new definitions
	if !currentSegment.RightNodeBound {
		currentSegment.Definitions = append(currentSegment.Definitions, currentSegment.RightNode)
	}

	if currentSegment.Expansion.Set {
		// This segment is part of a variable expansion

		if expansionFrame, err := s.query.Scope.PushFrame(); err != nil {
			return err
		} else {
			if currentSegment.LeftNodeBound {
				// If the left node is bound, this expansion will rebind it
				currentSegment.LeftNode.DetachFromFrame()
			}

			// Assign the new scope frame to the expansion
			currentSegment.Expansion.Value.Frame = expansionFrame
		}

		if boundProjections, err := buildVisibleScopeProjections(s.query.Scope, currentSegment.Definitions); err != nil {
			return err
		} else {
			currentSegment.Expansion.Value.Projection = boundProjections.Items
		}

		// Update the data type of the right node so that it reflects that it is now the terminal node of
		// an expansion
		currentSegment.RightNode.DataType = pgsql.ExpansionTerminalNode

		// If the edge is an expansion link the node as the right terminal node to the expansion
		if expansionBinding, found := currentSegment.Edge.FirstDependencyByType(pgsql.ExpansionPattern); !found {
			return errors.New("unable to find expansion context")
		} else {
			currentSegment.RightNode.Link(expansionBinding)
		}
	} else {
		if stepFrame, err := s.query.Scope.PushFrame(); err != nil {
			return err
		} else {
			currentSegment.Frame = stepFrame
		}

		if boundProjections, err := buildVisibleScopeProjections(s.query.Scope, currentSegment.Definitions); err != nil {
			return err
		} else {
			currentSegment.Projection = boundProjections.Items
		}
	}

	return nil
}

func (s *Translator) translateNodePatternToStep(part *PatternPart, bindingResult BindingResult) error {
	if part.IsTraversal {
		if numSteps := len(part.TraversalSteps); numSteps == 0 {
			// This is the traversal step's left node
			part.TraversalSteps = append(part.TraversalSteps, &PatternSegment{
				LeftNode:      bindingResult.Binding,
				LeftNodeBound: bindingResult.AlreadyBound,
			})
		} else {
			currentStep := part.TraversalSteps[numSteps-1]

			// Set the right node pattern identifier
			currentStep.RightNode = bindingResult.Binding
			currentStep.RightNodeBound = bindingResult.AlreadyBound

			// Finish setting up this traversal step
			return s.translateNodePatternSegmentWithTraversal(currentStep)
		}
	} else {
		return s.translateNodePatternSegment(part, bindingResult)
	}

	return nil
}

func consumeConstraintsFrom(visible *pgsql.IdentifierSet, trackers ...*ConstraintTracker) (*Constraint, error) {
	constraint := &Constraint{
		Dependencies: pgsql.NewIdentifierSet(),
	}

	for _, constraintTracker := range trackers {
		if trackedConstraint, err := constraintTracker.ConsumeSet(visible); err != nil {
			return nil, err
		} else if err := constraint.Merge(trackedConstraint); err != nil {
			return nil, err
		}
	}

	return constraint, nil
}

func (s *Translator) buildNodePattern(part *PatternPart) error {
	var (
		nextSelect pgsql.Select
	)

	if part.NodeSelect.Frame.Previous != nil && (s.query.CurrentPart().frame == nil || part.NodeSelect.Frame.Previous.Binding.Identifier != s.query.CurrentPart().frame.Binding.Identifier) {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{part.NodeSelect.Frame.Previous.Binding.Identifier},
			},
		})
	}

	nextSelect.Projection = part.NodeSelect.Select.Projection

	if constraints, err := consumeConstraintsFrom(part.NodeSelect.Frame.Visible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return err
	} else {
		if err := rewriteConstraintIdentifierReferences(part.NodeSelect.Frame, []*Constraint{constraints}); err != nil {
			return err
		}

		nextSelect.Where = constraints.Expression

		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(part.NodeSelect.Binding.Identifier),
			},
		})

		// Prepare the next select statement
		s.query.CurrentPart().Model.AddCTE(pgsql.CommonTableExpression{
			Alias: pgsql.TableAlias{
				Name: part.NodeSelect.Frame.Binding.Identifier,
			},
			Query: pgsql.Query{
				Body: nextSelect,
			},
		})
	}

	return nil
}
