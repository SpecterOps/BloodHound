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

package pgsql

import (
	"context"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"

	"github.com/specterops/bloodhound/cypher/models"
)

// KindMapper is an interface that represents a service that can map a given slice of graph.Kind to a slice of
// int16 numeric identifiers.
type KindMapper interface {
	MapKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error)
	AssertKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error)
}

// FormattingLiteral is a syntax node that is used as a transparent formatting syntax node. The formatter will
// take the string value and emit it as-is.
type FormattingLiteral string

func (s FormattingLiteral) AsExpression() Expression {
	return s
}

func (s FormattingLiteral) NodeType() string {
	return "formatting_literal"
}

func (s FormattingLiteral) String() string {
	return string(s)
}

// RecordShape defines a list of column identifiers that name a record's fields: (id, enrolled)
type RecordShape struct {
	Columns []Identifier
}

func (s RecordShape) NodeType() string {
	return "row_shape"
}

// TableAlias is an alias for a table with an optional record shape.
type TableAlias struct {
	Name  Identifier
	Shape models.Optional[RecordShape]
}

func (s TableAlias) NodeType() string {
	return "table_alias"
}

// Values represents a list of expressions to be taken as a value array:
// insert into table (id, enrolled) values (1, true);
type Values struct {
	Values []Expression
}

func (s Values) AsExpression() Expression {
	return s
}

func (s Values) AsSetExpression() SetExpression {
	return s
}

func (s Values) NodeType() string {
	return "values"
}

// Case represents a pattern matching syntax node:
//
// CASE
//
//	WHEN condition1 THEN result1
//	WHEN condition2 THEN result2
//	WHEN conditionN THEN resultN
//	ELSE result
//
// END;
type Case struct {
	Operand    Expression
	Conditions []Expression
	Then       []Expression
	Else       Expression
}

// InExpression represents a contains operation against a list of evaluated expressions:
// m.identifier in (val1, val2, ...)
type InExpression struct {
	Expression Expression
	List       []Expression
	Negated    bool
}

// InSubquery represents a contains expression against a suq-query.
// [not] in (<Select> ...)
type InSubquery struct {
	Expression Expression
	Query      Query
	Negated    bool
}

// Between represents a comparison expression between a low expression and high expression.
// <expr> [not] between <low> and <high>
type Between struct {
	Expression Expression
	Low        Expression
	High       Expression
	Negated    bool
}

// Variadic isrepresents a variadic expansion of the given expression.
type Variadic struct {
	Expression Expression
}

func (s Variadic) NodeType() string {
	return "variadic"
}

func (s Variadic) AsExpression() Expression {
	return s
}

// TypeCast wraps a given expression in a type cast that matches the given DataType.
type TypeCast struct {
	Expression Expression
	CastType   DataType
}

func (s TypeCast) NodeType() string {
	return "type_cast"
}

func (s TypeCast) AsExpression() Expression {
	return s
}

func (s TypeCast) TypeHint() DataType {
	return s.CastType
}

func NewTypeCast(expression Expression, dataType DataType) TypeHinted {
	if typeCast, isTypeCast := expression.(TypeCast); isTypeCast {
		typeCast.CastType = dataType
		return typeCast
	}

	return TypeCast{
		Expression: expression,
		CastType:   dataType,
	}
}

// Literal is a type-hinted literal SQL value.
type Literal struct {
	Value    any
	Null     bool
	CastType DataType
}

func NewLiteral(value any, dataType DataType) Literal {
	return Literal{
		Value:    value,
		CastType: dataType,
	}
}

func (s Literal) TypeHint() DataType {
	if s.CastType == UnsetDataType {
		return UnknownDataType
	}

	return s.CastType
}

func (s Literal) AsExpression() Expression {
	return s
}

func (s Literal) AsSelectItem() SelectItem {
	return s
}

