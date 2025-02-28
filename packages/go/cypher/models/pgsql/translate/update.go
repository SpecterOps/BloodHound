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

func (s *Translator) translateUpdates() error {
	currentQueryPart := s.query.CurrentPart()

	// Do not translate the update statements until all are collected
	currentQueryPart.numUpdatingClauses -= 1

	if currentQueryPart.numUpdatingClauses > 0 {
		return nil
	}

	for _, updateClause := range currentQueryPart.mutations.Updates.Values() {
		if stepFrame, err := s.query.Scope.PushFrame(); err != nil {
			return err
		} else {
			updateClause.Frame = stepFrame

			joinConstraint := &Constraint{
				Dependencies: pgsql.AsIdentifierSet(updateClause.TargetBinding.Identifier, updateClause.UpdateBinding.Identifier),
				Expression: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{updateClause.TargetBinding.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{updateClause.UpdateBinding.Identifier, pgsql.ColumnID},
				),
			}

			if err := RewriteFrameBindings(s.query.Scope, joinConstraint.Expression); err != nil {
				return err
			}

			updateClause.JoinConstraint = joinConstraint.Expression

			if boundProjections, err := buildVisibleProjections(s.query.Scope); err != nil {
				return err
			} else {
				// Zip through all projected identifiers and update their last projected frame
				for _, binding := range boundProjections.Bindings {
					binding.LastProjection = stepFrame
				}

				for _, selectItem := range boundProjections.Items {
					switch typedProjection := selectItem.(type) {
					case *pgsql.AliasedExpression:
						if !typedProjection.Alias.Set {
							return fmt.Errorf("expected aliased expression to have an alias set")
						} else if typedProjection.Alias.Value == updateClause.TargetBinding.Identifier {
							// This is the projection being replaced by the assignment
							if rewrittenProjections, err := buildProjection(updateClause.TargetBinding.Identifier, updateClause.UpdateBinding, s.query.Scope, s.query.Scope.ReferenceFrame()); err != nil {
								return err
							} else {
								updateClause.Projection = append(updateClause.Projection, rewrittenProjections...)
							}
						} else {
							updateClause.Projection = append(updateClause.Projection, typedProjection)
						}

					default:
						return fmt.Errorf("expected aliased expression as projection but got: %T", selectItem)
					}
				}
			}

		}
	}

	return s.buildUpdates()
}

