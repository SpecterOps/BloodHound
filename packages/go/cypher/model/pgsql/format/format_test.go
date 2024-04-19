package format

import (
	"testing"

	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/stretchr/testify/require"
)

func TestFormat_Delete(t *testing.T) {
	formattedQuery, err := Statement(pgsql.Delete{
		Table: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{"table"},
			Binding: pgsql.AsOptionalIdentifier("t"),
		},
		Where: pgsql.BinaryExpression{
			LOperand: pgsql.CompoundIdentifier{"t", "col1"},
			Operator: pgsql.Operator("<"),
			ROperand: pgsql.AsLiteral(4),
		},
	})

	require.Nil(t, err)
	require.Equal(t, "delete from table t where t.col1 < 4", formattedQuery.Value)
}

func TestFormat_Update(t *testing.T) {
	formattedQuery, err := Statement(pgsql.Update{
		Table: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{"table"},
			Binding: pgsql.AsOptionalIdentifier("t"),
		},
		Assignments: []pgsql.Assignment{{
			Identifier: "col1",
			Value:      pgsql.AsLiteral(1),
		}, {
			Identifier: "col2",
			Value:      pgsql.AsLiteral("12345"),
		}},
		Where: pgsql.BinaryExpression{
			LOperand: pgsql.CompoundIdentifier{"t", "col1"},
			Operator: pgsql.Operator("<"),
			ROperand: pgsql.AsLiteral(4),
		},
	})

	require.Nil(t, err)
	require.Equal(t, "update table t set col1 = 1, col2 = '12345' where t.col1 < 4", formattedQuery.Value)
}

