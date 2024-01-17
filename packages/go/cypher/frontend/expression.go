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
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/parser"
	"github.com/specterops/bloodhound/dawgs/graph"
)

// PropertiesVisitor
//
// oC_Properties: oC_MapLiteral | oC_Parameter | oC_LegacyParameter
type PropertiesVisitor struct {
	BaseVisitor

	Properties *model.Properties
}

func NewPropertiesVisitor() *PropertiesVisitor {
	return &PropertiesVisitor{
		Properties: model.NewProperties(),
	}
}

func (s *PropertiesVisitor) EnterOC_MapLiteral(ctx *parser.OC_MapLiteralContext) {
	s.ctx.Enter(NewMapLiteralVisitor())
}

func (s *PropertiesVisitor) ExitOC_MapLiteral(ctx *parser.OC_MapLiteralContext) {
	s.Properties.Map = s.ctx.Exit().(*MapLiteralVisitor).Map
}

func (s *PropertiesVisitor) EnterOC_Parameter(ctx *parser.OC_ParameterContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *PropertiesVisitor) ExitOC_Parameter(ctx *parser.OC_ParameterContext) {
	if symbolicName := s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name; symbolicName == "" {
		s.Properties.Parameter = &model.Parameter{
			Symbol: strings.TrimPrefix(ctx.GetText(), "$"),
		}
	} else {
		s.Properties.Parameter = &model.Parameter{
			Symbol: symbolicName,
		}
	}
}

type ExpressionVisitor struct {
	BaseVisitor

	Expression model.Expression
}

func NewExpressionVisitor() Visitor {
	return &ExpressionVisitor{}
}

func (s *ExpressionVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
}

func (s *ExpressionVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
}

func (s *ExpressionVisitor) EnterOC_NonArithmeticOperatorExpression(ctx *parser.OC_NonArithmeticOperatorExpressionContext) {
	s.ctx.Enter(&NonArithmeticOperatorExpressionVisitor{})
}

func (s *ExpressionVisitor) ExitOC_NonArithmeticOperatorExpression(ctx *parser.OC_NonArithmeticOperatorExpressionContext) {
	s.Expression = s.ctx.Exit().(*NonArithmeticOperatorExpressionVisitor).Expression
}

func (s *ExpressionVisitor) EnterOC_OrExpression(ctx *parser.OC_OrExpressionContext) {
	if len(ctx.AllOR()) > 0 {
		s.ctx.Enter(&JoiningVisitor{
			Joined: &model.Disjunction{},
		})
	}
}

func (s *ExpressionVisitor) ExitOC_OrExpression(ctx *parser.OC_OrExpressionContext) {
	if len(ctx.AllOR()) > 0 {
		visitor := s.ctx.Exit().(*JoiningVisitor)
		s.Expression = visitor.Joined
	}
}

func (s *ExpressionVisitor) EnterOC_XorExpression(ctx *parser.OC_XorExpressionContext) {
	if len(ctx.AllXOR()) > 0 {
		s.ctx.Enter(&JoiningVisitor{
			Joined: &model.ExclusiveDisjunction{},
		})
	}
}

func (s *ExpressionVisitor) ExitOC_XorExpression(ctx *parser.OC_XorExpressionContext) {
	if len(ctx.AllXOR()) > 0 {
		visitor := s.ctx.Exit().(*JoiningVisitor)
		s.Expression = visitor.Joined
	}
}

func (s *ExpressionVisitor) EnterOC_AndExpression(ctx *parser.OC_AndExpressionContext) {
	if len(ctx.AllAND()) > 0 {
		s.ctx.Enter(&JoiningVisitor{
			Joined: &model.Conjunction{},
		})
	}
}

func (s *ExpressionVisitor) ExitOC_AndExpression(ctx *parser.OC_AndExpressionContext) {
	if len(ctx.AllAND()) > 0 {
		visitor := s.ctx.Exit().(*JoiningVisitor)
		s.Expression = visitor.Joined
	}
}

func (s *ExpressionVisitor) EnterOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {
	if len(ctx.AllOC_PartialComparisonExpression()) > 0 {
		s.ctx.Enter(&ComparisonVisitor{})
	}
}

