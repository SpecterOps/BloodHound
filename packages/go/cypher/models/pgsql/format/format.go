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

package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

type OutputBuilder struct {
	MaterializeParameters bool
	StripLiterals         bool
	parameters            map[string]any
	builder               *strings.Builder
}

func NewOutputBuilder() *OutputBuilder {
	return &OutputBuilder{
		builder: &strings.Builder{},
	}
}

func (s *OutputBuilder) WithMaterializedParameters(parameters map[string]any) *OutputBuilder {
	s.MaterializeParameters = true
	s.parameters = parameters

	return s
}

func (s *OutputBuilder) HasOutput() bool {
	return s.builder.Len() != 0
}

func (s *OutputBuilder) Write(values ...any) {
	for _, value := range values {
		switch typedValue := value.(type) {
		case string:
			s.builder.WriteString(typedValue)

		case fmt.Stringer:
			s.builder.WriteString(typedValue.String())

		default:
			panic(fmt.Sprintf("invalid write parameter type: %T", value))
		}
	}
}

func (s *OutputBuilder) Build() string {
	return s.builder.String()
}

func formatSlice[T any, TS []T](builder *OutputBuilder, slice TS, dataType pgsql.DataType) error {
	builder.Write("array [")

	for idx, value := range slice {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatValue(builder, value); err != nil {
			return err
		}
	}

	builder.Write("]::", dataType.String())
	return nil
}

func formatValue(builder *OutputBuilder, value any) error {
	switch typedValue := value.(type) {
	case uint:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint8:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint16:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint32:
		builder.Write(strconv.FormatUint(uint64(typedValue), 10))

	case uint64:
		builder.Write(strconv.FormatUint(typedValue, 10))

	case int:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case []int:
		return formatSlice(builder, typedValue, pgsql.Int8Array)

	case int8:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case []int8:
		return formatSlice(builder, typedValue, pgsql.Int2Array)

	case int16:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case []int16:
		return formatSlice(builder, typedValue, pgsql.Int2Array)

	case int32:
		builder.Write(strconv.FormatInt(int64(typedValue), 10))

	case []int32:
		return formatSlice(builder, typedValue, pgsql.Int4Array)

	case int64:
		builder.Write(strconv.FormatInt(typedValue, 10))

	case []int64:
		return formatSlice(builder, typedValue, pgsql.Int8Array)

	case string:
		builder.Write("'", typedValue, "'")

	case bool:
		builder.Write(strconv.FormatBool(typedValue))

	case float32:
		builder.Write(strconv.FormatFloat(float64(typedValue), 'f', -1, 64))

	case float64:
		builder.Write(strconv.FormatFloat(typedValue, 'f', -1, 64))

	default:
		return fmt.Errorf("unsupported literal type: %T", value)
	}

	return nil
}

func formatLiteral(builder *OutputBuilder, literal pgsql.Literal) error {
	if !literal.Null {
		return formatValue(builder, literal.Value)
	}

	builder.Write("null")
	return nil
}

