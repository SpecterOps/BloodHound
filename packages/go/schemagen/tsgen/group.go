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

import (
	"bytes"
	"io"
)

// GroupType describes a Group's type.
type GroupType string

const (
	groupTypeDefinitions        GroupType = "definitions"
	groupTypeUnion              GroupType = "union"
	groupTypeSwitchParameter    GroupType = "switch_parameter"
	groupTypeFunctionParameters GroupType = "function_parameters"
	groupTypeCase               GroupType = "case"
	groupTypeDefault            GroupType = "default"
	groupTypeFunctionReturns    GroupType = "function_returns"
	groupTypeIndex              GroupType = "index"
	groupTypeList               GroupType = "list"
	groupTypeArgs               GroupType = "args"
	groupTypeBlock              GroupType = "block"
	groupTypeStatement          GroupType = "statement"
	groupTypeQualifiedID        GroupType = "qualified_id"
)

type Group struct {
	GroupType GroupType
	Code      []Node
	Tokens    GroupTokens
	MultiLine bool
	Null      bool
	Parent    *Group
}

func (s *Group) IsNull() bool {
	if s.Null {
		// A group is Null if and only if it itself is null and all of its items are null
		for _, code := range s.Code {
			if !code.IsNull() {
				return false
			}
		}

		return true
	}

	return false
}

func (s Group) NextTo(target Node) (Node, bool) {
	for idx, code := range s.Code {
		if code == target && idx < len(s.Code)-1 {
			return s.Code[idx+1], true
		}
	}

	return nil, false
}

func (s Group) PreviousTo(target Node) (Node, bool) {
	for idx, code := range s.Code {
		if code == target && idx > 0 {
			return s.Code[idx-1], true
		}
	}

	return nil, false
}

func (s Group) ParentNextTo(target Node) (Node, bool) {
	if s.Parent != nil {
		return s.Parent.NextTo(target)
	}

	return nil, false
}

func (s Group) ParentPreviousTo(target Node) (Node, bool) {
	if s.Parent != nil {
		return s.Parent.PreviousTo(target)
	}

	return nil, false
}

func (s *Group) Render(writer io.Writer) error {
	tokens := s.Tokens

	switch s.GroupType {
	case groupTypeBlock:
		if previousCode, hasPrevious := s.ParentPreviousTo(s); hasPrevious {
			switch typedCode := previousCode.(type) {
			case *Group:
				if typedCode.GroupType == groupTypeCase {
					tokens.Clear()
				}

			case token:
				if typedCode.Value == keywordDefault {
					tokens.Clear()
				}
			}
		}
	}

	if tokens.HasOpenToken() {
		if _, err := writer.Write([]byte(tokens.Open)); err != nil {
			return err
		}
	}

	if err := s.renderItems(writer); err != nil {
		return err
	}

	if !s.IsNull() && s.MultiLine && tokens.HasCloseToken() {
		// For multi-line blocks with a closing token, we insert a new line after the last item (but
		// not if all items were null). This is to ensure that if the statement finishes with a comment,
		// the closing token is not commented out.
		separator := literalNewline

		if tokens.Separator == literalComma {
			// We also insert add trailing comma if the separator was ",".
			separator = literalComma + literalNewline
		}

		if _, err := writer.Write([]byte(separator)); err != nil {
			return err
		}
	}

	if tokens.HasCloseToken() {
		if _, err := writer.Write([]byte(tokens.Close)); err != nil {
			return err
		}
	}

	switch s.GroupType {
	case groupTypeFunctionParameters:
		// Functions with parameters require a ':' after the parameters
		if _, err := writer.Write([]byte(literalColon)); err != nil {
			return err
		}
	}

	return nil
}

func (s Group) renderItems(writer io.Writer) error {
	first := true

	for idx, code := range s.Code {
		if code == nil || code.IsNull() {
			// Null() token produces no output but also no separator. Empty() token products no output but adds a
			// separator.
			continue
		}

		// Exceptions for omitting the separator prepend
		skipSeparator := false

		if group, isGroup := code.(*Group); isGroup {
			// Index groups that are preceded by an identifier should omit their separator so that the resulting render
			// appears as: "Identifier[<values>]" instead of: "Identifier [<values>]"
			if group.GroupType == groupTypeIndex && idx > 0 {
				if token, isToken := s.Code[idx-1].(token); isToken && token.Type == tokenTypeIdentifier {
					skipSeparator = true
				}
			}
		}

		if !first && !skipSeparator && s.Tokens.HasSeparatorToken() {
			// The separator token is added before each non-null item, but not before the first item.
			if _, err := writer.Write([]byte(s.Tokens.Separator)); err != nil {
				return err
			}
		}

		if s.MultiLine {
			// For multi-line blocks, we insert a new line before each non-null item.
			if _, err := writer.Write([]byte("\n")); err != nil {
				return err
			}
		}

		if err := code.Render(writer); err != nil {
			return err
		}

		first = false
	}

	return nil
}

func (s Group) String() string {
	buffer := bytes.Buffer{}

	if err := s.Render(&buffer); err != nil {
		panic(err)
	}

	return buffer.String()
}

func (s *Group) statement() *Group {
	if s.GroupType == groupTypeStatement {
		return s
	}

	newStatement := &Group{
		GroupType: groupTypeStatement,
		Tokens: GroupTokens{
			Separator: " ",
		},
	}

	s.pushGroup(EmptyHandler, newStatement)
	return newStatement
}

