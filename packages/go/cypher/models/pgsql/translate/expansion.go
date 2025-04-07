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

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
	"github.com/specterops/bloodhound/cypher/models/pgsql/pgd"
)

const translateDefaultMaxTraversalDepth int64 = 15

func expansionEdgeJoinCondition(traversalStep *TraversalStep) (pgsql.Expression, error) {
	return pgd.Equals(
		pgd.EntityID(traversalStep.LeftNode.Identifier),
		traversalStep.Expansion.Value.EdgeStartColumn,
	), nil
}

func expansionConstraints(traversalStep *TraversalStep) pgsql.Expression {
	return pgd.And(
		pgd.LessThan(
			pgd.Column(traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth),
			pgd.IntLiteral(traversalStep.Expansion.Value.MaxDepth.GetOr(translateDefaultMaxTraversalDepth)),
		),
		pgd.Not(
			pgd.Column(traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionIsCycle),
		),
	)
}

var (
	ErrUnsupportedExpansionDirection = errors.New("unsupported expansion direction")
)

type ExpansionBuilder struct {
	PrimerStatement     pgsql.Select
	RecursiveStatement  pgsql.Select
	ProjectionStatement pgsql.Select

	queryParameters map[string]any
	traversalStep   *TraversalStep
	model           *Expansion
}

func NewExpansionBuilder(queryParameters map[string]any, traversalStep *TraversalStep) (*ExpansionBuilder, error) {
	if !traversalStep.Expansion.Set {
		return nil, errors.New("traversal step must have expansion set")
	}

	return &ExpansionBuilder{
		queryParameters: queryParameters,
		traversalStep:   traversalStep,
		model:           traversalStep.Expansion.Value,
	}, nil
}

func nextFrontInsert(body pgsql.SetExpression) pgsql.Insert {
	return pgsql.Insert{
		Table: pgsql.TableReference{
			Name: expansionNextFront.AsCompoundIdentifier(),
		},
		Shape: expansionColumns(),
		Source: &pgsql.Query{
			Body: body,
		},
	}
}

func (s *ExpansionBuilder) prepareForwardFrontPrimerQuery(expansionModel *Expansion) pgsql.Select {
	nextQuery := pgsql.Select{
		Where: pgsql.OptionalAnd(expansionModel.PrimerNodeConstraints, expansionModel.EdgeConstraints),
	}

	nextQuery.Projection = []pgsql.SelectItem{
		s.model.EdgeStartColumn,
		s.model.EdgeEndColumn,
		pgd.IntLiteral(1),
	}

	if expansionModel.TerminalNodeSatisfactionProjection != nil {
		nextQuery.Projection = append(nextQuery.Projection, expansionModel.TerminalNodeSatisfactionProjection)
	} else {
		nextQuery.Projection = append(nextQuery.Projection, pgsql.ExistsExpression{
			Subquery: pgsql.Subquery{
				Query: pgsql.Query{
					Body: pgsql.Select{
						Projection: []pgsql.SelectItem{
							pgd.IntLiteral(1),
						},
						From: []pgsql.FromClause{{
							Source: pgsql.TableReference{
								Name: pgsql.TableEdge.AsCompoundIdentifier(),
							},
						}},
						Where: pgd.Equals(
							expansionModel.EdgeEndIdentifier,
							expansionModel.EdgeStartColumn,
						),
					},
				},
			},
			Negated: false,
		})
	}

	nextQuery.Projection = append(nextQuery.Projection,
		pgd.Equals(
			pgd.StartID(s.traversalStep.Edge.Identifier),
			pgd.EndID(s.traversalStep.Edge.Identifier),
		),
		pgd.ExpressionArrayLiteral(
			pgd.EntityID(s.traversalStep.Edge.Identifier),
		),
	)

	nextQueryFrom := pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
			Binding: models.ValueOptional(s.traversalStep.Edge.Identifier),
		},
	}

	if expansionModel.PrimerNodeConstraints != nil {
		nextQueryFrom.Joins = append(nextQueryFrom.Joins, pgsql.Join{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: s.traversalStep.Expansion.Value.PrimerNodeJoinCondition,
			},
		})
	}

	if expansionModel.TerminalNodeConstraints != nil {
		nextQueryFrom.Joins = append(nextQueryFrom.Joins, pgsql.Join{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: s.traversalStep.Expansion.Value.ExpansionNodeJoinCondition,
			},
		})
	}

	nextQuery.From = []pgsql.FromClause{nextQueryFrom}
	return nextQuery
}