func AsLiteral(value any) (Literal, error) {
	if value == nil {
		return Literal{
			Value: value,
			Null:  true,
		}, nil
	}

	if dataType, err := ValueToDataType(value); err != nil {
		return Literal{}, err
	} else if negotiatedValue, err := NegotiateValue(value); err != nil {
		return Literal{}, err
	} else {
		return Literal{
			Value:    negotiatedValue,
			CastType: dataType,
		}, nil
	}
}

func (s Literal) NodeType() string {
	return "literal"
}

type Future[U any] struct {
	SyntaxNode SyntaxNode
	Data       U
	DataType   DataType
}

func NewFuture[U any](data U, dataType DataType) *Future[U] {
	return &Future[U]{
		Data:     data,
		DataType: dataType,
	}
}

func (s Future[U]) Satisfied() bool {
	return s.SyntaxNode != nil
}

func (s Future[U]) Unwrap() SyntaxNode {
	return s.SyntaxNode
}

func (s Future[U]) TypeHint() DataType {
	return s.DataType
}

func (s Future[U]) NodeType() string {
	var (
		emptyU U
	)

	return fmt.Sprintf("syntax_node_future[%T]", emptyU)
}

func (s Future[U]) AsExpression() Expression {
	return s
}

type Subquery struct {
	Query Query
}

func (s Subquery) NodeType() string {
	return "subquery"
}

func (s Subquery) AsExpression() Expression {
	return s
}

// not <expr>
type UnaryExpression struct {
	Operator Expression
	Operand  Expression
}

func NewUnaryExpression(operator Operator, operand Expression) *UnaryExpression {
	return &UnaryExpression{
		Operator: operator,
		Operand:  operand,
	}
}

func (s UnaryExpression) AsExpression() Expression {
	return s
}

func (s UnaryExpression) NodeType() string {
	return "unary_expression"
}

type LiteralNodeValue struct {
	Value    any
	Null     bool
	CastType DataType
}

// <expr> > <expr>
// table.column > 12345
type BinaryExpression struct {
	Operator Operator
	LOperand Expression
	ROperand Expression
}

func NewBinaryExpression(left Expression, operator Operator, right Expression) *BinaryExpression {
	return &BinaryExpression{
		Operator: operator,
		LOperand: left,
		ROperand: right,
	}
}

func NewJSONTextFieldLookup(reference CompoundIdentifier, field Identifier) *BinaryExpression {
	return NewBinaryExpression(reference, OperatorJSONTextField, field)
}

func (s BinaryExpression) AsExpression() Expression {
	return s
}

func (s BinaryExpression) AsAssignment() Assignment {
	return s
}

func (s BinaryExpression) AsSelectItem() SelectItem {
	return s
}

func (s BinaryExpression) NodeType() string {
	return "binary_expression"
}

func NewPropertyLookup(identifier CompoundIdentifier, reference Literal) *BinaryExpression {
	return NewBinaryExpression(
		identifier,
		OperatorPropertyLookup,
		reference,
	)
}

type CompositeValue struct {
	Values   []Expression
	DataType DataType
}

func (s CompositeValue) NodeType() string {
	return "composite_value"
}

func (s CompositeValue) AsExpression() Expression {
	return s
}

func (s CompositeValue) AsSelectItem() SelectItem {
	return s
}

// (<expr>)
type Parenthetical struct {
	Expression Expression
}

func NewParenthetical(expr Expression) *Parenthetical {
	return &Parenthetical{
		Expression: expr,
	}
}

func (s *Parenthetical) AsSelectItem() SelectItem {
	return s
}

func (s *Parenthetical) NodeType() string {
	return "parenthetical"
}

func (s *Parenthetical) AsExpression() Expression {
	return s
}

type JoinType int

const (
	JoinTypeInner JoinType = iota
	JoinTypeLeftOuter
	JoinTypeRightOuter
	JoinTypeFullOuter
)

type JoinOperator struct {
	JoinType   JoinType
	Constraint Expression
}

func (s JoinOperator) NodeType() string {
	return "join_operator"
}

type OrderBy struct {
	Expression Expression
	Ascending  bool
}

