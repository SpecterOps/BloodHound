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
	"bytes"
	"errors"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/parser"
)

func DefaultCypherContext() *Context {
	return NewContext(
		NewUnsupportedOperationFilter(),
	)
}

func parseCypher(ctx *Context, input string) (*model.RegularQuery, error) {
	var (
		queryBuffer     = bytes.NewBufferString(input)
		lexer           = parser.NewCypherLexer(antlr.NewIoStream(queryBuffer))
		tokenStream     = antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		parserInst      = parser.NewCypherParser(tokenStream)
		parseTreeWalker = antlr.NewParseTreeWalker()
		queryVisitor    = &QueryVisitor{}
	)
	
	// Set up the lexer and parser to report errors to the context
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(ctx)

	parserInst.RemoveErrorListeners()
	parserInst.AddErrorListener(ctx)

	// Prime the context with the root level visitor
	ctx.Enter(queryVisitor)

	// Hand off to ANTLR
	parseTreeWalker.Walk(ctx, parserInst.OC_Cypher())

	// Collect errors
	return queryVisitor.Query, errors.Join(ctx.Errors...)
}

func ParseCypher(ctx *Context, input string) (*model.RegularQuery, error) {
	if formattedInput := strings.TrimSpace(input); len(formattedInput) == 0 {
		return nil, ErrInvalidInput
	} else {
		return parseCypher(ctx, formattedInput)
	}
}
