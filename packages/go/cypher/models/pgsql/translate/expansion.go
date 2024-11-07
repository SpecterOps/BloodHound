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
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
)

func expansionConstraints(expansionIdentifier pgsql.Identifier) pgsql.Expression {
	return pgsql.NewBinaryExpression(
		pgsql.NewBinaryExpression(
			pgsql.CompoundIdentifier{expansionIdentifier, expansionDepth},
			pgsql.OperatorLessThan,
			pgsql.NewLiteral(5, pgsql.Int),
		),
		pgsql.OperatorAnd,
		pgsql.UnaryExpression{
			Operator: pgsql.OperatorNot,
			Operand:  pgsql.CompoundIdentifier{expansionIdentifier, expansionIsCycle},
		},
	)
}

type ExpansionBuilder struct {
	PrimerStatement     pgsql.Select
	RecursiveStatement  pgsql.Select
	ProjectionStatement pgsql.Select
	Query               pgsql.Query
}

func (s ExpansionBuilder) BuildAllShortestPaths(primerIdentifier, recursiveIdentifier *pgsql.Parameter, expansionIdentifier pgsql.Identifier, maxDepth int) pgsql.Query {
	s.Query.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name:  expansionIdentifier,
			Shape: models.ValueOptional(expansionColumns()),
		},
		Query: pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Source: pgsql.FunctionCall{
						Function: pgsql.FunctionASPHarness,
						Parameters: []pgsql.Expression{
							primerIdentifier,
							recursiveIdentifier,
							pgsql.NewLiteral(maxDepth, pgsql.Int),
						},
					},
				}},
			},
		},
	})

	s.Query.Body = s.ProjectionStatement
	return s.Query
}

func (s ExpansionBuilder) Build(expansionIdentifier pgsql.Identifier) pgsql.Query {
	s.Query.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name:  expansionIdentifier,
			Shape: models.ValueOptional(expansionColumns()),
		},
		Query: pgsql.Query{
			Body: pgsql.SetOperation{
				LOperand: s.PrimerStatement,
				ROperand: s.RecursiveStatement,
				Operator: pgsql.OperatorUnion,
			},
		},
	})

	s.Query.Body = s.ProjectionStatement
	return s.Query
}