func (s *ExpressionVisitor) ExitOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {
	if len(ctx.AllOC_PartialComparisonExpression()) > 0 {
		result := s.ctx.Exit().(*ComparisonVisitor).Comparison
		s.Expression = result
	}
}

func (s *ExpressionVisitor) EnterOC_NotExpression(ctx *parser.OC_NotExpressionContext) {
	if len(ctx.AllNOT()) > 0 {
		s.ctx.Enter(&NegationVisitor{
			Negation: &model.Negation{},
		})
	}
}

func (s *ExpressionVisitor) ExitOC_NotExpression(ctx *parser.OC_NotExpressionContext) {
	if len(ctx.AllNOT()) > 0 {
		visitor := s.ctx.Exit().(*NegationVisitor)
		s.Expression = visitor.Negation
	}
}

func (s *ExpressionVisitor) EnterOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	s.ctx.Enter(&StringListNullPredicateExpressionVisitor{})
}

func (s *ExpressionVisitor) ExitOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	s.Expression = s.ctx.Exit().(*StringListNullPredicateExpressionVisitor).Expression
}

type tokenLiteralIterator struct {
	tokens []string
	index  int
}

func newTokenLiteralIterator(astNode TokenProvider) *tokenLiteralIterator {
	var tokens []string

	for idx := 0; idx < astNode.GetChildCount(); idx++ {
		nextChild := astNode.GetChild(idx)

		if terminalNode, typeOK := nextChild.(*antlr.TerminalNodeImpl); typeOK {
			formattedTerminalNodeText := strings.TrimSpace(terminalNode.GetText())

			if len(formattedTerminalNodeText) > 0 {
				tokens = append(tokens, formattedTerminalNodeText)
			}
		}
	}

	return &tokenLiteralIterator{
		tokens: tokens,
		index:  0,
	}
}

func (s *tokenLiteralIterator) HasTokens() bool {
	numTokens := len(s.tokens)
	return numTokens > 0 && s.index < numTokens
}

func (s *tokenLiteralIterator) NextToken() string {
	var nextToken string

	if s.index < len(s.tokens) {
		nextToken = s.tokens[s.index]
		s.index++
	}

	return nextToken
}

func (s *tokenLiteralIterator) NextOperator() model.Operator {
	nextOperator, _ := model.ParseOperator(s.NextToken())
	return nextOperator
}

type ArithmeticExpressionVisitor struct {
	BaseVisitor

	operatorTokens *tokenLiteralIterator
	Expression     model.Expression
}

func NewArithmeticExpressionVisitor(operatorTokens *tokenLiteralIterator) *ArithmeticExpressionVisitor {
	return &ArithmeticExpressionVisitor{
		operatorTokens: operatorTokens,
	}
}

func (s *ArithmeticExpressionVisitor) assignExpression(expression model.Expression) {
	if s.operatorTokens.HasTokens() {
		// If there's no expression in this visitor but there are collected operators then this AST node
		// is the left-most expression of the arithmetic operator expression
		if s.Expression == nil {
			s.Expression = &model.ArithmeticExpression{
				Left: expression,
			}
		} else {
			// This is a subsequent component of the arithmetic operator expression
			s.Expression.(*model.ArithmeticExpression).AddArithmeticExpressionPart(&model.PartialArithmeticExpression{
				Operator: s.operatorTokens.NextOperator(),
				Right:    expression,
			})
		}
	} else {
		// With no collected operators then it's assumed that this is purely a non-arithmetic operator expression
		s.Expression = expression
	}
}

func (s *ArithmeticExpressionVisitor) enterSubArithmeticExpression(astNode TokenProvider) {
	if operators := newTokenLiteralIterator(astNode); operators.HasTokens() {
		s.ctx.Enter(NewArithmeticExpressionVisitor(operators))
	}
}

func (s *ArithmeticExpressionVisitor) exitSubArithmeticExpression(astNode TokenProvider) {
	if operators := newTokenLiteralIterator(astNode); operators.HasTokens() {
		s.assignExpression(s.ctx.Exit().(*ArithmeticExpressionVisitor).Expression)
	}
}