func (s *ExpansionBuilder) prepareForwardFrontRecursiveQuery(expansionModel *Expansion) pgsql.Select {
	nextQuery := pgsql.Select{
		Where: expansionModel.EdgeConstraints,
	}

	nextQuery.Projection = []pgsql.SelectItem{
		pgd.Column(expansionModel.Frame.Binding.Identifier, expansionRootID),
		s.model.EdgeEndColumn,
		pgd.Add(
			pgd.Column(expansionModel.Frame.Binding.Identifier, expansionDepth),
			pgd.IntLiteral(1)),
	}

	if expansionModel.TerminalNodeSatisfactionProjection != nil {
		nextQuery.Projection = append(nextQuery.Projection, expansionModel.TerminalNodeSatisfactionProjection)
	} else {
		nextQuery.Projection = append(nextQuery.Projection, pgsql.ExistsExpression{
			Subquery: pgsql.Subquery{
				Query: pgsql.Query{
					Body: pgsql.Select{
						Projection: []pgsql.SelectItem{
							pgd.IntLiteral(1),
						},
						From: []pgsql.FromClause{{
							Source: pgsql.TableReference{
								Name: pgsql.TableEdge.AsCompoundIdentifier(),
							},
						}},
						Where: pgd.Equals(
							expansionModel.EdgeEndIdentifier,
							expansionModel.EdgeStartColumn,
						),
					},
				},
			},
			Negated: false,
		})
	}

	nextQuery.Projection = append(nextQuery.Projection, pgd.Equals(
		pgd.EntityID(s.traversalStep.Edge.Identifier),
		pgd.Any(pgd.Column(expansionModel.Frame.Binding.Identifier, expansionPath), pgsql.ExpansionPath),
	))

	nextQuery.Projection = append(nextQuery.Projection, pgd.Concatenate(
		pgd.Column(expansionModel.Frame.Binding.Identifier, expansionPath),
		pgd.EntityID(s.traversalStep.Edge.Identifier),
	))

	nextQueryFrom := pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{expansionForwardFront},
			Binding: models.ValueOptional(expansionModel.Frame.Binding.Identifier),
		},

		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(s.traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					s.model.EdgeStartColumn,
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	}

	if expansionModel.TerminalNodeConstraints != nil {
		nextQueryFrom.Joins = append(nextQueryFrom.Joins, pgsql.Join{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: s.traversalStep.Expansion.Value.ExpansionNodeJoinCondition,
			},
		})
	}

	nextQuery.From = []pgsql.FromClause{nextQueryFrom}
	return nextQuery
}

func (s *ExpansionBuilder) prepareBackwardFrontPrimerQuery(expansionModel *Expansion) pgsql.Select {
	nextQuery := pgsql.Select{
		Where: pgsql.OptionalAnd(expansionModel.TerminalNodeConstraints, expansionModel.EdgeConstraints),
	}

	nextQuery.Projection = []pgsql.SelectItem{
		s.model.EdgeEndColumn,
		s.model.EdgeStartColumn,
		pgd.IntLiteral(1),
	}

	if expansionModel.PrimerNodeSatisfactionProjection != nil {
		nextQuery.Projection = append(nextQuery.Projection, expansionModel.PrimerNodeSatisfactionProjection)
	} else {
		nextQuery.Projection = append(nextQuery.Projection, pgsql.ExistsExpression{
			Subquery: pgsql.Subquery{
				Query: pgsql.Query{
					Body: pgsql.Select{
						Projection: []pgsql.SelectItem{
							pgd.IntLiteral(1),
						},
						From: []pgsql.FromClause{{
							Source: pgsql.TableReference{
								Name: pgsql.TableEdge.AsCompoundIdentifier(),
							},
						}},
						Where: pgd.Equals(
							expansionModel.EdgeStartIdentifier,
							expansionModel.EdgeEndColumn,
						),
					},
				},
			},
			Negated: false,
		})
	}

	nextQuery.Projection = append(nextQuery.Projection,
		pgd.Equals(
			pgd.StartID(s.traversalStep.Edge.Identifier),
			pgd.EndID(s.traversalStep.Edge.Identifier),
		),
		pgd.ExpressionArrayLiteral(
			pgd.EntityID(s.traversalStep.Edge.Identifier),
		),
	)

	nextQueryFrom := pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
			Binding: models.ValueOptional(s.traversalStep.Edge.Identifier),
		},
	}

	if expansionModel.PrimerNodeConstraints != nil {
		nextQueryFrom.Joins = append(nextQueryFrom.Joins, pgsql.Join{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: s.traversalStep.Expansion.Value.PrimerNodeJoinCondition,
			},
		})
	}

	if expansionModel.TerminalNodeConstraints != nil {
		nextQueryFrom.Joins = append(nextQueryFrom.Joins, pgsql.Join{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: s.traversalStep.Expansion.Value.ExpansionNodeJoinCondition,
			},
		})
	}

	nextQuery.From = []pgsql.FromClause{nextQueryFrom}
	return nextQuery
}

