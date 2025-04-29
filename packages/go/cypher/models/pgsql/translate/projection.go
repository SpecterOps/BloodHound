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

	"github.com/specterops/bloodhound/cypher/models/cypher"

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

type BoundProjections struct {
	Items    pgsql.Projection
	Bindings []*BoundIdentifier
}

func rewriteConstraintIdentifierReferences(scope *Scope, frame *Frame, constraints []*Constraint) error {
	if frame.Previous == nil {
		return nil
	}

	for _, constraint := range constraints {
		if err := RewriteFrameBindings(scope, constraint.Expression); err != nil {
			return err
		}
	}

	return nil
}

func buildExternalProjection(scope *Scope, projections []*Projection) (pgsql.Projection, error) {
	var sqlProjection pgsql.Projection

	for _, projection := range projections {
		switch typedProjectionExpression := projection.SelectItem.(type) {
		case pgsql.Identifier:
			alias := projection.Alias.Value

			if projectedBinding, bound := scope.Lookup(typedProjectionExpression); !bound {
				return nil, fmt.Errorf("invalid identifier: %s", typedProjectionExpression)
			} else {
				if !projection.Alias.Set {
					alias = projectedBinding.Alias.Value
				}

				if builtProjection, err := buildProjection(alias, projectedBinding, scope, projectedBinding.LastProjection); err != nil {
					return nil, err
				} else {
					for _, buildProjectionItem := range builtProjection {
						sqlProjection = append(sqlProjection, buildProjectionItem)
					}
				}
			}

		default:
			builtProjection := projection.SelectItem

			if projection.Alias.Set {
				builtProjection = &pgsql.AliasedExpression{
					Expression: builtProjection,
					Alias:      projection.Alias,
				}
			}

			sqlProjection = append(sqlProjection, builtProjection)
		}
	}

	if err := RewriteFrameBindings(scope, sqlProjection); err != nil {
		return nil, err
	}

	// Lastly, return the projections while rewriting the given constraints
	return sqlProjection, nil
}

func buildInternalProjection(scope *Scope, projectedBindings []*BoundIdentifier) (BoundProjections, error) {
	var (
		boundProjections = BoundProjections{
			Bindings: projectedBindings,
		}
		projected = map[pgsql.Identifier]struct{}{}
	)

	for _, projectedBinding := range projectedBindings {
		if _, alreadyProjected := projected[projectedBinding.Identifier]; alreadyProjected {
			continue
		}

		projected[projectedBinding.Identifier] = struct{}{}

		// Build the identifier's projection
		if newSelectItems, err := buildProjection(projectedBinding.Identifier, projectedBinding, scope, projectedBinding.LastProjection); err != nil {
			return BoundProjections{}, err
		} else {
			boundProjections.Items = append(boundProjections.Items, newSelectItems...)
		}
	}

	// Lastly, return the projections while rewriting the given constraints
	return boundProjections, nil
}

func buildVisibleProjections(scope *Scope) (BoundProjections, error) {
	currentFrame := scope.CurrentFrame()

	if knownBindings, err := scope.LookupBindings(currentFrame.Known().Slice()...); err != nil {
		return BoundProjections{}, err
	} else {
		return buildInternalProjection(scope, knownBindings)
	}
}

