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

// Package tsgen is a code generation package for TypeScript inspired by https://github.com/dave/jennifer
package tsgen

import (
	"io"
	"os"
)

// Node represents a syntax node. The interface provides the bare minimum required to render any typed syntax element.
type Node interface {
	Render(writer io.Writer) error
	IsNull() bool
}

// CursorHandler is a function pointer type that represents a group's context. This definition allows a user to maintain
// a Cursor's fluent interface without having to bind scopes to local variables.
type CursorHandler func(cursor Cursor)

// EmptyHandler is no-op cursor handler.
func EmptyHandler(cursor Cursor) {
}

// Cursor is an interface with fluent functions for building out a syntax tree.
type Cursor interface {
	// ID creates an identifier syntax node.
	// Example:
	// golang
	// Let().ID("myvar").OP("=").Literal(1)
	//
	// typescript
	// let myvar = 1
	ID(symbol string) Cursor

	// Let creates a "let" keyword token.
	Let() Cursor

	// Dot creates a literal "." token that is distinct from all other tokens for the purposes of specifying member
	// access, representing qualified identifiers or delimiting module names.
	Dot() Cursor

	// Qualified creates a qualified identifier that has a qualifier prefix as well as an identifier symbol.
	// Example:
	// golang
	// Qualified("myvar", "myMember").OP("=").Literal(1)
	// Qualified("this.is.a.module", "exportedFunc").Parameters()
	//
	// typescript
	// myvar.myMember = 1
	// this.is.a.module.exportedFunc()
	Qualified(namespace, symbol string) Cursor

	// Export creates an "export" keyword token.
	Export() Cursor

	// Export creates an "import" keyword token.
	Import() Cursor

	// Export creates an "from" keyword token with an attached path literal to define where the preceding import
	// statement is being referenced from.
	From(importPath string) Cursor

	// Enum creates an "enum" keyword token.
	Enum() Cursor

	// Const creates a "const" keyword token.
	Const() Cursor

	// Type creates a "type" keyword token.
	Type() Cursor

	// Function creates a "function" keyword token.
	Function() Cursor

	Throw() Cursor
	Throws() Cursor
	New() Cursor

	// Switch creates switch group.
	Switch(handler CursorHandler) Cursor

	// Case creates case statement group.
	Case(handler CursorHandler) Cursor

	// Default creates a default case statement group.
	Default(handler CursorHandler) Cursor

	// Block creates a block scoped group.
	Block(handler ...CursorHandler) Cursor

	// Parameters creates a function parameter statement group.
	Parameters(handler ...CursorHandler) Cursor

	// Args creates a function call statement group.
	Args(handler ...CursorHandler) Cursor

	// Returns creates a returns statement group.
	Returns(handler CursorHandler) Cursor

	// Index creates a list statement group for addressing array indices or representing an array type.
	// Example:
	// golang
	// ID("a").Index()
	//
	// typescript
	// a[]
	Index(handler ...CursorHandler) Cursor

	// List creates a list statement group for creation of an array.
	// Example:
	// golang
	// ID("a").List(func(cursor tsgen.Cursor) {
	//   cursor.ID("item")
	// })
	//
	// typescript
	// a[]
	List(handler ...CursorHandler) Cursor

	// Defs creates a syntax group for groups of definition statements.
	// Example:
	// golang
	// Export().Enum().Defs(func (g CursorHandler) {
	//     g.ID("EnumValue1").OP("=").Literal("EnumValue")
	// })
	//
	// typescript
	// export enum {
	//     EnumValue1 = "EnumValue"
	// }
	//
	Defs(handler CursorHandler) Cursor

	// Union creates type union syntax group.
	// Example:
	// golang
	// root.Type().ID("UnionType").OP("=").Union(func(cursor tsgen.Cursor) {
	//    cursor.ID("MyType")
	//    cursor.Literal("A String")
	// })
	//
	// typescript
	// UnionType = MyType | "A String"
	//
	Union(handler CursorHandler) Cursor

	// OP creates an operator token with the given symbol.
	// Example:
	// golang
	// ID(a).OP("=").Literal(1).OP("+").Literal(1)
	//
	// typescript
	// a = 1 + 1
	//
	OP(symbol string) Cursor

	// Literal adds a literal token for the given value.
	Literal(value any) Cursor

	// Newline adds a formatting directive that results in an additional empty line in the rendered output.
	Newline() Cursor

	Node
}

// File structs are top-level Group definitions that can be written to a file.
type File struct {
	Name string
	Path string

	*Group
}

// Write writes the File and its top-level Group to File.Path using the given flags and os.FileMode.
func (s File) Write(flag int, perm os.FileMode) error {
	if fout, err := os.OpenFile(s.Path, flag, perm); err != nil {
		return err
	} else {
		defer fout.Close()

		if err := s.Render(fout); err != nil {
			return err
		}
	}

	return nil
}

// NewFile returns a new File with the given name and path.
func NewFile(name, path string) File {
	return File{
		Name: name,
		Path: path,

		Group: &Group{
			MultiLine: true,
		},
	}
}

func ID(symbol string) CursorHandler {
	return func(cursor Cursor) {
		cursor.ID(symbol)
	}
}

func Qualified(namespace, symbol string) CursorHandler {
	return func(cursor Cursor) {
		cursor.Qualified(namespace, symbol)
	}
}

func Parameter(symbol, typeName string) CursorHandler {
	return func(cursor Cursor) {
		cursor.ID(symbol).OP(":").ID(typeName)
	}
}

func List(of ...CursorHandler) CursorHandler {
	return func(cursor Cursor) {
		cursor.List(func(cursor Cursor) {
			for _, ofHandler := range of {
				ofHandler(cursor)
			}
		})
	}
}

func Literal(value any) CursorHandler {
	return func(cursor Cursor) {
		cursor.Literal(value)
	}
}
