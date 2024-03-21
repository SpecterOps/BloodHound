package format

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFormat_Delete(t *testing.T) {
	formattedQuery, err := Statement(Delete{
		Table: TableReference{
			Name:    CompoundIdentifier{"table"},
			Binding: AsOptionalIdentifier("t"),
		},
		Where: BinaryExpression{
			LeftOperand:  CompoundIdentifier{"t", "col1"},
			Operator:     Operator("<"),
			RightOperand: AsLiteral(4),
		},
	})

	require.Nil(t, err)
	require.Equal(t, "delete from table t where t.col1 < 4", formattedQuery.Value)
}

func TestFormat_Update(t *testing.T) {
	formattedQuery, err := Statement(Update{
		Table: TableReference{
			Name:    CompoundIdentifier{"table"},
			Binding: AsOptionalIdentifier("t"),
		},
		Assignments: []Assignment{{
			Identifier: "col1",
			Value:      AsLiteral(1),
		}, {
			Identifier: "col2",
			Value:      AsLiteral("12345"),
		}},
		Where: BinaryExpression{
			LeftOperand:  CompoundIdentifier{"t", "col1"},
			Operator:     Operator("<"),
			RightOperand: AsLiteral(4),
		},
	})

	require.Nil(t, err)
	require.Equal(t, "update table t set col1 = 1, col2 = '12345' where t.col1 < 4", formattedQuery.Value)
}

