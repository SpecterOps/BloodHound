package pgsql

import (
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cypher/model"
)

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

// TODO: Not super happy with this syntax node name but had trouble coming up with something more appropriate
type RowShape struct {
	Columns []Identifier
}

func (s RowShape) NodeType() string {
	return "row_shape"
}

type TableAlias struct {
	Name  Identifier
	Shape model.Optional[RowShape]
}

func (s TableAlias) NodeType() string {
	return "table_alias"
}

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

type Case struct {
	Operand    Expression
	Conditions []Expression
	Then       []Expression
	Else       Expression
}

// [not] exists(<query>)
type Exists struct {
	Query   Query
	Negated bool
}

// [not] in (val1, val2, ...)
type InExpression struct {
	Expression Expression
	List       []Expression
	Negated    bool
}

// [not] in (<Select> ...)
type InSubquery struct {
	Expression Expression
	Query      Query
	Negated    bool
}

// <expr> [not] between <low> and <high>
type Between struct {
	Expression Expression
	Low        Expression
	High       Expression
	Negated    bool
}

type Literal struct {
	Value    any
	Null     bool
	TypeHint DataType
}

func (s Literal) AsExpression() Expression {
	return s
}

func (s Literal) AsProjection() Projection {
	return s
}

func AsLiteral(value any) Literal {
	return Literal{
		Value: value,
	}
}

func (s Literal) NodeType() string {
	return "literal"
}

type Subquery struct {
	Query Query
}

// not <expr>
type UnaryExpression struct {
	Operator Expression
	Operand  Expression
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
	TypeHint DataType
}

// <expr> > <expr>
// table.column > 12345
type BinaryExpression struct {
	Rewritten     bool
	Operator      Expression
	LOperand      Expression
	LDependencies map[Identifier]struct{}
	ROperand      Expression
	RDependencies map[Identifier]struct{}
}

func NewBinaryExpression(left, operator, right Expression) *BinaryExpression {
	return &BinaryExpression{
		Operator:      operator,
		LOperand:      left,
		LDependencies: map[Identifier]struct{}{},
		ROperand:      right,
		RDependencies: map[Identifier]struct{}{},
	}
}

func (s BinaryExpression) CombinedDependencies() map[Identifier]struct{} {
	combined := make(map[Identifier]struct{}, len(s.LDependencies)+len(s.RDependencies))

	for key := range s.LDependencies {
		combined[key] = struct{}{}
	}

	for key := range s.RDependencies {
		combined[key] = struct{}{}
	}

	return combined
}

func (s BinaryExpression) AsExpression() Expression {
	return s
}

func (s BinaryExpression) AsProjection() Projection {
	return s
}

