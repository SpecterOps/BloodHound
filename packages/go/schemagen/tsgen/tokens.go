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

package tsgen

const (
	keywordSwitch   = "switch"
	keywordCase     = "case"
	keywordDefault  = "default"
	keywordConst    = "const"
	keywordReturn   = "return"
	keywordType     = "type"
	keywordExport   = "export"
	keywordImport   = "import"
	keywordFrom     = "from"
	keywordEnum     = "enum"
	keywordFunction = "function"
	keywordLet      = "let"
	keywordThrow    = "throw"
	keywordThrows   = "throws"
	keywordNew      = "new"

	literalEmptyString       = ""
	literalNewline           = "\n"
	literalSpace             = " "
	literalOpenParentheses   = "("
	literalClosedParentheses = ")"
	literalOpenBrace         = "{"
	literalClosedBrace       = "}"
	literalOpenBracket       = "["
	literalClosedBracket     = "]"
	literalSemiColon         = ";"
	literalColon             = ":"
	literalComma             = ","
	literalPeriod            = "."
	literalPipe              = "|"
)

type GroupTokens struct {
	Open      string
	Close     string
	Separator string
}

func (s *GroupTokens) Clear() {
	s.Open = ""
	s.Close = ""
}

func (s GroupTokens) HasOpenToken() bool {
	return s.Open != ""
}

func (s GroupTokens) HasCloseToken() bool {
	return s.Close != ""
}

func (s GroupTokens) HasSeparatorToken() bool {
	return s.Separator != ""
}

func SwitchParameterGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      literalOpenParentheses,
		Close:     literalClosedParentheses,
		Separator: literalSpace,
	}
}

func CaseGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      keywordCase + literalSpace,
		Close:     literalColon,
		Separator: literalComma,
	}
}

func DefaultGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      keywordDefault,
		Close:     literalColon,
		Separator: literalComma,
	}
}

func DefinitionGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      literalOpenBrace,
		Close:     literalClosedBrace,
		Separator: literalComma,
	}
}

func QualifiedIDGroupTokens() GroupTokens {
	return GroupTokens{
		Separator: literalPeriod,
	}
}

func UnionGroupTokens() GroupTokens {
	return GroupTokens{
		Separator: literalPipe,
	}
}

func FunctionParameterGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      literalOpenParentheses,
		Close:     literalClosedParentheses,
		Separator: literalComma,
	}
}

func ArgsGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      literalOpenParentheses,
		Close:     literalClosedParentheses,
		Separator: literalComma,
	}
}

func BlockGroupTokens() GroupTokens {
	return GroupTokens{
		Open:  literalOpenBrace,
		Close: literalClosedBrace,
	}
}

func IndexGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      literalOpenBracket,
		Close:     literalClosedBracket,
		Separator: literalComma,
	}
}

func ListGroupTokens() GroupTokens {
	return GroupTokens{
		Open:      literalOpenBracket,
		Close:     literalClosedBracket,
		Separator: literalComma,
	}
}

func ReturnsGroupTokens() GroupTokens {
	return GroupTokens{
		Separator: literalComma,
	}
}
