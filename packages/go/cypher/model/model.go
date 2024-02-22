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

package model

import (
	"sort"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type SortOrder string

func (s SortOrder) String() string {
	return string(s)
}

type Operator string

func (s Operator) String() string {
	return string(s)
}

type AssignmentOperator string

func (s AssignmentOperator) String() string {
	return string(s)
}

type Expression any

type ExpressionList interface {
	Add(expression Expression)
	AddSlice(expressions []Expression)
	Get(index int) Expression
	GetAll() []Expression
	Len() int
	IndexOf(expression Expression) int
	Remove(expression Expression) bool
	Replace(index int, expression Expression)
}

type expressionList struct {
	Expressions []Expression
}

func NewExpressionListFromSlice(slice []Expression) ExpressionList {
	return &expressionList{
		Expressions: slice,
	}
}

func NewExpressionList() ExpressionList {
	return &expressionList{}
}

func (s *expressionList) copy() expressionList {
	return expressionList{
		Expressions: Copy(s.Expressions),
	}
}

func (s *expressionList) IndexOf(expressionToFind Expression) int {
	for idx, expression := range s.Expressions {
		if expression == expressionToFind {
			return idx
		}
	}

	return -1
}

func (s *expressionList) Len() int {
	return len(s.Expressions)
}

func (s *expressionList) Remove(expressionToRemove Expression) bool {
	for idx, expression := range s.Expressions {
		if expression == expressionToRemove {
			s.Expressions = append(s.Expressions[:idx], s.Expressions[idx+1:]...)
			return true
		}
	}

	return false
}

func (s *expressionList) Add(expression Expression) {
	s.Expressions = append(s.Expressions, expression)
}

func (s *expressionList) AddSlice(expressions []Expression) {
	s.Expressions = append(s.Expressions, expressions...)
}

func (s *expressionList) Get(index int) Expression {
	return s.Expressions[index]
}

func (s *expressionList) GetAll() []Expression {
	return s.Expressions
}

func (s *expressionList) Replace(index int, expression Expression) {
	s.Expressions[index] = expression
}

type Fallible interface {
	AddError(err error)
	Errors() []error
}

func WithErrors[T Fallible](fallible T, errs ...error) T {
	for _, err := range errs {
		fallible.AddError(err)
	}

	return fallible
}

type errorContext struct {
	errors []error
}

func (s *errorContext) AddError(err error) {
	s.errors = append(s.errors, err)
}

func (s *errorContext) Errors() []error {
	return s.errors
}

///

type RegularQuery struct {
	SingleQuery *SingleQuery
}

func NewRegularQuery() *RegularQuery {
	return &RegularQuery{}
}

func (s *RegularQuery) copy() *RegularQuery {
	if s == nil {
		return nil
	}

	return &RegularQuery{
		SingleQuery: Copy(s.SingleQuery),
	}
}

type SingleQuery struct {
	SinglePartQuery *SinglePartQuery
	MultiPartQuery  *MultiPartQuery
}

func NewSingleQuery() *SingleQuery {
	return &SingleQuery{}
}

func (s *SingleQuery) copy() *SingleQuery {
	if s == nil {
		return nil
	}

	return &SingleQuery{
		SinglePartQuery: Copy(s.SinglePartQuery),
		MultiPartQuery:  Copy(s.MultiPartQuery),
	}
}

type Unwind struct {
	Expression Expression
	Binding    *Variable
}

func NewUnwind() *Unwind {
	return &Unwind{}
}

func (s *Unwind) copy() *Unwind {
	if s == nil {
		return nil
	}

	return &Unwind{
		Expression: Copy(s.Expression),
		Binding:    Copy(s.Binding),
	}
}

type ReadingClause struct {
	Match  *Match
	Unwind *Unwind
}

func NewReadingClause() *ReadingClause {
	return &ReadingClause{}
}

func (s *ReadingClause) copy() *ReadingClause {
	if s == nil {
		return nil
	}

	return &ReadingClause{
		Match:  Copy(s.Match),
		Unwind: Copy(s.Unwind),
	}
}

type MultiPartQueryPart struct {
	ReadingClauses  []*ReadingClause
	UpdatingClauses []*UpdatingClause
	With            *With
}

func NewMultiPartQueryPart() *MultiPartQueryPart {
	return &MultiPartQueryPart{}
}

func (s *MultiPartQueryPart) copy() *MultiPartQueryPart {
	if s == nil {
		return nil
	}

	return &MultiPartQueryPart{
		ReadingClauses:  Copy(s.ReadingClauses),
		UpdatingClauses: Copy(s.UpdatingClauses),
		With:            Copy(s.With),
	}
}

func (s *MultiPartQueryPart) AddReadingClause(clause *ReadingClause) {
	s.ReadingClauses = append(s.ReadingClauses, clause)
}

func (s *MultiPartQueryPart) AddUpdatingClause(clause *UpdatingClause) {
	s.UpdatingClauses = append(s.UpdatingClauses, clause)
}

type MultiPartQuery struct {
	Parts           []*MultiPartQueryPart
	SinglePartQuery *SinglePartQuery
}

func NewMultiPartQuery() *MultiPartQuery {
	return &MultiPartQuery{}
}

func (s *MultiPartQuery) copy() *MultiPartQuery {
	if s == nil {
		return nil
	}

	return &MultiPartQuery{
		Parts:           Copy(s.Parts),
		SinglePartQuery: Copy(s.SinglePartQuery),
	}
}

type With struct {
	Projection *Projection
	Where      *Where
}

func NewWith() *With {
	return &With{}
}

func (s *With) copy() *With {
	if s == nil {
		return nil
	}

	return &With{
		Projection: Copy(s.Projection),
		Where:      Copy(s.Where),
	}
}

type SinglePartQuery struct {
	errorContext

	ReadingClauses  []*ReadingClause
	UpdatingClauses []Expression
	Return          *Return
}

func NewSinglePartQuery() *SinglePartQuery {
	return &SinglePartQuery{}
}

func (s *SinglePartQuery) copy() *SinglePartQuery {
	if s == nil {
		return nil
	}

	return &SinglePartQuery{
		errorContext: errorContext{
			errors: s.errors,
		},

		ReadingClauses:  Copy(s.ReadingClauses),
		UpdatingClauses: Copy(s.UpdatingClauses),
		Return:          Copy(s.Return),
	}
}

func (s *SinglePartQuery) AddReadingClause(clause *ReadingClause) {
	s.ReadingClauses = append(s.ReadingClauses, clause)
}

func (s *SinglePartQuery) AddUpdatingClause(clause *UpdatingClause) {
	s.UpdatingClauses = append(s.UpdatingClauses, clause)
}

type PartialArithmeticExpression struct {
	Operator Operator
	Right    Expression
}

func NewArithmeticExpressionPart(operator Operator) *PartialArithmeticExpression {
	return &PartialArithmeticExpression{
		Operator: operator,
	}
}

func (s *PartialArithmeticExpression) copy() *PartialArithmeticExpression {
	return &PartialArithmeticExpression{
		Operator: s.Operator,
		Right:    Copy(s.Right),
	}
}

type ArithmeticExpression struct {
	Left     Expression
	Partials []*PartialArithmeticExpression
}

func NewArithmeticExpression() *ArithmeticExpression {
	return &ArithmeticExpression{}
}

func (s *ArithmeticExpression) copy() *ArithmeticExpression {
	return &ArithmeticExpression{
		Left:     Copy(s.Left),
		Partials: Copy(s.Partials),
	}
}

func (s *ArithmeticExpression) AddArithmeticExpressionPart(part *PartialArithmeticExpression) {
	s.Partials = append(s.Partials, part)
}

type Match struct {
	Optional bool
	Pattern  []*PatternPart
	Where    *Where
}

func NewMatch(optional bool) *Match {
	return &Match{
		Optional: optional,
	}
}

func (s *Match) copy() *Match {
	if s == nil {
		return nil
	}

	return &Match{
		Optional: s.Optional,
		Pattern:  Copy(s.Pattern),
		Where:    Copy(s.Where),
	}
}

func (s *Match) FirstRelationshipPattern() *RelationshipPattern {
	if s != nil && len(s.Pattern) > 0 {
		for _, patternElement := range s.Pattern[0].PatternElements {
			if relationshipPattern, isRelationshipPattern := patternElement.AsRelationshipPattern(); isRelationshipPattern {
				return relationshipPattern
			}
		}
	}

	return nil
}

type UpdatingClause struct {
	errorContext

	Clause Expression
}

func NewUpdatingClause[T *Delete | *Remove | *Set | *Create](clause T) *UpdatingClause {
	return &UpdatingClause{
		Clause: clause,
	}
}

func (s *UpdatingClause) copy() *UpdatingClause {
	if s == nil {
		return nil
	}

	return &UpdatingClause{
		errorContext: errorContext{
			errors: s.errors,
		},

		Clause: Copy(s.Clause),
	}
}

type Delete struct {
	Detach      bool
	Expressions []Expression
}

func NewDelete() *Delete {
	return &Delete{
		Detach: false,
	}
}

func (s *Delete) copy() *Delete {
	if s == nil {
		return nil
	}

	return &Delete{
		Detach:      s.Detach,
		Expressions: Copy(s.Expressions),
	}
}

type Remove struct {
	Items []*RemoveItem
}

func NewRemove() *Remove {
	return &Remove{}
}

func (s *Remove) copy() *Remove {
	if s == nil {
		return nil
	}

	return &Remove{
		Items: Copy(s.Items),
	}
}

type RemoveItem struct {
	KindMatcher Expression
	Property    *PropertyLookup
}

func NewRemoveItem() *RemoveItem {
	return &RemoveItem{}
}

func (s *RemoveItem) copy() *RemoveItem {
	if s == nil {
		return nil
	}

	return &RemoveItem{
		KindMatcher: Copy(s.KindMatcher),
		Property:    Copy(s.Property),
	}
}

type Set struct {
	Items []*SetItem
}

func NewSet() *Set {
	return &Set{}
}

func (s *Set) copy() *Set {
	if s == nil {
		return nil
	}

	return &Set{
		Items: Copy(s.Items),
	}
}

type SetItem struct {
	Left     Expression
	Operator AssignmentOperator
	Right    Expression
}

func NewSetItem() *SetItem {
	return &SetItem{}
}

func (s *SetItem) copy() *SetItem {
	if s == nil {
		return s
	}

	return &SetItem{
		Left:     Copy(s.Left),
		Operator: s.Operator,
		Right:    Copy(s.Right),
	}
}

type Create struct {
	errorContext

	Unique  bool
	Pattern []*PatternPart
}

func NewCreate() *Create {
	return &Create{}
}

func (s *Create) copy() *Create {
	if s == nil {
		return nil
	}

	return &Create{
		errorContext: errorContext{
			errors: s.errors,
		},

		Unique:  s.Unique,
		Pattern: Copy(s.Pattern),
	}
}

type IDInCollection struct {
	Variable   *Variable
	Expression Expression
}

func NewIDInCollection() *IDInCollection {
	return &IDInCollection{}
}

func (s *IDInCollection) copy() *IDInCollection {
	if s == nil {
		return s
	}

	return &IDInCollection{
		Variable:   Copy(s.Variable),
		Expression: Copy(s.Expression),
	}
}

type FilterExpression struct {
	Specifier *IDInCollection
	Where     *Where
}

func NewFilterExpression() *FilterExpression {
	return &FilterExpression{}
}

func (s *FilterExpression) copy() *FilterExpression {
	if s == nil {
		return s
	}

	return &FilterExpression{
		Specifier: Copy(s.Specifier),
		Where:     Copy(s.Where),
	}
}

type QuantifierType string

func (s QuantifierType) String() string {
	return string(s)
}

type Quantifier struct {
	Type   QuantifierType
	Filter *FilterExpression
}

func NewQuantifier(quantifierType QuantifierType) *Quantifier {
	return &Quantifier{
		Type: quantifierType,
	}
}

func (s *Quantifier) copy() *Quantifier {
	if s == nil {
		return s
	}

	return &Quantifier{
		Type:   s.Type,
		Filter: Copy(s.Filter),
	}
}

type RangeQuantifier struct {
	Value string
}

func NewRangeQuantifier(value string) *RangeQuantifier {
	return &RangeQuantifier{
		Value: value,
	}
}

type KindMatcher struct {
	Reference Expression
	Kinds     graph.Kinds
}

func NewKindMatcher(reference Expression, kinds graph.Kinds) *KindMatcher {
	return &KindMatcher{
		Reference: reference,
		Kinds:     kinds,
	}
}

func (s *KindMatcher) copy() *KindMatcher {
	if s == nil {
		return nil
	}

	return &KindMatcher{
		Reference: Copy(s.Reference),
		Kinds:     Copy(s.Kinds),
	}
}

type Literal struct {
	Value any
	Null  bool
}

func NewLiteral(value any, null bool) *Literal {
	return &Literal{
		Value: value,
		Null:  null,
	}
}

func NewStringLiteral(value string) *Literal {
	return NewLiteral("'"+value+"'", false)
}

func (s *Literal) copy() *Literal {
	return &Literal{
		Value: s.Value,
		Null:  s.Null,
	}
}

func (s *Literal) Set(value any) *Literal {
	s.Value = value
	s.Null = false

	return s
}

func (s *Literal) SetNull() *Literal {
	s.Null = true

	return s
}

type Parameter struct {
	Symbol string
	Value  any
}

func NewParameter(symbol string, value any) *Parameter {
	return &Parameter{
		Symbol: symbol,
		Value:  value,
	}
}

func (s *Parameter) copy() *Parameter {
	if s == nil {
		return nil
	}

	return &Parameter{
		Symbol: s.Symbol,
		Value:  s.Value,
	}
}

type MapLiteral map[string]Expression

func NewMapLiteral() MapLiteral {
	return MapLiteral{}
}

func (s MapLiteral) Keys() []any {
	keys := make([]any, 0, len(s))

	for key := range s {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i].(string), keys[j].(string)) > 0
	})

	return keys
}

