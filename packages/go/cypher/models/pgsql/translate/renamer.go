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

package translate

import (
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/walk"
)

func rewriteCompositeTypeReference(scopeIdentifier pgsql.Identifier, compositeReference pgsql.Identifier) pgsql.CompoundIdentifier {
	return pgsql.CompoundIdentifier{scopeIdentifier, compositeReference}
}

func rewriteCompositeTypeReferenceRoot(scopeIdentifier pgsql.Identifier, compositeReference pgsql.CompoundIdentifier) pgsql.CompoundIdentifier {
	rewritten := make(pgsql.CompoundIdentifier, 0, len(compositeReference))
	rewritten = append(rewritten, scopeIdentifier)

	return append(rewritten, compositeReference[1:]...)
}

func rewriteCompositeTypeFieldReference(scopeIdentifier pgsql.Identifier, compositeReference pgsql.CompoundIdentifier) pgsql.CompoundExpression {
	return pgsql.CompoundExpression{
		&pgsql.Parenthetical{
			Expression: pgsql.CompoundIdentifier{scopeIdentifier, compositeReference.Root()},
		},

		compositeReference[1:],
	}
}

// IdentifierRewriter is a PgSQL AST visitor that finds any matching identifier within the given target
// identifier set. Once a match is found, the matching identifiers are rewritten as a compound
// identifier leading with the current frame's identifier (scopeIdentifier) such that an identifier
// 'a' with scopeIdentifier 's0' becomes pgsql.CompoundIdentifier{scopeIdentifier, 'a'}.
//
// For field references of an identifier represented as compound identifiers, the compound identifier
// is wrapped in a parenthetical:
//
// // a.properties with scopeIdentifier 's0' becomes -> (s0.a).properties
//
//	pgsql.CompoundExpression{
//			&pgsql.Parenthetical{
//				Expression: pgsql.CompoundIdentifier{scopeIdentifier, fieldReference.Root()},
//			},
//
//			fieldReference[1:],
//		}
type IdentifierRewriter struct {
	walk.HierarchicalVisitor[pgsql.SyntaxNode]

	scopeIdentifier pgsql.Identifier
	targets         *pgsql.IdentifierSet
	stack           []pgsql.SyntaxNode
}

