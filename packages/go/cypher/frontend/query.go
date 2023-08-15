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

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/parser"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type VariableVisitor struct {
	BaseVisitor

	Variable *model.Variable
}

func NewVariableVisitor() *VariableVisitor {
	return &VariableVisitor{
		Variable: model.NewVariable(),
	}
}

func (s *VariableVisitor) EnterOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
	s.Variable.Symbol = ctx.GetText()
}

func (s *VariableVisitor) ExitOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
}

type ProjectionVisitor struct {
	BaseVisitor

	currentItem *model.ProjectionItem
	Projection  *model.Projection
}

func NewProjectionVisitor(ctx *parser.OC_ProjectionBodyContext) *ProjectionVisitor {
	var distinct = false

	if HasTokens(ctx, parser.CypherLexerDISTINCT) {
		distinct = true
	}

	return &ProjectionVisitor{
		Projection: model.NewProjection(distinct),
	}
}

func (s *ProjectionVisitor) EnterOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {
}

func (s *ProjectionVisitor) ExitOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {
	if HasTokens(ctx, parser.CypherLexerDISTINCT) {
		s.Projection.Distinct = true
	}
}

func (s *ProjectionVisitor) EnterOC_ProjectionItems(ctx *parser.OC_ProjectionItemsContext) {
	if operators := newTokenLiteralIterator(ctx); operators.HasTokens() {
		// Only look at the first token for a greedy projection
		switch token := operators.NextToken(); token {
		case TokenLiteralAsterisk:
			s.Projection.AddItem(model.NewGreedyProjectionItem())

		case TokenLiteralComma:
		default:
			s.ctx.AddErrors(fmt.Errorf("invalid token: %s", token))
		}
	}
}

func (s *ProjectionVisitor) ExitOC_ProjectionItems(ctx *parser.OC_ProjectionItemsContext) {
}

func (s *ProjectionVisitor) EnterOC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) {
	s.currentItem = model.NewProjectionItem()
}

func (s *ProjectionVisitor) ExitOC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) {
	s.Projection.AddItem(s.currentItem)
}

func (s *ProjectionVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.ctx.Enter(NewVariableVisitor())
}

func (s *ProjectionVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.Binding = s.ctx.Exit().(*VariableVisitor).Variable
}

func (s *ProjectionVisitor) EnterOC_Order(ctx *parser.OC_OrderContext) {
	s.Projection.Order = &model.Order{}
}

func (s *ProjectionVisitor) ExitOC_Order(ctx *parser.OC_OrderContext) {
}

func (s ProjectionVisitor) EnterOC_SortItem(ctx *parser.OC_SortItemContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s ProjectionVisitor) ExitOC_SortItem(ctx *parser.OC_SortItemContext) {
	var (
		expression = s.ctx.Exit().(*ExpressionVisitor).Expression
		sortItem   = &model.SortItem{
			Ascending:  true,
			Expression: expression,
		}
	)

	if ctx.GetToken(parser.CypherLexerDESC, 0) != nil || ctx.GetToken(parser.CypherLexerDESCENDING, 0) != nil {
		sortItem.Ascending = false
	}

	s.Projection.Order.AddItem(sortItem)
}

func (s *ProjectionVisitor) EnterOC_Skip(ctx *parser.OC_SkipContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *ProjectionVisitor) ExitOC_Skip(ctx *parser.OC_SkipContext) {
	expression := s.ctx.Exit().(*ExpressionVisitor).Expression
	s.Projection.Skip = &model.Skip{
		Value: expression,
	}
}

func (s *ProjectionVisitor) EnterOC_Limit(ctx *parser.OC_LimitContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *ProjectionVisitor) ExitOC_Limit(ctx *parser.OC_LimitContext) {
	expression := s.ctx.Exit().(*ExpressionVisitor).Expression
	s.Projection.Limit = &model.Limit{
		Value: expression,
	}
}

func (s *ProjectionVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *ProjectionVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.currentItem.Expression = s.ctx.Exit().(*ExpressionVisitor).Expression
}

