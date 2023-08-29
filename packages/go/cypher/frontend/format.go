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

package frontend

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

const strippedLiteral = "$STRIPPED"

func writeJoinedKinds(output io.Writer, delimiter string, kinds graph.Kinds) error {
	for idx, kind := range kinds {
		if idx > 0 {
			if _, err := io.WriteString(output, delimiter); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(output, kind.String()); err != nil {
			return err
		}
	}

	return nil
}

type Emitter interface {
	Write(query *model.RegularQuery, writer io.Writer) error
	WriteExpression(output io.Writer, expression model.Expression) error
}

type CypherEmitter struct {
	StripLiterals bool
}

func NewCypherEmitter(stripLiterals bool) Emitter {
	return CypherEmitter{
		StripLiterals: stripLiterals,
	}
}

func (s CypherEmitter) formatNodePattern(output io.Writer, nodePattern *model.NodePattern) error {
	if _, err := io.WriteString(output, "("); err != nil {
		return err
	}

	if _, err := io.WriteString(output, nodePattern.Binding); err != nil {
		return err
	}

	if len(nodePattern.Kinds) > 0 {
		if _, err := io.WriteString(output, ":"); err != nil {
			return err
		}

		if err := writeJoinedKinds(output, ":", nodePattern.Kinds); err != nil {
			return err
		}
	}

	if nodePattern.Properties != nil {
		if _, err := io.WriteString(output, " "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, nodePattern.Properties); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(output, ")"); err != nil {
		return err
	}

	return nil
}

func (s CypherEmitter) formatRelationshipPattern(output io.Writer, relationshipPattern *model.RelationshipPattern) error {
	switch relationshipPattern.Direction {
	case graph.DirectionOutbound:
		if _, err := io.WriteString(output, "-["); err != nil {
			return err
		}

	case graph.DirectionBoth:
		fallthrough

	case graph.DirectionInbound:
		if _, err := io.WriteString(output, "<-["); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(output, relationshipPattern.Binding); err != nil {
		return err
	}

	if len(relationshipPattern.Kinds) > 0 {
		if _, err := io.WriteString(output, ":"); err != nil {
			return err
		}

		if err := writeJoinedKinds(output, "|", relationshipPattern.Kinds); err != nil {
			return err
		}
	}

	if relationshipPattern.Range != nil {
		if _, err := io.WriteString(output, "*"); err != nil {
			return err
		}

		outputEllipsis := relationshipPattern.Range.StartIndex != nil || relationshipPattern.Range.EndIndex != nil

		if relationshipPattern.Range.StartIndex != nil {
			if _, err := io.WriteString(output, strconv.FormatInt(*relationshipPattern.Range.StartIndex, 10)); err != nil {
				return err
			}
		}

		if outputEllipsis {
			if _, err := io.WriteString(output, ".."); err != nil {
				return err
			}
		}

		if relationshipPattern.Range.EndIndex != nil {
			if _, err := io.WriteString(output, strconv.FormatInt(*relationshipPattern.Range.EndIndex, 10)); err != nil {
				return err
			}
		}
	}

	if relationshipPattern.Properties != nil {
		if _, err := io.WriteString(output, " "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, relationshipPattern.Properties); err != nil {
			return err
		}
	}

	switch relationshipPattern.Direction {
	case graph.DirectionInbound:
		if _, err := io.WriteString(output, "]-"); err != nil {
			return err
		}

	case graph.DirectionBoth:
		fallthrough

	case graph.DirectionOutbound:
		if _, err := io.WriteString(output, "]->"); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatPatternPart(output io.Writer, patternPart *model.PatternPart) error {
	if patternPart.Binding != "" {
		if _, err := io.WriteString(output, patternPart.Binding); err != nil {
			return err
		}

		if _, err := io.WriteString(output, " = "); err != nil {
			return err
		}
	}

	if patternPart.ShortestPathPattern {
		if _, err := io.WriteString(output, "shortestPath("); err != nil {
			return err
		}
	}

	if patternPart.AllShortestPathsPattern {
		if _, err := io.WriteString(output, "allShortestPaths("); err != nil {
			return err
		}
	}

	for idx, patternElement := range patternPart.PatternElements {
		if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
			// If this is another node pattern then output a delimiter
			if idx >= 1 && patternPart.PatternElements[idx-1].IsNodePattern() {
				if _, err := io.WriteString(output, ", "); err != nil {
					return err
				}
			}

			if err := s.formatNodePattern(output, nodePattern); err != nil {
				return err
			}
		} else if relationshipPattern, isRelationshipPattern := patternElement.AsRelationshipPattern(); isRelationshipPattern {
			if err := s.formatRelationshipPattern(output, relationshipPattern); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid pattern element: %T(%+v)", patternElement, patternElement)
		}
	}

	if patternPart.ShortestPathPattern || patternPart.AllShortestPathsPattern {
		if _, err := io.WriteString(output, ")"); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatProjection(output io.Writer, projection *model.Projection) error {
	if projection.Distinct {
		if _, err := io.WriteString(output, "distinct "); err != nil {
			return err
		}
	}

	for idx, projectionItem := range projection.Items {
		if idx > 0 {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		}

		if err := s.WriteExpression(output, projectionItem.Expression); err != nil {
			return err
		}

		if projectionItem.Binding != nil {
			if _, err := io.WriteString(output, " as "); err != nil {
				return err
			}
			if _, err := io.WriteString(output, projectionItem.Binding.Symbol); err != nil {
				return err
			}
		}
	}

	if projection.Order != nil {
		if _, err := io.WriteString(output, " order by "); err != nil {
			return err
		}

		for idx := 0; idx < len(projection.Order.Items); idx++ {
			if idx > 0 {
				if _, err := io.WriteString(output, ", "); err != nil {
					return err
				}
			}

			nextItem := projection.Order.Items[idx]

			if err := s.WriteExpression(output, nextItem.Expression); err != nil {
				return err
			}

			if nextItem.Ascending {
				if _, err := io.WriteString(output, " asc"); err != nil {
					return err
				}
			} else if _, err := io.WriteString(output, " desc"); err != nil {
				return err
			}
		}
	}

	if projection.Skip != nil {
		if _, err := io.WriteString(output, " skip "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, projection.Skip); err != nil {
			return err
		}
	}

	if projection.Limit != nil {
		if _, err := io.WriteString(output, " limit "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, projection.Limit); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatReturn(output io.Writer, returnClause *model.Return) error {
	if _, err := io.WriteString(output, " return "); err != nil {
		return err
	}

	if returnClause.Projection != nil {
		return s.formatProjection(output, returnClause.Projection)
	}

	return nil
}

func (s CypherEmitter) formatWhere(output io.Writer, whereClause *model.Where) error {
	if len(whereClause.Expressions) > 0 {
		if _, err := io.WriteString(output, " where "); err != nil {
			return err
		}
	}

	for _, expression := range whereClause.Expressions {
		if err := s.WriteExpression(output, expression); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatMapLiteral(output io.Writer, mapLiteral model.MapLiteral) error {
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

func (s CypherEmitter) formatLiteral(output io.Writer, literal *model.Literal) error {
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
		if _, err := io.WriteString(output, typedLiteral); err != nil {
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

	case model.MapLiteral:
		if err := s.formatMapLiteral(output, typedLiteral); err != nil {
			return err
		}

	case *model.ListLiteral:
		if _, err := io.WriteString(output, "["); err != nil {
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
		return fmt.Errorf("unexpected literal type for string formatting: %T", literal)
	}

	return nil
}

func (s CypherEmitter) WriteExpression(output io.Writer, expression model.Expression) error {
	switch typedExpression := expression.(type) {
	case *model.Negation:
		if _, err := io.WriteString(output, "not "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, typedExpression.Expression); err != nil {
			return err
		}

	case *model.IDInCollection:
		if err := s.WriteExpression(output, typedExpression.Variable); err != nil {
			return err
		}

		if _, err := io.WriteString(output, " in "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, typedExpression.Expression); err != nil {
			return err
		}

	case *model.FilterExpression:
		if err := s.WriteExpression(output, typedExpression.Specifier); err != nil {
			return err
		}

		if typedExpression.Where != nil {
			if err := s.formatWhere(output, typedExpression.Where); err != nil {
				return err
			}
		}

	case *model.Quantifier:
		if _, err := io.WriteString(output, typedExpression.Type.String()); err != nil {
			return err
		}

		if _, err := io.WriteString(output, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(output, typedExpression.Filter); err != nil {
			return err
		}

		if _, err := io.WriteString(output, ")"); err != nil {
			return err
		}

	case *model.Parenthetical:
		if _, err := io.WriteString(output, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(output, typedExpression.Expression); err != nil {
			return err
		}

		if _, err := io.WriteString(output, ")"); err != nil {
			return err
		}

	case *model.Disjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(output, " or "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(output, joinedExpression); err != nil {
				return err
			}
		}

	case *model.ExclusiveDisjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(output, " xor "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(output, joinedExpression); err != nil {
				return err
			}
		}

	case *model.Conjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(output, " and "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(output, joinedExpression); err != nil {
				return err
			}
		}

	case *model.PartialComparison:
		if _, err := io.WriteString(output, " "); err != nil {
			return err
		}

		if _, err := io.WriteString(output, typedExpression.Operator.String()); err != nil {
			return err
		}

		if _, err := io.WriteString(output, " "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, typedExpression.Right); err != nil {
			return err
		}

	case *model.Comparison:
		if err := s.WriteExpression(output, typedExpression.Left); err != nil {
			return err
		}

		for _, nextPart := range typedExpression.Partials {
			if err := s.WriteExpression(output, nextPart); err != nil {
				return err
			}
		}

	case *model.Properties:
		if typedExpression.Map != nil {
			if err := s.formatMapLiteral(output, typedExpression.Map); err != nil {
				return err
			}
		} else if err := s.WriteExpression(output, typedExpression.Parameter); err != nil {
			return err
		}

	case *model.Variable:
		if _, err := io.WriteString(output, typedExpression.Symbol); err != nil {
			return err
		}

	case *model.Parameter:
		if _, err := io.WriteString(output, "$"); err != nil {
			return err
		}

		if _, err := io.WriteString(output, typedExpression.Symbol); err != nil {
			return err
		}

	case *model.PropertyLookup:
		if err := s.WriteExpression(output, typedExpression.Atom); err != nil {
			return err
		}

		if _, err := io.WriteString(output, "."); err != nil {
			return err
		}

		if _, err := io.WriteString(output, strings.Join(typedExpression.Symbols, ".")); err != nil {
			return err
		}

	case *model.FunctionInvocation:
		if _, err := io.WriteString(output, strings.Join(typedExpression.Namespace, ".")); err != nil {
			return err
		}

		if _, err := io.WriteString(output, typedExpression.Name); err != nil {
			return err
		}

		if _, err := io.WriteString(output, "("); err != nil {
			return err
		}

		if typedExpression.Distinct {
			if _, err := io.WriteString(output, "distinct"); err != nil {
				return err
			}
		}

		for idx, subExpression := range typedExpression.Arguments {
			if idx > 0 {
				if _, err := io.WriteString(output, ", "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(output, subExpression); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(output, ")"); err != nil {
			return err
		}

	case graph.Kind:
		if _, err := io.WriteString(output, ":"); err != nil {
			return err
		}

		if _, err := io.WriteString(output, typedExpression.String()); err != nil {
			return err
		}

	case graph.Kinds:
		if _, err := io.WriteString(output, ":"); err != nil {
			return err
		}

		if err := writeJoinedKinds(output, ":", typedExpression); err != nil {
			return err
		}

	case *model.KindMatcher:
		if err := s.WriteExpression(output, typedExpression.Reference); err != nil {
			return err
		}

		for _, matcher := range typedExpression.Kinds {
			if _, err := io.WriteString(output, ":"); err != nil {
				return err
			}

			if _, err := io.WriteString(output, matcher.String()); err != nil {
				return err
			}
		}

	case *model.RangeQuantifier:
		if _, err := io.WriteString(output, typedExpression.Value); err != nil {
			return err
		}

	case model.Operator:
		if _, err := io.WriteString(output, typedExpression.String()); err != nil {
			return err
		}

	case *model.Skip:
		return s.WriteExpression(output, typedExpression.Value)

	case *model.Limit:
		return s.WriteExpression(output, typedExpression.Value)

	case *model.Literal:
		if !s.StripLiterals {
			return s.formatLiteral(output, typedExpression)
		} else {
			_, err := io.WriteString(output, strippedLiteral)
			return err
		}

	case []*model.PatternPart:
		return s.formatPattern(output, typedExpression)

	case *model.ArithmeticExpression:
		if err := s.WriteExpression(output, typedExpression.Left); err != nil {
			return err
		}

		for _, part := range typedExpression.Partials {
			if err := s.WriteExpression(output, part); err != nil {
				return err
			}
		}

	case *model.PartialArithmeticExpression:
		if _, err := io.WriteString(output, " "); err != nil {
			return err
		}

		if _, err := io.WriteString(output, typedExpression.Operator.String()); err != nil {
			return err
		}

		if _, err := io.WriteString(output, " "); err != nil {
			return err
		}

		return s.WriteExpression(output, typedExpression.Right)

	default:
		return fmt.Errorf("unexpected expression type for string formatting: %T", expression)
	}

	return nil
}

func (s CypherEmitter) formatRemove(output io.Writer, remove *model.Remove) error {
	if _, err := io.WriteString(output, "remove "); err != nil {
		return err
	}

	for idx, removeItem := range remove.Items {
		if idx > 0 {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		}

		var expression model.Expression

		if removeItem.KindMatcher != nil {
			expression = removeItem.KindMatcher
		} else if removeItem.Property != nil {
			expression = removeItem.Property
		} else {
			return fmt.Errorf("empty remove item")
		}

		if err := s.WriteExpression(output, expression); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatSet(output io.Writer, set *model.Set) error {
	if _, err := io.WriteString(output, "set "); err != nil {
		return err
	}

	for idx, setItem := range set.Items {
		if idx > 0 {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		}

		if err := s.WriteExpression(output, setItem.Left); err != nil {
			return err
		}

		switch setItem.Operator {
		case model.OperatorLabelAssignment:
		default:
			if _, err := io.WriteString(output, " "); err != nil {
				return err
			}

			if _, err := io.WriteString(output, setItem.Operator.String()); err != nil {
				return err
			}

			if _, err := io.WriteString(output, " "); err != nil {
				return err
			}
		}

		if err := s.WriteExpression(output, setItem.Right); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatDelete(output io.Writer, delete *model.Delete) error {
	if delete.Detach {
		if _, err := io.WriteString(output, "detach delete "); err != nil {
			return err
		}
	} else if _, err := io.WriteString(output, "delete "); err != nil {
		return err
	}

	for idx, expression := range delete.Expressions {
		if idx > 0 {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		}

		if err := s.WriteExpression(output, expression); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatPattern(output io.Writer, pattern []*model.PatternPart) error {
	for idx, patternPart := range pattern {
		if idx > 0 {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		}

		if err := s.formatPatternPart(output, patternPart); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatCreate(output io.Writer, create *model.Create) error {
	if _, err := io.WriteString(output, "create "); err != nil {
		return err
	}

	return s.formatPattern(output, create.Pattern)
}

func (s CypherEmitter) formatUpdatingClause(output io.Writer, updatingClause *model.UpdatingClause) error {
	switch typedClause := updatingClause.Clause.(type) {
	case *model.Create:
		return s.formatCreate(output, typedClause)

	case *model.Remove:
		return s.formatRemove(output, typedClause)

	case *model.Set:
		return s.formatSet(output, typedClause)

	case *model.Delete:
		return s.formatDelete(output, typedClause)

	default:
		return fmt.Errorf("unsupported updating clause type: %T", updatingClause)
	}
}

func (s CypherEmitter) formatReadingClause(output io.Writer, readingClause *model.ReadingClause) error {
	if readingClause.Match != nil {
		if readingClause.Match.Optional {
			if _, err := io.WriteString(output, "optional "); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(output, "match "); err != nil {
			return err
		}

		for idx, patternPart := range readingClause.Match.Pattern {
			if idx > 0 {
				if _, err := io.WriteString(output, ", "); err != nil {
					return err
				}
			}

			if err := s.formatPatternPart(output, patternPart); err != nil {
				return err
			}
		}

		if readingClause.Match.Where != nil {
			if err := s.formatWhere(output, readingClause.Match.Where); err != nil {
				return err
			}
		}
	} else if readingClause.Unwind != nil {
		if _, err := io.WriteString(output, "unwind "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, readingClause.Unwind.Expression); err != nil {
			return err
		}

		if _, err := io.WriteString(output, " as "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, readingClause.Unwind.Binding); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("reading clause has no match or unwind statement")
	}

	return nil
}

func (s CypherEmitter) formatSinglePartQuery(output io.Writer, singlePartQuery *model.SinglePartQuery) error {
	for idx, readingClause := range singlePartQuery.ReadingClauses {
		if idx > 0 {
			if _, err := io.WriteString(output, " "); err != nil {
				return err
			}
		}

		if err := s.formatReadingClause(output, readingClause); err != nil {
			return err
		}
	}

	if len(singlePartQuery.UpdatingClauses) > 0 {
		if len(singlePartQuery.ReadingClauses) > 0 {
			if _, err := io.WriteString(output, " "); err != nil {
				return err
			}
		}

		for idx, updatingClause := range singlePartQuery.UpdatingClauses {
			if idx > 0 {
				if _, err := io.WriteString(output, " "); err != nil {
					return err
				}
			}

			if err := s.formatUpdatingClause(output, updatingClause); err != nil {
				return err
			}
		}
	}

	if singlePartQuery.Return != nil {
		return s.formatReturn(output, singlePartQuery.Return)
	}

	return nil
}

func (s CypherEmitter) formatWith(output io.Writer, with *model.With) error {
	if _, err := io.WriteString(output, "with "); err != nil {
		return err
	}

	if err := s.formatProjection(output, with.Projection); err != nil {
		return err
	}

	if with.Where != nil {
		if err := s.formatWhere(output, with.Where); err != nil {
			return err
		}
	}

	return nil
}

func (s CypherEmitter) formatMultiPartQuery(output io.Writer, multiPartQuery *model.MultiPartQuery) error {
	for idx, multiPartQueryPart := range multiPartQuery.Parts {
		var (
			numReadingClauses  = len(multiPartQueryPart.ReadingClauses)
			numUpdatingClauses = len(multiPartQueryPart.UpdatingClauses)
		)

		if idx > 0 {
			if _, err := io.WriteString(output, " "); err != nil {
				return err
			}
		}

		for idx, readingClause := range multiPartQueryPart.ReadingClauses {
			if idx > 0 {
				if _, err := io.WriteString(output, " "); err != nil {
					return err
				}
			}

			if err := s.formatReadingClause(output, readingClause); err != nil {
				return err
			}
		}

		if len(multiPartQueryPart.UpdatingClauses) > 0 {
			if numReadingClauses > 0 {
				if _, err := io.WriteString(output, " "); err != nil {
					return err
				}
			}

			for idx, updatingClause := range multiPartQueryPart.UpdatingClauses {
				if idx > 0 {
					if _, err := io.WriteString(output, " "); err != nil {
						return err
					}
				}

				if err := s.formatUpdatingClause(output, updatingClause); err != nil {
					return err
				}
			}
		}

		if multiPartQueryPart.With != nil {
			if numReadingClauses+numUpdatingClauses > 0 {
				if _, err := io.WriteString(output, " "); err != nil {
					return err
				}
			}

			if err := s.formatWith(output, multiPartQueryPart.With); err != nil {
				return err
			}
		}
	}

	if multiPartQuery.SinglePartQuery != nil {
		if len(multiPartQuery.Parts) > 0 {
			if _, err := io.WriteString(output, " "); err != nil {
				return err
			}
		}

		return s.formatSinglePartQuery(output, multiPartQuery.SinglePartQuery)
	}

	return nil
}

func (s CypherEmitter) Write(regularQuery *model.RegularQuery, writer io.Writer) error {
	if regularQuery.SingleQuery != nil {
		if regularQuery.SingleQuery.MultiPartQuery != nil {
			if err := s.formatMultiPartQuery(writer, regularQuery.SingleQuery.MultiPartQuery); err != nil {
				return err
			}
		}

		if regularQuery.SingleQuery.SinglePartQuery != nil {
			if err := s.formatSinglePartQuery(writer, regularQuery.SingleQuery.SinglePartQuery); err != nil {
				return err
			}
		}
	}

	return nil
}
