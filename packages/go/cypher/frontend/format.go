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
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func joinKinds(delimiter string, kinds graph.Kinds) string {
	output := strings.Builder{}

	for idx, kind := range kinds {
		if idx > 0 {
			output.WriteString(delimiter)
		}

		output.WriteString(kind.String())
	}

	return output.String()
}

func FormatNodePattern(output *strings.Builder, nodePattern *model.NodePattern) error {
	output.WriteString("(")
	output.WriteString(nodePattern.Binding)

	if len(nodePattern.Kinds) > 0 {
		output.WriteString(":")
		output.WriteString(joinKinds(":", nodePattern.Kinds))
	}

	if nodePattern.Properties != nil {
		output.WriteString(" ")

		if err := FormatExpression(output, nodePattern.Properties); err != nil {
			return err
		}
	}

	output.WriteString(")")
	return nil
}

func FormatRelationshipPattern(output *strings.Builder, relationshipPattern *model.RelationshipPattern) error {
	switch relationshipPattern.Direction {
	case graph.DirectionOutbound:
		output.WriteString("-[")

	case graph.DirectionBoth:
		fallthrough

	case graph.DirectionInbound:
		output.WriteString("<-[")
	}

	output.WriteString(relationshipPattern.Binding)

	if len(relationshipPattern.Kinds) > 0 {
		output.WriteString(":")
		output.WriteString(joinKinds("|", relationshipPattern.Kinds))
	}

	if relationshipPattern.Range != nil {
		output.WriteString("*")

		outputEllipsis := relationshipPattern.Range.StartIndex != nil || relationshipPattern.Range.EndIndex != nil

		if relationshipPattern.Range.StartIndex != nil {
			output.WriteString(strconv.FormatInt(*relationshipPattern.Range.StartIndex, 10))
		}

		if outputEllipsis {
			output.WriteString("..")
		}

		if relationshipPattern.Range.EndIndex != nil {
			output.WriteString(strconv.FormatInt(*relationshipPattern.Range.EndIndex, 10))
		}
	}

	if relationshipPattern.Properties != nil {
		output.WriteString(" ")

		if err := FormatExpression(output, relationshipPattern.Properties); err != nil {
			return err
		}
	}

	switch relationshipPattern.Direction {
	case graph.DirectionInbound:
		output.WriteString("]-")

	case graph.DirectionBoth:
		fallthrough

	case graph.DirectionOutbound:
		output.WriteString("]->")
	}

	return nil
}

func FormatPatternPart(output *strings.Builder, patternPart *model.PatternPart) error {
	if patternPart.Binding != "" {
		output.WriteString(patternPart.Binding)
		output.WriteString(" = ")
	}

	if patternPart.ShortestPathPattern {
		output.WriteString("shortestPath(")
	}

	if patternPart.AllShortestPathsPattern {
		output.WriteString("allShortestPaths(")
	}

	for idx, patternElement := range patternPart.PatternElements {
		if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
			// If this is another node pattern then output a delimiter
			if idx >= 1 && patternPart.PatternElements[idx-1].IsNodePattern() {
				output.WriteString(", ")
			}

			if err := FormatNodePattern(output, nodePattern); err != nil {
				return err
			}
		} else if relationshipPattern, isRelationshipPattern := patternElement.AsRelationshipPattern(); isRelationshipPattern {
			if err := FormatRelationshipPattern(output, relationshipPattern); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid pattern element: %T(%+v)", patternElement, patternElement)
		}
	}

	if patternPart.ShortestPathPattern || patternPart.AllShortestPathsPattern {
		output.WriteString(")")
	}

	return nil
}