// Range literal expansion semantics
//
// * - Infinite expansion
//		()-[]->()...
// *.. - Infinite expansion
//		()-[]->()...
// *1.. - Infinite expansion returning one degree out
//		()-[]->()...
// *2.. - Infinite expansion returning two degrees out
//		()-[]->()-[]->()...
// *..2 - Expansion returning up to two degrees out
//		()-[]->()
//		()-[]->()-[]->()
// *2..2 - Expansion returning only two degrees out
//		()-[]->()-[]->()
//

// RangeLiteralVisitor
type RangeLiteralVisitor struct {
	BaseVisitor

	PatternRange *model.PatternRange
}

func (s *RangeLiteralVisitor) EnterOC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) {
	if value, err := strconv.ParseInt(ctx.GetText(), 10, 64); err != nil {
		s.ctx.AddErrors(fmt.Errorf("failed parsing range literal: %w", err))
	} else if s.PatternRange.StartIndex == nil {
		s.PatternRange.StartIndex = &value
	} else {
		s.PatternRange.EndIndex = &value
	}
}

func (s *RangeLiteralVisitor) ExitOC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) {
}

type WithVisitor struct {
	BaseVisitor

	With *model.With
}

func NewWithVisitor() *WithVisitor {
	return &WithVisitor{
		With: model.NewWith(),
	}
}

func (s *WithVisitor) EnterOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {
	s.ctx.Enter(NewProjectionVisitor(ctx))
}

func (s *WithVisitor) ExitOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {
	s.With.Projection = s.ctx.Exit().(*ProjectionVisitor).Projection
}

func (s *WithVisitor) EnterOC_Where(ctx *parser.OC_WhereContext) {
	s.ctx.Enter(NewWhereVisitor())
}

func (s *WithVisitor) ExitOC_Where(ctx *parser.OC_WhereContext) {
	s.With.Where = s.ctx.Exit().(*WhereVisitor).Where
}

type MatchVisitor struct {
	BaseVisitor

	Match *model.Match
}

func NewMatchVisitor(ctx *parser.OC_MatchContext) *MatchVisitor {
	optional := false

	if HasTokens(ctx, parser.CypherLexerOPTIONAL) {
		optional = true
	}

	return &MatchVisitor{
		Match: model.NewMatch(optional),
	}
}

func (s *MatchVisitor) EnterOC_Where(ctx *parser.OC_WhereContext) {
	s.ctx.Enter(NewWhereVisitor())
}

func (s *MatchVisitor) ExitOC_Where(ctx *parser.OC_WhereContext) {
	s.Match.Where = s.ctx.Exit().(*WhereVisitor).Where
}

func (s *MatchVisitor) EnterOC_Pattern(ctx *parser.OC_PatternContext) {
	s.ctx.Enter(&PatternVisitor{})
}

func (s *MatchVisitor) ExitOC_Pattern(ctx *parser.OC_PatternContext) {
	s.Match.Pattern = s.ctx.Exit().(*PatternVisitor).PatternParts
}

type UnwindVisitor struct {
	BaseVisitor

	Unwind *model.Unwind
}

func NewUnwindVisitor() *UnwindVisitor {
	return &UnwindVisitor{
		Unwind: model.NewUnwind(),
	}
}

func (s *UnwindVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *UnwindVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.Unwind.Expression = s.ctx.Exit().(*ExpressionVisitor).Expression
}

func (s *UnwindVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.ctx.Enter(NewVariableVisitor())
}

func (s *UnwindVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.Unwind.Binding = s.ctx.Exit().(*VariableVisitor).Variable
}

type ReadingClauseVisitor struct {
	BaseVisitor

	ReadingClause *model.ReadingClause
}

func NewReadingClauseVisitor() *ReadingClauseVisitor {
	return &ReadingClauseVisitor{
		ReadingClause: model.NewReadingClause(),
	}
}

func (s *ReadingClauseVisitor) EnterOC_Match(ctx *parser.OC_MatchContext) {
	s.ctx.Enter(NewMatchVisitor(ctx))
}

func (s *ReadingClauseVisitor) ExitOC_Match(ctx *parser.OC_MatchContext) {
	s.ReadingClause.Match = s.ctx.Exit().(*MatchVisitor).Match
}

func (s *ReadingClauseVisitor) EnterOC_Unwind(ctx *parser.OC_UnwindContext) {
	s.ctx.Enter(NewUnwindVisitor())
}

