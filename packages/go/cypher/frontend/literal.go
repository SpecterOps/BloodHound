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
)

type SymbolicNameOrReservedWordVisitor struct {
	BaseVisitor

	Name string
}

func (s *SymbolicNameOrReservedWordVisitor) EnterOC_LabelName(ctx *parser.OC_LabelNameContext) {
}

func (s *SymbolicNameOrReservedWordVisitor) ExitOC_LabelName(ctx *parser.OC_LabelNameContext) {
}

func (s *SymbolicNameOrReservedWordVisitor) EnterOC_SchemaName(ctx *parser.OC_SchemaNameContext) {
	s.Name = ctx.GetText()
}

func (s *SymbolicNameOrReservedWordVisitor) ExitOC_SchemaName(ctx *parser.OC_SchemaNameContext) {
}

func (s *SymbolicNameOrReservedWordVisitor) EnterOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
	s.Name = ctx.GetText()
}

func (s *SymbolicNameOrReservedWordVisitor) ExitOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
}

func (s *SymbolicNameOrReservedWordVisitor) EnterOC_ReservedWord(ctx *parser.OC_ReservedWordContext) {
	s.Name = ctx.GetText()
}

func (s *SymbolicNameOrReservedWordVisitor) ExitOC_ReservedWord(ctx *parser.OC_ReservedWordContext) {
}

type MapLiteralVisitor struct {
	BaseVisitor

	nextPropertyKey string
	Map             model.MapLiteral
}

func NewMapLiteralVisitor() *MapLiteralVisitor {
	return &MapLiteralVisitor{
		Map: model.MapLiteral{},
	}
}

func (s *MapLiteralVisitor) EnterOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.ctx.Enter(&SymbolicNameOrReservedWordVisitor{})
}

func (s *MapLiteralVisitor) ExitOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {
	s.nextPropertyKey = s.ctx.Exit().(*SymbolicNameOrReservedWordVisitor).Name
}

func (s *MapLiteralVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *MapLiteralVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.Map[s.nextPropertyKey] = s.ctx.Exit().(*ExpressionVisitor).Expression
}

type ListLiteralVisitor struct {
	BaseVisitor

	List *model.ListLiteral
}

func NewListLiteralVisitor() *ListLiteralVisitor {
	return &ListLiteralVisitor{
		List: model.NewListLiteral(),
	}
}

func (s *ListLiteralVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(NewExpressionVisitor())
}

func (s *ListLiteralVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	*s.List = append(*s.List, s.ctx.Exit().(*ExpressionVisitor).Expression)
}

type LiteralVisitor struct {
	BaseVisitor

	Literal *model.Literal
}

func NewLiteralVisitor() *LiteralVisitor {
	return &LiteralVisitor{
		Literal: &model.Literal{},
	}
}

func (s *LiteralVisitor) EnterOC_NumberLiteral(ctx *parser.OC_NumberLiteralContext) {
}

func (s *LiteralVisitor) ExitOC_NumberLiteral(ctx *parser.OC_NumberLiteralContext) {
}

func (s *LiteralVisitor) EnterOC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) {
	text := ctx.GetText()

	if parsedInt64, err := strconv.ParseInt(ctx.GetText(), 10, 64); err != nil {
		s.ctx.AddErrors(fmt.Errorf("invalid integer literal: %s - %w", text, err))
	} else {
		s.Literal.Set(parsedInt64)
	}
}

func (s *LiteralVisitor) ExitOC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) {
}

func (s *LiteralVisitor) EnterOC_BooleanLiteral(ctx *parser.OC_BooleanLiteralContext) {
	text := ctx.GetText()

	if parsedBool, err := strconv.ParseBool(text); err != nil {
		s.ctx.AddErrors(fmt.Errorf("invalid boolean literal: %s - %w", text, err))
	} else {
		s.Literal.Set(parsedBool)
	}
}

func (s *LiteralVisitor) ExitOC_BooleanLiteral(ctx *parser.OC_BooleanLiteralContext) {
}

func (s *LiteralVisitor) EnterOC_DoubleLiteral(ctx *parser.OC_DoubleLiteralContext) {
	text := ctx.GetText()

	if parsedFloat64, err := strconv.ParseFloat(text, 64); err != nil {
		s.ctx.AddErrors(fmt.Errorf("invalid double literal: %s - %w", text, err))
	} else {
		s.Literal.Set(parsedFloat64)
	}
}

func (s *LiteralVisitor) ExitOC_DoubleLiteral(ctx *parser.OC_DoubleLiteralContext) {
}

func (s *LiteralVisitor) EnterOC_MapLiteral(ctx *parser.OC_MapLiteralContext) {
	s.ctx.Enter(NewMapLiteralVisitor())
}

func (s *LiteralVisitor) ExitOC_MapLiteral(ctx *parser.OC_MapLiteralContext) {
	s.Literal.Set(s.ctx.Exit().(*MapLiteralVisitor).Map)
}

func (s *LiteralVisitor) EnterOC_ListLiteral(ctx *parser.OC_ListLiteralContext) {
	s.ctx.Enter(NewListLiteralVisitor())
}

func (s *LiteralVisitor) ExitOC_ListLiteral(ctx *parser.OC_ListLiteralContext) {
	s.Literal.Set(s.ctx.Exit().(*ListLiteralVisitor).List)
}