func (s *ExpansionBuilder) prepareBackwardFrontRecursiveQuery(expansionModel *Expansion) pgsql.Select {
	nextQuery := pgsql.Select{
		Where: expansionModel.EdgeConstraints,
	}

	nextQuery.Projection = []pgsql.SelectItem{
		pgd.Column(expansionModel.Frame.Binding.Identifier, expansionRootID),
		s.model.EdgeStartColumn,
		pgd.Add(
			pgd.Column(expansionModel.Frame.Binding.Identifier, expansionDepth),
			pgd.IntLiteral(1)),
	}

	if expansionModel.PrimerNodeSatisfactionProjection != nil {
		nextQuery.Projection = append(nextQuery.Projection, expansionModel.PrimerNodeSatisfactionProjection)
	} else {
		nextQuery.Projection = append(nextQuery.Projection, pgsql.ExistsExpression{
			Subquery: pgsql.Subquery{
				Query: pgsql.Query{
					Body: pgsql.Select{
						Projection: []pgsql.SelectItem{
							pgd.IntLiteral(1),
						},
						From: []pgsql.FromClause{{
							Source: pgsql.TableReference{
								Name: pgsql.TableEdge.AsCompoundIdentifier(),
							},
						}},
						Where: pgd.Equals(
							expansionModel.EdgeStartIdentifier,
							expansionModel.EdgeEndColumn,
						),
					},
				},
			},
			Negated: false,
		})
	}

	nextQuery.Projection = append(nextQuery.Projection, pgd.Equals(
		pgd.EntityID(s.traversalStep.Edge.Identifier),
		pgd.Any(pgd.Column(expansionModel.Frame.Binding.Identifier, expansionPath), pgsql.ExpansionPath),
	))

	nextQuery.Projection = append(nextQuery.Projection, pgd.Concatenate(
		pgd.EntityID(s.traversalStep.Edge.Identifier),
		pgd.Column(expansionModel.Frame.Binding.Identifier, expansionPath),
	))

	nextQueryFrom := pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{expansionBackwardFront},
			Binding: models.ValueOptional(expansionModel.Frame.Binding.Identifier),
		},

		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(s.traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					s.model.EdgeEndColumn,
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	}

	if expansionModel.PrimerNodeConstraints != nil {
		nextQueryFrom.Joins = append(nextQueryFrom.Joins, pgsql.Join{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: s.traversalStep.Expansion.Value.PrimerNodeJoinCondition,
			},
		})
	}

	nextQuery.From = []pgsql.FromClause{nextQueryFrom}
	return nextQuery
}