func (s BinaryExpression) NodeType() string {
	return "binary_expression"
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

func (s CompositeValue) AsProjection() Projection {
	return s
}

// (<expr>)
type Parenthetical struct {
	Expression Expression
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

type OrderBy struct {
	Expression Expression
	Ascending  bool
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
}

func NewAnyExpression(inner Expression) AnyExpression {
	return AnyExpression{
		Expression: inner,
	}
}

type FunctionCall struct {
	Distinct   bool
	Function   Identifier
	Parameters []Expression
	Over       *Window
}

func (s FunctionCall) AsExpression() Expression {
	return s
}

func (s FunctionCall) NodeType() string {
	return "function_call"
}

type Join struct {
	Table        TableReference
	JoinOperator JoinOperator
}

func (s *Join) NodeType() string {
	return "join"
}

type Identifier string

func (s Identifier) AsProjection() Projection {
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

func AsOptionalIdentifier(identifier Identifier) model.Optional[Identifier] {
	return model.ValueOptional(identifier)
}

type IdentifierSet map[Identifier]struct{}

func AsIdentifierSet(identifiers ...Identifier) IdentifierSet {
	newSet := make(IdentifierSet, len(identifiers))

	for _, identifier := range identifiers {
		newSet[identifier] = struct{}{}
	}

	return newSet
}

func (s IdentifierSet) Add(identifier Identifier) IdentifierSet {
	s[identifier] = struct{}{}
	return s
}

func (s IdentifierSet) Merge(other IdentifierSet) {
	for key := range other {
		s[key] = struct{}{}
	}
}

func (s IdentifierSet) Slice() []Identifier {
	identifiers := make([]Identifier, 0, len(s))

	for key := range s {
		identifiers = append(identifiers, key)
	}

	return identifiers
}

func (s IdentifierSet) Strings() []string {
	identifiers := make([]string, 0, len(s))

	for key := range s {
		identifiers = append(identifiers, key.String())
	}

	return identifiers
}

func (s IdentifierSet) CombinedKey() Identifier {
	// Pull the identifiers as a sorted slice
	identifierStrings := s.Strings()
	slices.Sort(identifierStrings)

	// Join the identifiers
	return Identifier(strings.Join(identifierStrings, ""))
}

// todo: `s` is said to satisfy `other` if every identifier in s is present in other.
func (s IdentifierSet) Satisfies(other IdentifierSet) bool {
	for identifier := range other {
		if _, satisfied := s[identifier]; !satisfied {
			return false
		}
	}

	return true
}

func (s IdentifierSet) Matches(other IdentifierSet) bool {
	return len(s) == len(other) && s.Satisfies(other)
}

type CompoundIdentifier []Identifier

func (s CompoundIdentifier) Replace(old, new Identifier) {
	for idx, identifier := range s {
		if identifier == old {
			s[idx] = new
		}
	}
}

func (s CompoundIdentifier) Root() Identifier {
	return s[0]
}

func (s CompoundIdentifier) Identifier() Identifier {
	return Identifier(strings.Join(s.Strings(), "."))
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

func (s CompoundIdentifier) Copy() CompoundIdentifier {
	copyInst := make(CompoundIdentifier, len(s))
	copy(copyInst, s)

	return copyInst
}

func (s CompoundIdentifier) AsExpression() Expression {
	return s
}

func (s CompoundIdentifier) AsProjection() Projection {
	return s
}

func (s CompoundIdentifier) NodeType() string {
	return "compound_identifier"
}

type TableReference struct {
	Name    CompoundIdentifier
	Binding model.Optional[Identifier]
}

func (s TableReference) AsExpression() Expression {
	return s
}

func (s TableReference) NodeType() string {
	return "table_reference"
}

type FromClause struct {
	Relation TableReference
	Joins    []Join
}

func (s FromClause) NodeType() string {
	return "from"
}

type AliasedExpression struct {
	Expression Expression
	Alias      Identifier
}

func (s AliasedExpression) NodeType() string {
	return "aliased_expression"
}

func (s AliasedExpression) AsExpression() Expression {
	return s
}

func (s AliasedExpression) AsProjection() Projection {
	return s
}

type Wildcard struct{}

func (s Wildcard) AsExpression() Expression {
	return s
}

func (s Wildcard) AsProjection() Projection {
	return s
}

func (s Wildcard) NodeType() string {
	return "wildcard"
}

type QualifiedWildcard struct {
	Qualifier string
}

type ArrayLiteral struct {
	Values   []Expression
	TypeHint DataType
}

func (s ArrayLiteral) AsExpression() Expression {
	return s
}

func (s ArrayLiteral) AsProjection() Projection {
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
	Columns    []Identifier
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

type Insert struct {
	Table      CompoundIdentifier
	Columns    []Identifier
	OnConflict *OnConflict
	Source     *Query
	Returning  []Projection
}

func (s Insert) AsStatement() Statement {
	return s
}

func (s Insert) NodeType() string {
	return "insert"
}

// <identifier> = <value>
type Assignment struct {
	Identifier Identifier
	Value      Expression
}

func (s Assignment) AsExpression() Expression {
	return s
}

func (s Assignment) NodeType() string {
	return "assignment"
}

type Delete struct {
	Table TableReference
	Where Expression
}

func (s Delete) AsStatement() Statement {
	return s
}

func (s Delete) NodeType() string {
	return "delete"
}

type Update struct {
	Table       TableReference
	Assignments []Assignment
	Where       Expression
}

func (s Update) AsStatement() Statement {
	return s
}

func (s Update) NodeType() string {
	return "update"
}

type Select struct {
	Distinct   bool
	Projection []Projection
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

// TODO: Should this embed/compose a BinaryExpression struct?
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

type CommonTableExpression struct {
	Alias        TableAlias
	Materialized *Materialized
	Query        Query
}

func (s CommonTableExpression) NodeType() string {
	return "common_table_expression"
}

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