type ListLiteral []Expression

func NewListLiteral() *ListLiteral {
	return &ListLiteral{}
}

func (s *ListLiteral) Keys() []any {
	keys := make([]any, len(*s))

	for idx := len(*s) - 1; idx >= 0; idx-- {
		keys[idx] = idx
	}

	return keys
}

type PatternRange struct {
	StartIndex *int64
	EndIndex   *int64
}

func NewPatternRange(start, end *int64) *PatternRange {
	return &PatternRange{
		StartIndex: start,
		EndIndex:   end,
	}
}

func (s *PatternRange) copy() *PatternRange {
	if s == nil {
		return nil
	}

	return &PatternRange{
		StartIndex: Copy(s.StartIndex),
		EndIndex:   Copy(s.EndIndex),
	}
}

type Negation struct {
	Expression Expression
}

func NewNegation(expression Expression) *Negation {
	return &Negation{
		Expression: expression,
	}
}

func (s *Negation) copy() *Negation {
	if s == nil {
		return nil
	}

	return &Negation{
		Expression: Copy(s.Expression),
	}
}

type Parenthetical struct {
	Expression Expression
}

func NewParenthetical(expression Expression) *Parenthetical {
	return &Parenthetical{
		Expression: expression,
	}
}

func (s *Parenthetical) copy() *Parenthetical {
	return &Parenthetical{
		Expression: Copy(s.Expression),
	}
}