func (s *ExpansionBuilder) BuildAllShortestPathsRoot() (pgsql.Query, error) {
	var (
		expansionModel             = s.traversalStep.Expansion.Value
		forwardFrontPrimerQuery    = s.prepareForwardFrontPrimerQuery(expansionModel)
		forwardFrontRecursiveQuery = s.prepareForwardFrontRecursiveQuery(expansionModel)
		projectionQuery            pgsql.Select
	)

	projectionQuery.Projection = expansionModel.Projection

	// Select the expansion components for the projection statement
	projectionQuery.From = []pgsql.FromClause{{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier},
			Binding: models.EmptyOptional[pgsql.Identifier](),
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{s.traversalStep.LeftNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{s.traversalStep.RightNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	}}

	if harnessParameters, err := s.allShortestPathsParameters(expansionModel, forwardFrontPrimerQuery, forwardFrontRecursiveQuery); err != nil {
		return pgsql.Query{}, err
	} else {
		query := pgsql.Query{
			CommonTableExpressions: &pgsql.With{},
			Body:                   projectionQuery,
		}

		query.AddCTE(pgsql.CommonTableExpression{
			Alias: pgsql.TableAlias{
				Name:  expansionModel.Frame.Binding.Identifier,
				Shape: models.ValueOptional(expansionColumns()),
			},
			Query: pgsql.Query{
				Body: pgsql.Select{
					Projection: []pgsql.SelectItem{
						pgsql.Wildcard{},
					},
					From: []pgsql.FromClause{{
						Source: pgsql.FunctionCall{
							Function:   pgsql.FunctionUnidirectionalASPHarness,
							Parameters: harnessParameters,
						},
					}},
				},
			},
		})

		return query, nil
	}
}

func (s *ExpansionBuilder) BuildBiDirectionalAllShortestPathsRoot() (pgsql.Query, error) {
	var (
		expansionModel              = s.traversalStep.Expansion.Value
		forwardFrontPrimerQuery     = s.prepareForwardFrontPrimerQuery(expansionModel)
		forwardFrontRecursiveQuery  = s.prepareForwardFrontRecursiveQuery(expansionModel)
		backwardFrontPrimerQuery    = s.prepareBackwardFrontPrimerQuery(expansionModel)
		backwardFrontRecursiveQuery = s.prepareBackwardFrontRecursiveQuery(expansionModel)
		projectionQuery             pgsql.Select
	)

	projectionQuery.Projection = expansionModel.Projection

	// Select the expansion components for the projection statement
	projectionQuery.From = []pgsql.FromClause{{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier},
			Binding: models.EmptyOptional[pgsql.Identifier](),
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{s.traversalStep.LeftNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(s.traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{s.traversalStep.RightNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	}}

	if harnessParameters, err := s.bidirectionalAllShortestPathsParameters(expansionModel, forwardFrontPrimerQuery, forwardFrontRecursiveQuery, backwardFrontPrimerQuery, backwardFrontRecursiveQuery); err != nil {
		return pgsql.Query{}, err
	} else {
		query := pgsql.Query{
			CommonTableExpressions: &pgsql.With{},
			Body:                   projectionQuery,
		}

		query.AddCTE(pgsql.CommonTableExpression{
			Alias: pgsql.TableAlias{
				Name:  expansionModel.Frame.Binding.Identifier,
				Shape: models.ValueOptional(expansionColumns()),
			},
			Query: pgsql.Query{
				Body: pgsql.Select{
					Projection: []pgsql.SelectItem{
						pgsql.Wildcard{},
					},
					From: []pgsql.FromClause{{
						Source: pgsql.FunctionCall{
							Function:   pgsql.FunctionBidirectionalASPHarness,
							Parameters: harnessParameters,
						},
					}},
				},
			},
		})

		return query, nil
	}
}

func (s *ExpansionBuilder) allShortestPathsParameters(expansionModel *Expansion, forwardFrontPrimerQuery pgsql.Select, forwardFrontRecursiveQuery pgsql.Select) ([]pgsql.Expression, error) {
	var (
		harnessParameters []pgsql.Expression
		formatFragment    = func(query pgsql.Select) (string, error) {
			return format.Statement(
				nextFrontInsert(query),
				format.NewOutputBuilder().WithMaterializedParameters(s.queryParameters))
		}
	)

	if formattedQuery, err := formatFragment(forwardFrontPrimerQuery); err != nil {
		return nil, err
	} else {
		// Put this in the translation's parameter bag which is transmitted down to the DB
		s.queryParameters[expansionModel.PrimerQueryParameter.Identifier.String()] = formattedQuery

		// Track this as a function parameter for the harness
		harnessParameters = append(harnessParameters, &pgsql.Parameter{
			Identifier: expansionModel.PrimerQueryParameter.Identifier,
			CastType:   pgsql.Text,
		})
	}

	if formattedQuery, err := formatFragment(forwardFrontRecursiveQuery); err != nil {
		return nil, err
	} else {
		s.queryParameters[expansionModel.RecursiveQueryParameter.Identifier.String()] = formattedQuery
		harnessParameters = append(harnessParameters, &pgsql.Parameter{
			Identifier: expansionModel.RecursiveQueryParameter.Identifier,
			CastType:   pgsql.Text,
		})
	}

	return append(harnessParameters, pgsql.NewLiteral(expansionModel.MaxDepth.GetOr(translateDefaultMaxTraversalDepth), pgsql.Int)), nil
}

func (s *ExpansionBuilder) bidirectionalAllShortestPathsParameters(expansionModel *Expansion, forwardFrontPrimerQuery pgsql.Select, forwardFrontRecursiveQuery pgsql.Select, backwardFrontPrimerQuery pgsql.Select, backwardFrontRecursiveQuery pgsql.Select) ([]pgsql.Expression, error) {
	var (
		harnessParameters []pgsql.Expression
		formatFragment    = func(query pgsql.Select) (string, error) {
			return format.Statement(
				nextFrontInsert(query),
				format.NewOutputBuilder().WithMaterializedParameters(s.queryParameters))
		}
	)

	if formattedQuery, err := formatFragment(forwardFrontPrimerQuery); err != nil {
		return nil, err
	} else {
		// Put this in the translation's parameter bag which is transmitted down to the DB
		s.queryParameters[expansionModel.PrimerQueryParameter.Identifier.String()] = formattedQuery

		// Track this as a function parameter for the harness
		harnessParameters = append(harnessParameters, &pgsql.Parameter{
			Identifier: expansionModel.PrimerQueryParameter.Identifier,
			CastType:   pgsql.Text,
		})
	}

	if formattedQuery, err := formatFragment(forwardFrontRecursiveQuery); err != nil {
		return nil, err
	} else {
		s.queryParameters[expansionModel.RecursiveQueryParameter.Identifier.String()] = formattedQuery
		harnessParameters = append(harnessParameters, &pgsql.Parameter{
			Identifier: expansionModel.RecursiveQueryParameter.Identifier,
			CastType:   pgsql.Text,
		})
	}

	if formattedQuery, err := formatFragment(backwardFrontPrimerQuery); err != nil {
		return nil, err
	} else {
		s.queryParameters[expansionModel.BackwardPrimerQueryParameter.Identifier.String()] = formattedQuery
		harnessParameters = append(harnessParameters, &pgsql.Parameter{
			Identifier: expansionModel.BackwardPrimerQueryParameter.Identifier,
			CastType:   pgsql.Text,
		})
	}

	if formattedQuery, err := formatFragment(backwardFrontRecursiveQuery); err != nil {
		return nil, err
	} else {
		s.queryParameters[expansionModel.BackwardRecursiveQueryParameter.Identifier.String()] = formattedQuery
		harnessParameters = append(harnessParameters, &pgsql.Parameter{
			Identifier: expansionModel.BackwardRecursiveQueryParameter.Identifier,
			CastType:   pgsql.Text,
		})
	}

	return append(harnessParameters, pgsql.NewLiteral(expansionModel.MaxDepth.GetOr(translateDefaultMaxTraversalDepth), pgsql.Int)), nil
}

func (s *ExpansionBuilder) BuildAllShortestPathsStep() pgsql.Query {
	return pgsql.Query{}
}

func (s *ExpansionBuilder) BuildAllShortestPathsQuery(primerIdentifier, recursiveIdentifier *pgsql.Parameter, expansionIdentifier pgsql.Identifier, maxDepth int64) pgsql.Query {
	query := pgsql.Query{
		CommonTableExpressions: &pgsql.With{},
		Body:                   s.ProjectionStatement,
	}

	query.AddCTE(pgsql.CommonTableExpression{
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
						Function: pgsql.FunctionUnidirectionalASPHarness,
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

	return query
}

func (s *ExpansionBuilder) Build(expansionIdentifier pgsql.Identifier) pgsql.Query {
	query := pgsql.Query{
		CommonTableExpressions: &pgsql.With{
			Recursive: true,
		},
		Body: s.ProjectionStatement,
	}

	query.AddCTE(pgsql.CommonTableExpression{
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

	return query
}

func (s *Translator) buildExpansionPatternRoot(traversalStep *TraversalStep, expansion *ExpansionBuilder) (pgsql.Query, error) {
	expansionModel := traversalStep.Expansion.Value

	expansion.ProjectionStatement.Projection = expansionModel.Projection
	expansion.PrimerStatement.Where = pgsql.OptionalAnd(expansionModel.PrimerNodeConstraints, expansionModel.EdgeConstraints)
	expansion.RecursiveStatement.Where = pgsql.OptionalAnd(expansionModel.EdgeConstraints, expansionModel.RecursiveConstraints)

	// If the left node was already bound at time of translation connect this expansion to the
	// previously materialized node
	if traversalStep.LeftNodeBound {
		expansion.PrimerStatement.From = append(expansion.PrimerStatement.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
			},
			Joins: []pgsql.Join{{
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
					Binding: models.ValueOptional(traversalStep.Edge.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType: pgsql.JoinTypeInner,
					Constraint: pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID},
						pgsql.OperatorEquals,
						rewriteCompositeTypeFieldReference(
							traversalStep.Frame.Previous.Binding.Identifier,
							pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
						)),
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: expansionModel.ExpansionNodeJoinCondition,
				},
			}},
		})
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
					Constraint: expansionModel.PrimerNodeJoinCondition,
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: expansionModel.ExpansionNodeJoinCondition,
				},
			}},
		})
	}

	// Make sure the recursive query has the expansion bound
	expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier},
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					expansionModel.EdgeStartColumn,
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: expansionModel.ExpansionNodeJoinCondition,
			},
		}},
	})

	// The current query part may not have a frame associated with it if is a single part query component
	if traversalStep.Frame.Previous != nil && (s.query.CurrentPart().Frame == nil || traversalStep.Frame.Previous.Binding.Identifier != s.query.CurrentPart().Frame.Binding.Identifier) {
		expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
			Source: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
				Binding: models.EmptyOptional[pgsql.Identifier](),
			},
		})
	}

	// Select the expansion components for the projection statement
	expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier},
			Binding: models.EmptyOptional[pgsql.Identifier](),
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.RightNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	})

	// If there are right node constraints, project them as part of the primer statement's projection
	if expansionModel.TerminalNodeSatisfactionProjection != nil {
		if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](expansionModel.TerminalNodeSatisfactionProjection); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.PrimerStatement.Projection = []pgsql.SelectItem{
				expansionModel.EdgeStartColumn,
				expansionModel.EdgeEndColumn,
				pgsql.NewLiteral(1, pgsql.Int),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					expansionModel.EdgeStartColumn,
					pgsql.OperatorEquals,
					expansionModel.EdgeEndColumn,
				),
				pgsql.ArrayLiteral{
					Values: []pgsql.Expression{
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					},
				},
			}

			expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
				expansionModel.EdgeEndColumn,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionPath}, pgsql.ExpansionPath),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}

			// Constraints that target the terminal node may crop up here where it's finally in scope. Additionally,
			// only accept paths that are marked satisfied from the recursive descent CTE
			if constraints, err := s.treeTranslator.IdentifierConstraints.ConsumeSet(expansionModel.Frame.Visible); err != nil {
				return pgsql.Query{}, err
			} else if projectionConstraints, err := ConjoinExpressions([]pgsql.Expression{pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionSatisfied}, constraints.Expression}); err != nil {
				return pgsql.Query{}, err
			} else {
				expansion.ProjectionStatement.Where = projectionConstraints
			}
		}
	} else {
		expansion.PrimerStatement.Projection = []pgsql.SelectItem{
			expansionModel.EdgeStartColumn,
			expansionModel.EdgeEndColumn,
			pgsql.NewLiteral(1, pgsql.Int),
			pgsql.NewLiteral(false, pgsql.Boolean),
			pgsql.NewBinaryExpression(
				expansionModel.EdgeStartColumn,
				pgsql.OperatorEquals,
				expansionModel.EdgeEndColumn,
			),
			pgsql.ArrayLiteral{
				Values: []pgsql.Expression{
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				},
			},
		}

		expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
			pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
			expansionModel.EdgeEndColumn,
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionDepth},
				pgsql.OperatorAdd,
				pgsql.NewLiteral(1, pgsql.Int),
			),
			pgsql.NewLiteral(false, pgsql.Boolean),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.NewAnyExpression(pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionPath}, pgsql.ExpansionPath),
			),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionPath},
				pgsql.OperatorConcatenate,
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
			),
		}
	}

	return expansion.Build(expansionModel.Frame.Binding.Identifier), nil
}