func (s *ReadingClauseVisitor) ExitOC_Unwind(ctx *parser.OC_UnwindContext) {
	s.ReadingClause.Unwind = s.ctx.Exit().(*UnwindVisitor).Unwind
}

type SinglePartQueryVisitor struct {
	BaseVisitor

	Query *model.SinglePartQuery
}

func NewSinglePartQueryVisitor() *SinglePartQueryVisitor {
	return &SinglePartQueryVisitor{
		Query: model.NewSinglePartQuery(),
	}
}

func (s *SinglePartQueryVisitor) EnterOC_Return(ctx *parser.OC_ReturnContext) {
	s.Query.Return = &model.Return{}
}

func (s *SinglePartQueryVisitor) ExitOC_Return(ctx *parser.OC_ReturnContext) {
}

func (s *SinglePartQueryVisitor) EnterOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {
	s.ctx.Enter(NewProjectionVisitor(ctx))
}

func (s *SinglePartQueryVisitor) ExitOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {
	s.Query.Return.Projection = s.ctx.Exit().(*ProjectionVisitor).Projection
}

func (s *SinglePartQueryVisitor) EnterOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {
	s.ctx.Enter(NewReadingClauseVisitor())
}

func (s *SinglePartQueryVisitor) ExitOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {
	s.Query.AddReadingClause(s.ctx.Exit().(*ReadingClauseVisitor).ReadingClause)
}

func (s *SinglePartQueryVisitor) EnterOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	s.ctx.Enter(NewUpdatingClauseVisitor())
}

func (s *SinglePartQueryVisitor) ExitOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	s.Query.AddUpdatingClause(s.ctx.Exit().(*UpdatingClauseVisitor).UpdatingClause)
}

type MultiPartQueryVisitor struct {
	BaseVisitor

	currentPart *model.MultiPartQueryPart
	Query       *model.MultiPartQuery
}

func NewMultiPartQueryVisitor() *MultiPartQueryVisitor {
	return &MultiPartQueryVisitor{
		Query: model.NewMultiPartQuery(),
	}
}

