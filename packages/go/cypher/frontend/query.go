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

	"github.com/specterops/bloodhound/cypher/models/cypher"

	"github.com/specterops/bloodhound/cypher/parser"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type VariableVisitor struct {
	BaseVisitor

	Variable *cypher.Variable
}

func NewVariableVisitor() *VariableVisitor {
	return &VariableVisitor{
		Variable: cypher.NewVariable(),
	}
}

func (s *VariableVisitor) EnterOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
	s.Variable.Symbol = ctx.GetText()
}

type ProjectionVisitor struct {
	BaseVisitor

	currentItem *cypher.ProjectionItem
	Projection  *cypher.Projection
}

func NewProjectionVisitor(ctx *parser.OC_ProjectionBodyContext) *ProjectionVisitor {
	var distinct = false

	if HasTokens(ctx, parser.CypherLexerDISTINCT) {
		distinct = true
	}

	return &ProjectionVisitor{
		Projection: cypher.NewProjection(distinct),
	}
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
			s.Projection.AddItem(cypher.NewGreedyProjectionItem())

		case TokenLiteralComma:
		default:
			s.ctx.AddErrors(fmt.Errorf("invalid token: %s", token))
		}
	}
}

func (s *ProjectionVisitor) EnterOC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) {
	s.currentItem = cypher.NewProjectionItem()
}

func (s *ProjectionVisitor) ExitOC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) {
	s.Projection.AddItem(s.currentItem)
}

func (s *ProjectionVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.ctx.Enter(NewVariableVisitor())
}

func (s *ProjectionVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.Alias = s.ctx.Exit().(*VariableVisitor).Variable
}

func (s *ProjectionVisitor) EnterOC_Order(ctx *parser.OC_OrderContext) {
	s.Projection.Order = &cypher.Order{}
}