func (s *Translator) buildAllShortestPathsExpansionRoot(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	var (
		expansion = ExpansionBuilder{
			Query: pgsql.Query{
				CommonTableExpressions: &pgsql.With{},
			},

			RecursiveStatement: pgsql.Select{
				Where: expansionConstraints(traversalStep.Expansion.Value.Binding.Identifier),
			},
		}
	)

	expansion.ProjectionStatement.Projection = traversalStep.Expansion.Value.Projection

	if terminalNode, err := traversalStep.TerminalNode(); err != nil {
		return pgsql.Query{}, err
	} else if rootNode, err := traversalStep.RootNode(); err != nil {
		return pgsql.Query{}, err
	} else if rootNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(rootNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if terminalNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(terminalNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else {
		// The exclusion below is done at this step in the process since the recursive descent portion of the query no longer has
		// a reference to `n0` and any dependent interaction between `n0` and `n1` would require an additional join. By not
		// consuming the remaining constraints for `n0` and `n1`, they become visible up in the outer select of the recursive CTE.
		recursiveVisible := traversalStep.Expansion.Value.Frame.Visible.Copy()
		recursiveVisible.Remove(rootNode.Identifier)

		if edgeConstraints, err := consumeConstraintsFrom(recursiveVisible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
			return pgsql.Query{}, err
		} else {
			// Set the edge constraints in the primer and recursive select where clauses
			expansion.PrimerStatement.Where = edgeConstraints.Expression
			expansion.RecursiveStatement.Where = pgsql.OptionalAnd(edgeConstraints.Expression, expansion.RecursiveStatement.Where)
		}

		if leftNodeJoinConstraint, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else if leftNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rootNodeConstraints.Expression, leftNodeJoinConstraint}); err != nil {
			return pgsql.Query{}, err
		} else if rightNodeJoinCondition, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.PrimerStatement.From = append(expansion.PrimerStatement.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
					Binding: models.ValueOptional(traversalStep.Edge.Identifier),
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: leftNodeJoinCondition,
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})

			// Make sure the recursive query has the expansion bound
			expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{"pathspace"},
					Binding: models.ValueOptional(traversalStep.Expansion.Value.Binding.Identifier),
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
						Binding: models.ValueOptional(traversalStep.Edge.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType: pgsql.JoinTypeInner,
						Constraint: pgsql.NewBinaryExpression(
							// TODO: Directional
							pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
							pgsql.OperatorEquals,
							pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, pgsql.ColumnNextID},
						),
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})

			if wrappedSelectJoinConstraint, err := ConjoinExpressions([]pgsql.Expression{
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					&pgsql.ArrayIndex{
						Expression: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
						Indexes: []pgsql.Expression{
							pgsql.FunctionCall{
								Function: pgsql.FunctionArrayLength,
								Parameters: []pgsql.Expression{
									pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
									pgsql.NewLiteral(1, pgsql.Int8),
								},
								CastType: pgsql.Int4,
							},
						},
					},
				),
				rightNodeJoinCondition}); err != nil {
				return pgsql.Query{}, err
			} else {
				// Select the expansion components for the projection statement
				expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
					Source: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier},
						Binding: models.EmptyOptional[pgsql.Identifier](),
					},
					Joins: []pgsql.Join{{
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
							Binding: models.ValueOptional(traversalStep.Edge.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType: pgsql.JoinTypeInner,
							Constraint: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
							),
						},
					}, {
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
							Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType: pgsql.JoinTypeInner,
							Constraint: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
							),
						},
					}, {
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
							Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType:   pgsql.JoinTypeInner,
							Constraint: wrappedSelectJoinConstraint,
						},
					}},
				})
			}
		}

		// If there are terminal constraints, project them as part of the projections
		if terminalNodeConstraints.Expression != nil {
			if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](terminalNodeConstraints.Expression); err != nil {
				return pgsql.Query{}, err
			} else {
				expansion.PrimerStatement.Projection = []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewLiteral(1, pgsql.Int),
					terminalCriteriaProjection,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					),
					pgsql.ArrayLiteral{
						Values: []pgsql.Expression{
							pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						},
					},
				}

				expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.NewLiteral(1, pgsql.Int),
					),
					terminalCriteriaProjection,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					),
				}

				// Make sure to only accept paths that are satisfied
				expansion.ProjectionStatement.Where = pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionSatisfied}
			}
		} else {
			expansion.PrimerStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				pgsql.NewLiteral(1, pgsql.Int),
				pgsql.ExistsExpression{
					Subquery: pgsql.Query{
						Body: pgsql.Select{
							Projection: []pgsql.SelectItem{
								pgsql.NewLiteral(1, pgsql.Int),
							},
							From: []pgsql.FromClause{{
								Source: pgsql.TableReference{
									Name:    pgsql.CompoundIdentifier{model.EdgeTable},
									Binding: models.ValueOptional(traversalStep.Edge.Identifier),
								},
							}},
							Where: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.RightNode.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
							),
						},
					},
					Negated: false,
				},
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				),
				pgsql.ArrayLiteral{
					Values: []pgsql.Expression{
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					},
				},
			}

			expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				pgsql.ExistsExpression{
					Subquery: pgsql.Query{
						Body: pgsql.Select{
							Projection: []pgsql.SelectItem{
								pgsql.NewLiteral(1, pgsql.Int),
							},
							From: []pgsql.FromClause{{
								Source: pgsql.TableReference{
									Name:    pgsql.CompoundIdentifier{model.EdgeTable},
									Binding: models.ValueOptional(traversalStep.Edge.Identifier),
								},
							}},
							Where: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.RightNode.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
							),
						},
					},
					Negated: false,
				},
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}
		}
	}

	var (
		primerInsert = pgsql.Insert{
			Table: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{"next_pathspace"},
			},
			Shape: expansionColumns(),
			Source: &pgsql.Query{
				Body: expansion.PrimerStatement,
			},
		}

		recursiveInsert = pgsql.Insert{
			Table: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{"next_pathspace"},
			},
			Shape: expansionColumns(),
			Source: &pgsql.Query{
				Body: expansion.RecursiveStatement,
			},
		}
	)

	// Create a new container for the parameter and its value
	if primerStatement, err := format.Statement(primerInsert, format.NewOutputBuilder().WithMaterializedParameters(s.translation.Parameters)); err != nil {
		return pgsql.Query{}, err
	} else if recursiveStatement, err := format.Statement(recursiveInsert, format.NewOutputBuilder().WithMaterializedParameters(s.translation.Parameters)); err != nil {
		return pgsql.Query{}, err
	} else if primerParameterBinding, err := s.query.Scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
		return pgsql.Query{}, err
	} else if recursiveParameterBinding, err := s.query.Scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
		return pgsql.Query{}, err
	} else if primerParameter, err := pgsql.AsParameter(primerParameterBinding.Identifier, primerStatement); err != nil {
		return pgsql.Query{}, err
	} else if primerValue, err := pgsql.NegotiateValue(primerStatement); err != nil {
		return pgsql.Query{}, err
	} else if recursiveParameter, err := pgsql.AsParameter(recursiveParameterBinding.Identifier, recursiveStatement); err != nil {
		return pgsql.Query{}, err
	} else if recursiveValue, err := pgsql.NegotiateValue(recursiveStatement); err != nil {
		return pgsql.Query{}, err
	} else {
		s.translation.Parameters[primerParameterBinding.Identifier.String()] = primerValue
		primerParameterBinding.Parameter = models.ValueOptional(primerParameter)

		s.translation.Parameters[recursiveParameterBinding.Identifier.String()] = recursiveValue
		recursiveParameterBinding.Parameter = models.ValueOptional(recursiveParameter)

		return expansion.BuildAllShortestPaths(primerParameter, recursiveParameter, traversalStep.Expansion.Value.Binding.Identifier, 5), nil
	}
}

