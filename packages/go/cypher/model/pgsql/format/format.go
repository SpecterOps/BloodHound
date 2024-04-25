package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type FormattedOutput struct {
	Value      string
	Parameters map[string]any
}

type OutputBuilder struct {
	builder    *strings.Builder
	parameters map[string]any
}

func NewOutputBuilder() OutputBuilder {
	return OutputBuilder{
		builder:    &strings.Builder{},
		parameters: map[string]any{},
	}
}

func (s OutputBuilder) Write(values ...any) {
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

func (s OutputBuilder) Build() FormattedOutput {
	return FormattedOutput{
		Value:      s.builder.String(),
		Parameters: s.parameters,
	}
}

func formatLiteral(builder OutputBuilder, literal pgsql.Literal) error {
	switch typedLiteral := literal.Value.(type) {
	case uint:
		builder.Write(strconv.FormatUint(uint64(typedLiteral), 10))
	case uint8:
		builder.Write(strconv.FormatUint(uint64(typedLiteral), 10))
	case uint16:
		builder.Write(strconv.FormatUint(uint64(typedLiteral), 10))
	case uint32:
		builder.Write(strconv.FormatUint(uint64(typedLiteral), 10))
	case uint64:
		builder.Write(strconv.FormatUint(typedLiteral, 10))
	case int:
		builder.Write(strconv.FormatInt(int64(typedLiteral), 10))
	case int8:
		builder.Write(strconv.FormatInt(int64(typedLiteral), 10))
	case int16:
		builder.Write(strconv.FormatInt(int64(typedLiteral), 10))
	case int32:
		builder.Write(strconv.FormatInt(int64(typedLiteral), 10))
	case int64:
		builder.Write(strconv.FormatInt(typedLiteral, 10))
	case string:
		builder.Write("'", typedLiteral, "'")
	case bool:
		builder.Write(strconv.FormatBool(typedLiteral))
	case graph.Kinds:
		builder.Write("array []::int2[]")
	default:
		return fmt.Errorf("unsupported literal type: %T", literal.Value)
	}

	return nil
}

func formatExpression(builder OutputBuilder, rootExpr pgsql.SyntaxNode) error {
	exprStack := []pgsql.SyntaxNode{
		rootExpr,
	}

	for len(exprStack) > 0 {
		nextExpr := exprStack[len(exprStack)-1]
		exprStack = exprStack[:len(exprStack)-1]

		switch typedNextExpr := nextExpr.(type) {
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

		case pgsql.FunctionCall:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))

			for idx := len(typedNextExpr.Parameters) - 1; idx >= 0; idx-- {
				exprStack = append(exprStack, typedNextExpr.Parameters[idx])

				if idx > 0 {
					exprStack = append(exprStack, pgsql.FormattingLiteral(","))
				}
			}

			if typedNextExpr.Distinct {
				exprStack = append(exprStack, pgsql.FormattingLiteral("distinct "))
			}

			exprStack = append(exprStack, pgsql.FormattingLiteral("("))
			exprStack = append(exprStack, typedNextExpr.Function)

		case pgsql.Operator:
			builder.Write(typedNextExpr)

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
				typedNextExpr.Value,
				pgsql.FormattingLiteral(" "),
				pgsql.Operator("="),
				pgsql.FormattingLiteral(" "),
				typedNextExpr.Identifier)

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
			if typedNextExpr.TypeHint != pgsql.UnsetDataType {
				exprStack = append(exprStack, pgsql.FormattingLiteral(typedNextExpr.TypeHint.String()), pgsql.FormattingLiteral("::"))
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

		case pgsql.AliasedExpression:
			if typedNextExpr.Alias != "" {
				exprStack = append(exprStack, typedNextExpr.Alias)
				exprStack = append(exprStack, pgsql.FormattingLiteral(" as "))
				exprStack = append(exprStack, typedNextExpr.Expression)
			} else {
				exprStack = append(exprStack, typedNextExpr.Expression)
			}

		case pgsql.AnyExpression:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			exprStack = append(exprStack, typedNextExpr.Expression)
			exprStack = append(exprStack, pgsql.FormattingLiteral("any ("))

		case pgsql.AllExpression:
			exprStack = append(exprStack, pgsql.FormattingLiteral(")"))
			exprStack = append(exprStack, typedNextExpr.Expression)
			exprStack = append(exprStack, pgsql.FormattingLiteral("all ("))

		default:
			return fmt.Errorf("unsupported expression type: %T", nextExpr)
		}
	}

	return nil
}