func TestFormat_Insert(t *testing.T) {
	formattedQuery, err := Statement(pgsql.Insert{
		Table:   pgsql.CompoundIdentifier{"table"},
		Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		Source: &pgsql.Query{
			Body: pgsql.Values{
				Values: []pgsql.Expression{pgsql.AsLiteral("1"), pgsql.AsLiteral(1), pgsql.AsLiteral(false)},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) values ('1', 1, false)", formattedQuery.Value)

	formattedQuery, err = Statement(pgsql.Insert{
		Table:   pgsql.CompoundIdentifier{"table"},
		Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Relation: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.Operator("="),
					ROperand: pgsql.AsLiteral("1234"),
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234'", formattedQuery.Value)

	formattedQuery, err = Statement(pgsql.Insert{
		Table:   pgsql.CompoundIdentifier{"table"},
		Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Relation: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.Operator("="),
					ROperand: pgsql.AsLiteral("1234"),
				},
			},
		},
		Returning: []pgsql.Projection{
			pgsql.Identifier("id"),
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' returning id", formattedQuery.Value)

	formattedQuery, err = Statement(pgsql.Insert{
		Table:   pgsql.CompoundIdentifier{"table"},
		Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Relation: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.Operator("="),
					ROperand: pgsql.AsLiteral("1234"),
				},
			},
		},
		OnConflict: &pgsql.OnConflict{
			Target: &pgsql.ConflictTarget{
				Constraint: pgsql.CompoundIdentifier{"other.hash_constraint"},
			},
			Action: pgsql.DoUpdate{
				Assignments: []pgsql.Assignment{{
					Identifier: "hit_count",
					Value: pgsql.BinaryExpression{
						LOperand: pgsql.Identifier("hit_count"),
						Operator: pgsql.Operator("+"),
						ROperand: pgsql.AsLiteral(1),
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.Identifier("hit_count"),
					Operator: pgsql.Operator("<"),
					ROperand: pgsql.AsLiteral(9999),
				},
			},
		},
		Returning: []pgsql.Projection{pgsql.Identifier("id"), pgsql.Identifier("hit_count")},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' on conflict on constraint other.hash_constraint do update set hit_count = hit_count + 1 where hit_count < 9999 returning id, hit_count", formattedQuery.Value)

	formattedQuery, err = Statement(pgsql.Insert{
		Table:   pgsql.CompoundIdentifier{"table"},
		Columns: []pgsql.Identifier{"col1", "col2", "col3"},
		Source: &pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{
					pgsql.Wildcard{},
				},
				From: []pgsql.FromClause{{
					Relation: pgsql.TableReference{
						Name: pgsql.CompoundIdentifier{"other"},
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"other", "col1"},
					Operator: pgsql.Operator("="),
					ROperand: pgsql.AsLiteral("1234"),
				},
			},
		},
		OnConflict: &pgsql.OnConflict{
			Target: &pgsql.ConflictTarget{
				Columns: pgsql.CompoundIdentifier{"hash"},
			},
			Action: pgsql.DoUpdate{
				Assignments: []pgsql.Assignment{{
					Identifier: "hit_count",
					Value: pgsql.BinaryExpression{
						LOperand: pgsql.Identifier("hit_count"),
						Operator: pgsql.Operator("+"),
						ROperand: pgsql.AsLiteral(1),
					},
				}},
				Where: pgsql.BinaryExpression{
					LOperand: pgsql.Identifier("hit_count"),
					Operator: pgsql.Operator("<"),
					ROperand: pgsql.AsLiteral(9999),
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' on conflict (hash) do update set hit_count = hit_count + 1 where hit_count < 9999", formattedQuery.Value)
}

func TestFormat_Query(t *testing.T) {
	query := pgsql.Query{
		Body: pgsql.Select{
			Distinct: false,
			Projection: []pgsql.Projection{
				pgsql.Wildcard{},
			},
			From: []pgsql.FromClause{{
				Relation: pgsql.TableReference{
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

	formattedQuery, err := Statement(query)
	require.Nil(t, err)
	require.Equal(t, "select * from table t where t.col1 > 1", formattedQuery.Value)
}

func TestFormat_Merge(t *testing.T) {
	formattedQuery, err := Statement(pgsql.Merge{
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
			Operator: pgsql.Operator("="),
			ROperand: pgsql.CompoundIdentifier{"s", "id"},
		},
		Actions: []pgsql.MergeAction{
			pgsql.MatchedUpdate{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.Operator(">"),
					ROperand: pgsql.CompoundIdentifier{"s", "value"},
				},
				Assignments: []pgsql.Assignment{{
					Identifier: "updated_at",
					Value: pgsql.FunctionCall{
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
				Assignments: []pgsql.Assignment{{
					Identifier: "value",
					Value:      pgsql.CompoundIdentifier{"s", "value"},
				}, {
					Identifier: "t.updated_at",
					Value: pgsql.FunctionCall{
						Function: "now",
					},
				}},
			},
			pgsql.MatchedDelete{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.Operator("="),
					ROperand: pgsql.CompoundIdentifier{"s", "value"},
				},
			},
			pgsql.UnmatchedAction{
				Predicate: pgsql.BinaryExpression{
					LOperand: pgsql.CompoundIdentifier{"t", "value"},
					Operator: pgsql.Operator("="),
					ROperand: pgsql.AsLiteral(0),
				},
				Columns: []pgsql.Identifier{"hit_count"},
				Values: pgsql.Values{
					Values: []pgsql.Expression{pgsql.AsLiteral(0)},
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "merge into table t using source s on t.source_id = s.id when matched and t.value > s.value then update set updated_at = now() when matched and t.value <= s.value then update set value = s.value, t.updated_at = now() when matched and t.value = s.value then delete when not matched and t.value = 0 then insert (hit_count) values (0)", formattedQuery.Value)
}

func TestFormat_CTEs(t *testing.T) {
	formattedQuery, err := Statement(pgsql.Query{
		CommonTableExpressions: &pgsql.With{
			Recursive: true,
			Expressions: []pgsql.CommonTableExpression{{
				Materialized: &pgsql.Materialized{
					Materialized: true,
				},
				Alias: pgsql.TableAlias{
					Name: "expansion_1",
					Columns: []pgsql.Identifier{
						"root_id",
						"next_id",
						"depth",
						"stop",
						"is_cycle",
						"path",
					},
				},
				Query: pgsql.Query{
					Body: pgsql.SetOperation{
						Operator: "union",
						All:      true,
						LeftOperand: pgsql.Select{
							Projection: []pgsql.Projection{
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
									Operator: pgsql.Operator("="),
									ROperand: pgsql.CompoundIdentifier{"r", "end_id"},
								},
								pgsql.ArrayLiteral{
									Values: []pgsql.Expression{
										pgsql.CompoundIdentifier{"r", "id"},
									},
								},
							},

							From: []pgsql.FromClause{{
								Relation: pgsql.TableReference{
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
											Operator: pgsql.Operator("="),
											ROperand: pgsql.CompoundIdentifier{"r", "start_id"},
										},
									},
								}},
							}},

							Where: pgsql.BinaryExpression{
								LOperand: pgsql.CompoundIdentifier{"a", "kind_ids"},
								Operator: pgsql.FunctionCall{
									Function: "operator",
									Parameters: []pgsql.Expression{
										pgsql.CompoundIdentifier{"pg_catalog", "&&"},
									},
								},
								ROperand: pgsql.ArrayLiteral{
									Values: []pgsql.Expression{
										pgsql.Literal{
											Value: 23,
										},
									},
									TypeHint: pgsql.Int2Array,
								},
							},
						},
						RightOperand: pgsql.Select{
							Projection: []pgsql.Projection{
								pgsql.CompoundIdentifier{"expansion_1", "root_id"},
								pgsql.CompoundIdentifier{"r", "end_id"},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"expansion_1", "depth"},
									Operator: pgsql.Operator("+"),
									ROperand: pgsql.Literal{
										Value: 1,
									},
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"b", "kind_ids"},
									Operator: pgsql.FunctionCall{
										Function: "operator",
										Parameters: []pgsql.Expression{
											pgsql.CompoundIdentifier{"pg_catalog", "&&"},
										},
									},
									ROperand: pgsql.ArrayLiteral{
										Values: []pgsql.Expression{
											pgsql.Literal{
												Value: 24,
											},
										},
										TypeHint: pgsql.Int2Array,
									},
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"r", "id"},
									Operator: pgsql.Operator("="),
									ROperand: pgsql.FunctionCall{
										Function: "any",
										Parameters: []pgsql.Expression{
											pgsql.CompoundIdentifier{"expansion_1", "path"},
										},
									},
								},
								pgsql.BinaryExpression{
									LOperand: pgsql.CompoundIdentifier{"expansion_1", "path"},
									Operator: pgsql.Operator("||"),
									ROperand: pgsql.CompoundIdentifier{"r", "id"},
								},
							},
							From: []pgsql.FromClause{{
								Relation: pgsql.TableReference{
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
											Operator: pgsql.Operator("="),
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
											Operator: pgsql.Operator("="),
											ROperand: pgsql.CompoundIdentifier{"r", "end_id"},
										},
									},
								}},
							}},
							Where: pgsql.BinaryExpression{
								LOperand: pgsql.UnaryExpression{
									Operator: pgsql.Operator("not"),
									Operand:  pgsql.CompoundIdentifier{"expansion_1", "is_cycle"},
								},
								Operator: pgsql.Operator("and"),
								ROperand: pgsql.UnaryExpression{
									Operator: pgsql.Operator("not"),
									Operand:  pgsql.CompoundIdentifier{"expansion_1", "stop"},
								},
							},
						},
					},
				},
			}},
		},
		Body: pgsql.Select{
			Projection: []pgsql.Projection{
				pgsql.CompoundIdentifier{"a", "properties"},
				pgsql.CompoundIdentifier{"b", "properties"},
			},
			From: []pgsql.FromClause{{
				Relation: pgsql.TableReference{
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
							Operator: pgsql.Operator("="),
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
							Operator: pgsql.Operator("="),
							ROperand: pgsql.CompoundIdentifier{"expansion_1", "next_id"},
						},
					},
				}},
			}},

			Where: pgsql.BinaryExpression{
				LOperand: pgsql.UnaryExpression{
					Operator: pgsql.Operator("not"),
					Operand:  pgsql.CompoundIdentifier{"expansion_1", "is_cycle"},
				},
				Operator: pgsql.Operator("and"),
				ROperand: pgsql.CompoundIdentifier{"expansion_1", "stop"},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "with recursive expansion_1(root_id, next_id, depth, stop, is_cycle, path) as materialized (select r.start_id, r.end_id, 1, false, r.start_id = r.end_id, array[r.id] from edge r join node a on a.id = r.start_id where a.kind_ids operator(pg_catalog.&&) array[23]::int2[] union all select expansion_1.root_id, r.end_id, expansion_1.depth + 1, b.kind_ids operator(pg_catalog.&&) array[24]::int2[], r.id = any(expansion_1.path), expansion_1.path || r.id from expansion_1 join edge r on r.start_id = expansion_1.next_id join node b on b.id = r.end_id where not expansion_1.is_cycle and not expansion_1.stop) select a.properties, b.properties from expansion_1 join node a on a.id = expansion_1.root_id join node b on b.id = expansion_1.next_id where not expansion_1.is_cycle and expansion_1.stop", formattedQuery.Value)
}