func (s *Translator) buildExpansionPatternStep(traversalStep *TraversalStep, expansion *ExpansionBuilder) (pgsql.Query, error) {
	expansionModel := traversalStep.Expansion.Value

	expansion.PrimerStatement = pgsql.Select{
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
	}

	expansion.RecursiveStatement = pgsql.Select{
		Projection: []pgsql.SelectItem{
			pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
			pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionDepth},
				pgsql.OperatorAdd,
				pgsql.NewLiteral(1, pgsql.Int),
			),
			pgsql.NewLiteral(false, pgsql.Boolean),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.NewAnyExpression(pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, pgsql.ColumnPath}, pgsql.ExpansionPath),
			),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, pgsql.ColumnPath},
				pgsql.OperatorConcatenate,
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
			),
		},
	}

	expansion.ProjectionStatement.Projection = expansionModel.Projection
	expansion.PrimerStatement.Where = pgsql.OptionalAnd(expansionModel.PrimerNodeConstraints, expansionModel.EdgeConstraints)
	expansion.RecursiveStatement.Where = pgsql.OptionalAnd(expansionModel.EdgeConstraints, expansionModel.RecursiveConstraints)

	expansion.PrimerStatement.From = append(expansion.PrimerStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: expansionModel.EdgeJoinCondition,
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: expansionModel.ExpansionNodeJoinCondition,
			},
		}},
	})

	// Make sure the recursive query has the expansion bound
	expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier},
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					expansionModel.EdgeStartColumn,
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: expansionModel.ExpansionNodeJoinCondition,
			},
		}},
	})

	// Select the expansion components for the projection statement
	expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{traversalStep.Frame.Previous.Binding.Identifier},
			Binding: models.EmptyOptional[pgsql.Identifier](),
		},
	})

	expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier},
			Binding: models.EmptyOptional[pgsql.Identifier](),
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.LeftNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.LeftNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.RightNode.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	})

	// If there are terminal constraints, project them as part of the recursive lookup
	if expansionModel.TerminalNodeSatisfactionProjection != nil {
		if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](expansionModel.TerminalNodeSatisfactionProjection); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionRootID},
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionPath}, pgsql.ExpansionPath),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{expansionModel.Frame.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}
		}
	}

	return pgsql.Query{
		Body: expansion.Build(expansionModel.Frame.Binding.Identifier),
	}, nil
}