type ExclusiveDisjunction struct {
	expressionList
}

func NewExclusiveDisjunction(expressions ...Expression) *ExclusiveDisjunction {
	return &ExclusiveDisjunction{
		expressionList{
			Expressions: expressions,
		},
	}
}

func (s *ExclusiveDisjunction) copy() *ExclusiveDisjunction {
	if s == nil {
		return nil
	}

	return &ExclusiveDisjunction{
		expressionList: Copy(s.expressionList),
	}
}

type Disjunction struct {
	expressionList
}

func NewDisjunction(expressions ...Expression) *Disjunction {
	return &Disjunction{
		expressionList: expressionList{
			Expressions: expressions,
		},
	}
}

func (s *Disjunction) copy() *Disjunction {
	if s == nil {
		return nil
	}

	return &Disjunction{
		expressionList: Copy(s.expressionList),
	}
}

type Conjunction struct {
	expressionList
}

func NewConjunction(expressions ...Expression) *Conjunction {
	return &Conjunction{
		expressionList{
			Expressions: expressions,
		},
	}
}

func (s *Conjunction) copy() *Conjunction {
	if s == nil {
		return nil
	}

	return &Conjunction{
		expressionList: Copy(s.expressionList),
	}
}

type FunctionInvocation struct {
	errorContext

	Distinct  bool
	Namespace []string
	Name      string
	Arguments []Expression
}

