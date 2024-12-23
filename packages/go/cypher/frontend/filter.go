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
	"github.com/specterops/bloodhound/cypher/parser"
)

// These are filter overrides to prevent cypher specified ops. Allows for customized parse filters through NewContext fn.

// TODO: Review if relying on a deny model is less secure than explicit allow
func DefaultCypherContext() *Context {
	return NewContext(
		&UpdatingClauseFilter{},
		&ExplicitProcedureInvocationFilter{},
		&ImplicitProcedureInvocationFilter{},
		&SpecifiedParametersFilter{},
	)
}

type ExplicitProcedureInvocationFilter struct {
	BaseVisitor
}

func (s *ExplicitProcedureInvocationFilter) EnterOC_ExplicitProcedureInvocation(ctx *parser.OC_ExplicitProcedureInvocationContext) {
	s.ctx.AddErrors(ErrProcedureInvocationNotSupported)
}

type ImplicitProcedureInvocationFilter struct {
	BaseVisitor
}

func (s *ImplicitProcedureInvocationFilter) EnterOC_ImplicitProcedureInvocation(ctx *parser.OC_ImplicitProcedureInvocationContext) {
	s.ctx.AddErrors(ErrProcedureInvocationNotSupported)
}

type SpecifiedParametersFilter struct {
	BaseVisitor
}

func (s *SpecifiedParametersFilter) EnterOC_Parameter(ctx *parser.OC_ParameterContext) {
	s.ctx.AddErrors(ErrUserSpecifiedParametersNotSupported)
}

type UpdatingClauseFilter struct {
	BaseVisitor
}

func (s *UpdatingClauseFilter) EnterOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {
	s.ctx.AddErrors(ErrUpdateClauseNotSupported)
}
