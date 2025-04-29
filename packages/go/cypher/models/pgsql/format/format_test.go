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

package format_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
	"github.com/stretchr/testify/require"
)

func mustAsLiteral(value any) pgsql.Literal {
	if literal, err := pgsql.AsLiteral(value); err != nil {
		panic(fmt.Sprintf("%v", err))
	} else {
		return literal
	}
}

func TestFormat_TypeCastedParenthetical(t *testing.T) {
	typeCastedParenthetical := pgsql.NewTypeCast(pgsql.NewParenthetical(pgsql.NewLiteral("str", pgsql.Text)), pgsql.Text)

	formattedQuery, err := format.Expression(typeCastedParenthetical, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "('str')::text", formattedQuery)
}

func TestFormat_Delete(t *testing.T) {
	formattedQuery, err := format.Statement(pgsql.Delete{
		From: []pgsql.TableReference{{
			Name:    pgsql.CompoundIdentifier{"table"},
			Binding: pgsql.AsOptionalIdentifier("t"),
		}},
		Where: models.ValueOptional[pgsql.Expression](pgsql.BinaryExpression{
			LOperand: pgsql.CompoundIdentifier{"t", "col1"},
			Operator: pgsql.OperatorLessThan,
			ROperand: pgsql.NewLiteral(4, pgsql.Int),
		}),
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "delete from table t where t.col1 < 4;", formattedQuery)
}

func TestFormat_Update(t *testing.T) {
	formattedQuery, err := format.Statement(pgsql.Update{
		Table: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{"table"},
			Binding: pgsql.AsOptionalIdentifier("t"),
		},
		Assignments: []pgsql.Assignment{pgsql.BinaryExpression{
			Operator: pgsql.OperatorAssignment,
			LOperand: pgsql.Identifier("col1"),
			ROperand: mustAsLiteral(1),
		}, pgsql.BinaryExpression{
			Operator: pgsql.OperatorAssignment,
			LOperand: pgsql.Identifier("col2"),
			ROperand: mustAsLiteral("12345"),
		}},
		Where: models.ValueOptional[pgsql.Expression](pgsql.BinaryExpression{
			LOperand: pgsql.CompoundIdentifier{"t", "col1"},
			Operator: pgsql.OperatorLessThan,
			ROperand: pgsql.NewLiteral(4, pgsql.Int),
		}),
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "update table t set col1 = 1, col2 = '12345' where t.col1 < 4;", formattedQuery)
}

func TestFormat_Insert(t *testing.T) {
	formattedQuery, err := format.Statement(pgsql.Insert{
		Table: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{"table"},
		},
		Shape: pgsql.RecordShape{
			Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		},
		Source: &pgsql.Query{
			Body: pgsql.Values{
				Values: []pgsql.Expression{mustAsLiteral("1"), mustAsLiteral(1), mustAsLiteral(false)},
			},
		},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) values ('1', 1, false);", formattedQuery)

	formattedQuery, err = format.Statement(pgsql.Insert{
		Table: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{"table"},
		},
		Shape: pgsql.RecordShape{
			Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Source: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.OperatorEquals,
					ROperand: mustAsLiteral("1234"),
				},
			},
		},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234';", formattedQuery)

	formattedQuery, err = format.Statement(pgsql.Insert{
		Table: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{"table"},
		},
		Shape: pgsql.RecordShape{
			Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Source: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.OperatorEquals,
					ROperand: mustAsLiteral("1234"),
				},
			},
		},
		Returning: []pgsql.SelectItem{
			pgsql.Identifier("id"),
		},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' returning id;", formattedQuery)

	formattedQuery, err = format.Statement(pgsql.Insert{
		Table: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{"table"},
		},
		Shape: pgsql.RecordShape{
			Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Source: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.OperatorEquals,
					ROperand: mustAsLiteral("1234"),
				},
			},
		},
		OnConflict: &pgsql.OnConflict{
			Target: &pgsql.ConflictTarget{
				Constraint: pgsql.CompoundIdentifier{"other.hash_constraint"},
			},
			Action: pgsql.DoUpdate{
				Assignments: []pgsql.Assignment{pgsql.BinaryExpression{
					Operator: pgsql.OperatorAssignment,
					LOperand: pgsql.Identifier("hit_count"),
					ROperand: pgsql.BinaryExpression{
						LOperand: pgsql.Identifier("hit_count"),
						Operator: pgsql.Operator("+"),
						ROperand: mustAsLiteral(1),
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.Identifier("hit_count"),
					Operator: pgsql.OperatorLessThan,
					ROperand: mustAsLiteral(9999),
				},
			},
		},
		Returning: []pgsql.SelectItem{pgsql.Identifier("id"), pgsql.Identifier("hit_count")},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' on conflict on constraint other.hash_constraint do update set hit_count = hit_count + 1 where hit_count < 9999 returning id, hit_count;", formattedQuery)

	formattedQuery, err = format.Statement(pgsql.Insert{
		Table: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{"table"},
		},
		Shape: pgsql.RecordShape{
			Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.SelectItem{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Source: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.OperatorEquals,
					ROperand: mustAsLiteral("1234"),
				},
			},
		},
		OnConflict: &pgsql.OnConflict{
			Target: &pgsql.ConflictTarget{
				Columns: []pgsql.Expression{pgsql.CompoundIdentifier{"hash"}},
			},
			Action: pgsql.DoUpdate{
				Assignments: []pgsql.Assignment{pgsql.BinaryExpression{
					Operator: pgsql.OperatorAssignment,
					LOperand: pgsql.Identifier("hit_count"),
					ROperand: pgsql.BinaryExpression{
						LOperand: pgsql.Identifier("hit_count"),
						Operator: pgsql.Operator("+"),
						ROperand: mustAsLiteral(1),
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.Identifier("hit_count"),
					Operator: pgsql.OperatorLessThan,
					ROperand: mustAsLiteral(9999),
				},
			},
		},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' on conflict (hash) do update set hit_count = hit_count + 1 where hit_count < 9999;", formattedQuery)
}

func TestFormat_Query(t *testing.T) {
	query := pgsql.Query{
		Body: pgsql.Select{
			Distinct: false,
			Projection: []pgsql.SelectItem{
				pgsql.Wildcard{},
			},
			From: []pgsql.FromClause{{
				Source: pgsql.TableReference{
					Name:    pgsql.CompoundIdentifier{"table"},
					Binding: pgsql.AsOptionalIdentifier("t"),
				},
			}},
			Where: pgsql.BinaryExpression{
				LOperand: pgsql.CompoundIdentifier{"t", "col1"},
				Operator: pgsql.Operator(">"),
				ROperand: pgsql.Literal{
					Value: 1,
				},
			},
		},
	}

	formattedQuery, err := format.Statement(query, format.NewOutputBuilder())
	require.Nil(t, err)
	require.Equal(t, "select * from table t where t.col1 > 1;", formattedQuery)
}

func TestFormat_Merge(t *testing.T) {
	formattedQuery, err := format.Statement(pgsql.Merge{
		Into: true,
		Table: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{"table"},
			Binding: pgsql.AsOptionalIdentifier("t"),
		},
		Source: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{"source"},
			Binding: pgsql.AsOptionalIdentifier("s"),
		},
		JoinTarget: pgsql.BinaryExpression{
			LOperand: pgsql.CompoundIdentifier{"t", "source_id"},
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{"s", "id"},
		},
		Actions: []pgsql.MergeAction{
			pgsql.MatchedUpdate{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.Operator(">"),
					ROperand: pgsql.CompoundIdentifier{"s", "value"},
				},
				Assignments: []pgsql.Assignment{pgsql.BinaryExpression{
					Operator: pgsql.OperatorAssignment,
					LOperand: pgsql.Identifier("updated_at"),
					ROperand: pgsql.FunctionCall{
						Function: "now",
					},
				}},
			},
			pgsql.MatchedUpdate{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.Operator("<="),
					ROperand: pgsql.CompoundIdentifier{"s", "value"},
				},
				Assignments: []pgsql.Assignment{pgsql.BinaryExpression{
					Operator: pgsql.OperatorAssignment,
					LOperand: pgsql.Identifier("value"),
					ROperand: pgsql.CompoundIdentifier{"s", "value"},
				}, pgsql.BinaryExpression{
					Operator: pgsql.OperatorAssignment,
					LOperand: pgsql.CompoundIdentifier{"t", "updated_at"},
					ROperand: pgsql.FunctionCall{
						Function: "now",
					},
				}},
			},
			pgsql.MatchedDelete{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.OperatorEquals,
					ROperand: pgsql.CompoundIdentifier{"s", "value"},
				},
			},
			pgsql.UnmatchedAction{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.OperatorEquals,
					ROperand: mustAsLiteral(0),
				},
				Columns: []pgsql.Identifier{"hit_count"},
				Values: pgsql.Values{
					Values: []pgsql.Expression{mustAsLiteral(0)},
				},
			},
		},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "merge into table t using source s on t.source_id = s.id when matched and t.value > s.value then update set updated_at = now() when matched and t.value <= s.value then update set value = s.value, t.updated_at = now() when matched and t.value = s.value then delete when not matched and t.value = 0 then insert (hit_count) values (0);", formattedQuery)
}