func NewSimpleFunctionInvocation(name string, arguments ...Expression) *FunctionInvocation {
	return &FunctionInvocation{
		Name:      name,
		Arguments: arguments,
	}
}

func (s *FunctionInvocation) copy() *FunctionInvocation {
	if s == nil {
		return nil
	}

	return &FunctionInvocation{
		errorContext: errorContext{
			errors: s.errors,
		},

		Distinct:  s.Distinct,
		Namespace: Copy(s.Namespace),
		Name:      s.Name,
		Arguments: Copy(s.Arguments),
	}
}

func (s *FunctionInvocation) AddArgument(arg Expression) {
	s.Arguments = append(s.Arguments, arg)
}

type Comparison struct {
	Left     Expression
	Partials []*PartialComparison
}

func NewComparison(left Expression, operator Operator, right Expression) *Comparison {
	return &Comparison{
		Left: left,
		Partials: []*PartialComparison{{
			Operator: operator,
			Right:    right,
		}},
	}
}

func (s *Comparison) copy() *Comparison {
	if s == nil {
		return nil
	}

	return &Comparison{
		Left:     Copy(s.Left),
		Partials: Copy(s.Partials),
	}
}

func (s *Comparison) AddPartialComparison(partial *PartialComparison) {
	s.Partials = append(s.Partials, partial)
}

