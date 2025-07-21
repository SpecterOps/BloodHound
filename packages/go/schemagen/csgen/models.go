// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package csgen

import (
	"fmt"
	"strings"
)

type Expression interface{}
type Symbol string

func (s Symbol) NodeType() string {
	return "symbol"
}

func (s Symbol) Children() []SyntaxNode {
	return []SyntaxNode{}
}

type Visibility int

const (
	VisibilityPrivate   Visibility = 1
	VisibilityPublic    Visibility = 2
	VisibilityInternal  Visibility = 3
	VisibilityProtected Visibility = 4
)

func (s Visibility) String() string {
	switch s {
	case VisibilityPrivate:
		return "private"
	case VisibilityPublic:
		return "public"
	case VisibilityInternal:
		return "internal"
	case VisibilityProtected:
		return "protected"
	}
	return ""
}

func (s Visibility) AsExpression() Expression {
	return s
}

func (s Visibility) NodeType() string {
	return "visibility"
}

func (s Visibility) Children() []SyntaxNode {
	return []SyntaxNode{}
}

type Operator string

func (s Operator) Children() []SyntaxNode {
	return []SyntaxNode{}
}

func (s Operator) NodeType() string {
	return "operator"
}

func (s Operator) String() string {
	return string(s)
}

const (
	OperatorEquals Operator = "="
	OperatorPlus   Operator = "+"
	OperatorMinus  Operator = "-"
)

type Type int

const (
	TypeString Type = 1
	TypeInt    Type = 2
	TypeFloat  Type = 3
)

func (s Type) String() string {
	switch s {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	}

	return ""
}

func (s Type) Children() []SyntaxNode {
	return []SyntaxNode{}
}

func (s Type) NodeType() string {
	return "type"
}

type Modifier int

const (
	ModifierStatic Modifier = 1
	ModifierConst  Modifier = 2
)

func (s Modifier) String() string {
	switch s {
	case ModifierStatic:
		return "static"
	case ModifierConst:
		return "const"
	}
	return ""
}

func (s Modifier) Children() []SyntaxNode {
	return []SyntaxNode{}
}

func (s Modifier) NodeType() string {
	return "modifier"
}

type Modifiers []Modifier

func (s Modifiers) String() string {
	builder := &strings.Builder{}
	for _, modifier := range s {
		builder.WriteString(modifier.String())
		builder.WriteString(" ")
	}

	return strings.TrimSpace(builder.String())
}

func (s Modifiers) NodeType() string {
	return "modifiers"
}

func (s Modifiers) Children() []SyntaxNode {
	return []SyntaxNode{}
}

type BinaryExpression struct {
	LeftOperand  SyntaxNode
	Operator     Operator
	RightOperand SyntaxNode
}

func (s BinaryExpression) NodeType() string {
	return "binary_expression"
}

func (s BinaryExpression) Children() []SyntaxNode {
	return []SyntaxNode{}
}

type ClassMemberAssignment struct {
	Visibility Visibility
	Modifier   Modifiers
	Type       Type
	Symbol     Symbol
}

func (s ClassMemberAssignment) NodeType() string {
	return "class_member_assignment"
}

func (s ClassMemberAssignment) Children() []SyntaxNode {
	return []SyntaxNode{}
}

type Literal struct {
	Value any
	Null  bool
}

func (s Literal) NodeType() string {
	return "literal"
}

func (s Literal) Children() []SyntaxNode {
	return []SyntaxNode{}
}

type FormattingLiteral string

func (s FormattingLiteral) AsExpression() Expression {
	return s
}

func (s FormattingLiteral) NodeType() string {
	return "formatting_literal"
}

func (s FormattingLiteral) String() string {
	return string(s)
}

func (s FormattingLiteral) Children() []SyntaxNode {
	return []SyntaxNode{}
}

const (
	FormattingLiteralSpace             FormattingLiteral = " "
	FormattingLiteralSemicolon         FormattingLiteral = ";"
	FormattingLiteralOpenCurlyBracket  FormattingLiteral = "{"
	FormattingLiteralCloseCurlyBracket FormattingLiteral = "}"
	FormattingLiteralNewline           FormattingLiteral = "\n"
)

// SyntaxNode is the top-level interface for any of the modeled C# AST syntax elements.
type SyntaxNode interface {
	NodeType() string
	Children() []SyntaxNode
}

type Namespace struct {
	Name  string
	Nodes []SyntaxNode
}

func (s Namespace) NodeType() string {
	return "namespace"
}

func (s Namespace) Children() []SyntaxNode {
	return s.Nodes
}

func (s Namespace) Enter() string {
	return fmt.Sprintf("namespace %s {\n", s.Name)
}

func (s Namespace) Exit() string {
	return "\n}"
}

type Class struct {
	Modifiers  Modifiers
	Visibility Visibility
	Name       string
	Nodes      []SyntaxNode
}

func (s Class) NodeType() string {
	return "class"
}

func (s Class) Children() []SyntaxNode {
	return s.Nodes
}

func (s Class) Enter() string {
	return fmt.Sprintf("%s %s class %s {\n", s.Visibility.String(), s.Modifiers.String(), s.Name)
}

func (s Class) Exit() string {
	return "\n}"
}
