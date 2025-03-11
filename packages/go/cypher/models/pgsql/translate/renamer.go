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

func rewriteCompositeTypeFieldReference(scopeIdentifier pgsql.Identifier, compositeReference pgsql.CompoundIdentifier) pgsql.RowColumnReference {
	return pgsql.RowColumnReference{
		Identifier: pgsql.CompoundIdentifier{scopeIdentifier, compositeReference.Root()},
		Column:     compositeReference[1],
	}
}

func rewriteIdentifierScopeReference(scope *Scope, identifier pgsql.Identifier) (pgsql.SelectItem, error) {
	if !pgsql.IsReservedIdentifier(identifier) {
		if binding, bound := scope.Lookup(identifier); bound {
			if binding.LastProjection != nil {
				return pgsql.CompoundIdentifier{binding.LastProjection.Binding.Identifier, identifier}, nil
			}
		}
	}

	// Return the original identifier if no rewrite is needed
	return identifier, nil
}

func rewriteCompoundIdentifierScopeReference(scope *Scope, identifier pgsql.CompoundIdentifier) (pgsql.SelectItem, error) {
	if binding, bound := scope.Lookup(identifier[0]); bound {
		if binding.LastProjection != nil {
			return pgsql.RowColumnReference{
				Identifier: pgsql.CompoundIdentifier{binding.LastProjection.Binding.Identifier, binding.Identifier},
				Column:     identifier[1],
			}, nil
		}
	}

	// Return the original identifier if no rewrite is needed
	return identifier, nil
}

type FrameBindingRewriter struct {
	walk.HierarchicalVisitor[pgsql.SyntaxNode]

	scope *Scope
}

func (s *FrameBindingRewriter) enter(node pgsql.SyntaxNode) error {
	switch typedExpression := node.(type) {
	case pgsql.Projection:
		for idx, projection := range typedExpression {
			switch typedProjection := projection.(type) {
			case pgsql.Identifier:
				if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedProjection); err != nil {
					return err
				} else {
					typedExpression[idx] = rewritten
				}

			case pgsql.CompoundIdentifier:
				if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedProjection); err != nil {
					return err
				} else {
					typedExpression[idx] = rewritten
				}
			}
		}

	case pgsql.CompositeValue:
		for idx, value := range typedExpression.Values {
			switch typedValue := value.(type) {
			case pgsql.Identifier:
				if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedValue); err != nil {
					return err
				} else {
					typedExpression.Values[idx] = rewritten
				}

			case pgsql.CompoundIdentifier:
				if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedValue); err != nil {
					return err
				} else {
					typedExpression.Values[idx] = rewritten
				}
			}
		}

	case pgsql.FunctionCall:
		for idx, parameter := range typedExpression.Parameters {
			switch typedParameter := parameter.(type) {
			case pgsql.Identifier:
				if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedParameter); err != nil {
					return err
				} else {
					// This is being done in case the top-level parameter is a value-type
					typedExpression.Parameters[idx] = rewritten
				}

			case pgsql.CompoundIdentifier:
				if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedParameter); err != nil {
					return err
				} else {
					// This is being done in case the top-level parameter is a value-type
					typedExpression.Parameters[idx] = rewritten
				}
			}
		}

	case *pgsql.ArrayIndex:
		switch typedArrayIndexInnerExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedArrayIndexInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedArrayIndexInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}
		}

		for idx, indexExpression := range typedExpression.Indexes {
			switch typedIndexExpression := indexExpression.(type) {
			case pgsql.Identifier:
				if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedIndexExpression); err != nil {
					return err
				} else {
					typedExpression.Indexes[idx] = rewritten
				}

			case pgsql.CompoundIdentifier:
				if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedIndexExpression); err != nil {
					return err
				} else {
					typedExpression.Indexes[idx] = rewritten
				}
			}
		}

	case *pgsql.Parenthetical:
		switch typedInnerExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}
		}

	case *pgsql.AliasedExpression:
		switch typedInnerExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}
		}

	case *pgsql.AnyExpression:
		switch typedInnerExpression := typedExpression.Expression.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedInnerExpression); err != nil {
				return err
			} else {
				typedExpression.Expression = rewritten
			}
		}

	case *pgsql.UnaryExpression:
		switch typedOperand := typedExpression.Operand.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedOperand); err != nil {
				return err
			} else {
				typedExpression.Operand = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedOperand); err != nil {
				return err
			} else {
				typedExpression.Operand = rewritten
			}
		}

	case *pgsql.BinaryExpression:
		switch typedLOperand := typedExpression.LOperand.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedLOperand); err != nil {
				return err
			} else {
				typedExpression.LOperand = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedLOperand); err != nil {
				return err
			} else {
				typedExpression.LOperand = rewritten
			}
		}

		switch typedROperand := typedExpression.ROperand.(type) {
		case pgsql.Identifier:
			if rewritten, err := rewriteIdentifierScopeReference(s.scope, typedROperand); err != nil {
				return err
			} else {
				typedExpression.ROperand = rewritten
			}

		case pgsql.CompoundIdentifier:
			if rewritten, err := rewriteCompoundIdentifierScopeReference(s.scope, typedROperand); err != nil {
				return err
			} else {
				typedExpression.ROperand = rewritten
			}
		}
	}

	return nil
}

func (s *FrameBindingRewriter) Enter(node pgsql.SyntaxNode) {
	if err := s.enter(node); err != nil {
		s.SetError(err)
	}
}

func (s *FrameBindingRewriter) exit(node pgsql.SyntaxNode) error {
	switch node.(type) {
	}

	return nil
}
func (s *FrameBindingRewriter) Exit(node pgsql.SyntaxNode) {
	if err := s.exit(node); err != nil {
		s.SetError(err)
	}
}

func RewriteFrameBindings(scope *Scope, expression pgsql.Expression) error {
	if expression == nil {
		return nil
	}

	return walk.WalkPgSQL(expression, &FrameBindingRewriter{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[pgsql.SyntaxNode](),
		scope:               scope,
	})
}