func buildProjection(alias pgsql.Identifier, projected *BoundIdentifier, scope *Scope, referenceFrame *Frame) ([]pgsql.SelectItem, error) {
	switch projected.DataType {
	case pgsql.ExpansionPath:
		if projected.LastProjection != nil {
			return []pgsql.SelectItem{
				&pgsql.AliasedExpression{
					Expression: pgsql.CompoundIdentifier{referenceFrame.Binding.Identifier, projected.Identifier},
					Alias:      pgsql.AsOptionalIdentifier(alias),
				},
			}, nil
		}

		return []pgsql.SelectItem{
			&pgsql.AliasedExpression{
				Expression: pgsql.CompoundIdentifier{scope.CurrentFrame().Binding.Identifier, pgsql.ColumnPath},
				Alias:      models.ValueOptional(alias),
			},
		}, nil

	case pgsql.PathComposite:
		var (
			parameterExpression    pgsql.Expression
			edgeReferences         []pgsql.Expression
			nodeReferences         []pgsql.Expression
			useEdgesToPathFunction = false
		)

		// Path composite components are encoded as dependencies on the bound identifier representing the
		// path. This is not ideal as it escapes normal translation flow as driven by the structure of the
		// originating cypher AST.
		for _, dependency := range projected.Dependencies {
			switch dependency.DataType {
			case pgsql.ExpansionPath:
				parameterExpression = pgsql.OptionalBinaryExpressionJoin(
					parameterExpression,
					pgsql.OperatorConcatenate,
					dependency.Identifier,
				)

				useEdgesToPathFunction = true

			case pgsql.EdgeComposite:
				edgeReferences = append(edgeReferences, rewriteCompositeTypeFieldReference(
					scope.CurrentFrameBinding().Identifier,
					pgsql.CompoundIdentifier{dependency.Identifier, pgsql.ColumnID},
				))

				useEdgesToPathFunction = true

			case pgsql.NodeComposite:
				nodeReferences = append(nodeReferences, rewriteCompositeTypeFieldReference(
					scope.CurrentFrameBinding().Identifier,
					pgsql.CompoundIdentifier{dependency.Identifier, pgsql.ColumnID},
				))

			default:
				return nil, fmt.Errorf("unsupported type for path rendering: %s", dependency.DataType)
			}
		}

		// The code below is covering a strange edge-case of cypher where a query may contain the following
		// form: match p = (n) return p
		//
		// In this case it is not appropriate to call the edges_to_path(...) function and instead a call to
		// the corresponding nodes_to_path(...) function must be authored.
		if useEdgesToPathFunction {
			// It's possible for a path to contain both edge ID references and expansion references. Expansions
			// are represented as a concatenation of arrays of edge IDs contained within the parameterExpression
			// variable. If there are edge ID references that are a part of the path then the individual edge
			// references must first be rewritten as an array and then further concatenated to the existing
			// path results.
			if len(edgeReferences) > 0 {
				parameterExpression = pgsql.OptionalBinaryExpressionJoin(
					parameterExpression,
					pgsql.OperatorConcatenate,
					pgsql.ArrayLiteral{
						Values:   edgeReferences,
						CastType: pgsql.Int8Array,
					},
				)
			}

			return []pgsql.SelectItem{
				&pgsql.AliasedExpression{
					Expression: pgsql.FunctionCall{
						Function: pgsql.FunctionEdgesToPath,
						Parameters: []pgsql.Expression{
							pgsql.Variadic{
								Expression: parameterExpression,
							},
						},
						CastType: pgsql.PathComposite,
					},
					Alias: pgsql.AsOptionalIdentifier(alias),
				},
			}, nil
		} else if len(nodeReferences) > 0 {
			return []pgsql.SelectItem{
				&pgsql.AliasedExpression{
					Expression: pgsql.FunctionCall{
						Function: pgsql.FunctionNodesToPath,
						Parameters: []pgsql.Expression{
							pgsql.Variadic{
								Expression: pgsql.ArrayLiteral{
									Values:   nodeReferences,
									CastType: pgsql.Int8Array,
								},
							},
						},
						CastType: pgsql.PathComposite,
					},
					Alias: pgsql.AsOptionalIdentifier(alias),
				},
			}, nil
		}

		return nil, fmt.Errorf("path variable does not contain valid components")

	case pgsql.ExpansionRootNode, pgsql.ExpansionTerminalNode:
		if projected.LastProjection != nil {
			return []pgsql.SelectItem{
				&pgsql.AliasedExpression{
					Expression: pgsql.CompoundIdentifier{referenceFrame.Binding.Identifier, projected.Identifier},
					Alias:      pgsql.AsOptionalIdentifier(alias),
				},
			}, nil
		}

		value := pgsql.CompositeValue{
			DataType: pgsql.NodeComposite,
		}

		for _, nodeTableColumn := range pgsql.NodeTableColumns {
			value.Values = append(value.Values, pgsql.CompoundIdentifier{projected.Identifier, nodeTableColumn})
		}

		// Change the type to the node composite now that this is projected
		projected.DataType = pgsql.NodeComposite

		// Create a new final projection that's aliased to the visible binding's identifier
		return []pgsql.SelectItem{
			&pgsql.AliasedExpression{
				Expression: value,
				Alias:      pgsql.AsOptionalIdentifier(alias),
			},
		}, nil

	case pgsql.NodeComposite:
		if projected.LastProjection != nil {
			return []pgsql.SelectItem{
				&pgsql.AliasedExpression{
					Expression: pgsql.CompoundIdentifier{referenceFrame.Binding.Identifier, projected.Identifier},
					Alias:      pgsql.AsOptionalIdentifier(alias),
				},
			}, nil
		}

		value := pgsql.CompositeValue{
			DataType: pgsql.NodeComposite,
		}

		for _, nodeTableColumn := range pgsql.NodeTableColumns {
			value.Values = append(value.Values, pgsql.CompoundIdentifier{projected.Identifier, nodeTableColumn})
		}

		// Create a new final projection that's aliased to the visible binding's identifier
		return []pgsql.SelectItem{
			&pgsql.AliasedExpression{
				Expression: value,
				Alias:      pgsql.AsOptionalIdentifier(alias),
			},
		}, nil

	case pgsql.ExpansionEdge:
		value := pgsql.CompositeValue{
			DataType: pgsql.EdgeComposite,
		}

		for _, edgeTableColumn := range pgsql.EdgeTableColumns {
			value.Values = append(value.Values, pgsql.CompoundIdentifier{projected.Identifier, edgeTableColumn})
		}

		// Change the type to the node composite now that this is projected
		projected.DataType = pgsql.EdgeComposite

		// Create a new final projection that's aliased to the visible binding's identifier
		return []pgsql.SelectItem{
			&pgsql.AliasedExpression{
				Expression: &pgsql.Parenthetical{
					Expression: pgsql.Select{
						Projection: []pgsql.SelectItem{
							pgsql.FunctionCall{
								Function:   pgsql.FunctionArrayAggregate,
								Parameters: []pgsql.Expression{value},
							},
						},
						From: []pgsql.FromClause{{
							Source: pgsql.TableReference{
								Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
								Binding: models.ValueOptional(projected.Identifier),
							},
							Joins: nil,
						}},
						Where: pgsql.NewBinaryExpression(
							pgsql.CompoundIdentifier{projected.Identifier, pgsql.ColumnID},
							pgsql.OperatorEquals,
							pgsql.NewAnyExpression(
								pgsql.CompoundIdentifier{scope.CurrentFrame().Binding.Identifier, pgsql.ColumnPath},
								pgsql.ExpansionPath,
							),
						),
					},
				},
				Alias: pgsql.AsOptionalIdentifier(alias),
			},
		}, nil

	case pgsql.EdgeComposite:
		if projected.LastProjection != nil {
			return []pgsql.SelectItem{
				&pgsql.AliasedExpression{
					Expression: pgsql.CompoundIdentifier{referenceFrame.Binding.Identifier, projected.Identifier},
					Alias:      pgsql.AsOptionalIdentifier(alias),
				},
			}, nil
		}

		value := pgsql.CompositeValue{
			DataType: pgsql.EdgeComposite,
		}

		for _, edgeTableColumn := range pgsql.EdgeTableColumns {
			value.Values = append(value.Values, pgsql.CompoundIdentifier{projected.Identifier, edgeTableColumn})
		}

		// Create a new final projection that's aliased to the visible binding's identifier
		return []pgsql.SelectItem{
			&pgsql.AliasedExpression{
				Expression: value,
				Alias:      pgsql.AsOptionalIdentifier(alias),
			},
		}, nil
	}

	// If this isn't a type that requires a unique projection, reflect the identifier as-is with its alias
	return []pgsql.SelectItem{
		&pgsql.AliasedExpression{
			Expression: pgsql.CompoundIdentifier{referenceFrame.Binding.Identifier, projected.Identifier},
			Alias:      pgsql.AsOptionalIdentifier(alias),
		},
	}, nil
}

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
		currentFrame          = s.scope.CurrentFrame()
		singlePartQuerySelect = pgsql.Select{}
	)

	singlePartQuerySelect.From = []pgsql.FromClause{{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{currentFrame.Binding.Identifier},
		},
	}}

	if projectionConstraint, err := s.treeTranslator.ConsumeAllConstraints(); err != nil {
		return err
	} else if projection, err := buildExternalProjection(s.scope, currentPart.projections.Items); err != nil {
		return err
	} else if err := RewriteFrameBindings(s.scope, projectionConstraint.Expression); err != nil {
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

	if len(currentPart.SortItems) > 0 {
		// If there are expressions in the order by of the current query part they will need to be visited to ensure
		// that frame references are rewritten
		for _, orderByExpression := range currentPart.SortItems {
			if err := RewriteFrameBindings(s.scope, orderByExpression); err != nil {
				return err
			}
		}

		currentPart.Model.OrderBy = currentPart.SortItems
	}

	return nil
}

