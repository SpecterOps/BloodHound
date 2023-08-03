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

type ParenthesizedExpressionVisitor struct {
	BaseVisitor

	Parenthetical *model.Parenthetical
}

func NewParenthesizedExpressionVisitor() *ParenthesizedExpressionVisitor {
	return &ParenthesizedExpressionVisitor{}
}

func (s *ParenthesizedExpressionVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *ParenthesizedExpressionVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.Parenthetical = &model.Parenthetical{
		Expression: s.ctx.Exit().(*ExpressionVisitor).Expression,
	}
}

type NegationVisitor struct {
	BaseVisitor

	Negation *model.Negation
}

func (s *NegationVisitor) EnterOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {
	if len(ctx.AllOC_PartialComparisonExpression()) > 0 {
		s.ctx.Enter(&ComparisonVisitor{})
	}
}

func (s *NegationVisitor) ExitOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {
	if len(ctx.AllOC_PartialComparisonExpression()) > 0 {
		result := s.ctx.Exit().(*ComparisonVisitor).Comparison
		s.Negation.Expression = result
	}
}

func (s *NegationVisitor) EnterOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	s.ctx.Enter(&StringListNullPredicateExpressionVisitor{})
}

func (s *NegationVisitor) ExitOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	result := s.ctx.Exit().(*StringListNullPredicateExpressionVisitor).Expression
	s.Negation.Expression = result
}

type JoiningVisitor struct {
	BaseVisitor

	Joined model.ExpressionList
}

func (s *JoiningVisitor) EnterOC_NotExpression(ctx *parser.OC_NotExpressionContext) {
	if len(ctx.AllNOT()) > 0 {
		s.ctx.Enter(&NegationVisitor{
			Negation: &model.Negation{},
		})
	}
}

func (s *JoiningVisitor) ExitOC_NotExpression(ctx *parser.OC_NotExpressionContext) {
	if len(ctx.AllNOT()) > 0 {
		visitor := s.ctx.Exit().(*NegationVisitor)
		s.Joined.Add(visitor.Negation)
	}
}

func (s *JoiningVisitor) EnterOC_OrExpression(ctx *parser.OC_OrExpressionContext) {
	if len(ctx.AllOR()) > 0 {
		s.ctx.Enter(&JoiningVisitor{
			Joined: &model.Disjunction{},
		})
	}
}

func (s *JoiningVisitor) ExitOC_OrExpression(ctx *parser.OC_OrExpressionContext) {
	if len(ctx.AllOR()) > 0 {
		visitor := s.ctx.Exit().(*JoiningVisitor)
		s.Joined.Add(visitor.Joined)
	}
}

func (s *JoiningVisitor) EnterOC_XorExpression(ctx *parser.OC_XorExpressionContext) {
	if len(ctx.AllXOR()) > 0 {
		s.ctx.Enter(&JoiningVisitor{
			Joined: &model.ExclusiveDisjunction{},
		})
	}
}

func (s *JoiningVisitor) ExitOC_XorExpression(ctx *parser.OC_XorExpressionContext) {
	if len(ctx.AllXOR()) > 0 {
		visitor := s.ctx.Exit().(*JoiningVisitor)
		s.Joined.Add(visitor.Joined)
	}
}

func (s *JoiningVisitor) EnterOC_AndExpression(ctx *parser.OC_AndExpressionContext) {
	if len(ctx.AllAND()) > 0 {
		s.ctx.Enter(&JoiningVisitor{
			Joined: &model.Conjunction{},
		})
	}
}

func (s *JoiningVisitor) ExitOC_AndExpression(ctx *parser.OC_AndExpressionContext) {
	if len(ctx.AllAND()) > 0 {
		visitor := s.ctx.Exit().(*JoiningVisitor)
		s.Joined.Add(visitor.Joined)
	}
}

func (s *JoiningVisitor) EnterOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {
	if len(ctx.AllOC_PartialComparisonExpression()) > 0 {
		s.ctx.Enter(&ComparisonVisitor{})
	}
}

func (s *JoiningVisitor) ExitOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {
	if len(ctx.AllOC_PartialComparisonExpression()) > 0 {
		result := s.ctx.Exit().(*ComparisonVisitor).Comparison
		s.Joined.Add(result)
	}
}

func (s *JoiningVisitor) EnterOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	s.ctx.Enter(&StringListNullPredicateExpressionVisitor{})
}

func (s *JoiningVisitor) ExitOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	result := s.ctx.Exit().(*StringListNullPredicateExpressionVisitor).Expression
	s.Joined.Add(result)
}