func (s *Translator) buildExpansionPatternRoot(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	if part.ShortestPath || part.AllShortestPaths {
		return s.buildAllShortestPathsExpansionRoot(part, traversalStep)
	}

	var (
		expansion = ExpansionBuilder{
			Query: pgsql.Query{
				CommonTableExpressions: &pgsql.With{
					Recursive: true,
				},
			},

			RecursiveStatement: pgsql.Select{
				Where: expansionConstraints(traversalStep.Expansion.Value.Binding.Identifier),
			},
		}
	)

	expansion.ProjectionStatement.Projection = traversalStep.Expansion.Value.Projection

	if terminalNode, err := traversalStep.TerminalNode(); err != nil {
		return pgsql.Query{}, err
	} else if rootNode, err := traversalStep.RootNode(); err != nil {
		return pgsql.Query{}, err
	} else if rootNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(rootNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if terminalNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(terminalNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else {
		// The exclusion below is done at this step in the process since the recursive descent portion of the query no longer has
		// a reference to `n0` and any dependent interaction between `n0` and `n1` would require an additional join. By not
		// consuming the remaining constraints for `n0` and `n1`, they become visible up in the outer select of the recursive CTE.
		recursiveVisible := traversalStep.Expansion.Value.Frame.Visible.Copy()
		recursiveVisible.Remove(rootNode.Identifier)

		if edgeConstraints, err := consumeConstraintsFrom(recursiveVisible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
			return pgsql.Query{}, err
		} else {
			// Set the edge constraints in the primer and recursive select where clauses
			expansion.PrimerStatement.Where = edgeConstraints.Expression
			expansion.RecursiveStatement.Where = pgsql.OptionalAnd(edgeConstraints.Expression, expansion.RecursiveStatement.Where)
		}

		if leftNodeJoinConstraint, err := leftNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else if leftNodeJoinCondition, err := ConjoinExpressions([]pgsql.Expression{rootNodeConstraints.Expression, leftNodeJoinConstraint}); err != nil {
			return pgsql.Query{}, err
		} else if rightNodeJoinCondition, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.PrimerStatement.From = append(expansion.PrimerStatement.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
					Binding: models.ValueOptional(traversalStep.Edge.Identifier),
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: leftNodeJoinCondition,
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})

			// Make sure the recursive query has the expansion bound
			expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier},
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
						Binding: models.ValueOptional(traversalStep.Edge.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType: pgsql.JoinTypeInner,
						Constraint: pgsql.NewBinaryExpression(
							// TODO: Directional
							pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
							pgsql.OperatorEquals,
							pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, pgsql.ColumnNextID},
						),
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})

			if wrappedSelectJoinConstraint, err := ConjoinExpressions([]pgsql.Expression{
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					&pgsql.ArrayIndex{
						Expression: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
						Indexes: []pgsql.Expression{
							pgsql.FunctionCall{
								Function: pgsql.FunctionArrayLength,
								Parameters: []pgsql.Expression{
									pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
									pgsql.NewLiteral(1, pgsql.Int8),
								},
								CastType: pgsql.Int4,
							},
						},
					},
				),
				rightNodeJoinCondition}); err != nil {
				return pgsql.Query{}, err
			} else {
				// Select the expansion components for the projection statement
				expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
					Source: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier},
						Binding: models.EmptyOptional[pgsql.Identifier](),
					},
					Joins: []pgsql.Join{{
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
							Binding: models.ValueOptional(traversalStep.Edge.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType: pgsql.JoinTypeInner,
							Constraint: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
							),
						},
					}, {
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
							Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType: pgsql.JoinTypeInner,
							Constraint: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
							),
						},
					}, {
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
							Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType:   pgsql.JoinTypeInner,
							Constraint: wrappedSelectJoinConstraint,
						},
					}},
				})
			}
		}

		// If there are terminal constraints, project them as part of the projections
		if terminalNodeConstraints.Expression != nil {
			if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](terminalNodeConstraints.Expression); err != nil {
				return pgsql.Query{}, err
			} else {
				expansion.PrimerStatement.Projection = []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewLiteral(1, pgsql.Int),
					terminalCriteriaProjection,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					),
					pgsql.ArrayLiteral{
						Values: []pgsql.Expression{
							pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						},
					},
				}

				expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.NewLiteral(1, pgsql.Int),
					),
					terminalCriteriaProjection,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					),
				}

				// Constraints that target the terminal node may crop up here where it's finally in scope. Additionally,
				// only accept paths that are marked satisfied from the recursive descent CTE
				if constraints, err := consumeConstraintsFrom(traversalStep.Expansion.Value.Frame.Visible, s.treeTranslator.IdentifierConstraints); err != nil {
					return pgsql.Query{}, err
				} else if projectionConstraints, err := ConjoinExpressions([]pgsql.Expression{pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionSatisfied}, constraints.Expression}); err != nil {
					return pgsql.Query{}, err
				} else {
					expansion.ProjectionStatement.Where = projectionConstraints
				}
			}
		} else {
			expansion.PrimerStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				pgsql.NewLiteral(1, pgsql.Int),
				pgsql.NewLiteral(false, pgsql.Boolean),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				),
				pgsql.ArrayLiteral{
					Values: []pgsql.Expression{
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					},
				},
			}

			expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				pgsql.NewLiteral(false, pgsql.Boolean),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}
		}
	}

	return expansion.Build(traversalStep.Expansion.Value.Binding.Identifier), nil
}

