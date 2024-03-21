// Copyright 2023 Specter Ops, Inc.
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

package pgsql

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/cypher"
	pgModel "github.com/specterops/bloodhound/cypher/model/pg"
	pgDriverModel "github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"io"
	"strconv"
)

const strippedLiteral = "$STRIPPED"

type KindMapper interface {
	MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds)
}

type Emitter struct {
	StripLiterals bool
	kindMapper    KindMapper
}

func NewEmitter(stripLiterals bool, kindMapper KindMapper) *Emitter {
	return &Emitter{
		StripLiterals: stripLiterals,
		kindMapper:    kindMapper,
	}
}

func (s *Emitter) formatMapLiteral(output io.Writer, mapLiteral cypher.MapLiteral) error {
	if _, err := io.WriteString(output, "{"); err != nil {
		return err
	}

	first := true
	for key, subExpression := range mapLiteral {
		if !first {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		} else {
			first = false
		}

		if _, err := io.WriteString(output, key); err != nil {
			return err
		}

		if _, err := io.WriteString(output, ": "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, subExpression); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(output, "}"); err != nil {
		return err
	}

	return nil
}

func (s *Emitter) formatLiteral(output io.Writer, literal *cypher.Literal) error {
	const literalNullToken = "null"

	// Check for a null literal first
	if literal.Null {
		if _, err := io.WriteString(output, literalNullToken); err != nil {
			return err
		}
		return nil
	}

	// Attempt to string format the literal value
	switch typedLiteral := literal.Value.(type) {
	case string:
		// Note: the cypher AST model expects literal strings to be wrapped in single quote characters (') so no
		// 		 additional formatting is done here
		if _, err := WriteStrings(output, typedLiteral); err != nil {
			return err
		}

	case graph.ID:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int8:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int16:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int32:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int64:
		if _, err := io.WriteString(output, strconv.FormatInt(typedLiteral, 10)); err != nil {
			return err
		}

	case int:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint8:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint16:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint32:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint64:
		if _, err := io.WriteString(output, strconv.FormatUint(typedLiteral, 10)); err != nil {
			return err
		}

	case uint:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case bool:
		if _, err := io.WriteString(output, strconv.FormatBool(typedLiteral)); err != nil {
			return err
		}

	case float32:
		if _, err := io.WriteString(output, strconv.FormatFloat(float64(typedLiteral), 'f', -1, 64)); err != nil {
			return err
		}

	case float64:
		if _, err := io.WriteString(output, strconv.FormatFloat(typedLiteral, 'f', -1, 64)); err != nil {
			return err
		}

	case cypher.MapLiteral:
		if err := s.formatMapLiteral(output, typedLiteral); err != nil {
			return err
		}

	case *cypher.ListLiteral:
		if _, err := io.WriteString(output, "array["); err != nil {
			return err
		}

		for idx, subExpression := range *typedLiteral {
			if idx > 0 {
				if _, err := io.WriteString(output, ", "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(output, subExpression); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(output, "]"); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unexpected literal type for string formatting: %T", literal.Value)
	}

	return nil
}

func (s *Emitter) writeReturn(writer io.Writer, returnClause *cypher.Return) error {
	if returnClause.Projection.Distinct {
		if _, err := WriteStrings(writer, "distinct "); err != nil {
			return err
		}
	}

	for idx, projectionItem := range returnClause.Projection.Items {
		if idx > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return nil
			}
		}

		if err := s.WriteExpression(writer, projectionItem); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeWhere(writer io.Writer, whereClause *cypher.Where) error {
	if len(whereClause.Expressions) > 0 {
		if _, err := io.WriteString(writer, " where "); err != nil {
			return err
		}
	}

	for _, expression := range whereClause.Expressions {
		if err := s.WriteExpression(writer, expression); err != nil {
			return err
		}
	}

	return nil
}

const traversalHeader = `with recursive `
const traversalTablePrefix = `pathspace_`
const traversalTableDef = `(next_node_id, depth, is_cycle, path) as (`

func expressionAsType[T any](expr cypher.Expression) (T, error) {
	if typed, isType := expr.(T); !isType {
		var empty T
		return empty, fmt.Errorf("expected type %T but received %T", empty, expr)
	} else {
		return typed, nil
	}
}

func formatVariableExpression(expr cypher.Expression, defaultBinding string) (string, error) {
	if expr == nil {
		return defaultBinding, nil
	}

	if variable, err := expressionAsType[*pgModel.AnnotatedVariable](expr); err != nil {
		return "", err
	} else {
		return variable.Symbol, nil
	}
}

const (
	defaultNodeBinding = "n"
	defaultEdgeBinding = "r"
)

// ()-[]
// ()<-[]

func (s *Emitter) startRelationshipPattern(writer io.Writer, idx int, nodeBinding string, nodePattern *cypher.NodePattern, relationshipBinding string, relationshipPattern *cypher.RelationshipPattern, where *cypher.Where) error {
	var (
		pathspaceTable = traversalTablePrefix + strconv.Itoa(idx)
	)

	// Each relationship element should be authored as a recursive CTE
	if _, err := WriteStrings(writer, traversalHeader, pathspaceTable, traversalTableDef); err != nil {
		return err
	}

	// Author the initial condition
	switch relationshipPattern.Direction {
	case graph.DirectionOutbound:
		if _, err := WriteStrings(writer, "select ", relationshipBinding, ".end_id, 0, false, array[", relationshipBinding, ".id] from edge ", relationshipBinding, " "); err != nil {
			return err
		}

		if nodePattern.Binding != nil {
			if _, err := WriteStrings(writer, "join node ", nodeBinding, " on ", nodeBinding, ".id = ", relationshipBinding, ".start_id"); err != nil {
				return err
			}
		}

	case graph.DirectionInbound:
		if _, err := WriteStrings(writer, "select ", relationshipBinding, ".start_id, 0, false, array[", relationshipBinding, ".id] from edge ", relationshipBinding, " "); err != nil {
			return err
		}

		if nodePattern.Binding != nil {
			if _, err := WriteStrings(writer, "join node ", nodeBinding, " on ", nodeBinding, ".id = ", relationshipBinding, ".end_id"); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported direction: %s(%d)", relationshipPattern.Direction, relationshipPattern.Direction)
	}

	return nil
}

// []->()
// []-()
func (s *Emitter) endRelationshipPattern(writer io.Writer, idx int, nodeBinding string, nodePattern *cypher.NodePattern, relationshipBinding string, relationshipPattern *cypher.RelationshipPattern, where *cypher.Where) error {
	var (
		// Make sure to use idx-1 since the end of a relationship pattern assumes that the previous index is the
		// relationship pattern element
		pathspaceTable = traversalTablePrefix + strconv.Itoa(idx-1)
	)

	if where != nil {
		if _, err := WriteStrings(writer, " where true "); err != nil {
			return err
		}
	}

	if _, err := WriteStrings(writer, " union all "); err != nil {
		return err
	}

	// Author the recursive portion of the query
	switch relationshipPattern.Direction {
	case graph.DirectionOutbound:
		if _, err := WriteStrings(writer, "select ", relationshipBinding, ".end_id, ", pathspaceTable, ".depth + 1, false, ", pathspaceTable, ".path || ", relationshipBinding, ".id from edge ", relationshipBinding, ", ", pathspaceTable); err != nil {
			return err
		}

		if nodePattern.Binding != nil {
			if _, err := WriteStrings(writer, " join node ", nodeBinding, " on ", nodeBinding, ".id = ", relationshipBinding, ".start_id"); err != nil {
				return err
			}
		}

	case graph.DirectionInbound:
		if _, err := WriteStrings(writer, "select ", relationshipBinding, ".start_id, ", pathspaceTable, ".depth + 1, false, ", pathspaceTable, ".path || ", relationshipBinding, ".id from edge ", relationshipBinding, ", ", pathspaceTable); err != nil {
			return err
		}

		if nodePattern.Binding != nil {
			if _, err := WriteStrings(writer, " join node ", nodeBinding, " on ", nodeBinding, ".id = ", relationshipBinding, ".end_id"); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported direction: %s(%d)", relationshipPattern.Direction, relationshipPattern.Direction)
	}

	if where != nil {
		if _, err := WriteStrings(writer, " where true "); err != nil {
			return err
		}
	}

	if _, err := WriteStrings(writer, ")"); err != nil {
		return err
	}

	return nil
}

func (s *Emitter) writePatternElements(writer io.Writer, patternElements []*cypher.PatternElement, where *cypher.Where) error {
	for idx := range patternElements {
		if _, isNodePattern := patternElements[idx].AsNodePattern(); isNodePattern {
			nodePattern, _ := patternElements[idx].AsNodePattern()

			if nodeBinding, err := formatVariableExpression(nodePattern.Binding, defaultNodeBinding); err != nil {
				return err
			} else {
				if idx > 0 {
					if relationshipPattern, isRelationshipPattern := patternElements[idx-1].AsRelationshipPattern(); isRelationshipPattern {
						if relationshipBinding, err := formatVariableExpression(relationshipPattern.Binding, defaultEdgeBinding); err != nil {
							return err
						} else if err := s.endRelationshipPattern(writer, idx, nodeBinding, nodePattern, relationshipBinding, relationshipPattern, where); err != nil {
							return err
						}
					} else {
						// TODO: Subsequent node patterns
					}
				}
			}
		} else {
			var (
				// The assumption is that the cypher parser would never allow something strange like a bare []->() or
				// a nil relationship pattern so errors are ignored here
				nodePattern, _         = patternElements[idx-1].AsNodePattern()
				relationshipPattern, _ = patternElements[idx].AsRelationshipPattern()
			)

			if nodeBinding, err := formatVariableExpression(nodePattern.Binding, defaultNodeBinding); err != nil {
				return err
			} else if relationshipBinding, err := formatVariableExpression(relationshipPattern.Binding, defaultEdgeBinding); err != nil {
				return err
			} else if err := s.startRelationshipPattern(writer, idx, nodeBinding, nodePattern, relationshipBinding, relationshipPattern, where); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Emitter) writeMatch(writer io.Writer, matchClause *cypher.Match) error {
	for idx, pattern := range matchClause.Pattern {
		if idx > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
		}

		if err := s.writePatternElements(writer, pattern.PatternElements, matchClause.Where); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeSelect(writer io.Writer, singlePartQuery *cypher.SinglePartQuery) error {
	if _, err := io.WriteString(writer, "select "); err != nil {
		return err
	}

	if singlePartQuery.Return != nil {
		if err := s.writeReturn(writer, singlePartQuery.Return); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(writer, " from "); err != nil {
		return err
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if readingClause.Match != nil {
			if err := s.writeMatch(writer, readingClause.Match); err != nil {
				return err
			}
		}
	}

	if singlePartQuery.Return != nil {
		if order := singlePartQuery.Return.Projection.Order; order != nil {
			if _, err := WriteStrings(writer, " order by "); err != nil {
				return err
			}

			for idx, orderItem := range order.Items {
				if idx > 0 {
					if _, err := WriteStrings(writer, ", "); err != nil {
						return err
					}
				}

				if err := s.WriteExpression(writer, orderItem.Expression); err != nil {
					return err
				}

				if orderItem.Ascending {
					if _, err := WriteStrings(writer, " asc"); err != nil {
						return err
					}
				} else {
					if _, err := WriteStrings(writer, " desc"); err != nil {
						return err
					}
				}
			}
		}

		if skip := singlePartQuery.Return.Projection.Skip; skip != nil {
			if _, err := WriteStrings(writer, " offset "); err != nil {
				return err
			}

			if err := s.WriteExpression(writer, skip.Value); err != nil {
				return err
			}
		}

		if limit := singlePartQuery.Return.Projection.Limit; limit != nil {
			if _, err := WriteStrings(writer, " limit "); err != nil {
				return err
			}

			if err := s.WriteExpression(writer, limit.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Emitter) writeDelete(writer io.Writer, singlePartQuery *cypher.SinglePartQuery, delete *pgModel.Delete) error {
	if delete.NodeDelete {
		if _, err := WriteStrings(writer, "delete from ", pgDriverModel.NodeTable, " as ", delete.Binding.Symbol); err != nil {
			return err
		}

		first := true

		for _, readingClause := range singlePartQuery.ReadingClauses {
			if matchClause := readingClause.Match; matchClause != nil {
				for _, pattern := range matchClause.Pattern {
					for _, patternElement := range pattern.PatternElements {
						if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
							switch typedBinding := nodePattern.Binding.(type) {
							case *pgModel.AnnotatedVariable:
								if typedBinding.Symbol == delete.Binding.Symbol {
									continue
								}
							}

							if !first {
								if _, err := WriteStrings(writer, ", "); err != nil {
									return err
								}
							} else {
								if _, err := WriteStrings(writer, " using "); err != nil {
									return err
								}

								first = false
							}

							if _, err := WriteStrings(writer, pgDriverModel.NodeTable, " as "); err != nil {
								return err
							}

							if err := s.WriteExpression(writer, nodePattern.Binding); err != nil {
								return err
							}
						} else {
							relationshipPattern, _ := patternElement.AsRelationshipPattern()

							switch typedBinding := relationshipPattern.Binding.(type) {
							case *pgModel.AnnotatedVariable:
								if typedBinding.Symbol == delete.Binding.Symbol {
									continue
								}
							}

							if !first {
								if _, err := WriteStrings(writer, ", "); err != nil {
									return err
								}
							} else {
								if _, err := WriteStrings(writer, " using "); err != nil {
									return err
								}

								first = false
							}

							if _, err := WriteStrings(writer, pgDriverModel.EdgeTable, " as "); err != nil {
								return err
							}

							if err := s.WriteExpression(writer, relationshipPattern.Binding); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	} else {
		if _, err := WriteStrings(writer, "delete from ", pgDriverModel.EdgeTable, " as ", delete.Binding.Symbol); err != nil {
			return err
		}

		first := true

		for _, readingClause := range singlePartQuery.ReadingClauses {
			if matchClause := readingClause.Match; matchClause != nil {
				for _, pattern := range matchClause.Pattern {
					for _, patternElement := range pattern.PatternElements {
						if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
							if !first {
								if _, err := WriteStrings(writer, ", "); err != nil {
									return err
								}
							} else {
								if _, err := WriteStrings(writer, " using "); err != nil {
									return err
								}

								first = false
							}

							if _, err := WriteStrings(writer, pgDriverModel.NodeTable, " as "); err != nil {
								return err
							}

							if err := s.WriteExpression(writer, nodePattern.Binding); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if matchClause := readingClause.Match; matchClause != nil {
			if matchClause.Where != nil {
				if err := s.writeWhere(writer, matchClause.Where); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Emitter) writeUpdates(writer io.Writer, singlePartQuery *cypher.SinglePartQuery) error {
	if _, err := io.WriteString(writer, "update "); err != nil {
		return err
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if matchClause := readingClause.Match; matchClause != nil {
			for idx, pattern := range matchClause.Pattern {
				if idx > 0 {
					if _, err := io.WriteString(writer, ", "); err != nil {
						return err
					}
				}

				if err := s.writePatternElements(writer, pattern.PatternElements, nil); err != nil {
					return err
				}
			}
		}
	}

	if _, err := WriteStrings(writer, " set "); err != nil {
		return err
	}

	for idx, item := range singlePartQuery.UpdatingClauses {
		if idx > 0 {
			if _, err := WriteStrings(writer, ", "); err != nil {
				return err
			}
		}

		switch typedUpdateItem := item.(type) {
		case *pgModel.Delete:
			if err := s.writeDelete(writer, singlePartQuery, typedUpdateItem); err != nil {
				return err
			}

		case *pgModel.PropertyMutation:
			// Can't use aliased names in the set clauses of the SQL statement so default to just the raw
			// column names
			if _, err := WriteStrings(writer, "properties = properties"); err != nil {
				return err
			}

			if typedUpdateItem.Additions != nil {
				if typedUpdateItem.Removals != nil {
					if _, err := WriteStrings(writer, " - "); err != nil {
						return err
					}

					if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
						return err
					}

					if _, err := WriteStrings(writer, "::text[]"); err != nil {
						return err
					}
				}

				if _, err := WriteStrings(writer, " || "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, typedUpdateItem.Additions); err != nil {
					return err
				}
			} else if typedUpdateItem.Removals != nil {
				if _, err := WriteStrings(writer, " - "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, "::text[]"); err != nil {
					return err
				}
			}

		case *pgModel.KindMutation:
			// Cypher and therefore this translation does not support kind mutation of relationships
			if typedUpdateItem.Variable.Type != pgModel.Node {
				return fmt.Errorf("unsupported SQL type for kind mutation: %s", typedUpdateItem.Variable.Type)
			}

			// Can't use aliased names in the set clauses of the SQL statement so default to just the raw
			// column names
			if _, err := WriteStrings(writer, "kind_ids = kind_ids"); err != nil {
				return err
			}

			if typedUpdateItem.Additions != nil {
				if typedUpdateItem.Removals != nil {
					if _, err := WriteStrings(writer, " - "); err != nil {
						return err
					}

					if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
						return err
					}
				}

				if _, err := WriteStrings(writer, " || "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, typedUpdateItem.Additions); err != nil {
					return err
				}
			} else if typedUpdateItem.Removals != nil {
				if _, err := WriteStrings(writer, " - "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("unsupported update clause item: %T", item)
		}
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if matchClause := readingClause.Match; matchClause != nil {
			if matchClause.Where != nil {
				if err := s.writeWhere(writer, matchClause.Where); err != nil {
					return err
				}
			}
		}
	}

	if singlePartQuery.Return != nil {
		if _, err := WriteStrings(writer, " returning "); err != nil {
			return err
		}

		if err := s.writeReturn(writer, singlePartQuery.Return); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeUpdatingClauses(writer io.Writer, singlePartQuery *cypher.SinglePartQuery) error {
	// Delete statements must be rendered as their own outputs
	numDeletes := 0

	for _, updateClause := range singlePartQuery.UpdatingClauses {
		switch typedClause := updateClause.(type) {
		case *pgModel.Delete:
			numDeletes++

			if err := s.writeDelete(writer, singlePartQuery, typedClause); err != nil {
				return err
			}
		}
	}

	if len(singlePartQuery.UpdatingClauses) > numDeletes {
		if err := s.writeUpdates(writer, singlePartQuery); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeSinglePartQuery(writer io.Writer, singlePartQuery *cypher.SinglePartQuery) error {
	if len(singlePartQuery.UpdatingClauses) > 0 {
		return s.writeUpdatingClauses(writer, singlePartQuery)
	} else {
		return s.writeSelect(writer, singlePartQuery)
	}
}

func (s *Emitter) writeSubquery(writer io.Writer, subquery *pgModel.Subquery) error {
	if _, err := io.WriteString(writer, "exists(select 1 from "); err != nil {
		return err
	}

	if err := s.writePatternElements(writer, subquery.PatternElements, nil); err != nil {
		return err
	}

	if subquery.Filter != nil {
		subQueryWhereClause := cypher.NewWhere()
		subQueryWhereClause.Add(subquery.Filter)

		if err := s.writeWhere(writer, subQueryWhereClause); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(writer, " limit 1)"); err != nil {
		return err
	}

	return nil
}

func (s *Emitter) WriteExpression(writer io.Writer, expression cypher.Expression) error {
	switch typedExpression := expression.(type) {
	case *pgModel.Subquery:
		if err := s.writeSubquery(writer, typedExpression); err != nil {
			return err
		}

	case *cypher.Negation:
		if _, err := io.WriteString(writer, "not "); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, typedExpression.Expression); err != nil {
			return err
		}

	case *cypher.Disjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(writer, " or "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(writer, joinedExpression); err != nil {
				return err
			}
		}

	case *cypher.Conjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(writer, " and "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(writer, joinedExpression); err != nil {
				return err
			}
		}

	case *cypher.Comparison:
		if err := s.WriteExpression(writer, typedExpression.Left); err != nil {
			return err
		}

		for _, nextPart := range typedExpression.Partials {
			if err := s.WriteExpression(writer, nextPart); err != nil {
				return err
			}
		}

	case *cypher.PartialComparison:
		if _, err := WriteStrings(writer, " ", typedExpression.Operator.String(), " "); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, typedExpression.Right); err != nil {
			return err
		}

	case *pgModel.AnnotatedLiteral:
		if err := s.WriteExpression(writer, &typedExpression.Literal); err != nil {
			return err
		}

	case *cypher.Literal:
		if !s.StripLiterals {
			return s.formatLiteral(writer, typedExpression)
		} else {
			_, err := io.WriteString(writer, strippedLiteral)
			return err
		}

	case *cypher.Variable:
		if _, err := io.WriteString(writer, typedExpression.Symbol); err != nil {
			return err
		}

	case *pgModel.AnnotatedVariable:
		if _, err := io.WriteString(writer, typedExpression.Symbol); err != nil {
			return err
		}

	case *pgModel.Entity:
		switch typedExpression.Binding.Type {
		case pgModel.Node:
			if _, err := WriteStrings(writer, "(", typedExpression.Binding.Symbol, ".id, ", typedExpression.Binding.Symbol, ".kind_ids, ", typedExpression.Binding.Symbol, ".properties)::nodeComposite"); err != nil {
				return err
			}

		case pgModel.Edge:
			if _, err := WriteStrings(writer, "(", typedExpression.Binding.Symbol, ".id, ", typedExpression.Binding.Symbol, ".start_id, ", typedExpression.Binding.Symbol, ".end_id, ", typedExpression.Binding.Symbol, ".kind_id, ", typedExpression.Binding.Symbol, ".properties)::edgeComposite"); err != nil {
				return err
			}

		case pgModel.Path:
			if _, err := WriteStrings(writer, "edges_to_path(", ")"); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported entity type %s", typedExpression.Binding.Type)
		}

	case *pgModel.NodeKindsReference:
		if err := s.WriteExpression(writer, typedExpression.Variable); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".kind_ids"); err != nil {
			return err
		}

	case *pgModel.EdgeKindReference:
		if err := s.WriteExpression(writer, typedExpression.Variable); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".kind_id"); err != nil {
			return err
		}

	case *pgModel.AnnotatedPropertyLookup:
		if _, err := io.WriteString(writer, "("); err != nil {
			return nil
		}

		if err := s.WriteExpression(writer, typedExpression.Atom); err != nil {
			return err
		}

		switch typedExpression.Type {
		case
			// We can't directly cast from JSONB types to time types since they require parsing first. The '->>'
			// operator coerces the underlying JSONB value to text before type casting
			pgModel.Date, pgModel.TimeWithTimeZone, pgModel.TimeWithoutTimeZone, pgModel.TimestampWithTimeZone, pgModel.TimestampWithoutTimeZone,

			// Text types also require the `->>' operator otherwise type casting clobbers itself
			pgModel.Text:

			if _, err := io.WriteString(writer, ".properties->>'"); err != nil {
				return nil
			}

		default:
			if _, err := io.WriteString(writer, ".properties->'"); err != nil {
				return nil
			}
		}

		if _, err := WriteStrings(writer, typedExpression.Symbols[0], "')::", typedExpression.Type.String()); err != nil {
			return nil
		}

	case *cypher.PropertyLookup:
		if err := s.WriteExpression(writer, typedExpression.Atom); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".properties->'", typedExpression.Symbols[0], "'"); err != nil {
			return nil
		}

	case *pgModel.AnnotatedKindMatcher:
		if err := s.WriteExpression(writer, typedExpression.Reference); err != nil {
			return err
		}

		if mappedKinds, missingKinds := s.kindMapper.MapKinds(typedExpression.Kinds); len(missingKinds) > 0 {
			return fmt.Errorf("query references the following undefined kinds: %v", missingKinds.Strings())
		} else {
			mappedKindStr := JoinInt(mappedKinds, ", ")

			switch typedExpression.Type {
			case pgModel.Node:
				if _, err := WriteStrings(writer, ".kind_ids operator(pg_catalog.&&) array[", mappedKindStr, "]::int2[]"); err != nil {
					return err
				}

			case pgModel.Edge:
				if _, err := WriteStrings(writer, ".kind_id = any(array[", mappedKindStr, "]::int2[])"); err != nil {
					return err
				}
			}
		}

	case *cypher.FunctionInvocation:
		if err := s.translateFunctionInvocation(writer, typedExpression); err != nil {
			return err
		}

	case *cypher.Parameter:
		if _, err := WriteStrings(writer, "@", typedExpression.Symbol); err != nil {
			return err
		}

	case *pgModel.AnnotatedParameter:
		if _, err := WriteStrings(writer, "@", typedExpression.Symbol); err != nil {
			return err
		}

	case *cypher.Parenthetical:
		if _, err := WriteStrings(writer, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, typedExpression.Expression); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ")"); err != nil {
			return err
		}

	case *pgModel.PropertiesReference:
		if err := s.WriteExpression(writer, typedExpression.Reference); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".properties"); err != nil {
			return err
		}

	case *cypher.ProjectionItem:
		if err := s.WriteExpression(writer, typedExpression.Expression); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, " as "); err != nil {
			return err
		}

		if typedExpression.Binding != nil {
			if err := s.WriteExpression(writer, typedExpression.Binding); err != nil {
				return err
			}
		} else {
			if _, err := WriteStrings(writer, "\""); err != nil {
				return err
			}

			switch typedProjectionExpression := typedExpression.Expression.(type) {
			case *pgModel.NodeKindsReference:
				if err := s.WriteExpression(writer, typedProjectionExpression); err != nil {
					return err
				}

			case *pgModel.EdgeKindReference:
				if err := s.WriteExpression(writer, typedProjectionExpression); err != nil {
					return err
				}

			case *cypher.FunctionInvocation:
				if err := s.WriteExpression(writer, typedProjectionExpression); err != nil {
					return err
				}

			case *cypher.PropertyLookup:
				if err := s.WriteExpression(writer, typedProjectionExpression.Atom); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".", typedProjectionExpression.Symbols[0]); err != nil {
					return err
				}

			case *pgModel.AnnotatedPropertyLookup:
				if err := s.WriteExpression(writer, typedProjectionExpression.Atom); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".", typedProjectionExpression.Symbols[0]); err != nil {
					return err
				}

			case *pgModel.AnnotatedVariable:
				if err := s.WriteExpression(writer, typedProjectionExpression.Symbol); err != nil {
					return err
				}

			case *pgModel.Entity:
				if err := s.WriteExpression(writer, typedProjectionExpression.Binding); err != nil {
					return err
				}

			default:
				return fmt.Errorf("unexpected projection item for binding formatting: %T", typedExpression.Expression)
			}

			if _, err := WriteStrings(writer, "\""); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected expression type for string formatting: %T", expression)
	}

	return nil
}

func (s *Emitter) translateFunctionInvocation(writer io.Writer, functionInvocation *cypher.FunctionInvocation) error {
	switch functionInvocation.Name {
	case cypherIdentityFunction:
		if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
			return err
		}

		if _, err := io.WriteString(writer, ".id"); err != nil {
			return err
		}

	case cypherDateFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::date"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "current_date"); err != nil {
			return err
		}

	case cypherTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::time with time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "current_time"); err != nil {
			return err
		}

	case cypherLocalTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::time without time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "localtime"); err != nil {
			return err
		}

	case cypherToLowerFunction:
		if _, err := WriteStrings(writer, pgsqlToLowerFunction, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ")"); err != nil {
			return err
		}

	case cypherDateTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::timestamp with time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "now()"); err != nil {
			return err
		}

	case cypherLocalDateTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::timestamp without time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "localtimestamp"); err != nil {
			return err
		}

	case cypherCountFunction:
		if _, err := WriteStrings(writer, "count("); err != nil {
			return err
		}

		for _, argument := range functionInvocation.Arguments {
			if err := s.WriteExpression(writer, argument); err != nil {
				return err
			}
		}

		if _, err := WriteStrings(writer, ")"); err != nil {
			return err
		}

	case pgsqlAnyFunction, pgsqlToJSONBFunction:
		if _, err := WriteStrings(writer, functionInvocation.Name, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
			return err
		}

		if _, err := io.WriteString(writer, ")"); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported function invocation %s", functionInvocation.Name)
	}

	return nil
}

func (s *Emitter) Write(regularQuery *cypher.RegularQuery, writer io.Writer) error {
	if regularQuery.SingleQuery != nil {
		if regularQuery.SingleQuery.MultiPartQuery != nil {
			return fmt.Errorf("not supported yet")
		}

		if regularQuery.SingleQuery.SinglePartQuery != nil {
			if err := s.writeSinglePartQuery(writer, regularQuery.SingleQuery.SinglePartQuery); err != nil {
				return err
			}
		}
	}

	return nil
}