func FormatProjection(output *strings.Builder, projection *model.Projection) error {
	if projection.Distinct {
		output.WriteString("distinct ")
	}

	for idx, projectionItem := range projection.Items {
		if idx > 0 {
			output.WriteString(", ")
		}

		if err := FormatExpression(output, projectionItem.Expression); err != nil {
			return err
		}

		if projectionItem.Binding != nil {
			output.WriteString(" as ")
			output.WriteString(projectionItem.Binding.Symbol)
		}
	}

	if projection.Order != nil {
		output.WriteString(" order by ")

		for idx := 0; idx < len(projection.Order.Items); idx++ {
			if idx > 0 {
				output.WriteString(", ")
			}

			nextItem := projection.Order.Items[idx]

			if err := FormatExpression(output, nextItem.Expression); err != nil {
				return err
			}

			if nextItem.Ascending {
				output.WriteString(" asc")
			} else {
				output.WriteString(" desc")
			}
		}
	}

	if projection.Skip != nil {
		output.WriteString(" skip ")

		if err := FormatExpression(output, projection.Skip); err != nil {
			return err
		}
	}

	if projection.Limit != nil {
		output.WriteString(" limit ")

		if err := FormatExpression(output, projection.Limit); err != nil {
			return err
		}
	}

	return nil
}

func FormatReturn(output *strings.Builder, returnClause *model.Return) error {
	output.WriteString(" return ")

	if returnClause.Projection != nil {
		return FormatProjection(output, returnClause.Projection)
	}

	return nil
}

func FormatWhere(output *strings.Builder, whereClause *model.Where) error {
	if len(whereClause.Expressions) > 0 {
		output.WriteString(" where ")
	}

	for _, expression := range whereClause.Expressions {
		if err := FormatExpression(output, expression); err != nil {
			return err
		}
	}

	return nil
}

func FormatMapLiteral(output *strings.Builder, mapLiteral model.MapLiteral) error {
	output.WriteString("{")

	first := true
	for key, subExpression := range mapLiteral {
		if !first {
			output.WriteString(", ")
		} else {
			first = false
		}

		output.WriteString(key)
		output.WriteString(": ")

		if err := FormatExpression(output, subExpression); err != nil {
			return err
		}
	}

	output.WriteString("}")
	return nil
}

func FormatLiteral(output *strings.Builder, literal *model.Literal) error {
	const literalNullToken = "null"

	// Check for a null literal first
	if literal.Null {
		output.WriteString(literalNullToken)
		return nil
	}

	// Attempt to string format the literal value
	switch typedLiteral := literal.Value.(type) {
	case string:
		output.WriteString(typedLiteral)

	case int8:
		output.WriteString(strconv.FormatInt(int64(typedLiteral), 10))

	case int16:
		output.WriteString(strconv.FormatInt(int64(typedLiteral), 10))

	case int32:
		output.WriteString(strconv.FormatInt(int64(typedLiteral), 10))

	case int64:
		output.WriteString(strconv.FormatInt(typedLiteral, 10))

	case int:
		output.WriteString(strconv.FormatInt(int64(typedLiteral), 10))

	case uint8:
		output.WriteString(strconv.FormatUint(uint64(typedLiteral), 10))

	case uint16:
		output.WriteString(strconv.FormatUint(uint64(typedLiteral), 10))

	case uint32:
		output.WriteString(strconv.FormatUint(uint64(typedLiteral), 10))

	case uint64:
		output.WriteString(strconv.FormatUint(typedLiteral, 10))

	case uint:
		output.WriteString(strconv.FormatUint(uint64(typedLiteral), 10))

	case bool:
		output.WriteString(strconv.FormatBool(typedLiteral))

	case float32:
		output.WriteString(strconv.FormatFloat(float64(typedLiteral), 'f', -1, 64))

	case float64:
		output.WriteString(strconv.FormatFloat(typedLiteral, 'f', -1, 64))

	case model.MapLiteral:
		if err := FormatMapLiteral(output, typedLiteral); err != nil {
			return err
		}

	case *model.ListLiteral:
		output.WriteString("[")

		for idx, subExpression := range *typedLiteral {
			if idx > 0 {
				output.WriteString(", ")
			}

			if err := FormatExpression(output, subExpression); err != nil {
				return err
			}
		}

		output.WriteString("]")

	default:
		return fmt.Errorf("unexpected literal type for string formatting: %T", literal)
	}

	return nil
}