func (s *Translator) translateTraversalPatternPartWithExpansion(isFirstTraversalStep bool, isShortestPath bool, traversalStep *TraversalStep) error {
	expansionModel := traversalStep.Expansion.Value

	// Translate the expansion's constraints - this has the side effect of making the pattern identifiers visible in
	// the current scope frame
	if err := s.translateExpansionConstraints(isFirstTraversalStep, traversalStep, expansionModel); err != nil {
		return err
	}

	// Export the path from the traversal's scope
	traversalStep.Frame.Export(expansionModel.PathBinding.Identifier)

	// Push a new frame that contains currently projected scope from the expansion recursive CTE
	if expansionFrame, err := s.scope.PushFrame(); err != nil {
		return err
	} else {
		expansionModel.Frame = expansionFrame
	}

	// Expansion edge join condition
	expansionModel.RecursiveConstraints = expansionConstraints(traversalStep)

	if err := RewriteFrameBindings(s.scope, expansionModel.RecursiveConstraints); err != nil {
		return err
	}

	// Remove the previous projections of the root and terminal node to reproject them after expansion
	traversalStep.LeftNode.Dematerialize()
	traversalStep.RightNode.Dematerialize()

	if boundProjections, err := buildVisibleProjections(s.scope); err != nil {
		return err
	} else {
		// Zip through all projected identifiers and update their last projected frame
		for _, binding := range boundProjections.Bindings {
			binding.MaterializedBy(expansionModel.Frame)
		}

		expansionModel.Projection = boundProjections.Items
	}

	if err := s.scope.PopFrame(); err != nil {
		return err
	}

	if boundProjections, err := buildVisibleProjections(s.scope); err != nil {
		return err
	} else {
		// Zip through all projected identifiers and update their last projected frame
		for _, binding := range boundProjections.Bindings {
			binding.MaterializedBy(traversalStep.Frame)
		}

		traversalStep.Projection = boundProjections.Items
	}

	if isShortestPath {
		if err := s.translateShortestPathTraversal(expansionModel); err != nil {
			return err
		}
	}

	return nil
}