func (s *IdentifierRewriter) enter(node pgsql.SyntaxNode) error {
	switch typedExpression := node.(type) {
	case pgsql.Projection:
		for idx, projection := range typedExpression {
			switch typedProjection := projection.(type) {
			case pgsql.Identifier:
				if s.targets == nil || s.targets.Contains(typedProjection) {
					typedExpression[idx] = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, pgsql.CompoundIdentifier{typedProjection})
				}

			case pgsql.CompoundIdentifier:
				if s.targets == nil || s.targets.Contains(typedProjection.Root()) {
					typedExpression[idx] = rewriteCompositeTypeFieldReference(s.scopeIdentifier, typedProjection)
				}
			}
		}

	case pgsql.CompositeValue:
		for idx, value := range typedExpression.Values {
			switch typedValue := value.(type) {
			case pgsql.Identifier:
				if s.targets == nil || s.targets.Contains(typedValue) {
					typedExpression.Values[idx] = rewriteCompositeTypeReference(s.scopeIdentifier, typedValue)
				}

			case pgsql.CompoundIdentifier:
				if s.targets == nil || s.targets.Contains(typedValue.Root()) {
					typedExpression.Values[idx] = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, typedValue)
				}
			}
		}

	case pgsql.FunctionCall:
		for idx, parameter := range typedExpression.Parameters {
			switch typedParameter := parameter.(type) {
			case pgsql.Identifier:
				if s.targets == nil || s.targets.Contains(typedParameter) {
					typedExpression.Parameters[idx] = rewriteCompositeTypeReference(s.scopeIdentifier, typedParameter)
				}

			case pgsql.CompoundIdentifier:
				if s.targets == nil || s.targets.Contains(typedParameter.Root()) {
					typedExpression.Parameters[idx] = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, typedParameter)
				}
			}
		}

	case *pgsql.ArrayIndex:
		switch typedArrayIndexExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if s.targets == nil || s.targets.Contains(typedArrayIndexExpression) {
				typedExpression.Expression = rewriteCompositeTypeReference(s.scopeIdentifier, typedArrayIndexExpression)
			}

		case pgsql.CompoundIdentifier:
			if s.targets == nil || s.targets.Contains(typedArrayIndexExpression.Root()) {
				typedExpression.Expression = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, typedArrayIndexExpression)
			}
		}

		for idx, indexExpression := range typedExpression.Indexes {
			switch typedIndexExpression := indexExpression.(type) {
			case pgsql.Identifier:
				if s.targets == nil || s.targets.Contains(typedIndexExpression) {
					typedExpression.Indexes[idx] = rewriteCompositeTypeReference(s.scopeIdentifier, typedIndexExpression)
				}

			case pgsql.CompoundIdentifier:
				if s.targets == nil || s.targets.Contains(typedIndexExpression.Root()) {
					typedExpression.Indexes[idx] = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, typedIndexExpression)
				}
			}
		}

	case *pgsql.Parenthetical:
		switch typedParentheticalExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if s.targets == nil || s.targets.Contains(typedParentheticalExpression) {
				typedExpression.Expression = rewriteCompositeTypeReference(s.scopeIdentifier, typedParentheticalExpression)
			}

		case pgsql.CompoundIdentifier:
			if s.targets == nil || s.targets.Contains(typedParentheticalExpression.Root()) {
				typedExpression.Expression = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, typedParentheticalExpression)
			}
		}

	case *pgsql.AliasedExpression:
		switch typedAliasedExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if s.targets == nil || s.targets.Contains(typedAliasedExpression) {
				typedExpression.Expression = rewriteCompositeTypeReference(s.scopeIdentifier, typedAliasedExpression)
			}

		case pgsql.CompoundIdentifier:
			if s.targets == nil || s.targets.Contains(typedAliasedExpression.Root()) {
				typedExpression.Expression = rewriteCompositeTypeReferenceRoot(s.scopeIdentifier, typedAliasedExpression)
			}
		}

	case *pgsql.BinaryExpression:
		switch typedLOperand := typedExpression.LOperand.(type) {
		case pgsql.Identifier:
			if s.targets == nil || s.targets.Contains(typedLOperand) {
				typedExpression.LOperand = rewriteCompositeTypeReference(s.scopeIdentifier, typedLOperand)
			}

		case pgsql.CompoundIdentifier:
			if s.targets == nil || s.targets.Contains(typedLOperand.Root()) {
				typedExpression.LOperand = rewriteCompositeTypeFieldReference(s.scopeIdentifier, typedLOperand)
			}
		}

		switch typedROperand := typedExpression.ROperand.(type) {
		case pgsql.Identifier:
			if s.targets == nil || s.targets.Contains(typedROperand) {
				typedExpression.ROperand = rewriteCompositeTypeReference(s.scopeIdentifier, typedROperand)
			}

		case pgsql.CompoundIdentifier:
			if s.targets == nil || s.targets.Contains(typedROperand.Root()) {
				typedExpression.ROperand = rewriteCompositeTypeFieldReference(s.scopeIdentifier, typedROperand)
			}
		}
	}

	s.stack = append(s.stack, node)
	return nil
}

func (s *IdentifierRewriter) Enter(node pgsql.SyntaxNode) {
	if err := s.enter(node); err != nil {
		s.SetError(err)
	}

	s.stack = append(s.stack, node)
}

func (s *IdentifierRewriter) exit(_ pgsql.SyntaxNode) error {
	s.stack = s.stack[:len(s.stack)-1]
	return nil
}

func (s *IdentifierRewriter) Exit(node pgsql.SyntaxNode) {
	if err := s.exit(node); err != nil {
		s.SetError(err)
	}
}

func NewIdentifierRewriter(scopeIdentifier pgsql.Identifier, targets *pgsql.IdentifierSet) walk.HierarchicalVisitor[pgsql.SyntaxNode] {
	return &IdentifierRewriter{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[pgsql.SyntaxNode](),
		scopeIdentifier:     scopeIdentifier,
		targets:             targets,
	}
}

func RewriteExpressionIdentifiers(expression pgsql.Expression, scopeIdentifier pgsql.Identifier, targets *pgsql.IdentifierSet) error {
	if expression == nil {
		return nil
	}

	return walk.WalkPgSQL(expression, NewIdentifierRewriter(scopeIdentifier, targets))
}

func RewriteExpressionCompoundIdentifier(expression pgsql.Expression, scopeIdentifier pgsql.Identifier, root pgsql.Identifier) error {
	return RewriteExpressionIdentifiers(expression, scopeIdentifier, pgsql.AsIdentifierSet(root))
}
