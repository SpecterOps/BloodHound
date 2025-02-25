// Copyright 2024 Specter Ops, Inc.
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

package pgsql

import "fmt"

// SyntaxNode is the top-level interface for any of the modeled PgSQL AST syntax elements.
type SyntaxNode interface {
	NodeType() string
}

// WrappedNode is an interface that allows for an expression to be wrapped to allow for side effects like futures.
type WrappedNode interface {
	SyntaxNode
}

// SyntaxNodeFuture is a SyntaxNode that can be satisfied out-of-tree and communicate if it has been satisfied to
// upstream consumers. This is useful when a SyntaxNode is required to be embedded in a built syntax tree but
// can not be formally built itself due to ordering or missing dependency information (in the case of an
// existential subquery).
type SyntaxNodeFuture interface {
	SyntaxNode
	Unwrap() SyntaxNode
	Satisfied() bool
}

// Statement is a syntax node that does not evaluate to a value.
type Statement interface {
	SyntaxNode
	AsStatement() Statement
}

// Assignment is an expression that can be used as an assignment.
type Assignment interface {
	SyntaxNode
	AsAssignment() Assignment
}

// Expression one or more syntax nodes that evaluate to a value.
type Expression interface {
	SyntaxNode
	AsExpression() Expression
}

// TypeHinted is an expression that contains a DataType hint that can be accessed by handlers.
type TypeHinted interface {
	Expression
	TypeHint() DataType
}

// SelectItem is an expression that can be used as a projection.
type SelectItem interface {
	Expression
	AsSelectItem() SelectItem
}

// MergeAction is an expression that can be used as the action of a merge statement.
type MergeAction interface {
	Expression
	AsMergeAction() MergeAction
}

// SetExpression is an expression that represents a query body expression.
type SetExpression interface {
	Expression
	AsSetExpression() SetExpression
}

// ConflictAction is an expression that is evaluated as part of an OnConflict expression.
type ConflictAction interface {
	Expression
	AsConflictAction() ConflictAction
}

// As is a helper function that takes a SyntaxNode and attempts to cast it to the type represented by T. If the
// node passed is equal to nil then the empty value of T is returned. If the typed node does not convert to the
// type represented by T then an error is returned.
func As[T any](node SyntaxNode) (T, error) {
	var empty T

	if node == nil {
		return empty, nil
	}

	if typedNode, isT := node.(T); isT {
		return typedNode, nil
	}

	return empty, fmt.Errorf("node type %T does not convert to expected type %T", node, empty)
}
