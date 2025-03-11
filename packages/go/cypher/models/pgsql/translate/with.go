// Copyright 2025 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func (s *Translator) translateWith() error {
	currentPart := s.query.CurrentPart()

	if !currentPart.HasProjections() {
		currentPart.Frame.Exported.Clear()
	} else {
		var (
			projectedItems = pgsql.NewIdentifierSet()

			// aggregatedItems contains a set of symbols of projected aggregate functions.
			aggregatedItems = NewSymbols()

			// groupByItems is a set of symbols (identifiers and compound identifiers) that the query is expected to
			// group by. This is built by exclusion of all aggregated items.
			groupByItems = NewSymbols()
		)

		for _, projectionItem := range currentPart.projections.Items {
			if err := RewriteFrameBindings(s.scope, projectionItem.SelectItem); err != nil {
				return err
			}
		}

		// If an aggregation function is being used this invokes an implicit group by of non-function projections
		for _, projectionItem := range currentPart.projections.Items {
			switch typedSelectItem := projectionItem.SelectItem.(type) {
			case pgsql.FunctionCall:
				if pgsql.IsAggregateFunction(typedSelectItem.Function) {
					for _, parameter := range typedSelectItem.Parameters {
						if symbols, err := SymbolsFor(parameter); err != nil {
							return err
						} else {
							aggregatedItems.AddSymbols(symbols)
						}
					}

					continue
				}
			}

			if selectItemSymbols, err := SymbolsFor(projectionItem.SelectItem); err != nil {
				return err
			} else {
				groupByItems.Add(selectItemSymbols.NotIn(aggregatedItems))
			}
		}

		if projectionConstraint, err := s.treeTranslator.ConsumeSet(s.scope.CurrentFrame().Known().RemoveSet(aggregatedItems.RootIdentifiers())); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.scope, projectionConstraint.Expression); err != nil {
			return err
		} else {
			currentPart.projections.Constraints = projectionConstraint.Expression
		}

		for idx, projectionItem := range currentPart.projections.Items {
			switch typedSelectItem := projectionItem.SelectItem.(type) {
			case *pgsql.BinaryExpression:
				return fmt.Errorf("binary expression not supported in with statement")

			case pgsql.CompoundIdentifier:
				return fmt.Errorf("compound identifier not supported in with statement")

			case pgsql.Identifier:
				if binding, isBound := s.scope.Lookup(typedSelectItem); !isBound {
					return fmt.Errorf("unable to lookup identifer %s for with statement", typedSelectItem)
				} else {
					// Track this projected item for scope pruning
					projectedItems.Add(binding.Identifier)

					// Create a new projection that maps the identifier
					currentPart.projections.Items[idx] = &Projection{
						SelectItem: pgsql.CompoundIdentifier{
							binding.LastProjection.Binding.Identifier, typedSelectItem,
						},
						Alias: pgsql.AsOptionalIdentifier(binding.Identifier),
					}

					// Assign the frame to the binding's last projection backref
					binding.MaterializedBy(currentPart.Frame)

					// Reveal and export the identifier in the current multipart query part's frame
					currentPart.Frame.Reveal(binding.Identifier)
					currentPart.Frame.Export(binding.Identifier)
				}

			default:
				// If this is not an identifier then check if the alias is specified. If the alias is specified, this
				// is a pure export (left-hand side is some other expression) and a new bound identifier is being
				// introduced.
				if projectionItem.Alias.Set {
					if binding, isBound := s.scope.AliasedLookup(projectionItem.Alias.Value); !isBound {
						return fmt.Errorf("unable to lookup alias %s for with statement", projectionItem.Alias.Value)
					} else {
						// Track this projected item for scope pruning
						projectedItems.Add(binding.Identifier)

						// Assign the frame to the binding's last projection backref
						binding.LastProjection = currentPart.Frame

						// Reveal and export the identifier in the current multipart query part's frame
						currentPart.Frame.Reveal(binding.Identifier)
						currentPart.Frame.Export(binding.Identifier)

						// Rewrite this projection's alias to use the internal binding
						projectionItem.Alias.Value = binding.Identifier
					}
				}
			}
		}

		if !aggregatedItems.IsEmpty() {
			groupByItems.EachIdentifier(func(next pgsql.Identifier) bool {
				currentPart.projections.GroupBy = append(currentPart.projections.GroupBy, next)
				return true
			})

			groupByItems.EachCompoundIdentifier(func(next pgsql.CompoundIdentifier) bool {
				currentPart.projections.GroupBy = append(currentPart.projections.GroupBy, next)
				return true
			})
		}

		if err := s.scope.PruneDefinitions(projectedItems); err != nil {
			return err
		}

		// Prune scope to only what's being exported by the with statement
		currentPart.Frame.Visible = projectedItems.Copy()
		currentPart.Frame.Exported = projectedItems.Copy()
	}

	return nil
}
