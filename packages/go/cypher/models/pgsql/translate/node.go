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
	cypher "github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) translateNodePattern(nodePattern *cypher.NodePattern) error {
	var (
		queryPart   = s.query.CurrentPart()
		patternPart = queryPart.currentPattern.CurrentPart()
	)

	if bindingResult, err := s.bindPatternExpression(nodePattern, pgsql.NodeComposite); err != nil {
		return err
	} else if err := s.translateNodePatternToStep(nodePattern, patternPart, bindingResult); err != nil {
		return err
	}

	return nil
}

func (s *Translator) translateNodePatternSegment(nodePattern *cypher.NodePattern, part *PatternPart, bindingResult BindingResult) error {
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
		// Update the data type of the right node so that it reflects that it is now the terminal node of
		// an expansion
		currentSegment.RightNode.DataType = pgsql.ExpansionTerminalNode
	}

	return nil
}

func (s *Translator) translateNodePatternToStep(nodePattern *cypher.NodePattern, part *PatternPart, bindingResult BindingResult) error {
	currentQueryPart := s.query.CurrentPart()

	// Check the IR for any collected properties
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

	// Check for kind constraints
	if len(nodePattern.Kinds) > 0 {
		if kindIDs, err := s.kindMapper.MapKinds(s.ctx, nodePattern.Kinds); err != nil {
			return fmt.Errorf("failed to translate kinds: %w", err)
		} else if kindIDsLiteral, err := pgsql.AsLiteral(kindIDs); err != nil {
			return err
		} else if err := s.treeTranslator.Constrain(pgsql.NewIdentifierSet().Add(bindingResult.Binding.Identifier), pgsql.NewBinaryExpression(
			pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnKindIDs},
			pgsql.OperatorPGArrayOverlap,
			kindIDsLiteral,
		)); err != nil {
			return err
		}
	}

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
		return s.translateNodePatternSegment(nodePattern, part, bindingResult)
	}

	return nil
}

func (s *Translator) buildNodePattern(part *PatternPart) error {
	var (
		partFrame  = part.NodeSelect.Frame
		nextSelect pgsql.Select
	)

	// The current query part may not have a frame associated with it if is a single part query component
	if previousFrame, hasPrevious := previousValidFrame(s.query, partFrame); hasPrevious {
		nextSelect.From = append(nextSelect.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{previousFrame.Binding.Identifier},
			},
		})
	}

	nextSelect.Projection = part.NodeSelect.Select.Projection

	if part.NodeSelect.Constraint != nil {
		nextSelect.Where = part.NodeSelect.Constraint.Expression
	}

	nextSelect.From = append(nextSelect.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
			Binding: models.ValueOptional(part.NodeSelect.Binding.Identifier),
		},
	})

	// Prepare the next select statement
	s.query.CurrentPart().Model.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name: partFrame.Binding.Identifier,
		},
		Query: pgsql.Query{
			Body: nextSelect,
		},
	})

	return nil
}