func NewOrderBy(ascending bool) *OrderBy {
	return &OrderBy{
		Ascending: ascending,
	}
}

func (s OrderBy) AsExpression() Expression {
	return s
}

func (s OrderBy) NodeType() string {
	return "order_by"
}

type WindowFrameUnit int

const (
	WindowFrameUnitRows WindowFrameUnit = iota
	WindowFrameUnitRange
	WindowFrameUnitGroups
)

type WindowFrameBoundaryType int

const (
	WindowFrameBoundaryTypeCurrentRow WindowFrameBoundaryType = iota
	WindowFrameBoundaryTypePreceding
	WindowFrameBoundaryTypeFollowing
)

type WindowFrameBoundary struct {
	BoundaryType    WindowFrameBoundaryType
	BoundaryLiteral *Literal
}

type WindowFrame struct {
	Unit          WindowFrameUnit
	StartBoundary WindowFrameBoundary
	EndBoundary   *WindowFrameBoundary
}

type Window struct {
	PartitionBy []Expression
	OrderBy     []OrderBy
	WindowFrame *WindowFrame
}

type AllExpression struct {
	Expression
}

func NewAllExpression(inner Expression) AllExpression {
	return AllExpression{
		Expression: inner,
	}
}

type AnyExpression struct {
	Expression
	CastType DataType
}

func NewAnyExpression(inner Expression, castType DataType) *AnyExpression {
	return &AnyExpression{
		Expression: inner,
		CastType:   castType,
	}
}

func NewAnyExpressionHinted(inner TypeHinted) *AnyExpression {
	return &AnyExpression{
		Expression: inner,
		CastType:   inner.TypeHint(),
	}
}

func (s AnyExpression) AsExpression() Expression {
	return s
}

func (s AnyExpression) NodeType() string {
	return "any"
}

func (s AnyExpression) TypeHint() DataType {
	return s.CastType
}

type Parameter struct {
	Identifier Identifier
	CastType   DataType
}

func (s Parameter) AsSelectItem() SelectItem {
	return s
}

func (s Parameter) NodeType() string {
	return "parameter"
}

func (s Parameter) AsExpression() Expression {
	return s
}

func (s Parameter) TypeHint() DataType {
	if s.CastType == UnsetDataType {
		return UnknownDataType
	}

	return s.CastType
}

func AsParameter(identifier Identifier, value any) (*Parameter, error) {
	parameter := &Parameter{
		Identifier: identifier,
	}

	if value != nil {
		if dataType, err := ValueToDataType(value); err != nil {
			return parameter, err
		} else {
			parameter.CastType = dataType
		}
	}

	return parameter, nil
}

type FunctionCall struct {
	Bare       bool
	Distinct   bool
	Function   Identifier
	Parameters []Expression
	Over       *Window
	CastType   DataType
}

func (s FunctionCall) AsAssignment() Assignment {
	return s
}

func (s FunctionCall) AsSelectItem() SelectItem {
	return s
}

func (s FunctionCall) AsExpression() Expression {
	return s
}

func (s FunctionCall) NodeType() string {
	return "function_call"
}

func (s FunctionCall) TypeHint() DataType {
	return s.CastType
}

type Join struct {
	Table        TableReference
	JoinOperator JoinOperator
}

func (s Join) NodeType() string {
	return "join"
}

type Identifier string

func AsIdentifiers(strs ...string) []Identifier {
	identifiers := make([]Identifier, len(strs))

	for idx, str := range strs {
		identifiers[idx] = Identifier(str)
	}

	return identifiers
}

func (s Identifier) AsCompoundIdentifier() CompoundIdentifier {
	return CompoundIdentifier{s}
}

func (s Identifier) AsSelectItem() SelectItem {
	return s
}

func (s Identifier) AsExpression() Expression {
	return s
}

func (s Identifier) NodeType() string {
	return "identifier"
}

func (s Identifier) String() string {
	return string(s)
}

func (s Identifier) Matches(others ...Identifier) bool {
	for _, other := range others {
		if s == other {
			return true
		}
	}

	return false
}