func (s *Translator) translateExpansionConstraints(isFirstTraversalStep bool, step *TraversalStep, expansionModel *Expansion) error {
	if constraints, err := consumePatternConstraints(isFirstTraversalStep, recursivePattern, step, s.treeTranslator.IdentifierConstraints); err != nil {
		return err
	} else {
		// If one side of the expansion has constraints but the other does not this may be an opportunity to reorder the traversal
		// to start with tighter search bounds
		if err := constraints.OptimizePatternConstraintBalance(s.scope, step); err != nil {
			return err
		}

		// Left node
		if leftNodeJoinCondition, err := leftNodeTraversalStepConstraint(step); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.scope, leftNodeJoinCondition); err != nil {
			return err
		} else {
			expansionModel.PrimerNodeJoinCondition = leftNodeJoinCondition
		}

		if constraints.LeftNode.Expression != nil {
			if err := RewriteFrameBindings(s.scope, constraints.LeftNode.Expression); err != nil {
				return err
			}

			expansionModel.PrimerNodeConstraints = constraints.LeftNode.Expression

			if primerCriteriaProjection, err := pgsql.As[pgsql.SelectItem](expansionModel.PrimerNodeConstraints); err != nil {
				return err
			} else {
				expansionModel.PrimerNodeSatisfactionProjection = primerCriteriaProjection
			}
		}

		// Expansion edge constraints
		if constraints.Edge.Expression != nil {
			expansionModel.EdgeConstraints = constraints.Edge.Expression

			if err := RewriteFrameBindings(s.scope, expansionModel.EdgeConstraints); err != nil {
				return err
			}
		}

		if !isFirstTraversalStep {
			if edgeJoinCondition, err := expansionEdgeJoinCondition(step); err != nil {
				return err
			} else if err := RewriteFrameBindings(s.scope, edgeJoinCondition); err != nil {
				return err
			} else {
				expansionModel.EdgeJoinCondition = edgeJoinCondition
			}
		}

		// Right node
		if rightNodeJoinCondition, err := rightNodeTraversalStepJoinCondition(step); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.scope, rightNodeJoinCondition); err != nil {
			return err
		} else {
			expansionModel.ExpansionNodeJoinCondition = rightNodeJoinCondition
		}

		if constraints.RightNode.Expression != nil {
			if err := RewriteFrameBindings(s.scope, constraints.RightNode.Expression); err != nil {
				return err
			} else {
				expansionModel.TerminalNodeConstraints = constraints.RightNode.Expression

				if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](expansionModel.TerminalNodeConstraints); err != nil {
					return err
				} else {
					expansionModel.TerminalNodeSatisfactionProjection = terminalCriteriaProjection
				}
			}
		}
	}

	return nil
}