func (s *Translator) buildUpdates() error {
	for _, identifierMutation := range s.query.CurrentPart().mutations.Updates.Values() {
		sqlUpdate := pgsql.Update{
			From: []pgsql.FromClause{{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{identifierMutation.Frame.Previous.Binding.Identifier},
				}},
			},
		}

		switch identifierMutation.UpdateBinding.DataType {
		case pgsql.NodeComposite:
			sqlUpdate.Table = pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(identifierMutation.UpdateBinding.Identifier),
			}

		case pgsql.EdgeComposite:
			sqlUpdate.Table = pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(identifierMutation.UpdateBinding.Identifier),
			}

		default:
			return fmt.Errorf("invalid identifier data type for update: %s", identifierMutation.UpdateBinding.Identifier)
		}

		var (
			kindAssignments      models.Optional[pgsql.Expression]
			kindRemovals         models.Optional[pgsql.Expression]
			propertyAssignments  models.Optional[pgsql.Expression]
			propertyRemovals     models.Optional[pgsql.Expression]
			kindColumnIdentifier = pgsql.ColumnKindID
		)

		if identifierMutation.UpdateBinding.DataType.MatchesOneOf(pgsql.NodeComposite, pgsql.ExpansionRootNode, pgsql.ExpansionTerminalNode) {
			kindColumnIdentifier = pgsql.ColumnKindIDs
		}

		if len(identifierMutation.KindAssignments) > 0 {
			if kindIDs, err := s.kindMapper.MapKinds(s.ctx, identifierMutation.KindAssignments); err != nil {
				s.SetError(fmt.Errorf("failed to translate kinds: %w", err))
			} else {
				arrayLiteral := pgsql.ArrayLiteral{
					Values:   make([]pgsql.Expression, len(kindIDs)),
					CastType: pgsql.Int2Array,
				}

				for idx, kindID := range kindIDs {
					arrayLiteral.Values[idx] = pgsql.NewLiteral(kindID, pgsql.Int2)
				}

				kindAssignments = models.ValueOptional(arrayLiteral.AsExpression())
			}
		}

		if len(identifierMutation.KindRemovals) > 0 {
			if kindIDs, err := s.kindMapper.MapKinds(s.ctx, identifierMutation.KindRemovals); err != nil {
				s.SetError(fmt.Errorf("failed to translate kinds: %w", err))
			} else {
				arrayLiteral := pgsql.ArrayLiteral{
					Values:   make([]pgsql.Expression, len(kindIDs)),
					CastType: pgsql.Int2Array,
				}

				for idx, kindID := range kindIDs {
					arrayLiteral.Values[idx] = pgsql.NewLiteral(kindID, pgsql.Int2)
				}

				kindRemovals = models.ValueOptional(arrayLiteral.AsExpression())
			}
		}

		if identifierMutation.PropertyAssignments.Len() > 0 {
			jsonObjectFunction := pgsql.FunctionCall{
				Function: pgsql.FunctionJSONBBuildObject,
				CastType: pgsql.JSONB,
			}

			for _, propertyAssignment := range identifierMutation.PropertyAssignments.Values() {
				if err := RewriteFrameBindings(s.query.Scope, propertyAssignment.ValueExpression); err != nil {
					return err
				}

				if propertyLookup, isPropertyLookup := asPropertyLookup(propertyAssignment.ValueExpression); isPropertyLookup {
					// Ensure that property lookups in JSONB build functions use the JSONB field type
					propertyLookup.Operator = pgsql.OperatorJSONField
				}

				jsonObjectFunction.Parameters = append(jsonObjectFunction.Parameters,
					pgsql.NewLiteral(propertyAssignment.Field, pgsql.Text),
					propertyAssignment.ValueExpression,
				)
			}

			propertyAssignments = models.ValueOptional(jsonObjectFunction.AsExpression())
		}

		if identifierMutation.Removals.Len() > 0 {
			fieldRemovalArray := pgsql.ArrayLiteral{
				CastType: pgsql.TextArray,
			}

			for _, propertyRemoval := range identifierMutation.Removals.Values() {
				fieldRemovalArray.Values = append(fieldRemovalArray.Values, pgsql.NewLiteral(propertyRemoval.Field, pgsql.Text))
			}

			propertyRemovals = models.ValueOptional(fieldRemovalArray.AsExpression())
		}

		if kindAssignments.Set {
			if err := RewriteFrameBindings(s.query.Scope, kindAssignments.Value); err != nil {
				return err
			}

			if kindRemovals.Set {
				sqlUpdate.Assignments = []pgsql.Assignment{
					pgsql.NewBinaryExpression(
						kindColumnIdentifier,
						pgsql.OperatorAssignment,
						pgsql.FunctionCall{
							Function: pgsql.FunctionIntArrayUnique,
							Parameters: []pgsql.Expression{
								pgsql.FunctionCall{
									Function: pgsql.FunctionIntArraySort,
									Parameters: []pgsql.Expression{
										pgsql.NewBinaryExpression(
											pgsql.NewBinaryExpression(
												pgsql.CompoundIdentifier{identifierMutation.UpdateBinding.Identifier, kindColumnIdentifier},
												pgsql.OperatorSubtract,
												kindRemovals.Value,
											),
											pgsql.OperatorConcatenate,
											kindAssignments.Value,
										),
									},
									CastType: pgsql.Int2Array,
								},
							},
							CastType: pgsql.Int2Array,
						},
					),
				}
			} else {
				sqlUpdate.Assignments = []pgsql.Assignment{
					pgsql.NewBinaryExpression(
						kindColumnIdentifier,
						pgsql.OperatorAssignment,
						pgsql.FunctionCall{
							Function: pgsql.FunctionIntArrayUnique,
							Parameters: []pgsql.Expression{
								pgsql.FunctionCall{
									Function: pgsql.FunctionIntArraySort,
									Parameters: []pgsql.Expression{
										pgsql.NewBinaryExpression(
											pgsql.CompoundIdentifier{identifierMutation.UpdateBinding.Identifier, kindColumnIdentifier},
											pgsql.OperatorConcatenate,
											kindAssignments.Value,
										),
									},
									CastType: pgsql.Int2Array,
								},
							},
							CastType: pgsql.Int2Array,
						},
					),
				}
			}
		} else if kindRemovals.Set {
			sqlUpdate.Assignments = []pgsql.Assignment{pgsql.NewBinaryExpression(
				kindColumnIdentifier,
				pgsql.OperatorAssignment,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{identifierMutation.UpdateBinding.Identifier, kindColumnIdentifier},
					pgsql.OperatorSubtract,
					kindRemovals.Value,
				),
			)}
		}

		if propertyAssignments.Set {
			if err := RewriteFrameBindings(s.query.Scope, propertyAssignments.Value); err != nil {
				return err
			}

			if propertyRemovals.Set {
				sqlUpdate.Assignments = []pgsql.Assignment{pgsql.NewBinaryExpression(
					pgsql.ColumnProperties,
					pgsql.OperatorAssignment,
					pgsql.NewBinaryExpression(
						pgsql.NewBinaryExpression(
							pgsql.CompoundIdentifier{identifierMutation.UpdateBinding.Identifier, pgsql.ColumnProperties},
							pgsql.OperatorSubtract,
							propertyRemovals.Value,
						),
						pgsql.OperatorConcatenate,
						propertyAssignments.Value,
					),
				)}
			} else {
				sqlUpdate.Assignments = []pgsql.Assignment{pgsql.NewBinaryExpression(
					pgsql.ColumnProperties,
					pgsql.OperatorAssignment,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{identifierMutation.UpdateBinding.Identifier, pgsql.ColumnProperties},
						pgsql.OperatorConcatenate,
						propertyAssignments.Value,
					),
				)}
			}
		} else if propertyRemovals.Set {
			sqlUpdate.Assignments = []pgsql.Assignment{pgsql.NewBinaryExpression(
				pgsql.ColumnProperties,
				pgsql.OperatorAssignment,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{identifierMutation.UpdateBinding.Identifier, pgsql.ColumnProperties},
					pgsql.OperatorSubtract,
					propertyRemovals.Value,
				),
			)}
		}

		sqlUpdate.Returning = identifierMutation.Projection
		sqlUpdate.Where = models.ValueOptional(identifierMutation.JoinConstraint)

		s.query.CurrentPart().Model.AddCTE(pgsql.CommonTableExpression{
			Alias: pgsql.TableAlias{
				Name: identifierMutation.Frame.Binding.Identifier,
			},
			Query: pgsql.Query{
				Body: sqlUpdate,
			},
		})
	}

	return nil
}