type ArrayIndex struct {
	Expression Expression
	Indexes    []Expression
}

func (s ArrayIndex) NodeType() string {
	return "array_index"
}

func (s ArrayIndex) AsExpression() Expression {
	return s
}

type RowColumnReference struct {
	Identifier Expression
	Column     Identifier
}

func (s RowColumnReference) NodeType() string {
	return "row_member_reference"
}

func (s RowColumnReference) AsExpression() Expression {
	return s
}

func (s RowColumnReference) AsSelectItem() SelectItem {
	return s
}

type CompoundIdentifier []Identifier

func (s CompoundIdentifier) Root() Identifier {
	return s[0]
}

func (s CompoundIdentifier) HasField() bool {
	return len(s) > 1
}

func (s CompoundIdentifier) Field() Identifier {
	return s[1]
}

func (s CompoundIdentifier) String() string {
	return strings.Join(s.Strings(), ".")
}

func (s CompoundIdentifier) Strings() []string {
	strCopy := make([]string, len(s))

	for idx, identifier := range s {
		strCopy[idx] = identifier.String()
	}

	return strCopy
}

func (s CompoundIdentifier) IsBlank() bool {
	return len(s) == 0
}

func (s CompoundIdentifier) Copy() CompoundIdentifier {
	copyInst := make(CompoundIdentifier, len(s))
	copy(copyInst, s)

	return copyInst
}

func (s CompoundIdentifier) AsExpression() Expression {
	return s
}

func (s CompoundIdentifier) AsSelectItem() SelectItem {
	return s
}

func (s CompoundIdentifier) NodeType() string {
	return "compound_identifier"
}

func AsCompoundIdentifier[T string | Identifier](parts ...T) CompoundIdentifier {
	compoundIdentifier := make(CompoundIdentifier, len(parts))

	for idx, part := range parts {
		switch typedPart := any(part).(type) {
		case string:
			compoundIdentifier[idx] = Identifier(typedPart)

		case Identifier:
			compoundIdentifier[idx] = typedPart
		}
	}

	return compoundIdentifier
}

type TableReference struct {
	Name    CompoundIdentifier
	Binding models.Optional[Identifier]
}

func (s TableReference) AsExpression() Expression {
	return s
}

func (s TableReference) NodeType() string {
	return "table_reference"
}

type FromClause struct {
	Source Expression
	Joins  []Join
}

func (s FromClause) NodeType() string {
	return "from"
}

type AliasedExpression struct {
	Expression Expression
	Alias      models.Optional[Identifier]
}

func (s AliasedExpression) NodeType() string {
	return "aliased_expression"
}

func (s AliasedExpression) AsExpression() Expression {
	return s
}

func (s AliasedExpression) AsSelectItem() SelectItem {
	return s
}

type Wildcard struct{}

func (s Wildcard) AsExpression() Expression {
	return s
}

func (s Wildcard) AsSelectItem() SelectItem {
	return s
}

func (s Wildcard) NodeType() string {
	return "wildcard"
}

type ArrayExpression struct {
	Expression Expression
}

func (s ArrayExpression) NodeType() string {
	return "array_expression"
}

func (s ArrayExpression) AsExpression() Expression {
	return s
}

type KindListLiteral struct {
	Values graph.Kinds
}

func (s KindListLiteral) NodeType() string {
	return "kind_list_literal"
}

func (s KindListLiteral) AsExpression() Expression {
	return s
}

type ArrayLiteral struct {
	Values   []Expression
	CastType DataType
}

func NewArrayLiteral[T any](values []T, castType DataType) (ArrayLiteral, error) {
	valuesCopy := make([]Expression, len(values))

	for idx, value := range values {
		switch typedValue := any(value).(type) {
		case Expression:
			valuesCopy[idx] = typedValue

		default:
			// Assume if it isn't an expression that it may be a bare literal and require wrapping
			if newLiteral, err := AsLiteral(value); err != nil {
				return ArrayLiteral{}, err
			} else {
				valuesCopy[idx] = newLiteral
			}
		}
	}

	return ArrayLiteral{
		Values:   valuesCopy,
		CastType: castType,
	}, nil
}

