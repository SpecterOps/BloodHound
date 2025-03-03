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
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

const translateDefaultMaxTraversalDepth int64 = 10

func expansionConstraints(expansionIdentifier pgsql.Identifier, minTraversalDepth models.Optional[int64], maxTraversalDepth models.Optional[int64]) pgsql.Expression {
	return pgsql.NewBinaryExpression(
		pgsql.NewBinaryExpression(
			pgsql.CompoundIdentifier{expansionIdentifier, expansionDepth},
			pgsql.OperatorLessThan,
			pgsql.NewLiteral(maxTraversalDepth.GetOr(translateDefaultMaxTraversalDepth), pgsql.Int),
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

func (s ExpansionBuilder) BuildAllShortestPaths(primerIdentifier, recursiveIdentifier *pgsql.Parameter, expansionIdentifier pgsql.Identifier, maxDepth int64) pgsql.Query {
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
		rootIDColumnIdentifier pgsql.SelectItem
		nextIDColumnIdentifier pgsql.SelectItem

		expansion = ExpansionBuilder{
			Query: pgsql.Query{
				CommonTableExpressions: &pgsql.With{},
			},

			RecursiveStatement: pgsql.Select{
				Where: expansionConstraints(traversalStep.Expansion.Value.Frame.Binding.Identifier, traversalStep.Expansion.Value.MinDepth, traversalStep.Expansion.Value.MaxDepth),
			},
		}
	)

	switch traversalStep.Direction {
	case graph.DirectionInbound:
		rootIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID}
		nextIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID}

	case graph.DirectionOutbound:
		rootIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID}
		nextIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID}

	default:
		return pgsql.Query{}, fmt.Errorf("unsupported expansion direction: %s", traversalStep.Direction.String())
	}

	expansion.ProjectionStatement.Projection = traversalStep.Expansion.Value.Projection

	expansion.PrimerStatement.Where = traversalStep.Expansion.Value.PrimerConstraints
	expansion.RecursiveStatement.Where = traversalStep.Expansion.Value.RecursiveConstraints

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
				Constraint: traversalStep.Expansion.Value.LeftNodeJoinCondition,
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
			},
		}},
	})

	// Make sure the recursive query has the expansion bound
	expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{"pathspace"},
			Binding: models.ValueOptional(traversalStep.Expansion.Value.Frame.Binding.Identifier),
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					rootIDColumnIdentifier,
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
			},
		}},
	})

	// Select the expansion components for the projection statement
	expansion.ProjectionStatement.From = append(expansion.ProjectionStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier},
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	})

	// If there are terminal constraints, project them as part of the projections
	if traversalStep.Expansion.Value.TerminalNodeConstraints != nil {
		if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](traversalStep.Expansion.Value.TerminalNodeConstraints); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.PrimerStatement.Projection = []pgsql.SelectItem{
				rootIDColumnIdentifier,
				nextIDColumnIdentifier,
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
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
				nextIDColumnIdentifier,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath}),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}

			// Constraints that target the terminal node may crop up here where it's finally in scope. Additionally,
			// only accept paths that are marked satisfied from the recursive descent CTE
			if constraints, err := s.treeTranslator.IdentifierConstraints.ConsumeSet(traversalStep.Expansion.Value.Frame.Visible); err != nil {
				return pgsql.Query{}, err
			} else if projectionConstraints, err := ConjoinExpressions([]pgsql.Expression{pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionSatisfied}, constraints.Expression}); err != nil {
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
			pgsql.ExistsExpression{
				Subquery: pgsql.Subquery{
					Query: pgsql.Query{
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
			pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
			pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth},
				pgsql.OperatorAdd,
				pgsql.NewLiteral(1, pgsql.Int),
			),
			pgsql.ExistsExpression{
				Subquery: pgsql.Subquery{
					Query: pgsql.Query{
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
				},
				Negated: false,
			},
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath}),
			),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath},
				pgsql.OperatorConcatenate,
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
			),
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

		var traversalDepthLimit = translateDefaultMaxTraversalDepth

		if traversalStep.Expansion.Value.MaxDepth.Set {
			traversalDepthLimit = traversalStep.Expansion.Value.MaxDepth.Value
		}

		return expansion.BuildAllShortestPaths(primerParameter, recursiveParameter, traversalStep.Expansion.Value.Frame.Binding.Identifier, traversalDepthLimit), nil
	}
}

