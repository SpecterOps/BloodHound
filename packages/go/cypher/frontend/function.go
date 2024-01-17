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

type NamespaceVisitor struct {
	BaseVisitor

	Namespace []string
}

func (s *NamespaceVisitor) EnterOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
	s.Namespace = append(s.Namespace, ctx.GetText())
}

func (s *NamespaceVisitor) ExitOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
}

type FunctionInvocationVisitor struct {
	BaseVisitor

	FunctionInvocation *model.FunctionInvocation
}

func NewFunctionInvocationVisitor(ctx *parser.OC_FunctionInvocationContext) *FunctionInvocationVisitor {
	return &FunctionInvocationVisitor{
		FunctionInvocation: &model.FunctionInvocation{
			Distinct: HasTokens(ctx, parser.CypherLexerDISTINCT),
		},
	}
}

func (s *FunctionInvocationVisitor) EnterOC_FunctionName(ctx *parser.OC_FunctionNameContext) {
}

func (s *FunctionInvocationVisitor) ExitOC_FunctionName(ctx *parser.OC_FunctionNameContext) {
}

func (s *FunctionInvocationVisitor) EnterOC_Namespace(ctx *parser.OC_NamespaceContext) {
	s.ctx.Enter(&NamespaceVisitor{})
}

func (s *FunctionInvocationVisitor) ExitOC_Namespace(ctx *parser.OC_NamespaceContext) {
	s.FunctionInvocation.Namespace = s.ctx.Exit().(*NamespaceVisitor).Namespace
}

func (s *FunctionInvocationVisitor) EnterOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
	s.FunctionInvocation.Name = ctx.GetText()
}

func (s *FunctionInvocationVisitor) ExitOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {
}

func (s *FunctionInvocationVisitor) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {
	s.ctx.Enter(&ExpressionVisitor{})
}

func (s *FunctionInvocationVisitor) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {
	result := s.ctx.Exit().(*ExpressionVisitor).Expression
	s.FunctionInvocation.Arguments = append(s.FunctionInvocation.Arguments, result)
}