func (s ArrayLiteral) TypeHint() DataType {
	return s.CastType
}

func (s ArrayLiteral) AsExpression() Expression {
	return s
}

func (s ArrayLiteral) AsSelectItem() SelectItem {
	return s
}

func (s ArrayLiteral) NodeType() string {
	return "array"
}

type MatchedUpdate struct {
	Predicate   Expression
	Assignments []Assignment
}

func (s MatchedUpdate) NodeType() string {
	return "matched_update"
}

func (s MatchedUpdate) AsExpression() Expression {
	return s
}

func (s MatchedUpdate) AsMergeAction() MergeAction {
	return s
}

type MatchedDelete struct {
	Predicate Expression
}

func (s MatchedDelete) NodeType() string {
	return "matched_delete"
}

func (s MatchedDelete) AsExpression() Expression {
	return s
}

func (s MatchedDelete) AsMergeAction() MergeAction {
	return s
}

type UnmatchedAction struct {
	Predicate Expression
	Columns   []Identifier
	Values    Values
}

func (s UnmatchedAction) NodeType() string {
	return "unmatched_action"
}

func (s UnmatchedAction) AsExpression() Expression {
	return s
}

func (s UnmatchedAction) AsMergeAction() MergeAction {
	return s
}

type Merge struct {
	Into       bool
	Table      TableReference
	Source     TableReference
	JoinTarget Expression
	Actions    []MergeAction
}

func (s Merge) NodeType() string {
	return "merge"
}

func (s Merge) AsStatement() Statement {
	return s
}

type ConflictTarget struct {
	Columns    []Expression
	Constraint CompoundIdentifier
}

func (s ConflictTarget) NodeType() string {
	return "conflict_target"
}

func (s ConflictTarget) AsExpression() Expression {
	return s
}

type DoNothing struct{}

type DoUpdate struct {
	Assignments []Assignment
	Where       Expression
}

func (s DoUpdate) NodeType() string {
	return "do_update"
}

func (s DoUpdate) AsExpression() Expression {
	return s
}

func (s DoUpdate) AsConflictAction() ConflictAction {
	return s
}

// OnConflict is an expression that outlines the target columns of a conflict and the action to be taken
// when a conflict is encountered.
type OnConflict struct {
	Target *ConflictTarget
	Action ConflictAction
}

func (s OnConflict) NodeType() string {
	return "on_conflict"
}

func (s OnConflict) AsExpression() Expression {
	return s
}

// Insert is a SQL statement that is evaluated to insert data into a target table.
type Insert struct {
	Table      TableReference
	Shape      RecordShape
	OnConflict *OnConflict
	Source     *Query
	Returning  []SelectItem
}

func (s Insert) AsStatement() Statement {
	return s
}

func (s Insert) NodeType() string {
	return "insert"
}

// Delete is a SQL statement that is evaluated to remove data from a target table.
type Delete struct {
	From      []TableReference
	Using     []FromClause
	Where     models.Optional[Expression]
	Returning []SelectItem
}

func (s Delete) AsExpression() Expression {
	return s
}

func (s Delete) AsSetExpression() SetExpression {
	return s
}

func (s Delete) AsStatement() Statement {
	return s
}

func (s Delete) NodeType() string {
	return "delete"
}

// Update is a SQL statement that is evaluated to update data in a target table.
type Update struct {
	Table       TableReference
	Assignments []Assignment
	From        []FromClause
	Where       models.Optional[Expression]
	Returning   []SelectItem
}

func (s Update) AsExpression() Expression {
	return s
}

func (s Update) AsSetExpression() SetExpression {
	return s
}

func (s Update) AsStatement() Statement {
	return s
}

func (s Update) NodeType() string {
	return "update"
}