func (s *MultiPartQueryVisitor) EnterOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {
	s.currentPart = model.NewMultiPartQueryPart()
	s.Query.Parts = append(s.Query.Parts, s.currentPart)

	s.ctx.Enter(NewReadingClauseVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {
	if s.currentPart != nil {
		s.currentPart.AddReadingClause(s.ctx.Exit().(*ReadingClauseVisitor).ReadingClause)
	} else {
		s.ctx.AddErrors(ErrInvalidInput)
	}
}

func (s *MultiPartQueryVisitor) EnterOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	s.ctx.Enter(NewUpdatingClauseVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	if s.currentPart != nil {
		s.currentPart.AddUpdatingClause(s.ctx.Exit().(*UpdatingClauseVisitor).UpdatingClause)
	} else {
		s.ctx.AddErrors(ErrInvalidInput)
	}
}

func (s *MultiPartQueryVisitor) EnterOC_With(ctx *parser.OC_WithContext) {
	s.ctx.Enter(NewWithVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_With(ctx *parser.OC_WithContext) {
	if s.currentPart != nil {
		s.currentPart.With = s.ctx.Exit().(*WithVisitor).With
	} else {
		s.ctx.AddErrors(ErrInvalidInput)
	}
}

func (s *MultiPartQueryVisitor) EnterOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {
	s.ctx.Enter(NewSinglePartQueryVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {
	s.Query.SinglePartQuery = s.ctx.Exit().(*SinglePartQueryVisitor).Query
}

type QueryVisitor struct {
	BaseVisitor

	Query *model.RegularQuery
}

func (s *QueryVisitor) EnterOC_Cypher(ctx *parser.OC_CypherContext) {
}

func (s *QueryVisitor) ExitOC_Cypher(ctx *parser.OC_CypherContext) {
}

func (s *QueryVisitor) EnterOC_Query(ctx *parser.OC_QueryContext) {
}

func (s *QueryVisitor) ExitOC_Query(ctx *parser.OC_QueryContext) {
}

func (s *QueryVisitor) EnterOC_Statement(ctx *parser.OC_StatementContext) {
}

func (s *QueryVisitor) ExitOC_Statement(ctx *parser.OC_StatementContext) {
}

func (s *QueryVisitor) EnterOC_QueryOptions(ctx *parser.OC_QueryOptionsContext) {
}

func (s *QueryVisitor) ExitOC_QueryOptions(ctx *parser.OC_QueryOptionsContext) {
}

func (s *QueryVisitor) EnterOC_RegularQuery(ctx *parser.OC_RegularQueryContext) {
	s.Query = model.NewRegularQuery()
}

func (s *QueryVisitor) ExitOC_RegularQuery(ctx *parser.OC_RegularQueryContext) {
}

func (s *QueryVisitor) EnterOC_SingleQuery(ctx *parser.OC_SingleQueryContext) {
	s.Query.SingleQuery = model.NewSingleQuery()
}

func (s *QueryVisitor) ExitOC_SingleQuery(ctx *parser.OC_SingleQueryContext) {
}

func (s *QueryVisitor) EnterOC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) {
	s.ctx.Enter(NewMultiPartQueryVisitor())
}

func (s *QueryVisitor) ExitOC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) {
	s.Query.SingleQuery.MultiPartQuery = s.ctx.Exit().(*MultiPartQueryVisitor).Query
}

func (s *QueryVisitor) EnterOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {
	s.ctx.Enter(NewSinglePartQueryVisitor())
}

func (s *QueryVisitor) ExitOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {
	s.Query.SingleQuery.SinglePartQuery = s.ctx.Exit().(*SinglePartQueryVisitor).Query
}

type RemoveVisitor struct {
	BaseVisitor

	currentItem *model.RemoveItem
	Remove      *model.Remove
}

func (s *RemoveVisitor) EnterOC_RemoveItem(ctx *parser.OC_RemoveItemContext) {
	s.currentItem = &model.RemoveItem{}
}

func (s *RemoveVisitor) ExitOC_RemoveItem(ctx *parser.OC_RemoveItemContext) {
	s.Remove.Items = append(s.Remove.Items, s.currentItem)
}

func (s *RemoveVisitor) EnterOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.ctx.Enter(&NodeLabelsVisitor{})
}

func (s *RemoveVisitor) ExitOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.currentItem.KindMatcher.(*model.KindMatcher).Kinds = s.ctx.Exit().(*NodeLabelsVisitor).Kinds
}

func (s *RemoveVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.KindMatcher = &model.KindMatcher{}

	s.ctx.Enter(NewVariableVisitor())
}

func (s *RemoveVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.KindMatcher.(*model.KindMatcher).Reference = s.ctx.Exit().(*VariableVisitor).Variable
}

func (s *RemoveVisitor) EnterOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.ctx.Enter(&PropertyExpressionVisitor{
		PropertyLookup: &model.PropertyLookup{},
	})
}

func (s *RemoveVisitor) ExitOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.currentItem.Property = s.ctx.Exit().(*PropertyExpressionVisitor).PropertyLookup
}

type DeleteVisitor struct {
	BaseVisitor

	Delete *model.Delete
}

func (s *DeleteVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s *DeleteVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.Delete.Expressions = append(s.Delete.Expressions, s.ctx.Exit().(*ExpressionVisitor).Expression)
}

type CreateVisitor struct {
	BaseVisitor

	Create *model.Create
}

func (s *CreateVisitor) EnterOC_Pattern(ctx *parser.OC_PatternContext) {
	s.ctx.Enter(&PatternVisitor{})
}

func (s *CreateVisitor) ExitOC_Pattern(ctx *parser.OC_PatternContext) {
	s.Create.Pattern = s.ctx.Exit().(*PatternVisitor).PatternParts
}

type UpdatingClauseVisitor struct {
	BaseVisitor

	UpdatingClause *model.UpdatingClause
}

func NewUpdatingClauseVisitor() *UpdatingClauseVisitor {
	return &UpdatingClauseVisitor{
		UpdatingClause: &model.UpdatingClause{},
	}
}

