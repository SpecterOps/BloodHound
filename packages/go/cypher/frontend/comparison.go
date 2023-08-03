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

	"github.com/antlr4-go/antlr/v4"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/parser"
)

// oC_PartialComparisonExpression
//                           :  ( '=' SP? oC_StringListNullPredicateExpression )
//                               | ( '<>' SP? oC_StringListNullPredicateExpression )
//                               | ( '<' SP? oC_StringListNullPredicateExpression )
//                               | ( '>' SP? oC_StringListNullPredicateExpression )
//                               | ( '<=' SP? oC_StringListNullPredicateExpression )
//                               | ( '>=' SP? oC_StringListNullPredicateExpression )
//                               ;

// PartialComparisonVisitor
type PartialComparisonVisitor struct {
	BaseVisitor

	PartialComparison *model.PartialComparison
}

func NewPartialComparisonVisitor() *PartialComparisonVisitor {
	return &PartialComparisonVisitor{
		PartialComparison: &model.PartialComparison{},
	}
}

func (s *PartialComparisonVisitor) EnterOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	s.ctx.Enter(&StringListNullPredicateExpressionVisitor{})
}

func (s *PartialComparisonVisitor) ExitOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	result := s.ctx.Exit().(*StringListNullPredicateExpressionVisitor).Expression
	s.PartialComparison.Right = result
}

// oC_StringListNullPredicateExpression ( SP? oC_PartialComparisonExpression )*
type ComparisonVisitor struct {
	BaseVisitor

	Comparison *model.Comparison
}

func (s *ComparisonVisitor) EnterOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	s.ctx.Enter(&StringListNullPredicateExpressionVisitor{})
}

func (s *ComparisonVisitor) ExitOC_StringListNullPredicateExpression(ctx *parser.OC_StringListNullPredicateExpressionContext) {
	result := s.ctx.Exit().(*StringListNullPredicateExpressionVisitor).Expression

	s.Comparison = &model.Comparison{
		Left: result,
	}
}

func (s *ComparisonVisitor) EnterOC_PartialComparisonExpression(ctx *parser.OC_PartialComparisonExpressionContext) {
	partialComparisonVisitor := NewPartialComparisonVisitor()

	switch operatorChild := ctx.GetChild(0).(type) {
	case *antlr.TerminalNodeImpl:
		if operator, err := model.ParseOperator(operatorChild.GetText()); err != nil {
			s.ctx.AddErrors(err)
		} else {
			partialComparisonVisitor.PartialComparison.Operator = operator
		}

	default:
		s.ctx.AddErrors(fmt.Errorf("expected oC_PartialComparisonExpression to contain an operator token at branch position 0 but saw: %T %+v", operatorChild, operatorChild))
	}

	s.ctx.Enter(partialComparisonVisitor)
}

func (s *ComparisonVisitor) ExitOC_PartialComparisonExpression(ctx *parser.OC_PartialComparisonExpressionContext) {
	s.Comparison.AddPartialComparison(s.ctx.Exit().(*PartialComparisonVisitor).PartialComparison)
}