func (s *ArithmeticExpressionVisitor) EnterOC_MultiplyDivideModuloExpression(ctx *parser.OC_MultiplyDivideModuloExpressionContext) {
	s.enterSubArithmeticExpression(ctx)
}

func (s *ArithmeticExpressionVisitor) ExitOC_MultiplyDivideModuloExpression(ctx *parser.OC_MultiplyDivideModuloExpressionContext) {
	s.exitSubArithmeticExpression(ctx)
}

func (s *ArithmeticExpressionVisitor) EnterOC_PowerOfExpression(ctx *parser.OC_PowerOfExpressionContext) {
	s.enterSubArithmeticExpression(ctx)
}

func (s *ArithmeticExpressionVisitor) ExitOC_PowerOfExpression(ctx *parser.OC_PowerOfExpressionContext) {
	s.exitSubArithmeticExpression(ctx)
}

func (s *ArithmeticExpressionVisitor) EnterOC_UnaryAddOrSubtractExpression(ctx *parser.OC_UnaryAddOrSubtractExpressionContext) {
	s.enterSubArithmeticExpression(ctx)
}

func (s *ArithmeticExpressionVisitor) ExitOC_UnaryAddOrSubtractExpression(ctx *parser.OC_UnaryAddOrSubtractExpressionContext) {
	s.exitSubArithmeticExpression(ctx)
}

func (s *ArithmeticExpressionVisitor) EnterOC_NonArithmeticOperatorExpression(ctx *parser.OC_NonArithmeticOperatorExpressionContext) {
	s.ctx.Enter(&NonArithmeticOperatorExpressionVisitor{})
}

func (s *ArithmeticExpressionVisitor) ExitOC_NonArithmeticOperatorExpression(ctx *parser.OC_NonArithmeticOperatorExpressionContext) {
	s.assignExpression(s.ctx.Exit().(*NonArithmeticOperatorExpressionVisitor).Expression)
}

// StringListNullPredicateExpressionVisitor
// oC_AddOrSubtractExpression ( oC_StringPredicateExpression | oC_ListPredicateExpression | oC_NullPredicateExpression )*
type StringListNullPredicateExpressionVisitor struct {
	BaseVisitor

	Expression model.Expression
}

func (s *StringListNullPredicateExpressionVisitor) EnterOC_AddOrSubtractExpression(ctx *parser.OC_AddOrSubtractExpressionContext) {
	s.ctx.Enter(NewArithmeticExpressionVisitor(newTokenLiteralIterator(ctx)))
}

func (s *StringListNullPredicateExpressionVisitor) ExitOC_AddOrSubtractExpression(ctx *parser.OC_AddOrSubtractExpressionContext) {
	expression := s.ctx.Exit().(*ArithmeticExpressionVisitor).Expression

	switch typedExpression := s.Expression.(type) {
	case *model.Comparison:
		typedExpression.LastPartial().Right = expression

	default:
		s.Expression = expression
	}
}

func (s *StringListNullPredicateExpressionVisitor) EnterOC_RegularExpression(ctx *parser.OC_RegularExpressionContext) {
	s.Expression = &model.Comparison{
		Left: s.Expression,
		Partials: []*model.PartialComparison{{
			Operator: model.OperatorRegexMatch,
		}},
	}
}

func (s *StringListNullPredicateExpressionVisitor) ExitOC_RegularExpression(ctx *parser.OC_RegularExpressionContext) {
}

func (s *StringListNullPredicateExpressionVisitor) EnterOC_ListPredicateExpression(ctx *parser.OC_ListPredicateExpressionContext) {
	if ctx.GetToken(parser.CypherLexerIN, 0) != nil {
		s.Expression = &model.Comparison{
			Left: s.Expression,
			Partials: []*model.PartialComparison{{
				Operator: model.OperatorIn,
			}},
		}
	}
}

func (s *StringListNullPredicateExpressionVisitor) ExitOC_ListPredicateExpression(ctx *parser.OC_ListPredicateExpressionContext) {
}

