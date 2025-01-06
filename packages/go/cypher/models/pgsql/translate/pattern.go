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
)

type BindingResult struct {
	Binding      *BoundIdentifier
	AlreadyBound bool
}

func (s *Translator) bindPatternExpression(scope *Scope, cypherExpression cypher.Expression, dataType pgsql.DataType) (BindingResult, error) {
	if cypherBinding, hasCypherBinding, err := extractIdentifierFromCypherExpression(cypherExpression); err != nil {
		return BindingResult{}, err
	} else if existingBinding, bound := scope.AliasedLookup(cypherBinding); bound {
		return BindingResult{
			Binding:      existingBinding,
			AlreadyBound: true,
		}, nil
	} else if binding, err := scope.DefineNew(dataType); err != nil {
		return BindingResult{}, err
	} else {
		if hasCypherBinding {
			scope.Alias(cypherBinding, binding)
		}

		return BindingResult{
			Binding:      binding,
			AlreadyBound: false,
		}, nil
	}
}

func (s *Translator) translatePatternPart(scope *Scope, patternPart *cypher.PatternPart) error {
	// We expect this to be a node select if there aren't enough pattern elements for a traversal
	newPatternPart := s.pattern.NewPart()
	newPatternPart.IsTraversal = len(patternPart.PatternElements) > 1
	newPatternPart.ShortestPath = patternPart.ShortestPathPattern
	newPatternPart.AllShortestPaths = patternPart.AllShortestPathsPattern

	if cypherBinding, hasCypherSymbol, err := extractIdentifierFromCypherExpression(patternPart); err != nil {
		return err
	} else if hasCypherSymbol {
		if pathBinding, err := scope.DefineNew(pgsql.PathComposite); err != nil {
			return err
		} else {
			// Generate an alias for this binding
			scope.Alias(cypherBinding, pathBinding)

			// Record the new binding in the traversal pattern being built
			newPatternPart.PatternBinding = models.ValueOptional(pathBinding)
		}
	}

	return nil
}

func (s *Translator) buildPatternPart(scope *Scope, part *PatternPart) error {
	if part.IsTraversal {
		return s.buildPattern(scope, part)
	} else {
		return s.buildNodePattern(scope, part)
	}
}

func (s *Translator) buildPattern(scope *Scope, part *PatternPart) error {
	for idx, traversalStep := range part.TraversalSteps {
		if traversalStep.Expansion.Set {
			if idx > 0 {
				if traversalStepQuery, err := s.buildExpansionPatternStep(part, traversalStep); err != nil {
					return err
				} else {
					s.query.Model.AddCTE(pgsql.CommonTableExpression{
						Alias: pgsql.TableAlias{
							Name: traversalStep.Expansion.Value.Frame.Binding.Identifier,
						},
						Query: traversalStepQuery,
					})
				}
			} else {
				if traversalStepQuery, err := s.buildExpansionPatternRoot(part, traversalStep); err != nil {
					return err
				} else {
					s.query.Model.AddCTE(pgsql.CommonTableExpression{
						Alias: pgsql.TableAlias{
							Name: traversalStep.Expansion.Value.Frame.Binding.Identifier,
						},
						Query: traversalStepQuery,
					})
				}
			}
		} else if idx > 0 {
			if traversalStepQuery, err := s.buildTraversalPatternStep(part, traversalStep); err != nil {
				return err
			} else {
				s.query.Model.AddCTE(pgsql.CommonTableExpression{
					Alias: pgsql.TableAlias{
						Name: traversalStep.Frame.Binding.Identifier,
					},
					Query: traversalStepQuery,
				})
			}
		} else {
			if traversalStepQuery, err := s.buildTraversalPatternRoot(part, traversalStep); err != nil {
				return err
			} else {
				s.query.Model.AddCTE(pgsql.CommonTableExpression{
					Alias: pgsql.TableAlias{
						Name: traversalStep.Frame.Binding.Identifier,
					},
					Query: traversalStepQuery,
				})
			}
		}
	}

	return nil
}