func TestFormat_CTEs(t *testing.T) {
	formattedQuery, err := format.Statement(pgsql.Query{
		CommonTableExpressions: &pgsql.With{
			Recursive: true,
			Expressions: []pgsql.CommonTableExpression{{
				Materialized: models.ValueOptional(pgsql.Materialized{
					Materialized: true,
				}),
				Alias: pgsql.TableAlias{
					Name: "expansion_1",
					Shape: models.ValueOptional(pgsql.RecordShape{
						Columns: []pgsql.Identifier{
							"root_id",
							"next_id",
							"depth",
							"stop",
							"is_cycle",
							"path",
						},
					}),
				},
				Query: pgsql.Query{
					Body: pgsql.SetOperation{
						Operator: "union",
						All:      true,
						LOperand: pgsql.Select{
							Projection: []pgsql.SelectItem{
								pgsql.CompoundIdentifier{"r", "start_id"},
								pgsql.CompoundIdentifier{"r", "end_id"},
								pgsql.Literal{
									Value: 1,
								},
								pgsql.Literal{
									Value: false,
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"r", "start_id"},
									Operator: pgsql.OperatorEquals,
									ROperand: pgsql.CompoundIdentifier{"r", "end_id"},
								},
								pgsql.ArrayLiteral{
									Values: []pgsql.Expression{
										pgsql.CompoundIdentifier{"r", "id"},
									},
								},
							},

							From: []pgsql.FromClause{{
								Source: pgsql.TableReference{
									Name:    pgsql.CompoundIdentifier{"edge"},
									Binding: pgsql.AsOptionalIdentifier("r"),
								},

								Joins: []pgsql.Join{{
									Table: pgsql.TableReference{
										Name:    pgsql.CompoundIdentifier{"node"},
										Binding: pgsql.AsOptionalIdentifier("a"),
									},
									JoinOperator: pgsql.JoinOperator{
										JoinType: pgsql.JoinTypeInner,
										Constraint: pgsql.BinaryExpression{
											LOperand: pgsql.CompoundIdentifier{"a", "id"},
											Operator: pgsql.OperatorEquals,
											ROperand: pgsql.CompoundIdentifier{"r", "start_id"},
										},
									},
								}},
							}},

							Where: pgsql.BinaryExpression{
								LOperand: pgsql.CompoundIdentifier{"a", "kind_ids"},
								Operator: pgsql.OperatorPGArrayOverlap,
								ROperand: pgsql.ArrayLiteral{
									Values: []pgsql.Expression{
										pgsql.Literal{
											Value: 23,
										},
									},
									CastType: pgsql.Int2Array,
								},
							},
						},
						ROperand: pgsql.Select{
							Projection: []pgsql.SelectItem{
								pgsql.CompoundIdentifier{"expansion_1", "root_id"},
								pgsql.CompoundIdentifier{"r", "end_id"},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"expansion_1", "depth"},
									Operator: pgsql.OperatorAdd,
									ROperand: pgsql.Literal{
										Value: 1,
									},
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"b", "kind_ids"},
									Operator: pgsql.OperatorPGArrayOverlap,
									ROperand: pgsql.ArrayLiteral{
										Values: []pgsql.Expression{
											pgsql.Literal{
												Value: 24,
											},
										},
										CastType: pgsql.Int2Array,
									},
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"r", "id"},
									Operator: pgsql.OperatorEquals,
									ROperand: pgsql.FunctionCall{
										Function: "any",
										Parameters: []pgsql.Expression{
											pgsql.CompoundIdentifier{"expansion_1", "path"},
										},
									},
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"expansion_1", "path"},
									Operator: pgsql.OperatorConcatenate,
									ROperand: pgsql.CompoundIdentifier{"r", "id"},
								},
							},
							From: []pgsql.FromClause{{
								Source: pgsql.TableReference{
									Name: pgsql.CompoundIdentifier{"expansion_1"},
								},
								Joins: []pgsql.Join{{
									Table: pgsql.TableReference{
										Name:    pgsql.CompoundIdentifier{"edge"},
										Binding: pgsql.AsOptionalIdentifier("r"),
									},
									JoinOperator: pgsql.JoinOperator{
										JoinType: pgsql.JoinTypeInner,
										Constraint: pgsql.BinaryExpression{
											LOperand: pgsql.CompoundIdentifier{"r", "start_id"},
											Operator: pgsql.OperatorEquals,
											ROperand: pgsql.CompoundIdentifier{"expansion_1", "next_id"},
										},
									},
								}, {
									Table: pgsql.TableReference{
										Name:    pgsql.CompoundIdentifier{"node"},
										Binding: pgsql.AsOptionalIdentifier("b"),
									},
									JoinOperator: pgsql.JoinOperator{
										JoinType: pgsql.JoinTypeInner,
										Constraint: pgsql.BinaryExpression{
											LOperand: pgsql.CompoundIdentifier{"b", "id"},
											Operator: pgsql.OperatorEquals,
											ROperand: pgsql.CompoundIdentifier{"r", "end_id"},
										},
									},
								}},
							}},
							Where: pgsql.BinaryExpression{
								LOperand: pgsql.UnaryExpression{
									Operator: pgsql.OperatorNot,
									Operand:  pgsql.CompoundIdentifier{"expansion_1", "is_cycle"},
								},
								Operator: pgsql.Operator("and"),
								ROperand: pgsql.UnaryExpression{
									Operator: pgsql.OperatorNot,
									Operand:  pgsql.CompoundIdentifier{"expansion_1", "stop"},
								},
							},
						},
					},
				},
			}},
		},
		Body: pgsql.Select{
			Projection: []pgsql.SelectItem{
				pgsql.CompoundIdentifier{"a", "properties"},
				pgsql.CompoundIdentifier{"b", "properties"},
			},
			From: []pgsql.FromClause{{
				Source: pgsql.TableReference{
					Name: pgsql.CompoundIdentifier{"expansion_1"},
				},
				Joins: []pgsql.Join{{
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{"node"},
						Binding: pgsql.AsOptionalIdentifier("a"),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType: pgsql.JoinTypeInner,
						Constraint: pgsql.BinaryExpression{
							LOperand: pgsql.CompoundIdentifier{"a", "id"},
							Operator: pgsql.OperatorEquals,
							ROperand: pgsql.CompoundIdentifier{"expansion_1", "root_id"},
						},
					},
				}, {
					Table: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{"node"},
						Binding: pgsql.AsOptionalIdentifier("b"),
					},
					JoinOperator: pgsql.JoinOperator{
						JoinType: pgsql.JoinTypeInner,
						Constraint: pgsql.BinaryExpression{
							LOperand: pgsql.CompoundIdentifier{"b", "id"},
							Operator: pgsql.OperatorEquals,
							ROperand: pgsql.CompoundIdentifier{"expansion_1", "next_id"},
						},
					},
				}},
			}},

			Where: pgsql.BinaryExpression{
				LOperand: pgsql.UnaryExpression{
					Operator: pgsql.OperatorNot,
					Operand:  pgsql.CompoundIdentifier{"expansion_1", "is_cycle"},
				},
				Operator: pgsql.OperatorAnd,
				ROperand: pgsql.CompoundIdentifier{"expansion_1", "stop"},
			},
		},
	}, format.NewOutputBuilder())

	require.Nil(t, err)
	require.Equal(t, "with recursive expansion_1(root_id, next_id, depth, stop, is_cycle, path) as materialized (select r.start_id, r.end_id, 1, false, r.start_id = r.end_id, array [r.id] from edge r join node a on a.id = r.start_id where a.kind_ids operator (pg_catalog.&&) array [23]::int2[] union all select expansion_1.root_id, r.end_id, expansion_1.depth + 1, b.kind_ids operator (pg_catalog.&&) array [24]::int2[], r.id = any(expansion_1.path), expansion_1.path || r.id from expansion_1 join edge r on r.start_id = expansion_1.next_id join node b on b.id = r.end_id where not expansion_1.is_cycle and not expansion_1.stop) select a.properties, b.properties from expansion_1 join node a on a.id = expansion_1.root_id join node b on b.id = expansion_1.next_id where not expansion_1.is_cycle and expansion_1.stop;", formattedQuery)
}