func (s *StringListNullPredicateExpressionVisitor) EnterOC_NullPredicateExpression(ctx *parser.OC_NullPredicateExpressionContext) {
	literalNull := &model.Literal{
		Null: true,
	}

	if HasTokens(ctx, parser.CypherLexerIS) {
		if HasTokens(ctx, parser.CypherLexerNOT) {
			s.Expression = &model.Comparison{
				Left: s.Expression,
				Partials: []*model.PartialComparison{{
					Operator: model.OperatorIsNot,
					Right:    literalNull,
				}},
			}
		} else {
			s.Expression = &model.Comparison{
				Left: s.Expression,
				Partials: []*model.PartialComparison{{
					Operator: model.OperatorIs,
					Right:    literalNull,
				}},
			}
		}
	}
}

func (s *StringListNullPredicateExpressionVisitor) ExitOC_NullPredicateExpression(ctx *parser.OC_NullPredicateExpressionContext) {
}

func (s *StringListNullPredicateExpressionVisitor) EnterOC_StringPredicateExpression(ctx *parser.OC_StringPredicateExpressionContext) {
	if HasTokens(ctx, parser.CypherLexerSTARTS, parser.CypherLexerWITH) {
		s.Expression = &model.Comparison{
			Left: s.Expression,
			Partials: []*model.PartialComparison{{
				Operator: model.OperatorStartsWith,
			}},
		}
	} else if HasTokens(ctx, parser.CypherLexerENDS, parser.CypherLexerWITH) {
		s.Expression = &model.Comparison{
			Left: s.Expression,
			Partials: []*model.PartialComparison{{
				Operator: model.OperatorEndsWith,
			}},
		}
	} else if HasTokens(ctx, parser.CypherLexerCONTAINS) {
		s.Expression = &model.Comparison{
			Left: s.Expression,
			Partials: []*model.PartialComparison{{
				Operator: model.OperatorContains,
			}},
		}
	}
}

func (s *StringListNullPredicateExpressionVisitor) ExitOC_StringPredicateExpression(ctx *parser.OC_StringPredicateExpressionContext) {
}

// oC_Atom ( ( SP? oC_ListOperatorExpression ) | ( SP? oC_PropertyLookup ) )* ( SP? oC_NodeLabels )?
type NonArithmeticOperatorExpressionVisitor struct {
	BaseVisitor

	Expression      model.Expression
	PropertyKeyName string
}

func (s *NonArithmeticOperatorExpressionVisitor) EnterOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
	s.Expression = &model.KindMatcher{
		Reference: s.Expression,
	}
}

func (s *NonArithmeticOperatorExpressionVisitor) ExitOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {
}

func (s *NonArithmeticOperatorExpressionVisitor) EnterOC_NodeLabel(ctx *parser.OC_NodeLabelContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *NonArithmeticOperatorExpressionVisitor) ExitOC_NodeLabel(ctx *parser.OC_NodeLabelContext) {
	matcher := s.Expression.(*model.KindMatcher)
	matcher.Kinds = append(matcher.Kinds, graph.StringKind(s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name))
}

func (s *NonArithmeticOperatorExpressionVisitor) EnterOC_Atom(ctx *parser.OC_AtomContext) {
	if !HasTokens(ctx, parser.CypherLexerCOUNT) {
		s.ctx.Enter(NewAtomVisitor())
	}
}

func (s *NonArithmeticOperatorExpressionVisitor) ExitOC_Atom(ctx *parser.OC_AtomContext) {
	if HasTokens(ctx, parser.CypherLexerCOUNT) {
		s.Expression = &model.FunctionInvocation{
			Name:      "count",
			Arguments: []model.Expression{model.GreedyRangeQuantifier},
		}
	} else {
		s.Expression = s.ctx.Exit().(*AtomVisitor).Atom
	}
}

func (s *NonArithmeticOperatorExpressionVisitor) EnterOC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) {
	s.Expression = &model.PropertyLookup{
		Atom: s.Expression,
	}
}

func (s *NonArithmeticOperatorExpressionVisitor) ExitOC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) {
	s.Expression.(*model.PropertyLookup).AddLookupSymbol(s.PropertyKeyName)
}

func (s *NonArithmeticOperatorExpressionVisitor) EnterOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *NonArithmeticOperatorExpressionVisitor) ExitOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.PropertyKeyName = s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name
}