func (s *Comparison) FirstPartial() *PartialComparison {
	if len(s.Partials) > 0 {
		return s.Partials[0]
	}

	return nil
}

func (s *Comparison) LastPartial() *PartialComparison {
	if len(s.Partials) > 0 {
		return s.Partials[len(s.Partials)-1]
	}

	return nil
}

type PartialComparison struct {
	Operator Operator
	Right    Expression
}

func (s *PartialComparison) copy() *PartialComparison {
	return &PartialComparison{
		Operator: s.Operator,
		Right:    Copy(s.Right),
	}
}

type Variable struct {
	Symbol string
}

func NewVariable() *Variable {
	return &Variable{}
}

func NewVariableWithSymbol(symbol string) *Variable {
	return &Variable{
		Symbol: symbol,
	}
}

func (s *Variable) copy() *Variable {
	if s == nil {
		return nil
	}

	return &Variable{
		Symbol: s.Symbol,
	}
}

type ProjectionItem struct {
	Expression Expression
	Binding    Expression
}

func NewProjectionItem() *ProjectionItem {
	return &ProjectionItem{}
}

func NewGreedyProjectionItem() *ProjectionItem {
	return &ProjectionItem{
		Expression: NewVariableWithSymbol(TokenLiteralAsterisk),
	}
}

func (s *ProjectionItem) copy() *ProjectionItem {
	if s == nil {
		return nil
	}

	return &ProjectionItem{
		Expression: Copy(s.Expression),
		Binding:    Copy(s.Binding),
	}
}

type PropertyLookup struct {
	Atom    Expression
	Symbols []string
}

func (s *PropertyLookup) copy() *PropertyLookup {
	if s == nil {
		return nil
	}

	return &PropertyLookup{
		Atom:    Copy(s.Atom),
		Symbols: Copy(s.Symbols),
	}
}

func NewPropertyLookup(variableSymbol string, propertySymbols ...string) *PropertyLookup {
	return &PropertyLookup{
		Atom: &Variable{
			Symbol: variableSymbol,
		},
		Symbols: propertySymbols,
	}
}

func (s *PropertyLookup) AddLookupSymbol(symbol string) {
	s.Symbols = append(s.Symbols, symbol)
}

type PatternElement struct {
	Element Expression
}

func (s *PatternElement) copy() *PatternElement {
	return &PatternElement{
		Element: Copy(s.Element),
	}
}

func (s *PatternElement) IsNodePattern() bool {
	_, isNodePattern := s.AsNodePattern()
	return isNodePattern
}

func (s *PatternElement) AsNodePattern() (*NodePattern, bool) {
	nodePattern, isNodePattern := s.Element.(*NodePattern)
	return nodePattern, isNodePattern
}

func (s *PatternElement) IsRelationshipPattern() bool {
	_, isRelationshipPattern := s.AsRelationshipPattern()
	return isRelationshipPattern
}

func (s *PatternElement) AsRelationshipPattern() (*RelationshipPattern, bool) {
	relationshipPattern, isRelationshipPattern := s.Element.(*RelationshipPattern)
	return relationshipPattern, isRelationshipPattern
}

type Properties struct {
	Map       MapLiteral
	Parameter *Parameter
}

func NewProperties() *Properties {
	return &Properties{}
}

// NodePattern Type
//
// Kinds is a conjunction of types for the given node.
// e.g. (n:K1:K2) may be rendered as (n) where n:K1 and n:K2

// NodePattern
type NodePattern struct {
	Binding    Expression
	Kinds      graph.Kinds
	Properties Expression
}

func (s *NodePattern) copy() *NodePattern {
	if s == nil {
		return nil
	}

	return &NodePattern{
		Binding:    s.Binding,
		Kinds:      Copy(s.Kinds),
		Properties: Copy(s.Properties),
	}
}

func (s *NodePattern) AddKind(kind graph.Kind) {
	s.Kinds = append(s.Kinds, kind)
}

// RelationshipPattern Type
//
// Kinds is a disjunction of types for the given edge.
//	e.g. [r:K1|K2] may be rendered as `()-[r]-() where type(r) = K1 or type(r) = K2`