func FormatExpression(output *strings.Builder, expression model.Expression) error {
	switch typedExpression := expression.(type) {
	case *model.Negation:
		output.WriteString("not ")

		if err := FormatExpression(output, typedExpression.Expression); err != nil {
			return err
		}

	case *model.IDInCollection:
		if err := FormatExpression(output, typedExpression.Variable); err != nil {
			return err
		}

		output.WriteString(" in ")

		if err := FormatExpression(output, typedExpression.Expression); err != nil {
			return err
		}

	case *model.FilterExpression:
		if err := FormatExpression(output, typedExpression.Specifier); err != nil {
			return err
		}

		if typedExpression.Where != nil {
			if err := FormatWhere(output, typedExpression.Where); err != nil {
				return err
			}
		}

	case *model.Quantifier:
		output.WriteString(typedExpression.Type.String())
		output.WriteString("(")

		if err := FormatExpression(output, typedExpression.Filter); err != nil {
			return err
		}

		output.WriteString(")")

	case *model.Parenthetical:
		output.WriteString("(")

		if err := FormatExpression(output, typedExpression.Expression); err != nil {
			return err
		}

		output.WriteString(")")

	case *model.Disjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				output.WriteString(" or ")
			}

			if err := FormatExpression(output, joinedExpression); err != nil {
				return err
			}
		}

	case *model.ExclusiveDisjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				output.WriteString(" xor ")
			}

			if err := FormatExpression(output, joinedExpression); err != nil {
				return err
			}
		}

	case *model.Conjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				output.WriteString(" and ")
			}

			if err := FormatExpression(output, joinedExpression); err != nil {
				return err
			}
		}

	case *model.PartialComparison:
		output.WriteString(" ")
		output.WriteString(typedExpression.Operator.String())
		output.WriteString(" ")

		if err := FormatExpression(output, typedExpression.Right); err != nil {
			return err
		}

	case *model.Comparison:
		if err := FormatExpression(output, typedExpression.Left); err != nil {
			return err
		}

		for _, nextPart := range typedExpression.Partials {
			if err := FormatExpression(output, nextPart); err != nil {
				return err
			}
		}

	case *model.Properties:
		if typedExpression.Map != nil {
			if err := FormatMapLiteral(output, typedExpression.Map); err != nil {
				return err
			}
		} else if err := FormatExpression(output, typedExpression.Parameter); err != nil {
			return err
		}

	case *model.Variable:
		output.WriteString(typedExpression.Symbol)

	case *model.Parameter:
		output.WriteString("$")
		output.WriteString(typedExpression.Symbol)

	case *model.PropertyLookup:
		if err := FormatExpression(output, typedExpression.Atom); err != nil {
			return err
		}

		output.WriteString(".")
		output.WriteString(strings.Join(typedExpression.Symbols, "."))

	case *model.FunctionInvocation:
		output.WriteString(strings.Join(typedExpression.Namespace, "."))
		output.WriteString(typedExpression.Name)
		output.WriteString("(")

		if typedExpression.Distinct {
			output.WriteString("distinct")
		}

		for idx, subExpression := range typedExpression.Arguments {
			if idx > 0 {
				output.WriteString(", ")
			}

			if err := FormatExpression(output, subExpression); err != nil {
				return err
			}
		}

		output.WriteString(")")

	case graph.Kind:
		output.WriteString(":")
		output.WriteString(typedExpression.String())

	case graph.Kinds:
		output.WriteString(":")
		output.WriteString(strings.Join(typedExpression.Strings(), ":"))

	case *model.KindMatcher:
		if err := FormatExpression(output, typedExpression.Reference); err != nil {
			return err
		}

		for _, matcher := range typedExpression.Kinds {
			output.WriteString(":")
			output.WriteString(matcher.String())
		}

	case *model.RangeQuantifier:
		output.WriteString(typedExpression.Value)

	case model.Operator:
		output.WriteString(typedExpression.String())

	case *model.Skip:
		return FormatExpression(output, typedExpression.Value)

	case *model.Limit:
		return FormatExpression(output, typedExpression.Value)

	case *model.Literal:
		return FormatLiteral(output, typedExpression)

	case []*model.PatternPart:
		return FormatPattern(output, typedExpression)

	case *model.ArithmeticExpression:
		if err := FormatExpression(output, typedExpression.Left); err != nil {
			return err
		}

		for _, part := range typedExpression.Partials {
			if err := FormatExpression(output, part); err != nil {
				return err
			}
		}

	case *model.PartialArithmeticExpression:
		output.WriteString(" ")
		output.WriteString(typedExpression.Operator.String())
		output.WriteString(" ")

		return FormatExpression(output, typedExpression.Right)

	default:
		return fmt.Errorf("unexpected expression type for string formatting: %T", expression)
	}

	return nil
}