func (s *UpdatingClauseVisitor) EnterOC_Create(ctx *parser.OC_CreateContext) {
	s.ctx.Enter(&CreateVisitor{
		Create: &model.Create{
			Unique: HasTokens(ctx, parser.CypherLexerUNIQUE),
		},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Create(ctx *parser.OC_CreateContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*CreateVisitor).Create
}

func (s *UpdatingClauseVisitor) EnterOC_Delete(ctx *parser.OC_DeleteContext) {
	s.ctx.Enter(&DeleteVisitor{
		Delete: &model.Delete{
			Detach: HasTokens(ctx, parser.CypherLexerDETACH),
		},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Delete(ctx *parser.OC_DeleteContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*DeleteVisitor).Delete
}

func (s *UpdatingClauseVisitor) EnterOC_Remove(ctx *parser.OC_RemoveContext) {
	s.ctx.Enter(&RemoveVisitor{
		Remove: &model.Remove{},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Remove(ctx *parser.OC_RemoveContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*RemoveVisitor).Remove
}

func (s *UpdatingClauseVisitor) EnterOC_Set(ctx *parser.OC_SetContext) {
	s.ctx.Enter(&SetVisitor{
		Set: &model.Set{},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Set(ctx *parser.OC_SetContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*SetVisitor).Set
}

type NodeLabelsVisitor struct {
	BaseVisitor

	Kinds graph.Kinds
}

func (s *NodeLabelsVisitor) EnterOC_NodeLabel(ctx *parser.OC_NodeLabelContext) {
}

func (s *NodeLabelsVisitor) ExitOC_NodeLabel(ctx *parser.OC_NodeLabelContext) {
}

func (s *NodeLabelsVisitor) EnterOC_LabelName(ctx *parser.OC_LabelNameContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *NodeLabelsVisitor) ExitOC_LabelName(ctx *parser.OC_LabelNameContext) {
	kind := graph.StringKind(s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name)
	s.Kinds = append(s.Kinds, kind)
}

type SetVisitor struct {
	BaseVisitor

	currentItem *model.SetItem
	Set         *model.Set
}

func (s *SetVisitor) EnterOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.ctx.Enter(&NodeLabelsVisitor{})
}

func (s *SetVisitor) ExitOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.currentItem.Right = s.ctx.Exit().(*NodeLabelsVisitor).Kinds
}

func (s *SetVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s *SetVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.currentItem.Right = s.ctx.Exit().(*ExpressionVisitor).Expression
}

// TODO: Break this out into a SetItem visitor
func (s *SetVisitor) EnterOC_SetItem(ctx *parser.OC_SetItemContext) {
	if HasTokens(ctx, TokenTypeEquals) {
		s.currentItem = &model.SetItem{
			Operator: model.OperatorAssignment,
		}
	} else if HasTokens(ctx, TokenTypeAdditionAssignment) {
		s.currentItem = &model.SetItem{
			Operator: model.OperatorAdditionAssignment,
		}
	} else {
		// Assume that this means we're assigning a label
		s.currentItem = &model.SetItem{
			Operator: model.OperatorLabelAssignment,
		}
	}
}

func (s *SetVisitor) ExitOC_SetItem(ctx *parser.OC_SetItemContext) {
	s.Set.Items = append(s.Set.Items, s.currentItem)
}

func (s *SetVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.ctx.Enter(NewVariableVisitor())
}

func (s *SetVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.Left = s.ctx.Exit().(*VariableVisitor).Variable
}

func (s *SetVisitor) EnterOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.ctx.Enter(&PropertyExpressionVisitor{
		PropertyLookup: &model.PropertyLookup{},
	})
}

func (s *SetVisitor) ExitOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.currentItem.Left = s.ctx.Exit().(*PropertyExpressionVisitor).PropertyLookup
}

type PropertyExpressionVisitor struct {
	BaseVisitor

	PropertyLookup *model.PropertyLookup
}

func (s *PropertyExpressionVisitor) EnterOC_Atom(ctx *parser.OC_AtomContext) {
	s.ctx.Enter(&AtomVisitor{})
}

func (s *PropertyExpressionVisitor) ExitOC_Atom(ctx *parser.OC_AtomContext) {
	s.PropertyLookup.Atom = s.ctx.Exit().(*AtomVisitor).Atom
}

func (s *PropertyExpressionVisitor) EnterOC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) {
}

func (s *PropertyExpressionVisitor) ExitOC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) {
}

func (s *PropertyExpressionVisitor) EnterOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *PropertyExpressionVisitor) ExitOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.PropertyLookup.AddLookupSymbol(s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name)
}