func (s *Translator) buildExpansionPatternStep(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	var (
		expansion = ExpansionBuilder{
			Query: pgsql.Query{
				CommonTableExpressions: &pgsql.With{
					Recursive: true,
				},
			},

			PrimerStatement: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewLiteral(1, pgsql.Int),
					pgsql.NewLiteral(false, pgsql.Boolean),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					),
					pgsql.ArrayLiteral{
						Values: []pgsql.Expression{
							pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						},
					},
				},
			},

			RecursiveStatement: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.NewLiteral(1, pgsql.Int),
					),
					pgsql.NewLiteral(false, pgsql.Boolean),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, pgsql.ColumnPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, pgsql.ColumnPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					),
				},

				Where: expansionConstraints(traversalStep.Expansion.Value.Binding.Identifier),
			},
		}
	)

	expansion.ProjectionStatement.Projection = traversalStep.Expansion.Value.Projection

	if terminalNode, err := traversalStep.TerminalNode(); err != nil {
		return pgsql.Query{}, err
	} else if terminalNodeConstraints, err := consumeConstraintsFrom(pgsql.AsIdentifierSet(terminalNode.Identifier), s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else if edgeConstraints, err := consumeConstraintsFrom(traversalStep.Expansion.Value.Frame.Visible, s.treeTranslator.IdentifierConstraints, part.Constraints); err != nil {
		return pgsql.Query{}, err
	} else {
		if rightNodeJoinCondition, err := rightNodeTraversalStepConstraint(traversalStep); err != nil {
			return pgsql.Query{}, err
		} else {
			if err := rewriteIdentifierReferences(traversalStep.Expansion.Value.Frame, []pgsql.Expression{edgeConstraints.Expression, rightNodeJoinCondition}); err != nil {
				return pgsql.Query{}, err
			}

			expansion.PrimerStatement.From = append(expansion.PrimerStatement.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Previous.Binding.Identifier},
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
						Binding: models.ValueOptional(traversalStep.Edge.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: edgeConstraints.Expression,
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})

			// Make sure the recursive query has the expansion bound
			expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier},
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
						Binding: models.ValueOptional(traversalStep.Edge.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType: pgsql.JoinTypeInner,
						Constraint: pgsql.NewBinaryExpression(
							// TODO: Directional
							pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
							pgsql.OperatorEquals,
							pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, pgsql.ColumnNextID},
						),
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
						Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType:   pgsql.JoinTypeInner,
						Constraint: rightNodeJoinCondition,
					},
				}},
			})

			if wrappedSelectJoinConstraint, err := ConjoinExpressions([]pgsql.Expression{
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					&pgsql.ArrayIndex{
						Expression: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
						Indexes: []pgsql.Expression{
							pgsql.FunctionCall{
								Function: pgsql.FunctionArrayLength,
								Parameters: []pgsql.Expression{
									pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
									pgsql.NewLiteral(1, pgsql.Int8),
								},
								CastType: pgsql.Int4,
							},
						},
					},
				),
				rightNodeJoinCondition}); err != nil {
				return pgsql.Query{}, err
			} else {
				// Select the expansion components for the projection statement
				expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
					Source: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Previous.Binding.Identifier},
						Binding: models.EmptyOptional[pgsql.Identifier](),
					},
				})

				expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
					Source: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier},
						Binding: models.EmptyOptional[pgsql.Identifier](),
					},
					Joins: []pgsql.Join{{
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
							Binding: models.ValueOptional(traversalStep.Edge.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType: pgsql.JoinTypeInner,
							Constraint: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
							),
						},
					}, {
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
							Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType: pgsql.JoinTypeInner,
							Constraint: pgsql.NewBinaryExpression(
								pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
								pgsql.OperatorEquals,
								pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
							),
						},
					}, {
						Table: pgsql.TableReference{
							Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
							Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
						},
						JoinOperator: pgsql.JoinOperator{
							JoinType:   pgsql.JoinTypeInner,
							Constraint: wrappedSelectJoinConstraint,
						},
					}},
				})
			}
		}

		// If there are terminal constraints, project them as part of the recursive lookup
		if terminalNodeConstraints.Expression != nil {
			if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](terminalNodeConstraints.Expression); err != nil {
				return pgsql.Query{}, err
			} else {
				expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionRootID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.NewLiteral(1, pgsql.Int),
					),
					terminalCriteriaProjection,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Binding.Identifier, expansionPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					),
				}
			}
		}
	}

	return pgsql.Query{
		Body: expansion.Build(traversalStep.Expansion.Value.Binding.Identifier),
	}, nil
}
