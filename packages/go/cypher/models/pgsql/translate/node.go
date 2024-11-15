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
	"strings"

	"github.com/specterops/bloodhound/cypher/models"
	cypher "github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) translateNodePattern(scope *Scope, nodePattern *cypher.NodePattern, part *PatternPart) error {
	if bindingResult, err := s.bindPatternExpression(scope, nodePattern, pgsql.NodeComposite); err != nil {
		return err
	} else if err := s.translateNodePatternToStep(scope, part, bindingResult); err != nil {
		return err
	} else {
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

		if len(nodePattern.Kinds) > 0 {
			if kindIDs, missingKinds := s.kindMapper.MapKinds(nodePattern.Kinds); len(missingKinds) > 0 {
				s.SetErrorf("unable to map kinds: %s", strings.Join(missingKinds.Strings(), ", "))
			} else if kindIDsLiteral, err := pgsql.AsLiteral(kindIDs); err != nil {
				s.SetError(err)
			} else {
				expression := pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{bindingResult.Binding.Identifier, pgsql.ColumnKindIDs},
					pgsql.OperatorPGArrayOverlap,
					kindIDsLiteral,
				)

				if err := part.Constraints.Constrain(pgsql.AsIdentifierSet(bindingResult.Binding.Identifier), expression); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Translator) translateNodePatternToStep(scope *Scope, part *PatternPart, bindingResult BindingResult) error {
	if part.IsTraversal {
		if numSteps := len(part.TraversalSteps); numSteps == 0 {
			// This is the traversal step's left node
			newTraversalSegment := &PatternSegment{
				LeftNode: bindingResult.Binding,
			}

			// If the left node has not been bound before add it to the pattern part's list of new definitions
			if !bindingResult.AlreadyBound {
				newTraversalSegment.Definitions = append(newTraversalSegment.Definitions, bindingResult.Binding)
			}

			part.TraversalSteps = append(part.TraversalSteps, newTraversalSegment)
		} else {
			currentStep := part.TraversalSteps[numSteps-1]

			// Set the right node pattern identifier
			currentStep.RightNode = bindingResult.Binding

			// If the right node has not been bound before add it to the pattern part's list of new definitions
			if !bindingResult.AlreadyBound {
				currentStep.Definitions = append(currentStep.Definitions, bindingResult.Binding)
			}

			// This is part of a continuing pattern element chain. Inspect the previous edge pattern to see if this
			// is the terminal node of an expansion.
			if currentStep.Expansion.Set {
				if stepFrame, err := scope.PushFrame(); err != nil {
					return err
				} else {
					currentStep.Expansion.Value.Frame = stepFrame
				}

				if boundProjections, err := buildVisibleScopeProjections(scope, currentStep.Definitions); err != nil {
					return err
				} else {
					currentStep.Expansion.Value.Projection = boundProjections.Items
				}

				bindingResult.Binding.DataType = pgsql.ExpansionTerminalNode

				// If the edge is an expansion link the node as the right terminal node to the expansion
				if expansionBinding, found := currentStep.Edge.FirstDependencyByType(pgsql.ExpansionPattern); !found {
					return fmt.Errorf("unable to find expansion context for node: %s", bindingResult.Binding.Identifier)
				} else {
					bindingResult.Binding.Link(expansionBinding)
				}
			} else {
				if stepFrame, err := scope.PushFrame(); err != nil {
					return err
				} else {
					currentStep.Frame = stepFrame
				}

				if boundProjections, err := buildVisibleScopeProjections(scope, currentStep.Definitions); err != nil {
					return err
				} else {
					currentStep.Projection = boundProjections.Items
				}
			}
		}
	} else {
		// If this isn't a traversal
		if nodeFrame, err := scope.PushFrame(); err != nil {
			return err
		} else {
			part.NodeSelect.Frame = nodeFrame
		}

		// Track if the node select is introducing a new variable into the scope
		if bindingResult.AlreadyBound {
			part.NodeSelect.IsDefinition = false

			if boundProjections, err := buildVisibleScopeProjections(scope, nil); err != nil {
				return err
			} else {
				part.NodeSelect.Select.Projection = boundProjections.Items
			}
		} else {
			part.NodeSelect.IsDefinition = true

			if boundProjections, err := buildVisibleScopeProjections(scope, []*BoundIdentifier{bindingResult.Binding}); err != nil {
				return err
			} else {
				part.NodeSelect.Select.Projection = boundProjections.Items
			}
		}

		part.NodeSelect.Binding = bindingResult.Binding
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

func (s *Translator) buildNodePattern(scope *Scope, part *PatternPart) error {
	var (
		nextSelect pgsql.Select
	)

	if part.NodeSelect.Frame.Previous != nil {
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
		s.query.Model.AddCTE(pgsql.CommonTableExpression{
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
