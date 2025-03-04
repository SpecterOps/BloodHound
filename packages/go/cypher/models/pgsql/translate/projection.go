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
			parameterExpression pgsql.Expression
			edgeReferences      []pgsql.Expression
		)

		for _, dependency := range projected.Dependencies {
			switch dependency.DataType {
			case pgsql.ExpansionPath:
				parameterExpression = pgsql.OptionalBinaryExpressionJoin(
					parameterExpression,
					pgsql.OperatorConcatenate,
					dependency.Identifier,
				)

			case pgsql.EdgeComposite:
				edgeReferences = append(edgeReferences, rewriteCompositeTypeFieldReference(
					scope.CurrentFrameBinding().Identifier,
					pgsql.CompoundIdentifier{dependency.Identifier, pgsql.ColumnID},
				))

			default:
				return nil, fmt.Errorf("unsupported nested composite type for pathcomposite: %s", dependency.DataType)
			}
		}

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
				Expression: pgsql.Parenthetical{
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