func (s *Translator) buildExpansionPatternRoot(part *PatternPart, traversalStep *PatternSegment) (pgsql.Query, error) {
	if part.ShortestPath || part.AllShortestPaths {
		return s.buildAllShortestPathsExpansionRoot(part, traversalStep)
	}

	var (
		rootIDColumnIdentifier pgsql.SelectItem
		nextIDColumnIdentifier pgsql.SelectItem
		expansion              = ExpansionBuilder{
			Query: pgsql.Query{
				CommonTableExpressions: &pgsql.With{
					Recursive: true,
				},
			},

			RecursiveStatement: pgsql.Select{},
		}
	)

	switch traversalStep.Direction {
	case graph.DirectionInbound:
		rootIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID}
		nextIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID}

	case graph.DirectionOutbound:
		rootIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID}
		nextIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID}

	default:
		return pgsql.Query{}, fmt.Errorf("unsupported expansion direction: %s", traversalStep.Direction.String())
	}

	expansion.ProjectionStatement.Projection = traversalStep.Expansion.Value.Projection

	expansion.PrimerStatement.Where = traversalStep.Expansion.Value.PrimerConstraints
	expansion.RecursiveStatement.Where = traversalStep.Expansion.Value.RecursiveConstraints

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
					Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
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
					Constraint: traversalStep.Expansion.Value.LeftNodeJoinCondition,
				},
			}, {
				Table: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
					Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
				},
				JoinOperator: pgsql.JoinOperator{
					JoinType:   pgsql.JoinTypeInner,
					Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
				},
			}},
		})
	}

	// Make sure the recursive query has the expansion bound
	expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier},
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					rootIDColumnIdentifier,
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
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
			Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier},
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	})

	// If there are right node constraints, project them as part of the primer statement's projection
	if traversalStep.Expansion.Value.TerminalNodeConstraints != nil {
		if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](traversalStep.Expansion.Value.TerminalNodeConstraints); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.PrimerStatement.Projection = []pgsql.SelectItem{
				rootIDColumnIdentifier,
				nextIDColumnIdentifier,
				pgsql.NewLiteral(1, pgsql.Int),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					rootIDColumnIdentifier,
					pgsql.OperatorEquals,
					nextIDColumnIdentifier,
				),
				pgsql.ArrayLiteral{
					Values: []pgsql.Expression{
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					},
				},
			}

			expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
				nextIDColumnIdentifier,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath}),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}

			// Constraints that target the terminal node may crop up here where it's finally in scope. Additionally,
			// only accept paths that are marked satisfied from the recursive descent CTE
			if constraints, err := s.treeTranslator.IdentifierConstraints.ConsumeSet(traversalStep.Expansion.Value.Frame.Visible); err != nil {
				return pgsql.Query{}, err
			} else if projectionConstraints, err := ConjoinExpressions([]pgsql.Expression{pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionSatisfied}, constraints.Expression}); err != nil {
				return pgsql.Query{}, err
			} else {
				expansion.ProjectionStatement.Where = projectionConstraints
			}
		}
	} else {
		expansion.PrimerStatement.Projection = []pgsql.SelectItem{
			rootIDColumnIdentifier,
			nextIDColumnIdentifier,
			pgsql.NewLiteral(1, pgsql.Int),
			pgsql.NewLiteral(false, pgsql.Boolean),
			pgsql.NewBinaryExpression(
				rootIDColumnIdentifier,
				pgsql.OperatorEquals,
				nextIDColumnIdentifier,
			),
			pgsql.ArrayLiteral{
				Values: []pgsql.Expression{
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				},
			},
		}

		expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
			pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
			nextIDColumnIdentifier,
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth},
				pgsql.OperatorAdd,
				pgsql.NewLiteral(1, pgsql.Int),
			),
			pgsql.NewLiteral(false, pgsql.Boolean),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath}),
			),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath},
				pgsql.OperatorConcatenate,
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
			),
		}
	}

	return expansion.Build(traversalStep.Expansion.Value.Frame.Binding.Identifier), nil
}

func (s *Translator) buildExpansionPatternStep(traversalStep *PatternSegment) (pgsql.Query, error) {
	var (
		nextIDColumnIdentifier pgsql.SelectItem
		expansion              = ExpansionBuilder{
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.NewLiteral(1, pgsql.Int),
					),
					pgsql.NewLiteral(false, pgsql.Boolean),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, pgsql.ColumnPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, pgsql.ColumnPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					),
				},
			},
		}
	)

	switch traversalStep.Direction {
	case graph.DirectionInbound:
		nextIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnStartID}

	case graph.DirectionOutbound:
		nextIDColumnIdentifier = pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID}

	default:
		return pgsql.Query{}, fmt.Errorf("unsupported expansion direction: %s", traversalStep.Direction.String())
	}

	expansion.ProjectionStatement.Projection = traversalStep.Expansion.Value.Projection

	expansion.PrimerStatement.Where = traversalStep.Expansion.Value.PrimerConstraints
	expansion.RecursiveStatement.Where = traversalStep.Expansion.Value.RecursiveConstraints

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
				Constraint: traversalStep.Expansion.Value.ExpansionEdgeConstraints,
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
			},
		}},
	})

	// Make sure the recursive query has the expansion bound
	expansion.RecursiveStatement.From = append(expansion.RecursiveStatement.From, pgsql.FromClause{
		Source: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier},
		},
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: models.ValueOptional(traversalStep.Edge.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.OptionalAnd(
					traversalStep.Expansion.Value.PrimerConstraints,
					pgsql.NewBinaryExpression(
						nextIDColumnIdentifier,
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionNextID},
					),
				),
			},
		}, {
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: models.ValueOptional(traversalStep.RightNode.Identifier),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType:   pgsql.JoinTypeInner,
				Constraint: traversalStep.Expansion.Value.ExpansionNodeConstraints,
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
			Name:    pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier},
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
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
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionNextID},
				),
			},
		}},
	})

	// If there are terminal constraints, project them as part of the recursive lookup
	if traversalStep.Expansion.Value.TerminalNodeConstraints != nil {
		if terminalCriteriaProjection, err := pgsql.As[pgsql.SelectItem](traversalStep.Expansion.Value.TerminalNodeConstraints); err != nil {
			return pgsql.Query{}, err
		} else {
			expansion.RecursiveStatement.Projection = []pgsql.SelectItem{
				pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionRootID},
				pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnEndID},
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionDepth},
					pgsql.OperatorAdd,
					pgsql.NewLiteral(1, pgsql.Int),
				),
				terminalCriteriaProjection,
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath}),
				),
				pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.Expansion.Value.Frame.Binding.Identifier, expansionPath},
					pgsql.OperatorConcatenate,
					pgsql.CompoundIdentifier{traversalStep.Edge.Identifier, pgsql.ColumnID},
				),
			}
		}
	}

	return pgsql.Query{
		Body: expansion.Build(traversalStep.Expansion.Value.Frame.Binding.Identifier),
	}, nil
}