func formatNode(builder *OutputBuilder, rootExpr pgsql.SyntaxNode) error {
	exprStack := []pgsql.SyntaxNode{
		rootExpr,
	}

	for len(exprStack) > 0 {
		nextExpr := exprStack[len(exprStack)-1]
		exprStack = exprStack[:len(exprStack)-1]

		switch typedNextExpr := nextExpr.(type) {
		case pgsql.Query:
			if err := formatSetExpression(builder, typedNextExpr); err != nil {
				return err
			}

		case pgsql.Select:
			if err := formatSelect(builder, typedNextExpr); err != nil {
				return err
			}

		case pgsql.OrderBy:
			// Don't emit 'asc' since it's the default
			if !typedNextExpr.Ascending {
				exprStack = append(exprStack, pgsql.FormattingLiteral(" desc"))
			}

			exprStack = append(exprStack, typedNextExpr.Expression)

		case pgsql.Wildcard:
			builder.Write("*")

		case pgsql.Literal:
			if err := formatLiteral(builder, typedNextExpr); err != nil {
				return err
			}

		case pgsql.Materialized:
			if typedNextExpr.Materialized {
				exprStack = append(exprStack, pgsql.FormattingLiteral("materialized"))
			} else {
				exprStack = append(exprStack, pgsql.FormattingLiteral("not materialized"))
			}

		case *pgsql.FunctionCall:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.FunctionCall:
			if typedNextExpr.CastType.IsKnown() {
				exprStack = append(exprStack, typedNextExpr.CastType, pgsql.FormattingLiteral("::"))
			}

			if !typedNextExpr.Bare {
				exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			}

			for idx := len(typedNextExpr.Parameters) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr.Parameters[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral(", "))
				}
			}

			if typedNextExpr.Distinct {
				exprStack = append(exprStack, pgsql.FormattingLiteral("distinct "))
			}

			if !typedNextExpr.Bare {
				exprStack = append(exprStack, pgsql.FormattingLiteral("("))
			}

			exprStack = append(exprStack, typedNextExpr.Function)

		case pgsql.Operator:
			builder.Write(typedNextExpr.String())

		case pgsql.Identifier:
			builder.Write(typedNextExpr)

		case pgsql.CompoundIdentifier:
			for idx := len(typedNextExpr) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral("."))
				}
			}

		case pgsql.FormattingLiteral:
			builder.Write(typedNextExpr)

		case *pgsql.UnaryExpression:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.UnaryExpression:
			exprStack = append(exprStack,
				typedNextExpr.Operand,
				pgsql.FormattingLiteral(" "),
				typedNextExpr.Operator,
			)

		case *pgsql.BinaryExpression:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.BinaryExpression:
			// Push the operands and operator onto the stack in reverse order
			exprStack = append(exprStack,
				typedNextExpr.ROperand,
				pgsql.FormattingLiteral(" "),
				typedNextExpr.Operator,
				pgsql.FormattingLiteral(" "),
				typedNextExpr.LOperand,
			)

		case pgsql.TableReference:
			if typedNextExpr.Binding.Set {
				exprStack = append(exprStack, typedNextExpr.Binding.Value, pgsql.FormattingLiteral(" "))
			}

			exprStack = append(exprStack, typedNextExpr.Name)

		case pgsql.Assignment:
			exprStack = append(exprStack,
				typedNextExpr,
				pgsql.FormattingLiteral(" "),
				pgsql.Operator("="),
				pgsql.FormattingLiteral(" "),
				typedNextExpr)

		case pgsql.Values:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))

			for idx := len(typedNextExpr.Values) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr.Values[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral(", "))
				}
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("values ("))

		case pgsql.ArrayLiteral:
			if typedNextExpr.CastType != pgsql.UnsetDataType {
				if arrayCastType, err := typedNextExpr.CastType.ToArrayType(); err != nil {
					return err
				} else {
					exprStack = append(exprStack, pgsql.FormattingLiteral(arrayCastType.String()), pgsql.FormattingLiteral("::"))
				}
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("]"))

			for idx := len(typedNextExpr.Values) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr.Values[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral(", "))
				}
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("array ["))

		case pgsql.DataType:
			exprStack = append(exprStack, pgsql.FormattingLiteral(typedNextExpr.String()))

		case pgsql.CompositeValue:
			exprStack = append(exprStack, typedNextExpr.DataType)
			exprStack = append(exprStack, pgsql.FormattingLiteral("::"))
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))

			for idx := len(typedNextExpr.Values) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr.Values[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral(", "))
				}
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("("))

		case pgsql.DoUpdate:
			if typedNextExpr.Where != nil {
				exprStack = append(exprStack, typedNextExpr.Where)
				exprStack = append(exprStack, pgsql.FormattingLiteral(" where "))
			}

			if len(typedNextExpr.Assignments) > 0 {
				for idx := len(typedNextExpr.Assignments) - 1; idx >= 0; idx-- {
					exprStack = append(exprStack, typedNextExpr.Assignments[idx])
				}

				exprStack = append(exprStack, pgsql.FormattingLiteral(" set "))
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("do update"))

		case pgsql.ConflictTarget:
			if len(typedNextExpr.Columns) > 0 {
				if len(typedNextExpr.Constraint) > 0 {
					return fmt.Errorf("conflict target has both columns and an 'on constraint' expression set")
				}

				exprStack = append(exprStack, pgsql.FormattingLiteral(")"))

				for idx := len(typedNextExpr.Columns) - 1; idx >= 0; idx-- {
					exprStack = append(exprStack, typedNextExpr.Columns[idx])

					if idx > 0 {
						exprStack = append(exprStack, pgsql.FormattingLiteral(", "))
					}
				}

				exprStack = append(exprStack, pgsql.FormattingLiteral("("))
			}

			if len(typedNextExpr.Constraint) > 0 {
				if len(typedNextExpr.Columns) > 0 {
					return fmt.Errorf("conflict target has both columns and an 'on constraint' expression set")
				}

				exprStack = append(exprStack, typedNextExpr.Constraint, pgsql.FormattingLiteral("on constraint "))
			}

		case *pgsql.AliasedExpression:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.AliasedExpression:
			if typedNextExpr.Alias.Set {
				exprStack = append(exprStack, typedNextExpr.Alias.Value)
				exprStack = append(exprStack, pgsql.FormattingLiteral(" as "))
				exprStack = append(exprStack, typedNextExpr.Expression)
			} else {
				exprStack = append(exprStack, typedNextExpr.Expression)
			}

		case *pgsql.AnyExpression:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			exprStack = append(exprStack, typedNextExpr.Expression)
			exprStack = append(exprStack, pgsql.FormattingLiteral("any ("))

		case pgsql.AllExpression:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			exprStack = append(exprStack, typedNextExpr.Expression)
			exprStack = append(exprStack, pgsql.FormattingLiteral("all ("))

		case pgsql.ArrayExpression:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			exprStack = append(exprStack, typedNextExpr.Expression)

			exprStack = append(exprStack, pgsql.FormattingLiteral("array("))

		case pgsql.ArrayIndex:
			exprStack = append(exprStack, pgsql.FormattingLiteral("]"))

			for idx := len(typedNextExpr.Indexes) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr.Indexes[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral(", "))
				}
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("["))
			exprStack = append(exprStack, typedNextExpr.Expression)

		case *pgsql.ArrayIndex:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.TypeCast:
			switch typedCastedExpr := typedNextExpr.Expression.(type) {
			case *pgsql.BinaryExpression:
				if typedCastedExpr.Operator == pgsql.OperatorJSONTextField && typedNextExpr.CastType == pgsql.Text {
					// Avoid formatting property lookups wrapped in text type casts
					exprStack = append(exprStack, typedNextExpr.Expression)
				} else {
					exprStack = append(exprStack, pgsql.FormattingLiteral(typedNextExpr.CastType), pgsql.FormattingLiteral(")::"))
					exprStack = append(exprStack, typedNextExpr.Expression)
					exprStack = append(exprStack, pgsql.FormattingLiteral("("))
				}

			case pgsql.Parenthetical:
				// Avoid formatting type-casted parenthetical statements as (('test'))::text - this should instead look like ('test')::text
				exprStack = append(exprStack, pgsql.FormattingLiteral(typedNextExpr.CastType), pgsql.FormattingLiteral("::"))
				exprStack = append(exprStack, typedNextExpr.Expression)

			default:
				exprStack = append(exprStack, pgsql.FormattingLiteral(typedNextExpr.CastType), pgsql.FormattingLiteral(")::"))
				exprStack = append(exprStack, typedNextExpr.Expression)
				exprStack = append(exprStack, pgsql.FormattingLiteral("("))
			}

		case pgsql.Parenthetical:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			exprStack = append(exprStack, typedNextExpr.Expression)
			exprStack = append(exprStack, pgsql.FormattingLiteral("("))

		case *pgsql.Parenthetical:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.Parameter:
			if builder.MaterializeParameters {
				if parameterValue, hasParameter := builder.parameters[typedNextExpr.Identifier.String()]; !hasParameter {
					return fmt.Errorf("invalid parameter %s", typedNextExpr.Identifier.String())
				} else if parameterLiteral, err := pgsql.AsLiteral(parameterValue); err != nil {
					return fmt.Errorf("invalid parameter value for %s: %v", typedNextExpr.Identifier.String(), err)
				} else {
					exprStack = append(exprStack, parameterLiteral)
				}
			} else {
				if typedNextExpr.CastType != pgsql.UnsetDataType {
					exprStack = append(exprStack, typedNextExpr.CastType, pgsql.FormattingLiteral("::"))
				}

				exprStack = append(exprStack, typedNextExpr.Identifier, pgsql.FormattingLiteral("@"))
			}

		case *pgsql.Parameter:
			exprStack = append(exprStack, *typedNextExpr)

		case pgsql.Variadic:
			exprStack = append(exprStack, typedNextExpr.Expression, pgsql.FormattingLiteral("variadic "))

		case pgsql.RowColumnReference:
			exprStack = append(exprStack, typedNextExpr.Column, pgsql.FormattingLiteral(")."), typedNextExpr.Identifier, pgsql.FormattingLiteral("("))

		case pgsql.ExistsExpression:
			exprStack = append(exprStack, typedNextExpr.Subquery, pgsql.FormattingLiteral("exists "))

			if typedNextExpr.Negated {
				exprStack = append(exprStack, pgsql.FormattingLiteral("not "))
			}

		case pgsql.ProjectionFrom:
			for idx, projection := range typedNextExpr.Projection {
				if idx > 0 {
					builder.Write(", ")
				}

				if err := formatNode(builder, projection); err != nil {
					return err
				}
			}

			if len(typedNextExpr.From) > 0 {
				builder.Write(" from ")

				if err := formatFromClauses(builder, typedNextExpr.From); err != nil {
					return err
				}
			}

		case pgsql.Subquery:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"), typedNextExpr.Query, pgsql.FormattingLiteral("("))

		case pgsql.SyntaxNodeFuture:
			exprStack = append(exprStack, typedNextExpr.Unwrap())

		default:
			return fmt.Errorf("unable to format pgsql node type: %T", nextExpr)
		}
	}

	return nil
}