func TestFormat_Insert(t *testing.T) {
	formattedQuery, err := Statement(Insert{
		Table:   CompoundIdentifier{"table"},
		Columns: []Identifier{"col1", "col2", "col3"},
		Source: &Query{
			Body: Values{
				Values: []Expression{AsLiteral("1"), AsLiteral(1), AsLiteral(false)},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) values ('1', 1, false)", formattedQuery.Value)

	formattedQuery, err = Statement(Insert{
		Table:   CompoundIdentifier{"table"},
		Columns: []Identifier{"col1", "col2", "col3"},
		Source: &Query{
			Body: Select{
				Projection: []Projection{
					Wildcard{},
				},
				From: []FromClause{{
					Relation: TableReference{
						Name: CompoundIdentifier{"other"},
					},
				}},
				Where: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"other", "col1"},
					Operator:     Operator("="),
					RightOperand: AsLiteral("1234"),
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234'", formattedQuery.Value)

	formattedQuery, err = Statement(Insert{
		Table:   CompoundIdentifier{"table"},
		Columns: []Identifier{"col1", "col2", "col3"},
		Source: &Query{
			Body: Select{
				Projection: []Projection{
					Wildcard{},
				},
				From: []FromClause{{
					Relation: TableReference{
						Name: CompoundIdentifier{"other"},
					},
				}},
				Where: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"other", "col1"},
					Operator:     Operator("="),
					RightOperand: AsLiteral("1234"),
				},
			},
		},
		Returning: []Projection{
			Identifier("id"),
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' returning id", formattedQuery.Value)

	formattedQuery, err = Statement(Insert{
		Table:   CompoundIdentifier{"table"},
		Columns: []Identifier{"col1", "col2", "col3"},
		Source: &Query{
			Body: Select{
				Projection: []Projection{
					Wildcard{},
				},
				From: []FromClause{{
					Relation: TableReference{
						Name: CompoundIdentifier{"other"},
					},
				}},
				Where: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"other", "col1"},
					Operator:     Operator("="),
					RightOperand: AsLiteral("1234"),
				},
			},
		},
		OnConflict: &OnConflict{
			Target: &ConflictTarget{
				Constraint: CompoundIdentifier{"other.hash_constraint"},
			},
			Action: DoUpdate{
				Assignments: []Assignment{{
					Identifier: "hit_count",
					Value: BinaryExpression{
						LeftOperand:  Identifier("hit_count"),
						Operator:     Operator("+"),
						RightOperand: AsLiteral(1),
					},
				}},
				Where: BinaryExpression{
					LeftOperand:  Identifier("hit_count"),
					Operator:     Operator("<"),
					RightOperand: AsLiteral(9999),
				},
			},
		},
		Returning: []Projection{Identifier("id"), Identifier("hit_count")},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' on conflict on constraint other.hash_constraint do update set hit_count = hit_count + 1 where hit_count < 9999 returning id, hit_count", formattedQuery.Value)

	formattedQuery, err = Statement(Insert{
		Table:   CompoundIdentifier{"table"},
		Columns: []Identifier{"col1", "col2", "col3"},
		Source: &Query{
			Body: Select{
				Projection: []Projection{
					Wildcard{},
				},
				From: []FromClause{{
					Relation: TableReference{
						Name: CompoundIdentifier{"other"},
					},
				}},
				Where: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"other", "col1"},
					Operator:     Operator("="),
					RightOperand: AsLiteral("1234"),
				},
			},
		},
		OnConflict: &OnConflict{
			Target: &ConflictTarget{
				Columns: CompoundIdentifier{"hash"},
			},
			Action: DoUpdate{
				Assignments: []Assignment{{
					Identifier: "hit_count",
					Value: BinaryExpression{
						LeftOperand:  Identifier("hit_count"),
						Operator:     Operator("+"),
						RightOperand: AsLiteral(1),
					},
				}},
				Where: BinaryExpression{
					LeftOperand:  Identifier("hit_count"),
					Operator:     Operator("<"),
					RightOperand: AsLiteral(9999),
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "insert into table (col1, col2, col3) select * from other where other.col1 = '1234' on conflict (hash) do update set hit_count = hit_count + 1 where hit_count < 9999", formattedQuery.Value)
}

func TestFormat_Query(t *testing.T) {
	query := Query{
		Body: Select{
			Distinct: false,
			Projection: []Projection{
				Wildcard{},
			},
			From: []FromClause{{
				Relation: TableReference{
					Name:    CompoundIdentifier{"table"},
					Binding: AsOptionalIdentifier("t"),
				},
			}},
			Where: BinaryExpression{
				LeftOperand: CompoundIdentifier{"t", "col1"},
				Operator:    Operator(">"),
				RightOperand: Literal{
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
	formattedQuery, err := Statement(Merge{
		Into: true,
		Table: TableReference{
			Name:    CompoundIdentifier{"table"},
			Binding: AsOptionalIdentifier("t"),
		},
		Source: TableReference{
			Name:    CompoundIdentifier{"source"},
			Binding: AsOptionalIdentifier("s"),
		},
		JoinTarget: BinaryExpression{
			LeftOperand:  CompoundIdentifier{"t", "source_id"},
			Operator:     Operator("="),
			RightOperand: CompoundIdentifier{"s", "id"},
		},
		Actions: []MergeAction{
			MatchedUpdate{
				Predicate: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"t", "value"},
					Operator:     Operator(">"),
					RightOperand: CompoundIdentifier{"s", "value"},
				},
				Assignments: []Assignment{{
					Identifier: "updated_at",
					Value: FunctionCall{
						Function: "now",
					},
				}},
			},
			MatchedUpdate{
				Predicate: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"t", "value"},
					Operator:     Operator("<="),
					RightOperand: CompoundIdentifier{"s", "value"},
				},
				Assignments: []Assignment{{
					Identifier: "value",
					Value:      CompoundIdentifier{"s", "value"},
				}, {
					Identifier: "t.updated_at",
					Value: FunctionCall{
						Function: "now",
					},
				}},
			},
			MatchedDelete{
				Predicate: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"t", "value"},
					Operator:     Operator("="),
					RightOperand: CompoundIdentifier{"s", "value"},
				},
			},
			UnmatchedAction{
				Predicate: BinaryExpression{
					LeftOperand:  CompoundIdentifier{"t", "value"},
					Operator:     Operator("="),
					RightOperand: AsLiteral(0),
				},
				Columns: []Identifier{"hit_count"},
				Values: Values{
					Values: []Expression{AsLiteral(0)},
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "merge into table t using source s on t.source_id = s.id when matched and t.value > s.value then update set updated_at = now() when matched and t.value <= s.value then update set value = s.value, t.updated_at = now() when matched and t.value = s.value then delete when not matched and t.value = 0 then insert (hit_count) values (0)", formattedQuery.Value)
}

func TestFormat_CTEs(t *testing.T) {
	formattedQuery, err := Statement(Query{
		CommonTableExpressions: &With{
			Recursive: true,
			Expressions: []CommonTableExpression{{
				Materialized: &Materialized{
					Materialized: true,
				},
				Alias: TableAlias{
					Name: "expansion_1",
					Columns: []Identifier{
						"root_id",
						"next_id",
						"depth",
						"stop",
						"is_cycle",
						"path",
					},
				},
				Query: Query{
					Body: SetOperation{
						Operator: "union",
						All:      true,
						LeftOperand: Select{
							Projection: []Projection{
								CompoundIdentifier{"r", "start_id"},
								CompoundIdentifier{"r", "end_id"},
								Literal{
									Value: 1,
								},
								Literal{
									Value: false,
								},
								BinaryExpression{
									LeftOperand:  CompoundIdentifier{"r", "start_id"},
									Operator:     Operator("="),
									RightOperand: CompoundIdentifier{"r", "end_id"},
								},
								ArrayLiteral{
									Values: []Expression{
										CompoundIdentifier{"r", "id"},
									},
								},
							},

							From: []FromClause{{
								Relation: TableReference{
									Name:    CompoundIdentifier{"edge"},
									Binding: AsOptionalIdentifier("r"),
								},

								Joins: []Join{{
									Table: TableReference{
										Name:    CompoundIdentifier{"node"},
										Binding: AsOptionalIdentifier("a"),
									},
									JoinOperator: JoinOperator{
										JoinType: JoinTypeInner,
										Constraint: BinaryExpression{
											LeftOperand:  CompoundIdentifier{"a", "id"},
											Operator:     Operator("="),
											RightOperand: CompoundIdentifier{"r", "start_id"},
										},
									},
								}},
							}},

							Where: BinaryExpression{
								LeftOperand: CompoundIdentifier{"a", "kind_ids"},
								Operator: FunctionCall{
									Function: "operator",
									Parameters: []Expression{
										CompoundIdentifier{"pg_catalog", "&&"},
									},
								},
								RightOperand: ArrayLiteral{
									Values: []Expression{
										Literal{
											Value: 23,
										},
									},
									TypeHint: Int2Array,
								},
							},
						},
						RightOperand: Select{
							Projection: []Projection{
								CompoundIdentifier{"expansion_1", "root_id"},
								CompoundIdentifier{"r", "end_id"},
								BinaryExpression{
									LeftOperand: CompoundIdentifier{"expansion_1", "depth"},
									Operator:    Operator("+"),
									RightOperand: Literal{
										Value: 1,
									},
								},
								BinaryExpression{
									LeftOperand: CompoundIdentifier{"b", "kind_ids"},
									Operator: FunctionCall{
										Function: "operator",
										Parameters: []Expression{
											CompoundIdentifier{"pg_catalog", "&&"},
										},
									},
									RightOperand: ArrayLiteral{
										Values: []Expression{
											Literal{
												Value: 24,
											},
										},
										TypeHint: Int2Array,
									},
								},
								BinaryExpression{
									LeftOperand: CompoundIdentifier{"r", "id"},
									Operator:    Operator("="),
									RightOperand: FunctionCall{
										Function: "any",
										Parameters: []Expression{
											CompoundIdentifier{"expansion_1", "path"},
										},
									},
								},
								BinaryExpression{
									LeftOperand:  CompoundIdentifier{"expansion_1", "path"},
									Operator:     Operator("||"),
									RightOperand: CompoundIdentifier{"r", "id"},
								},
							},
							From: []FromClause{{
								Relation: TableReference{
									Name: CompoundIdentifier{"expansion_1"},
								},
								Joins: []Join{{
									Table: TableReference{
										Name:    CompoundIdentifier{"edge"},
										Binding: AsOptionalIdentifier("r"),
									},
									JoinOperator: JoinOperator{
										JoinType: JoinTypeInner,
										Constraint: BinaryExpression{
											LeftOperand:  CompoundIdentifier{"r", "start_id"},
											Operator:     Operator("="),
											RightOperand: CompoundIdentifier{"expansion_1", "next_id"},
										},
									},
								}, {
									Table: TableReference{
										Name:    CompoundIdentifier{"node"},
										Binding: AsOptionalIdentifier("b"),
									},
									JoinOperator: JoinOperator{
										JoinType: JoinTypeInner,
										Constraint: BinaryExpression{
											LeftOperand:  CompoundIdentifier{"b", "id"},
											Operator:     Operator("="),
											RightOperand: CompoundIdentifier{"r", "end_id"},
										},
									},
								}},
							}},
							Where: BinaryExpression{
								LeftOperand: UnaryExpression{
									Operator: Operator("not"),
									Operand:  CompoundIdentifier{"expansion_1", "is_cycle"},
								},
								Operator: Operator("and"),
								RightOperand: UnaryExpression{
									Operator: Operator("not"),
									Operand:  CompoundIdentifier{"expansion_1", "stop"},
								},
							},
						},
					},
				},
			}},
		},
		Body: Select{
			Projection: []Projection{
				CompoundIdentifier{"a", "properties"},
				CompoundIdentifier{"b", "properties"},
			},
			From: []FromClause{{
				Relation: TableReference{
					Name: CompoundIdentifier{"expansion_1"},
				},
				Joins: []Join{{
					Table: TableReference{
						Name:    CompoundIdentifier{"node"},
						Binding: AsOptionalIdentifier("a"),
					},
					JoinOperator: JoinOperator{
						JoinType: JoinTypeInner,
						Constraint: BinaryExpression{
							LeftOperand:  CompoundIdentifier{"a", "id"},
							Operator:     Operator("="),
							RightOperand: CompoundIdentifier{"expansion_1", "root_id"},
						},
					},
				}, {
					Table: TableReference{
						Name:    CompoundIdentifier{"node"},
						Binding: AsOptionalIdentifier("b"),
					},
					JoinOperator: JoinOperator{
						JoinType: JoinTypeInner,
						Constraint: BinaryExpression{
							LeftOperand:  CompoundIdentifier{"b", "id"},
							Operator:     Operator("="),
							RightOperand: CompoundIdentifier{"expansion_1", "next_id"},
						},
					},
				}},
			}},

			Where: BinaryExpression{
				LeftOperand: UnaryExpression{
					Operator: Operator("not"),
					Operand:  CompoundIdentifier{"expansion_1", "is_cycle"},
				},
				Operator:     Operator("and"),
				RightOperand: CompoundIdentifier{"expansion_1", "stop"},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, "with recursive expansion_1(root_id, next_id, depth, stop, is_cycle, path) as materialized (select r.start_id, r.end_id, 1, false, r.start_id = r.end_id, array[r.id] from edge r join node a on a.id = r.start_id where a.kind_ids operator(pg_catalog.&&) array[23]::int2[] union all select expansion_1.root_id, r.end_id, expansion_1.depth + 1, b.kind_ids operator(pg_catalog.&&) array[24]::int2[], r.id = any(expansion_1.path), expansion_1.path || r.id from expansion_1 join edge r on r.start_id = expansion_1.next_id join node b on b.id = r.end_id where not expansion_1.is_cycle and not expansion_1.stop) select a.properties, b.properties from expansion_1 join node a on a.id = expansion_1.root_id join node b on b.id = expansion_1.next_id where not expansion_1.is_cycle and expansion_1.stop", formattedQuery.Value)
}