func (s ProjectionVisitor) EnterOC_SortItem(ctx *parser.OC_SortItemContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s ProjectionVisitor) ExitOC_SortItem(ctx *parser.OC_SortItemContext) {
	var (
		expression = s.ctx.Exit().(*ExpressionVisitor).Expression
		sortItem   = &cypher.SortItem{
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
	s.Projection.Skip = &cypher.Skip{
		Value: expression,
	}
}

func (s *ProjectionVisitor) EnterOC_Limit(ctx *parser.OC_LimitContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *ProjectionVisitor) ExitOC_Limit(ctx *parser.OC_LimitContext) {
	expression := s.ctx.Exit().(*ExpressionVisitor).Expression
	s.Projection.Limit = &cypher.Limit{
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

	PatternRange *cypher.PatternRange
}

func NewRangeLiteralVisitor() *RangeLiteralVisitor {
	return &RangeLiteralVisitor{
		PatternRange: &cypher.PatternRange{},
	}
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

type WithVisitor struct {
	BaseVisitor

	With *cypher.With
}

func NewWithVisitor() *WithVisitor {
	return &WithVisitor{
		With: cypher.NewWith(),
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

	Match *cypher.Match
}

func NewMatchVisitor(ctx *parser.OC_MatchContext) *MatchVisitor {
	optional := false

	if HasTokens(ctx, parser.CypherLexerOPTIONAL) {
		optional = true
	}

	return &MatchVisitor{
		Match: cypher.NewMatch(optional),
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

	Unwind *cypher.Unwind
}

func NewUnwindVisitor() *UnwindVisitor {
	return &UnwindVisitor{
		Unwind: cypher.NewUnwind(),
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
	s.Unwind.Variable = s.ctx.Exit().(*VariableVisitor).Variable
}

type ReadingClauseVisitor struct {
	BaseVisitor

	ReadingClause *cypher.ReadingClause
}

func NewReadingClauseVisitor() *ReadingClauseVisitor {
	return &ReadingClauseVisitor{
		ReadingClause: cypher.NewReadingClause(),
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

	Query *cypher.SinglePartQuery
}

func NewSinglePartQueryVisitor() *SinglePartQueryVisitor {
	return &SinglePartQueryVisitor{
		Query: cypher.NewSinglePartQuery(),
	}
}

func (s *SinglePartQueryVisitor) EnterOC_Return(ctx *parser.OC_ReturnContext) {
	s.Query.Return = &cypher.Return{}
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
	s.ctx.HasMutation = true
	s.Query.AddUpdatingClause(s.ctx.Exit().(*UpdatingClauseVisitor).UpdatingClause)
}

type MultiPartQueryVisitor struct {
	BaseVisitor

	Query   *cypher.MultiPartQuery
	partIdx int
}

func NewMultiPartQueryVisitor() *MultiPartQueryVisitor {
	return &MultiPartQueryVisitor{
		Query:   cypher.NewMultiPartQuery(),
		partIdx: 0,
	}
}

func (s *MultiPartQueryVisitor) EnterOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {
	// If the part index is equal to the length of parts then this signifies that a new query part
	// is required. We do not advance the index here - this is done with the following `with`
	// cypher AST component
	if len(s.Query.Parts) == s.partIdx {
		s.Query.Parts = append(s.Query.Parts, cypher.NewMultiPartQueryPart())
	}

	s.ctx.Enter(NewReadingClauseVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {
	s.Query.CurrentPart().AddReadingClause(s.ctx.Exit().(*ReadingClauseVisitor).ReadingClause)
}

func (s *MultiPartQueryVisitor) EnterOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	if len(s.Query.Parts) == s.partIdx {
		s.Query.Parts = append(s.Query.Parts, cypher.NewMultiPartQueryPart())
	}

	s.ctx.Enter(NewUpdatingClauseVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	// Make sure to mark that this multipart query part contains a mutation (non-read operation). This
	// field is being set to make it easier for downstream consumers of the openCypher AST to identify
	// if this query contains a mutation.
	s.ctx.HasMutation = true

	s.Query.CurrentPart().AddUpdatingClause(s.ctx.Exit().(*UpdatingClauseVisitor).UpdatingClause)
}

func (s *MultiPartQueryVisitor) EnterOC_With(ctx *parser.OC_WithContext) {
	if len(s.Query.Parts) == s.partIdx {
		s.Query.Parts = append(s.Query.Parts, cypher.NewMultiPartQueryPart())
	}

	s.ctx.Enter(NewWithVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_With(ctx *parser.OC_WithContext) {
	s.Query.CurrentPart().With = s.ctx.Exit().(*WithVisitor).With

	// Advance the part index so a new multipart query part gets allocated for the next reading
	// or updating clause
	s.partIdx += 1
}

func (s *MultiPartQueryVisitor) EnterOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {
	s.ctx.Enter(NewSinglePartQueryVisitor())
}

func (s *MultiPartQueryVisitor) ExitOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {
	s.Query.SinglePartQuery = s.ctx.Exit().(*SinglePartQueryVisitor).Query
}

type QueryVisitor struct {
	BaseVisitor

	Query *cypher.RegularQuery
}

func (s *QueryVisitor) EnterOC_RegularQuery(ctx *parser.OC_RegularQueryContext) {
	s.Query = cypher.NewRegularQuery()
}

func (s *QueryVisitor) EnterOC_SingleQuery(ctx *parser.OC_SingleQueryContext) {
	s.Query.SingleQuery = cypher.NewSingleQuery()
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

	currentItem *cypher.RemoveItem
	Remove      *cypher.Remove
}

func (s *RemoveVisitor) EnterOC_RemoveItem(ctx *parser.OC_RemoveItemContext) {
	s.currentItem = &cypher.RemoveItem{}
}

func (s *RemoveVisitor) ExitOC_RemoveItem(ctx *parser.OC_RemoveItemContext) {
	s.Remove.Items = append(s.Remove.Items, s.currentItem)
}

func (s *RemoveVisitor) EnterOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.ctx.Enter(&NodeLabelsVisitor{})
}

func (s *RemoveVisitor) ExitOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.currentItem.KindMatcher.Kinds = s.ctx.Exit().(*NodeLabelsVisitor).Kinds
}

func (s *RemoveVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.KindMatcher = &cypher.KindMatcher{}

	s.ctx.Enter(NewVariableVisitor())
}

func (s *RemoveVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.currentItem.KindMatcher.Reference = s.ctx.Exit().(*VariableVisitor).Variable
}

func (s *RemoveVisitor) EnterOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.ctx.Enter(&PropertyExpressionVisitor{
		PropertyLookup: &cypher.PropertyLookup{},
	})
}

func (s *RemoveVisitor) ExitOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.currentItem.Property = s.ctx.Exit().(*PropertyExpressionVisitor).PropertyLookup
}

type DeleteVisitor struct {
	BaseVisitor

	Delete *cypher.Delete
}

func (s *DeleteVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s *DeleteVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.Delete.Expressions = append(s.Delete.Expressions, s.ctx.Exit().(*ExpressionVisitor).Expression)
}

type CreateVisitor struct {
	BaseVisitor

	Create *cypher.Create
}

func (s *CreateVisitor) EnterOC_Pattern(ctx *parser.OC_PatternContext) {
	s.ctx.Enter(&PatternVisitor{})
}

func (s *CreateVisitor) ExitOC_Pattern(ctx *parser.OC_PatternContext) {
	s.Create.Pattern = s.ctx.Exit().(*PatternVisitor).PatternParts
}

type MergeVisitor struct {
	BaseVisitor

	Merge *cypher.Merge
}

func (s *MergeVisitor) EnterOC_PatternPart(ctx *parser.OC_PatternPartContext) {
	s.ctx.Enter(&PatternPartVisitor{
		PatternPart: &cypher.PatternPart{},
	})
}

func (s *MergeVisitor) ExitOC_PatternPart(ctx *parser.OC_PatternPartContext) {
	s.Merge.PatternPart = s.ctx.Exit().(*PatternPartVisitor).PatternPart
}

func (s *MergeVisitor) EnterOC_MergeAction(ctx *parser.OC_MergeActionContext) {
	s.ctx.Enter(&MergeActionVisitor{
		MergeAction: &cypher.MergeAction{
			OnCreate: HasTokens(ctx, parser.CypherLexerON, parser.CypherLexerCREATE),
			OnMatch:  HasTokens(ctx, parser.CypherLexerON, parser.CypherLexerMATCH),
		},
	})
}

func (s *MergeVisitor) ExitOC_MergeAction(ctx *parser.OC_MergeActionContext) {
	s.Merge.MergeActions = append(s.Merge.MergeActions, s.ctx.Exit().(*MergeActionVisitor).MergeAction)
}

type MergeActionVisitor struct {
	BaseVisitor

	MergeAction *cypher.MergeAction
}

func (s *MergeActionVisitor) EnterOC_Set(ctx *parser.OC_SetContext) {
	s.ctx.Enter(&SetVisitor{
		Set: &cypher.Set{},
	})
}

func (s *MergeActionVisitor) ExitOC_Set(ctx *parser.OC_SetContext) {
	s.MergeAction.Set = s.ctx.Exit().(*SetVisitor).Set
}

type UpdatingClauseVisitor struct {
	BaseVisitor

	UpdatingClause *cypher.UpdatingClause
}

func NewUpdatingClauseVisitor() *UpdatingClauseVisitor {
	return &UpdatingClauseVisitor{
		UpdatingClause: &cypher.UpdatingClause{},
	}
}

func (s *UpdatingClauseVisitor) EnterOC_Create(ctx *parser.OC_CreateContext) {
	s.ctx.Enter(&CreateVisitor{
		Create: &cypher.Create{
			Unique: HasTokens(ctx, parser.CypherLexerUNIQUE),
		},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Create(ctx *parser.OC_CreateContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*CreateVisitor).Create
}

func (s *UpdatingClauseVisitor) EnterOC_Delete(ctx *parser.OC_DeleteContext) {
	s.ctx.Enter(&DeleteVisitor{
		Delete: &cypher.Delete{
			Detach: HasTokens(ctx, parser.CypherLexerDETACH),
		},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Delete(ctx *parser.OC_DeleteContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*DeleteVisitor).Delete
}

func (s *UpdatingClauseVisitor) EnterOC_Remove(ctx *parser.OC_RemoveContext) {
	s.ctx.Enter(&RemoveVisitor{
		Remove: &cypher.Remove{},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Remove(ctx *parser.OC_RemoveContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*RemoveVisitor).Remove
}

func (s *UpdatingClauseVisitor) EnterOC_Set(ctx *parser.OC_SetContext) {
	s.ctx.Enter(&SetVisitor{
		Set: &cypher.Set{},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Set(ctx *parser.OC_SetContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*SetVisitor).Set
}

func (s *UpdatingClauseVisitor) EnterOC_Merge(ctx *parser.OC_MergeContext) {
	s.ctx.Enter(&MergeVisitor{
		Merge: &cypher.Merge{},
	})
}

func (s *UpdatingClauseVisitor) ExitOC_Merge(ctx *parser.OC_MergeContext) {
	s.UpdatingClause.Clause = s.ctx.Exit().(*MergeVisitor).Merge
}

type NodeLabelsVisitor struct {
	BaseVisitor

	Kinds graph.Kinds
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

	currentItem *cypher.SetItem
	Set         *cypher.Set
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
		s.currentItem = &cypher.SetItem{
			Operator: cypher.OperatorAssignment,
		}
	} else if HasTokens(ctx, TokenTypeAdditionAssignment) {
		s.currentItem = &cypher.SetItem{
			Operator: cypher.OperatorAdditionAssignment,
		}
	} else {
		// Assume that this means we're assigning a label
		s.currentItem = &cypher.SetItem{
			Operator: cypher.OperatorLabelAssignment,
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
		PropertyLookup: &cypher.PropertyLookup{},
	})
}

func (s *SetVisitor) ExitOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {
	s.currentItem.Left = s.ctx.Exit().(*PropertyExpressionVisitor).PropertyLookup
}

type PropertyExpressionVisitor struct {
	BaseVisitor

	PropertyLookup *cypher.PropertyLookup
}

func (s *PropertyExpressionVisitor) EnterOC_Atom(ctx *parser.OC_AtomContext) {
	s.ctx.Enter(&AtomVisitor{})
}

func (s *PropertyExpressionVisitor) ExitOC_Atom(ctx *parser.OC_AtomContext) {
	s.PropertyLookup.Atom = s.ctx.Exit().(*AtomVisitor).Atom
}

func (s *PropertyExpressionVisitor) EnterOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *PropertyExpressionVisitor) ExitOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.PropertyLookup.SetSymbol(s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name)
}
