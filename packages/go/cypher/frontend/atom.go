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
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/parser"
)

type IDInCollectionVisitor struct {
	BaseVisitor

	IDInCollection *model.IDInCollection
}

func NewIDInCollectionVisitor() *IDInCollectionVisitor {
	return &IDInCollectionVisitor{
		IDInCollection: model.NewIDInCollection(),
	}
}

func (s *IDInCollectionVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.ctx.Enter(NewVariableVisitor())
}

func (s *IDInCollectionVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	s.IDInCollection.Variable = s.ctx.Exit().(*VariableVisitor).Variable
}

func (s *IDInCollectionVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *IDInCollectionVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.IDInCollection.Expression = s.ctx.Exit().(*ExpressionVisitor).Expression
}

type FilterExpressionVisitor struct {
	BaseVisitor

	FilterExpression *model.FilterExpression
}

func NewFilterExpressionVisitor() *FilterExpressionVisitor {
	return &FilterExpressionVisitor{
		FilterExpression: model.NewFilterExpression(),
	}
}

func (s *FilterExpressionVisitor) EnterOC_IdInColl(ctx *parser.OC_IdInCollContext) {
	s.ctx.Enter(NewIDInCollectionVisitor())
}

func (s *FilterExpressionVisitor) ExitOC_IdInColl(ctx *parser.OC_IdInCollContext) {
	s.FilterExpression.Specifier = s.ctx.Exit().(*IDInCollectionVisitor).IDInCollection
}

func (s *FilterExpressionVisitor) EnterOC_Where(ctx *parser.OC_WhereContext) {
	s.ctx.Enter(NewWhereVisitor())
}

func (s *FilterExpressionVisitor) ExitOC_Where(ctx *parser.OC_WhereContext) {
	s.FilterExpression.Where = s.ctx.Exit().(*WhereVisitor).Where
}

type QuantifierVisitor struct {
	BaseVisitor

	Quantifier *model.Quantifier
}

func NewQuantifierVisitor(ctx *parser.OC_QuantifierContext) *QuantifierVisitor {
	quantifierType := model.QuantifierTypeInvalid

	if HasTokens(ctx, parser.CypherParserALL) {
		quantifierType = model.QuantifierTypeAll
	} else if HasTokens(ctx, parser.CypherParserANY) {
		quantifierType = model.QuantifierTypeAny
	} else if HasTokens(ctx, parser.CypherParserNONE) {
		quantifierType = model.QuantifierTypeNone
	} else if HasTokens(ctx, parser.CypherParserSINGLE) {
		quantifierType = model.QuantifierTypeSingle
	}

	return &QuantifierVisitor{
		Quantifier: model.NewQuantifier(quantifierType),
	}
}

func (s *QuantifierVisitor) EnterOC_FilterExpression(ctx *parser.OC_FilterExpressionContext) {
	s.ctx.Enter(NewFilterExpressionVisitor())
}

func (s *QuantifierVisitor) ExitOC_FilterExpression(ctx *parser.OC_FilterExpressionContext) {
	s.Quantifier.Filter = s.ctx.Exit().(*FilterExpressionVisitor).FilterExpression
}

// AtomVisitor
//
// oC_Atom
//
//	:  oC_Literal
//	    | oC_Parameter
//	    | oC_CaseExpression
//	    | ( COUNT SP? '(' SP? '*' SP? ')' )
//	    | oC_ListComprehension
//	    | oC_PatternComprehension
//	    | oC_Quantifier
//	    | oC_PatternPredicate
//	    | oC_ParenthesizedExpression
//	    | oC_FunctionInvocation
//	    | oC_ExistentialSubquery
//	    | oC_Variable
//	    ;
type AtomVisitor struct {
	BaseVisitor

	Atom model.Expression
}

func NewAtomVisitor() Visitor {
	return &AtomVisitor{}
}

func (s *AtomVisitor) EnterOC_ParenthesizedExpression(ctx *parser.OC_ParenthesizedExpressionContext) {
	s.ctx.Enter(NewParenthesizedExpressionVisitor())
}

func (s *AtomVisitor) ExitOC_ParenthesizedExpression(ctx *parser.OC_ParenthesizedExpressionContext) {
	s.Atom = s.ctx.Exit().(*ParenthesizedExpressionVisitor).Parenthetical
}

func (s *AtomVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s *AtomVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.Atom = s.ctx.Exit().(*ExpressionVisitor).Expression
}

func (s *AtomVisitor) EnterOC_PatternPredicate(ctx *parser.OC_PatternPredicateContext) {
	s.ctx.Enter(NewPatternPredicateVisitor())
}

func (s *AtomVisitor) ExitOC_PatternPredicate(ctx *parser.OC_PatternPredicateContext) {
	s.Atom = s.ctx.Exit().(*PatternPredicateVisitor).PatternPredicate
}

func (s *AtomVisitor) EnterOC_Quantifier(ctx *parser.OC_QuantifierContext) {
	s.ctx.Enter(NewQuantifierVisitor(ctx))
}

func (s *AtomVisitor) ExitOC_Quantifier(ctx *parser.OC_QuantifierContext) {
	s.Atom = s.ctx.Exit().(*QuantifierVisitor).Quantifier
}

func (s *AtomVisitor) EnterOC_Literal(ctx *parser.OC_LiteralContext) {
	// String and null are special types in the cypher grammar and will not have downstream state transitions
	if ctx.NULL() == nil && ctx.StringLiteral() == nil {
		s.ctx.Enter(NewLiteralVisitor())
	}
}

func (s *AtomVisitor) ExitOC_Literal(ctx *parser.OC_LiteralContext) {
	if ctx.NULL() != nil {
		s.Atom = &model.Literal{
			Null: true,
		}
	} else if ctx.StringLiteral() != nil {
		s.Atom = &model.Literal{
			Value: ctx.GetText(),
		}
	} else {
		s.Atom = s.ctx.Exit().(*LiteralVisitor).Literal
	}
}

func (s *AtomVisitor) EnterOC_Variable(ctx *parser.OC_VariableContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *AtomVisitor) ExitOC_Variable(ctx *parser.OC_VariableContext) {
	result := s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name

	s.Atom = &model.Variable{
		Symbol: result,
	}
}

func (s *AtomVisitor) EnterOC_FunctionInvocation(ctx *parser.OC_FunctionInvocationContext) {
	s.ctx.Enter(NewFunctionInvocationVisitor(ctx))
}

func (s *AtomVisitor) ExitOC_FunctionInvocation(ctx *parser.OC_FunctionInvocationContext) {
	s.Atom = s.ctx.Exit().(*FunctionInvocationVisitor).FunctionInvocation
}