// Projection is an alias for a slice of SelectItem instances.
type Projection []SelectItem

func (s Projection) AsSyntaxNodes() []SyntaxNode {
	nodes := make([]SyntaxNode, len(s))

	for idx, value := range s {
		nodes[idx] = value
	}

	return nodes
}

func (s Projection) AsExpression() Expression {
	return s
}

func (s Projection) NodeType() string {
	return "projection"
}

type ProjectionFrom struct {
	Projection Projection
	From       []FromClause
}

func (s ProjectionFrom) NodeType() string {
	return "projection from"
}

func (s ProjectionFrom) AsExpression() Expression {
	return s
}

// Select is a SQL expression that is evaluated to fetch data.
type Select struct {
	Distinct   bool
	Projection Projection
	From       []FromClause
	Where      Expression
	GroupBy    []Expression
	Having     Expression
}

func (s Select) AsExpression() Expression {
	return s
}

func (s Select) AsSetExpression() SetExpression {
	return s
}

func (s Select) NodeType() string {
	return "select"
}

// SetOperation is a binary expression that represents an operation to be applied to two set expressions:
//
// select 1
// union
// select 2;
type SetOperation struct {
	Operator Operator
	LOperand SetExpression
	ROperand SetExpression
	All      bool
	Distinct bool
}

func (s SetOperation) AsExpression() Expression {
	return s
}

func (s SetOperation) AsSetExpression() SetExpression {
	return s
}

func (s SetOperation) NodeType() string {
	return "set_operation"
}

// ExistsExpression is an expression that is wrapped by an exists unary expression that can be optionally negated.
//
// [not] exists(<query>)
type ExistsExpression struct {
	Subquery Subquery
	Negated  bool
}

func (s ExistsExpression) AsSelectItem() SelectItem {
	return s
}

func (s ExistsExpression) NodeType() string {
	return "exists_expression"
}

func (s ExistsExpression) AsExpression() Expression {
	return s
}

// CommonTableExpression is a named component of a `WITH` query and provides a way to write auxiliary statements
// for use in a larger query
type CommonTableExpression struct {
	Alias        TableAlias
	Materialized models.Optional[Materialized]
	Query        Query
}

func (s CommonTableExpression) NodeType() string {
	return "common_table_expression"
}

// Materialized is an expression that represents a query hint for common table expressions that informs
// the query planner if it should prefer to materialize the given CTE.
type Materialized struct {
	Materialized bool
}

func (s Materialized) AsExpression() Expression {
	return s
}

func (s Materialized) AsSetExpression() SetExpression {
	return s
}

func (s Materialized) NodeType() string {
	return "materialized"
}

// With is a statement that provides a way to write auxiliary statements for use in a larger query.
type With struct {
	Recursive   bool
	Expressions []CommonTableExpression
}

func (s With) NodeType() string {
	return "with"
}

// [with <CTE>] select * from table;
type Query struct {
	CommonTableExpressions *With
	Body                   SetExpression
	OrderBy                []*OrderBy
	Offset                 models.Optional[Expression]
	Limit                  models.Optional[Expression]
}

func (s Query) AddCTE(cte CommonTableExpression) {
	s.CommonTableExpressions.Expressions = append(s.CommonTableExpressions.Expressions, cte)
}

func (s Query) AsExpression() Expression {
	return s
}

func (s Query) AsSetExpression() SetExpression {
	return s
}

func (s Query) AsStatement() Statement {
	return s
}

func (s Query) NodeType() string {
	return "query"
}

func OptionalBinaryExpressionJoin(optional Expression, operator Operator, conjoined Expression) Expression {
	if optional == nil {
		return conjoined
	}

	return NewBinaryExpression(
		conjoined,
		operator,
		optional,
	)
}

func OptionalAnd(leftOperand Expression, rightOperand Expression) Expression {
	if leftOperand == nil {
		return rightOperand
	} else if rightOperand == nil {
		return leftOperand
	}

	return NewBinaryExpression(leftOperand, OperatorAnd, rightOperand)
}
