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
	"github.com/antlr4-go/antlr/v4"
	"github.com/specterops/bloodhound/cypher/parser"
	"strings"
)

type TokenProvider interface {
	GetToken(ttype int, i int) antlr.TerminalNode

	GetChildCount() int
	GetChild(int) antlr.Tree

	GetText() string
}

func HasTokens(ctx TokenProvider, tokens ...int) bool {
	for _, nextToken := range tokens {
		if ctx.GetToken(nextToken, 0) == nil {
			return false
		}
	}

	return true
}

// Because the grammar doesn't bind names for certain tokens this is a low-cost, low-effort way of mapping them to more
// human-readable variable names.

// findTokenRuleIndex looks up a given token literal against an instance of the generated ANTLR openCypher parser and
// returns the token's rule index. If the token literal can not be found this function returns an index of -1.
func findTokenRuleIndex(tokenLiteral string) int {
	// Force initialization to ensure that the static data is populated. The static data struct instance is protected
	// by a sync.Once instance and is therefore both idempotent and safe to call here.
	parser.CypherLexerInit()

	searchLiteralName := "'" + tokenLiteral + "'"

	for idx, literalName := range parser.CypherLexerLexerStaticData.LiteralNames {
		if literalName == searchLiteralName {
			return idx
		}
	}

	return -1
}

// TokenRuleLiteralName returns the string value of the literal name of a given token rule index.
func TokenRuleLiteralName(index int) string {
	if index < 0 || index > len(parser.CypherLexerLexerStaticData.LiteralNames) {
		return "INVALID"
	}

	return strings.Trim(parser.CypherLexerLexerStaticData.LiteralNames[index], "'")
}

var (
	// TokenType rule indexes
	TokenTypeEquals             = findTokenRuleIndex("=")
	TokenTypeAdditionAssignment = findTokenRuleIndex("+=")
	TokenTypePlusSign           = findTokenRuleIndex("+")
	TokenTypeHyphen             = findTokenRuleIndex("-")
	TokenTypeSlash              = findTokenRuleIndex("/")
	TokenTypePercentSign        = findTokenRuleIndex("%")
	TokenTypeAsterisk           = findTokenRuleIndex("*")
	TokenTypeComma              = findTokenRuleIndex(",")
	TokenTypeCaret              = findTokenRuleIndex("^")

	// TokenLiteral string values
	TokenLiteralEquals             = TokenRuleLiteralName(TokenTypeEquals)
	TokenLiteralAdditionAssignment = TokenRuleLiteralName(TokenTypeAdditionAssignment)
	TokenLiteralPlusSign           = TokenRuleLiteralName(TokenTypePlusSign)
	TokenLiteralHyphen             = TokenRuleLiteralName(TokenTypeHyphen)
	TokenLiteralSlash              = TokenRuleLiteralName(TokenTypeSlash)
	TokenLiteralPercentSign        = TokenRuleLiteralName(TokenTypePercentSign)
	TokenLiteralAsterisk           = TokenRuleLiteralName(TokenTypeAsterisk)
	TokenLiteralComma              = TokenRuleLiteralName(TokenTypeComma)
	TokenLiteralCaret              = TokenRuleLiteralName(TokenTypeCaret)

	OperatorTokensTypes = []int{
		TokenTypeEquals,
		TokenTypeAdditionAssignment,
		TokenTypePlusSign,
		TokenTypeHyphen,
		TokenTypeSlash,
		TokenTypePercentSign,
		TokenTypeAsterisk,
		TokenTypeCaret,
	}
)