func Expression(expression pgsql.SyntaxNode) (FormattedOutput, error) {
	var (
		builder = NewOutputBuilder()
		err     = formatExpression(builder, expression)
	)

	return builder.Build(), err
}

func formatSelect(builder OutputBuilder, selectStmt pgsql.Select) error {
	builder.Write("select ")

	for idx, projection := range selectStmt.Projection {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatExpression(builder, projection); err != nil {
			return err
		}
	}

	builder.Write(" ")

	if len(selectStmt.From) > 0 {
		builder.Write("from ")

		for idx, fromClause := range selectStmt.From {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatExpression(builder, fromClause.Relation); err != nil {
				return err
			}

			if len(fromClause.Joins) > 0 {
				builder.Write(" ")

				for idx, join := range fromClause.Joins {
					if idx > 0 {
						builder.Write(" ")
					}

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

					if err := formatExpression(builder, join.Table); err != nil {
						return err
					}

					builder.Write(" on ")

					if err := formatExpression(builder, join.JoinOperator.Constraint); err != nil {
						return err
					}
				}
			}
		}
	}

	if selectStmt.Where != nil {
		builder.Write(" where ")

		if err := formatExpression(builder, selectStmt.Where); err != nil {
			return err
		}
	}

	return nil
}

func formatTableAlias(builder OutputBuilder, tableAlias pgsql.TableAlias) error {
	builder.Write(tableAlias.Name)

	if tableAlias.Shape.Set {
		builder.Write("(")

		for idx, column := range tableAlias.Shape.Value.Columns {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatExpression(builder, column); err != nil {
				return err
			}
		}

		builder.Write(")")
	}

	return nil
}