func (s *Translator) translateShortestPathTraversal(expansionModel *Expansion) error {
	// If this query is a shortest-path look up, the translator will have to use a function harness for
	// traversal. As such, query fragments for the traversal harness will have to be passed by the parameters
	// defined below.
	if primerQueryParameter, err := s.scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
		return err
	} else {
		expansionModel.PrimerQueryParameter = primerQueryParameter
	}

	if recursiveQueryParameter, err := s.scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
		return err
	} else {
		expansionModel.RecursiveQueryParameter = recursiveQueryParameter
	}

	// Bidirectional BFS searches require an additional set of query fragments to represent the backward traversal
	// front of the search.
	if expansionModel.CanExecuteBidirectionalSearch() {
		if reversePrimerQueryParameter, err := s.scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
			return err
		} else {
			expansionModel.BackwardPrimerQueryParameter = reversePrimerQueryParameter
		}

		if reverseRecursiveQueryParameter, err := s.scope.DefineNew(pgsql.ParameterIdentifier); err != nil {
			return err
		} else {
			expansionModel.BackwardRecursiveQueryParameter = reverseRecursiveQueryParameter
		}
	}

	return nil
}

func (s *Translator) translateNonTraversalPatternPart(part *PatternPart) error {
	if nextFrame, err := s.scope.PushFrame(); err != nil {
		return err
	} else {
		part.NodeSelect.Frame = nextFrame

		nextFrame.Export(part.NodeSelect.Binding.Identifier)

		if constraint, err := s.treeTranslator.IdentifierConstraints.ConsumeSet(nextFrame.Known()); err != nil {
			return err
		} else if err := RewriteFrameBindings(s.scope, constraint.Expression); err != nil {
			return err
		} else {
			part.NodeSelect.Constraints = constraint.Expression
		}

		if boundProjections, err := buildVisibleProjections(s.scope); err != nil {
			return err
		} else {
			// Zip through all projected identifiers and update their last projected frame
			for _, binding := range boundProjections.Bindings {
				binding.MaterializedBy(nextFrame)
			}

			part.NodeSelect.Select.Projection = boundProjections.Items
		}
	}

	return nil
}