func Expression(expression pgsql.SyntaxNode, builder *OutputBuilder) (string, error) {
	if err := formatNode(builder, expression); err != nil {
		return "", err
	}

	return builder.Build(), nil
}

func formatSelect(builder *OutputBuilder, selectStmt pgsql.Select) error {
	builder.Write("select ")

	for idx, projection := range selectStmt.Projection {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatNode(builder, projection); err != nil {
			return err
		}
	}

	if len(selectStmt.From) > 0 {
		builder.Write(" from ")

		if err := formatFromClauses(builder, selectStmt.From); err != nil {
			return err
		}
	}

	if selectStmt.Where != nil {
		builder.Write(" where ")

		if err := formatNode(builder, selectStmt.Where); err != nil {
			return err
		}
	}

	if len(selectStmt.GroupBy) > 0 {
		builder.Write(" group by ")

		if err := formatGroupBy(builder, selectStmt.GroupBy); err != nil {
			return err
		}
	}

	return nil
}

func formatGroupBy(builder *OutputBuilder, groupByExpressions []pgsql.Expression) error {
	for idx, groupByExpression := range groupByExpressions {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatNode(builder, groupByExpression); err != nil {
			return err
		}
	}

	return nil
}

func formatFromClauses(builder *OutputBuilder, fromClauses []pgsql.FromClause) error {
	for idx, fromClause := range fromClauses {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatNode(builder, fromClause.Source); err != nil {
			return err
		}

		for _, join := range fromClause.Joins {
			builder.Write(" ")

			switch join.JoinOperator.JoinType {
			case pgsql.JoinTypeInner:
				// A bare join keyword is also an alias for an inner join

			case pgsql.JoinTypeLeftOuter:
				builder.Write("left outer ")

			case pgsql.JoinTypeRightOuter:
				builder.Write("right outer ")

			case pgsql.JoinTypeFullOuter:
				builder.Write("full outer ")

			default:
				return fmt.Errorf("unsupported join type: %d", join.JoinOperator.JoinType)
			}

			builder.Write("join ")

			if err := formatNode(builder, join.Table); err != nil {
				return err
			}

			if join.JoinOperator.Constraint != nil {
				builder.Write(" on ")

				if err := formatNode(builder, join.JoinOperator.Constraint); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func formatTableAlias(builder *OutputBuilder, tableAlias pgsql.TableAlias) error {
	builder.Write(tableAlias.Name)

	if tableAlias.Shape.Set {
		builder.Write("(")

		for idx, column := range tableAlias.Shape.Value.Columns {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatNode(builder, column); err != nil {
				return err
			}
		}

		builder.Write(")")
	}

	return nil
}

func formatCommonTableExpressions(builder *OutputBuilder, commonTableExpressions pgsql.With) error {
	builder.Write("with ")

	if commonTableExpressions.Recursive {
		builder.Write("recursive ")
	}

	for idx, commonTableExpression := range commonTableExpressions.Expressions {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatTableAlias(builder, commonTableExpression.Alias); err != nil {
			return err
		}

		builder.Write(" as ")

		if commonTableExpression.Materialized.Set {
			if err := formatNode(builder, commonTableExpression.Materialized.Value); err != nil {
				return err
			}

			builder.Write(" ")
		}

		builder.Write("(")

		if err := formatSetExpression(builder, commonTableExpression.Query); err != nil {
			return err
		}

		builder.Write(")")
	}

	// Leave a trailing space after formatting CTEs for the subsequent query body
	builder.Write(" ")

	return nil
}

func formatSetExpression(builder *OutputBuilder, expression pgsql.SetExpression) error {
	switch typedSetExpression := expression.(type) {
	case pgsql.Query:
		if typedSetExpression.CommonTableExpressions != nil {
			if err := formatCommonTableExpressions(builder, *typedSetExpression.CommonTableExpressions); err != nil {
				return err
			}
		}

		if err := formatSetExpression(builder, typedSetExpression.Body); err != nil {
			return err
		}

		if len(typedSetExpression.OrderBy) > 0 {
			builder.Write(" order by ")

			for idx, orderBy := range typedSetExpression.OrderBy {
				if idx > 0 {
					builder.Write(", ")
				}

				if err := formatNode(builder, orderBy); err != nil {
					return err
				}
			}
		}

		if typedSetExpression.Offset.Set {
			builder.Write(" offset ")

			if err := formatNode(builder, typedSetExpression.Offset.Value); err != nil {
				return err
			}
		}

		if typedSetExpression.Limit.Set {
			builder.Write(" limit ")

			if err := formatNode(builder, typedSetExpression.Limit.Value); err != nil {
				return err
			}
		}

		builder.Write()

	case pgsql.Select:
		return formatSelect(builder, typedSetExpression)

	case pgsql.Delete:
		return formatDeleteStatement(builder, typedSetExpression)

	case pgsql.SetOperation:
		if typedSetExpression.All && typedSetExpression.Distinct {
			return fmt.Errorf("set operation for query may not be both ALL and DISTINCT")
		}

		if err := formatSetExpression(builder, typedSetExpression.LOperand); err != nil {
			return err
		}

		builder.Write(" ")

		if err := formatNode(builder, typedSetExpression.Operator); err != nil {
			return err
		}

		builder.Write(" ")

		if typedSetExpression.All {
			builder.Write("all ")
		}

		if typedSetExpression.Distinct {
			builder.Write("distinct ")
		}

		if err := formatSetExpression(builder, typedSetExpression.ROperand); err != nil {
			return err
		}

	case pgsql.Values:
		return formatNode(builder, typedSetExpression)

	case pgsql.Update:
		return formatUpdateStatement(builder, typedSetExpression)

	default:
		return fmt.Errorf("unsupported set expression type %T", expression)
	}

	return nil
}

func formatMergeStatement(builder *OutputBuilder, merge pgsql.Merge) error {
	builder.Write("merge ")

	if merge.Into {
		builder.Write("into ")
	}

	if err := formatNode(builder, merge.Table); err != nil {
		return err
	}

	builder.Write(" using ")

	if err := formatNode(builder, merge.Source); err != nil {
		return err
	}

	builder.Write(" on ")

	if err := formatNode(builder, merge.JoinTarget); err != nil {
		return err
	}

	builder.Write(" ")

	for idx, mergeAction := range merge.Actions {
		if idx > 0 {
			builder.Write(" ")
		}

		builder.Write("when ")

		switch typedMergeAction := mergeAction.(type) {
		case pgsql.MatchedUpdate:
			builder.Write("matched")

			// Predicate is optional
			if typedMergeAction.Predicate != nil {
				builder.Write(" and ")

				if err := formatNode(builder, typedMergeAction.Predicate); err != nil {
					return err
				}
			}

			builder.Write(" then update set ")

			for idx, assignment := range typedMergeAction.Assignments {
				if idx > 0 {
					builder.Write(", ")
				}

				if err := formatNode(builder, assignment); err != nil {
					return err
				}
			}

		case pgsql.MatchedDelete:
			builder.Write("matched")

			// Predicate is optional
			if typedMergeAction.Predicate != nil {
				builder.Write(" and ")

				if err := formatNode(builder, typedMergeAction.Predicate); err != nil {
					return err
				}
			}

			builder.Write(" then delete")

		case pgsql.UnmatchedAction:
			builder.Write("not matched")

			// Predicate is optional
			if typedMergeAction.Predicate != nil {
				builder.Write(" and ")

				if err := formatNode(builder, typedMergeAction.Predicate); err != nil {
					return err
				}
			}

			builder.Write(" then insert (")

			for idx, column := range typedMergeAction.Columns {
				if idx > 0 {
					builder.Write(", ")
				}

				if err := formatNode(builder, column); err != nil {
					return err
				}
			}

			builder.Write(") ")

			if err := formatNode(builder, typedMergeAction.Values); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown merge action type: %T", mergeAction)
		}
	}

	return nil
}

func formatInsertStatement(builder *OutputBuilder, insert pgsql.Insert) error {
	builder.Write("insert into ")

	if err := formatNode(builder, insert.Table); err != nil {
		return err
	}

	if len(insert.Shape.Columns) > 0 {
		builder.Write(" (")

		for idx, column := range insert.Shape.Columns {
			if idx > 0 {
				builder.Write(", ")
			}

			builder.Write(column)
		}

		builder.Write(")")
	}

	builder.Write(" ")

	if insert.Source != nil {
		if err := formatSetExpression(builder, *insert.Source); err != nil {
			return err
		}
	}

	if insert.OnConflict != nil {
		builder.Write(" on conflict ")

		if insert.OnConflict.Target != nil {
			if err := formatNode(builder, *insert.OnConflict.Target); err != nil {
				return err
			}

			builder.Write(" ")
		}

		if err := formatNode(builder, insert.OnConflict.Action); err != nil {
			return err
		}
	}

	if len(insert.Returning) > 0 {
		builder.Write(" returning ")

		for idx, projection := range insert.Returning {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatNode(builder, projection); err != nil {
				return err
			}
		}
	}

	return nil
}

func formatUpdateStatement(builder *OutputBuilder, update pgsql.Update) error {
	builder.Write("update ")

	if err := formatNode(builder, update.Table); err != nil {
		return err
	}

	builder.Write(" set ")

	for idx, assignment := range update.Assignments {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatNode(builder, assignment); err != nil {
			return err
		}
	}

	if len(update.From) > 0 {
		builder.Write(" from ")

		if err := formatFromClauses(builder, update.From); err != nil {
			return err
		}
	}

	if update.Where.Set {
		builder.Write(" where ")

		if err := formatNode(builder, update.Where.Value); err != nil {
			return err
		}
	}

	if len(update.Returning) > 0 {
		builder.Write(" returning ")

		for idx, projection := range update.Returning {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatNode(builder, projection); err != nil {
				return err
			}
		}
	}

	return nil
}

func formatDeleteStatement(builder *OutputBuilder, sqlDelete pgsql.Delete) error {
	builder.Write("delete from ")

	for idx, tableRef := range sqlDelete.From {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatNode(builder, tableRef); err != nil {
			return err
		}
	}

	if len(sqlDelete.Using) > 0 {
		builder.Write(" using ")

		if err := formatFromClauses(builder, sqlDelete.Using); err != nil {
			return err
		}
	}

	if sqlDelete.Where.Set {
		builder.Write(" where ")

		if err := formatNode(builder, sqlDelete.Where.Value); err != nil {
			return err
		}
	}

	if len(sqlDelete.Returning) > 0 {
		builder.Write(" returning ")

		for idx, projection := range sqlDelete.Returning {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatNode(builder, projection); err != nil {
				return err
			}
		}
	}

	return nil
}

func Statement(statement pgsql.Statement, builder *OutputBuilder) (string, error) {
	switch typedStatement := statement.(type) {
	case pgsql.Merge:
		if err := formatMergeStatement(builder, typedStatement); err != nil {
			return "", err
		}

	case pgsql.Query:
		if err := formatSetExpression(builder, typedStatement); err != nil {
			return "", err
		}

	case pgsql.Insert:
		if err := formatInsertStatement(builder, typedStatement); err != nil {
			return "", err
		}

	case pgsql.Update:
		if err := formatUpdateStatement(builder, typedStatement); err != nil {
			return "", err
		}

	case pgsql.Delete:
		if err := formatDeleteStatement(builder, typedStatement); err != nil {
			return "", err
		}

	default:
		return "", fmt.Errorf("unsupported PgSQL statement type: %T", statement)
	}

	builder.Write(";")
	return builder.Build(), nil
}

func SyntaxNode(node pgsql.SyntaxNode) (string, error) {
	builder := NewOutputBuilder()

	switch typedNode := node.(type) {
	case pgsql.Statement:
		return Statement(typedNode, builder)

	case pgsql.Expression:
		return Expression(typedNode, builder)

	default:
		return "", fmt.Errorf("unknown SQL AST type: %T", node)
	}
}

type Formatted struct {
	Statement  string
	Parameters map[string]any
}