func FormatRemove(output *strings.Builder, remove *model.Remove) error {
	output.WriteString("remove ")

	for idx, removeItem := range remove.Items {
		if idx > 0 {
			output.WriteString(", ")
		}

		var expression model.Expression

		if removeItem.KindMatcher != nil {
			expression = removeItem.KindMatcher
		} else if removeItem.Property != nil {
			expression = removeItem.Property
		} else {
			return fmt.Errorf("empty remove item")
		}

		if err := FormatExpression(output, expression); err != nil {
			return err
		}
	}

	return nil
}

func FormatSet(output *strings.Builder, set *model.Set) error {
	output.WriteString("set ")

	for idx, setItem := range set.Items {
		if idx > 0 {
			output.WriteString(", ")
		}

		if err := FormatExpression(output, setItem.Left); err != nil {
			return err
		}

		switch setItem.Operator {
		case model.OperatorLabelAssignment:
		default:
			output.WriteString(" ")
			output.WriteString(setItem.Operator.String())
			output.WriteString(" ")
		}

		if err := FormatExpression(output, setItem.Right); err != nil {
			return err
		}
	}

	return nil
}

func FormatDelete(output *strings.Builder, delete *model.Delete) error {
	if delete.Detach {
		output.WriteString("detach delete ")
	} else {
		output.WriteString("delete ")
	}

	for idx, expression := range delete.Expressions {
		if idx > 0 {
			output.WriteString(", ")
		}

		if err := FormatExpression(output, expression); err != nil {
			return err
		}
	}

	return nil
}

func FormatPattern(output *strings.Builder, pattern []*model.PatternPart) error {
	for idx, patternPart := range pattern {
		if idx > 0 {
			output.WriteString(", ")
		}

		if err := FormatPatternPart(output, patternPart); err != nil {
			return err
		}
	}

	return nil
}

func FormatCreate(output *strings.Builder, create *model.Create) error {
	output.WriteString("create ")
	return FormatPattern(output, create.Pattern)
}

func FormatUpdatingClause(output *strings.Builder, updatingClause *model.UpdatingClause) error {
	switch typedClause := updatingClause.Clause.(type) {
	case *model.Create:
		return FormatCreate(output, typedClause)

	case *model.Remove:
		return FormatRemove(output, typedClause)

	case *model.Set:
		return FormatSet(output, typedClause)

	case *model.Delete:
		return FormatDelete(output, typedClause)

	default:
		return fmt.Errorf("unsupported updating clause type: %T", updatingClause)
	}
}