func formatCommonTableExpressions(builder OutputBuilder, commonTableExpressions pgsql.With) error {
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

		if commonTableExpression.Materialized != nil {
			if err := formatExpression(builder, *commonTableExpression.Materialized); err != nil {
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

func formatSetExpression(builder OutputBuilder, expression pgsql.SetExpression) error {
	switch typedSetExpression := expression.(type) {
	case pgsql.Query:
		if typedSetExpression.CommonTableExpressions != nil {
			if err := formatCommonTableExpressions(builder, *typedSetExpression.CommonTableExpressions); err != nil {
				return err
			}
		}

		return formatSetExpression(builder, typedSetExpression.Body)

	case pgsql.Select:
		return formatSelect(builder, typedSetExpression)

	case pgsql.SetOperation:
		if typedSetExpression.All && typedSetExpression.Distinct {
			return fmt.Errorf("set operation for query may not be both ALL and DISTINCT")
		}

		if err := formatSetExpression(builder, typedSetExpression.LOperand); err != nil {
			return err
		}

		builder.Write(" ")

		if err := formatExpression(builder, typedSetExpression.Operator); err != nil {
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
		if err := formatExpression(builder, typedSetExpression); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported set expression type %T", expression)
	}

	return nil
}

func formatMergeStatement(builder OutputBuilder, merge pgsql.Merge) error {
	builder.Write("merge ")

	if merge.Into {
		builder.Write("into ")
	}

	if err := formatExpression(builder, merge.Table); err != nil {
		return err
	}

	builder.Write(" using ")

	if err := formatExpression(builder, merge.Source); err != nil {
		return err
	}

	builder.Write(" on ")

	if err := formatExpression(builder, merge.JoinTarget); err != nil {
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

				if err := formatExpression(builder, typedMergeAction.Predicate); err != nil {
					return err
				}
			}

			builder.Write(" then update set ")

			for idx, assignment := range typedMergeAction.Assignments {
				if idx > 0 {
					builder.Write(", ")
				}

				if err := formatExpression(builder, assignment); err != nil {
					return err
				}
			}

		case pgsql.MatchedDelete:
			builder.Write("matched")

			// Predicate is optional
			if typedMergeAction.Predicate != nil {
				builder.Write(" and ")

				if err := formatExpression(builder, typedMergeAction.Predicate); err != nil {
					return err
				}
			}

			builder.Write(" then delete")

		case pgsql.UnmatchedAction:
			builder.Write("not matched")

			// Predicate is optional
			if typedMergeAction.Predicate != nil {
				builder.Write(" and ")

				if err := formatExpression(builder, typedMergeAction.Predicate); err != nil {
					return err
				}
			}

			builder.Write(" then insert (")

			for idx, column := range typedMergeAction.Columns {
				if idx > 0 {
					builder.Write(", ")
				}

				if err := formatExpression(builder, column); err != nil {
					return err
				}
			}

			builder.Write(") ")

			if err := formatExpression(builder, typedMergeAction.Values); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown merge action type: %T", mergeAction)
		}
	}

	return nil
}

func formatInsertStatement(builder OutputBuilder, insert pgsql.Insert) error {
	builder.Write("insert into ")

	if err := formatExpression(builder, insert.Table); err != nil {
		return err
	}

	if len(insert.Columns) > 0 {
		builder.Write(" (")

		for idx, column := range insert.Columns {
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
			if err := formatExpression(builder, *insert.OnConflict.Target); err != nil {
				return err
			}

			builder.Write(" ")
		}

		if err := formatExpression(builder, insert.OnConflict.Action); err != nil {
			return err
		}
	}

	if len(insert.Returning) > 0 {
		builder.Write(" returning ")

		for idx, projection := range insert.Returning {
			if idx > 0 {
				builder.Write(", ")
			}

			if err := formatExpression(builder, projection); err != nil {
				return err
			}
		}
	}

	return nil
}

func formatUpdateStatement(builder OutputBuilder, update pgsql.Update) error {
	builder.Write("update ")

	if err := formatExpression(builder, update.Table); err != nil {
		return err
	}

	builder.Write(" set ")

	for idx, assignment := range update.Assignments {
		if idx > 0 {
			builder.Write(", ")
		}

		if err := formatExpression(builder, assignment); err != nil {
			return err
		}
	}

	if update.Where != nil {
		builder.Write(" where ")

		if err := formatExpression(builder, update.Where); err != nil {
			return err
		}
	}

	return nil
}

func formatDeleteStatement(builder OutputBuilder, delete pgsql.Delete) error {
	builder.Write("delete from ")

	if err := formatExpression(builder, delete.Table); err != nil {
		return err
	}

	if delete.Where != nil {
		builder.Write(" where ")

		if err := formatExpression(builder, delete.Where); err != nil {
			return err
		}
	}

	return nil
}

func Statement(statement pgsql.Statement) (FormattedOutput, error) {
	builder := NewOutputBuilder()

	switch typedStatement := statement.(type) {
	case pgsql.Merge:
		if err := formatMergeStatement(builder, typedStatement); err != nil {
			return FormattedOutput{}, err
		}

	case pgsql.Query:
		if err := formatSetExpression(builder, typedStatement); err != nil {
			return FormattedOutput{}, err
		}

	case pgsql.Insert:
		if err := formatInsertStatement(builder, typedStatement); err != nil {
			return FormattedOutput{}, err
		}

	case pgsql.Update:
		if err := formatUpdateStatement(builder, typedStatement); err != nil {
			return FormattedOutput{}, err
		}

	case pgsql.Delete:
		if err := formatDeleteStatement(builder, typedStatement); err != nil {
			return FormattedOutput{}, err
		}

	default:
		return FormattedOutput{}, fmt.Errorf("unsupported PgSQL statement type: %T", statement)
	}

	builder.Write(";")
	return builder.Build(), nil
}

func SyntaxNode(node pgsql.SyntaxNode) (FormattedOutput, error) {
	switch typedNode := node.(type) {
	case pgsql.Statement:
		return Statement(typedNode)

	case pgsql.Expression:
		return Expression(typedNode)

	default:
		return FormattedOutput{}, fmt.Errorf("unknown SQL AST type: %T", node)
	}
}