func (s *Group) pushToken(tokenType tokenType, value any) *Group {
	statement := s.statement()
	statement.Code = append(statement.Code, token{
		Type:  tokenType,
		Value: value,
	})

	return statement
}

func (s *Group) pushGroup(handler CursorHandler, group *Group) {
	// Allow the user to finish realizing the tree
	if handler != nil {
		handler(group)
	}

	// Bind the sub-group to us for reverse lookups
	group.Parent = s

	s.Code = append(s.Code, group)
}

func (s *Group) runHandlers(handlers []CursorHandler, groupType GroupType, groupTokens GroupTokens, multiline bool) {
	if len(handlers) > 0 {
		for _, handler := range handlers {
			s.pushGroup(handler, &Group{
				GroupType: groupType,
				Tokens:    groupTokens,
				MultiLine: multiline,
			})
		}
	} else {
		s.pushGroup(nil, &Group{
			GroupType: groupType,
			Tokens:    groupTokens,
			MultiLine: multiline,
		})
	}
}

func (s *Group) Throw() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordThrow)
}

func (s *Group) Throws() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordThrows)
}

func (s *Group) New() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordNew)
}

func (s *Group) Qualified(namespace, symbol string) Cursor {
	statement := s.statement()
	statement.pushGroup(EmptyHandler, &Group{
		GroupType: groupTypeQualifiedID,
		Code: []Node{
			token{
				Type:  tokenTypeQualifier,
				Value: namespace,
			},
			token{
				Type:  tokenTypeIdentifier,
				Value: symbol,
			},
		},
		Tokens: QualifiedIDGroupTokens(),
	})

	return statement
}

func (s *Group) Let() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordLet)
}

func (s *Group) Dot() Cursor {
	return s.pushToken(tokenTypeAccessor, literalPeriod)
}

func (s *Group) Switch(handler CursorHandler) Cursor {
	statement := s.statement()
	statement.pushToken(tokenTypeKeyword, keywordSwitch)
	statement.pushGroup(handler, &Group{
		GroupType: groupTypeSwitchParameter,
		Tokens:    SwitchParameterGroupTokens(),
	})

	return statement
}

func (s *Group) Case(handler CursorHandler) Cursor {
	statement := s.statement()
	statement.pushGroup(handler, &Group{
		GroupType: groupTypeCase,
		Tokens:    CaseGroupTokens(),
	})

	return statement
}

func (s *Group) Default(handler CursorHandler) Cursor {
	statement := s.statement()
	statement.pushGroup(handler, &Group{
		GroupType: groupTypeDefault,
		Tokens:    DefaultGroupTokens(),
	})

	return statement
}

func (s *Group) ID(value string) Cursor {
	return s.pushToken(tokenTypeIdentifier, value)
}

func (s *Group) Export() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordExport)
}

func (s *Group) Import() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordImport)
}

func (s *Group) From(importPath string) Cursor {
	return s.pushToken(tokenTypeKeyword, keywordFrom).pushToken(tokenTypeLiteral, importPath)
}

func (s *Group) Enum() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordEnum)
}

func (s *Group) Defs(handler CursorHandler) Cursor {
	s.pushGroup(handler, &Group{
		GroupType: groupTypeDefinitions,
		Tokens:    DefinitionGroupTokens(),
		MultiLine: true,
	})

	return s
}

func (s *Group) Union(handler CursorHandler) Cursor {
	s.pushGroup(handler, &Group{
		GroupType: groupTypeUnion,
		Tokens:    UnionGroupTokens(),
	})

	return s
}

func (s *Group) Function() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordFunction)
}

func (s *Group) Block(handlers ...CursorHandler) Cursor {
	s.runHandlers(handlers, groupTypeBlock, BlockGroupTokens(), true)
	return s
}

func (s *Group) Args(handlers ...CursorHandler) Cursor {
	s.runHandlers(handlers, groupTypeArgs, ArgsGroupTokens(), false)
	return s
}

func (s *Group) Parameters(handlers ...CursorHandler) Cursor {
	s.runHandlers(handlers, groupTypeFunctionParameters, FunctionParameterGroupTokens(), false)
	return s
}

func (s *Group) Returns(handlers CursorHandler) Cursor {
	statement := s.pushToken(tokenTypeKeyword, keywordReturn)
	statement.pushGroup(handlers, &Group{
		GroupType: groupTypeFunctionReturns,
		Tokens:    ReturnsGroupTokens(),
	})

	return s
}

func (s *Group) Index(handlers ...CursorHandler) Cursor {
	s.runHandlers(handlers, groupTypeIndex, IndexGroupTokens(), false)
	return s
}

func (s *Group) List(handlers ...CursorHandler) Cursor {
	s.runHandlers(handlers, groupTypeList, ListGroupTokens(), false)
	return s
}

func (s *Group) Const() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordConst)
}

func (s *Group) Type() Cursor {
	return s.pushToken(tokenTypeKeyword, keywordType)
}

func (s *Group) OP(value string) Cursor {
	return s.pushToken(tokenTypeOperator, value)
}

func (s *Group) Literal(value any) Cursor {
	return s.pushToken(tokenTypeLiteral, value)
}

func (s *Group) Newline() Cursor {
	return s.pushToken(tokenTypeFormatting, literalEmptyString)
}