func (s *Translator) translateProjectionItem(scope *Scope, projectionItem *cypher.ProjectionItem) error {
	if alias, hasAlias, err := extractIdentifierFromCypherExpression(projectionItem); err != nil {
		return err
	} else if nextExpression, err := s.treeTranslator.PopOperand(); err != nil {
		return err
	} else if selectItem, isProjection := nextExpression.(pgsql.SelectItem); !isProjection {
		s.SetErrorf("invalid type for select item: %T", nextExpression)
	} else {
		if identifiers, err := ExtractSyntaxNodeReferences(selectItem); err != nil {
			return err
		} else if identifiers.Len() > 0 {
			// Identifier lookups will require a scope reference
			s.query.CurrentPart().projections.Frame = s.scope.CurrentFrame()
		}

		switch typedSelectItem := unwrapParenthetical(selectItem).(type) {
		case pgsql.Identifier:
			// If this is an identifier then assume the identifier as the projection alias since the translator
			// rewrites all identifiers
			if !hasAlias {
				if boundSelectItem, bound := scope.Lookup(typedSelectItem); !bound {
					return fmt.Errorf("invalid identifier: %s", typedSelectItem)
				} else {
					s.query.CurrentPart().CurrentProjection().SetAlias(boundSelectItem.Aliased())
				}
			}

		case *pgsql.BinaryExpression:
			// Binary expressions are used when properties are returned from a result projection
			// e.g. match (n) return n.prop
			if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(typedSelectItem); isPropertyLookup {
				// Ensure that projections maintain the raw JSONB type of the field
				propertyLookup.Operator = pgsql.OperatorJSONField
			}

		default:
			if hasAlias {
				if inferredType, err := InferExpressionType(typedSelectItem); err != nil {
					return err
				} else if _, isBound := s.scope.AliasedLookup(alias); !isBound {
					if newBinding, err := s.scope.DefineNew(inferredType); err != nil {
						return err
					} else {
						// This binding is its own alias
						s.scope.Alias(alias, newBinding)
					}
				}
			}
		}

		if hasAlias {
			s.query.CurrentPart().CurrentProjection().SetAlias(alias)
		}

		s.query.CurrentPart().CurrentProjection().SelectItem = selectItem
	}

	return nil
}

func (s *Translator) prepareProjection(projection *cypher.Projection) error {
	currentPart := s.query.CurrentPart()
	currentPart.PrepareProjections(projection.Distinct)

	if projection.Skip != nil {
		if cypherLiteral, isLiteral := projection.Skip.Value.(*cypher.Literal); !isLiteral {
			return fmt.Errorf("expected a literal skip value but received: %T", projection.Skip.Value)
		} else if pgLiteral, err := pgsql.AsLiteral(cypherLiteral.Value); err != nil {
			return err
		} else {
			currentPart.Skip = models.ValueOptional[pgsql.Expression](pgLiteral)
		}
	}

	if projection.Limit != nil {
		if cypherLiteral, isLiteral := projection.Limit.Value.(*cypher.Literal); !isLiteral {
			return fmt.Errorf("expected a literal limit value but received: %T", projection.Limit.Value)
		} else if pgLiteral, err := pgsql.AsLiteral(cypherLiteral.Value); err != nil {
			return err
		} else {
			currentPart.Limit = models.ValueOptional[pgsql.Expression](pgLiteral)
		}
	}

	return nil
}
