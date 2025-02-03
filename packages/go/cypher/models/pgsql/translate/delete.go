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

func (s *Translator) translateDelete(scope *Scope, cypherDelete *cypher.Delete) error {
	for range cypherDelete.Expressions {
		if expression, err := s.treeTranslator.Pop(); err != nil {
			return err
		} else {
			switch typedExpression := expression.(type) {
			case pgsql.Identifier:
				if deleteFrame, err := scope.PushFrame(); err != nil {
					return err
				} else {
					if identifierDeletion, err := s.query.CurrentPart().mutations.AddDeletion(scope, typedExpression, deleteFrame); err != nil {
						return err
					} else if boundProjections, err := buildVisibleScopeProjections(scope, nil); err != nil {
						return err
					} else {
						for _, selectItem := range boundProjections.Items {
							switch typedProjection := selectItem.(type) {
							case *pgsql.AliasedExpression:
								if !typedProjection.Alias.Set {
									return fmt.Errorf("expected aliased expression to have an alias set")
								} else if typedProjection.Alias.Value == typedExpression {
									// This is the projection being replaced by the assignment
									if rewrittenProjections, err := buildProjection(typedExpression, identifierDeletion.UpdateBinding, scope, scope.ReferenceFrame()); err != nil {
										return err
									} else {
										identifierDeletion.Projection = append(identifierDeletion.Projection, rewrittenProjections...)
									}
								} else {
									// Reflects this scope binding to the next query part
									identifierDeletion.Projection = append(identifierDeletion.Projection, typedProjection)
								}

							default:
								return fmt.Errorf("expected aliased expression as projection but got: %T", selectItem)
							}
						}
					}
				}

			default:
				return fmt.Errorf("unsupported delete expression: %T", expression)
			}
		}
	}

	return nil
}

func (s *Translator) buildDeletions(scope *Scope) error {
	for _, identifierDeletion := range s.query.CurrentPart().mutations.Deletions.Values() {
		var (
			sqlDelete = pgsql.Delete{
				Using: []pgsql.FromClause{{
					Source: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{identifierDeletion.Frame.Previous.Binding.Identifier},
					},
				}},
			}

			joinConstraint = &Constraint{
				Dependencies: pgsql.AsIdentifierSet(identifierDeletion.TargetBinding.Identifier, identifierDeletion.UpdateBinding.Identifier),
				Expression: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{identifierDeletion.TargetBinding.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{identifierDeletion.UpdateBinding.Identifier, pgsql.ColumnID},
				),
			}
		)

		switch identifierDeletion.UpdateBinding.DataType {
		case pgsql.NodeComposite:
			sqlDelete.From = append(sqlDelete.From, pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(identifierDeletion.UpdateBinding.Identifier),
			})

		case pgsql.EdgeComposite:
			sqlDelete.From = append(sqlDelete.From, pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(identifierDeletion.UpdateBinding.Identifier),
			})

		default:
			return fmt.Errorf("invalid identifier data type for deletion: %s", identifierDeletion.UpdateBinding.Identifier)
		}

		if err := rewriteConstraintIdentifierReferences(identifierDeletion.Frame, []*Constraint{joinConstraint}); err != nil {
			return err
		}

		sqlDelete.Returning = identifierDeletion.Projection
		sqlDelete.Where = models.ValueOptional(joinConstraint.Expression)

		s.query.CurrentPart().Model.AddCTE(pgsql.CommonTableExpression{
			Alias: pgsql.TableAlias{
				Name: scope.CurrentFrameBinding().Identifier,
			},
			Query: pgsql.Query{
				Body: sqlDelete,
			},
		})
	}

	return nil
}