func FormatReadingClause(output *strings.Builder, readingClause *model.ReadingClause) error {
	if readingClause.Match != nil {
		if readingClause.Match.Optional {
			output.WriteString("optional ")
		}

		output.WriteString("match ")

		for idx, patternPart := range readingClause.Match.Pattern {
			if idx > 0 {
				output.WriteString(", ")
			}

			if err := FormatPatternPart(output, patternPart); err != nil {
				return err
			}
		}

		if readingClause.Match.Where != nil {
			if err := FormatWhere(output, readingClause.Match.Where); err != nil {
				return err
			}
		}
	} else if readingClause.Unwind != nil {
		output.WriteString("unwind ")

		if err := FormatExpression(output, readingClause.Unwind.Expression); err != nil {
			return err
		}

		output.WriteString(" as ")

		if err := FormatExpression(output, readingClause.Unwind.Binding); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("reading clause has no match or unwind statement")
	}

	return nil
}

func FormatSinglePartQuery(output *strings.Builder, singlePartQuery *model.SinglePartQuery) error {
	for idx, readingClause := range singlePartQuery.ReadingClauses {
		if idx > 0 {
			output.WriteString(" ")
		}

		if err := FormatReadingClause(output, readingClause); err != nil {
			return err
		}
	}

	if len(singlePartQuery.UpdatingClauses) > 0 {
		if len(singlePartQuery.ReadingClauses) > 0 {
			output.WriteString(" ")
		}

		for idx, updatingClause := range singlePartQuery.UpdatingClauses {
			if idx > 0 {
				output.WriteString(" ")
			}

			if err := FormatUpdatingClause(output, updatingClause); err != nil {
				return err
			}
		}
	}

	if singlePartQuery.Return != nil {
		return FormatReturn(output, singlePartQuery.Return)
	}

	return nil
}

func FormatWith(output *strings.Builder, with *model.With) error {
	output.WriteString("with ")

	if err := FormatProjection(output, with.Projection); err != nil {
		return err
	}

	if with.Where != nil {
		if err := FormatWhere(output, with.Where); err != nil {
			return err
		}
	}

	return nil
}

func FormatMultiPartQuery(output *strings.Builder, multiPartQuery *model.MultiPartQuery) error {
	for idx, multiPartQueryPart := range multiPartQuery.Parts {
		var (
			numReadingClauses  = len(multiPartQueryPart.ReadingClauses)
			numUpdatingClauses = len(multiPartQueryPart.UpdatingClauses)
		)

		if idx > 0 {
			output.WriteString(" ")
		}

		for idx, readingClause := range multiPartQueryPart.ReadingClauses {
			if idx > 0 {
				output.WriteString(" ")
			}

			if err := FormatReadingClause(output, readingClause); err != nil {
				return err
			}
		}

		if len(multiPartQueryPart.UpdatingClauses) > 0 {
			if numReadingClauses > 0 {
				output.WriteString(" ")
			}

			for idx, updatingClause := range multiPartQueryPart.UpdatingClauses {
				if idx > 0 {
					output.WriteString(" ")
				}

				if err := FormatUpdatingClause(output, updatingClause); err != nil {
					return err
				}
			}
		}

		if multiPartQueryPart.With != nil {
			if numReadingClauses+numUpdatingClauses > 0 {
				output.WriteString(" ")
			}

			if err := FormatWith(output, multiPartQueryPart.With); err != nil {
				return err
			}
		}
	}

	if multiPartQuery.SinglePartQuery != nil {
		if len(multiPartQuery.Parts) > 0 {
			output.WriteString(" ")
		}

		return FormatSinglePartQuery(output, multiPartQuery.SinglePartQuery)
	}

	return nil
}

func FormatRegularQuery(regularQuery *model.RegularQuery) (string, error) {
	output := &strings.Builder{}

	if regularQuery.SingleQuery != nil {
		if regularQuery.SingleQuery.MultiPartQuery != nil {
			if err := FormatMultiPartQuery(output, regularQuery.SingleQuery.MultiPartQuery); err != nil {
				return "", err
			}
		}

		if regularQuery.SingleQuery.SinglePartQuery != nil {
			if err := FormatSinglePartQuery(output, regularQuery.SingleQuery.SinglePartQuery); err != nil {
				return "", err
			}
		}
	}

	return output.String(), nil
}