// RelationshipPattern
type RelationshipPattern struct {
	Binding    Expression
	Kinds      graph.Kinds
	Direction  graph.Direction
	Range      *PatternRange
	Properties Expression
}

func (s *RelationshipPattern) copy() *RelationshipPattern {
	if s == nil {
		return nil
	}

	return &RelationshipPattern{
		Binding:    s.Binding,
		Kinds:      Copy(s.Kinds),
		Direction:  s.Direction,
		Range:      Copy(s.Range),
		Properties: Copy(s.Properties),
	}
}

func (s *RelationshipPattern) AddKind(kind graph.Kind) {
	s.Kinds = append(s.Kinds, kind)
}

type Where struct {
	expressionList
}

func NewWhere() *Where {
	return &Where{}
}

func (s *Where) copy() *Where {
	if s == nil {
		return nil
	}

	return &Where{
		expressionList: Copy(s.expressionList),
	}
}

type Not struct {
	Expression Expression
}

type SortItem struct {
	Ascending  bool
	Expression Expression
}

func (s *SortItem) copy() *SortItem {
	return &SortItem{
		Ascending:  s.Ascending,
		Expression: Copy(s.Expression),
	}
}

type Order struct {
	Items []*SortItem
}

func (s *Order) copy() *Order {
	if s == nil {
		return nil
	}

	return &Order{
		Items: Copy(s.Items),
	}
}

func (s *Order) AddItem(item *SortItem) {
	s.Items = append(s.Items, item)
}

type Projection struct {
	Distinct bool
	All      bool
	Order    *Order
	Skip     *Skip
	Limit    *Limit
	Items    []Expression
}

func NewProjection(distinct bool) *Projection {
	return &Projection{
		Distinct: distinct,
	}
}

func (s *Projection) copy() *Projection {
	if s == nil {
		return nil
	}

	return &Projection{
		Distinct: s.Distinct,
		All:      s.All,
		Order:    Copy(s.Order),
		Skip:     Copy(s.Skip),
		Limit:    Copy(s.Limit),
		Items:    Copy(s.Items),
	}
}

func (s *Projection) AddItem(item *ProjectionItem) {
	s.Items = append(s.Items, item)
}

type Return struct {
	Projection *Projection
}

func NewReturn() *Return {
	return &Return{}
}

func (s *Return) copy() *Return {
	if s == nil {
		return nil
	}

	return &Return{
		Projection: s.Projection.copy(),
	}
}

type PatternPart struct {
	Binding                 Expression
	ShortestPathPattern     bool
	AllShortestPathsPattern bool
	PatternElements         []*PatternElement
}

func NewPatternPart() *PatternPart {
	return &PatternPart{}
}

func (s *PatternPart) copy() *PatternPart {
	if s == nil {
		return nil
	}

	return &PatternPart{
		Binding:                 s.Binding,
		ShortestPathPattern:     s.ShortestPathPattern,
		AllShortestPathsPattern: s.AllShortestPathsPattern,
		PatternElements:         Copy(s.PatternElements),
	}
}

func (s *PatternPart) CurrentElement() any {
	if numElements := len(s.PatternElements); numElements == 0 {
		return nil
	} else {
		return s.PatternElements[numElements-1]
	}
}

func (s *PatternPart) AddPatternElements(nextElements ...Expression) *PatternPart {
	for _, nextElement := range nextElements {
		s.PatternElements = append(s.PatternElements, &PatternElement{
			Element: nextElement,
		})
	}

	return s
}

type Limit struct {
	Value Expression
}

func (s *Limit) copy() *Limit {
	if s == nil {
		return nil
	}

	return &Limit{
		Value: Copy(s.Value),
	}
}

type Skip struct {
	Value Expression
}

func NewSkip(value Expression) *Skip {
	return &Skip{
		Value: value,
	}
}

func (s *Skip) copy() *Skip {
	if s == nil {
		return nil
	}

	return &Skip{
		Value: Copy(s.Value),
	}
}

type PatternPredicate struct {
	PatternElements []*PatternElement
}

func NewPatternPredicate() *PatternPredicate {
	return &PatternPredicate{}
}

func (s *PatternPredicate) AddElement(element Expression) {
	s.PatternElements = append(s.PatternElements, &PatternElement{
		Element: element,
	})
}

func (s *PatternPredicate) copy() *PatternPredicate {
	if s == nil {
		return nil
	}

	return &PatternPredicate{
		PatternElements: Copy(s.PatternElements),
	}
}
