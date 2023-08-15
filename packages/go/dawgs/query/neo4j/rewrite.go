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

package neo4j

import (
	"fmt"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/query"
)

func RemoveEmptyExpressionLists(stack *model.WalkStack, element model.Expression) error {
	var (
		shouldRemove  = false
		shouldReplace = false

		replacementExpression model.Expression
	)

	switch typedElement := element.(type) {
	case model.ExpressionList:
		shouldRemove = typedElement.Len() == 0

	case *model.Parenthetical:
		switch typedParentheticalElement := typedElement.Expression.(type) {
		case model.ExpressionList:
			numExpressions := typedParentheticalElement.Len()

			shouldRemove = numExpressions == 0
			shouldReplace = numExpressions == 1

			if shouldReplace {
				// Dump the parenthetical and the joined expression by grabbing the only element in the joined
				// expression for replacement
				replacementExpression = typedParentheticalElement.Get(0)
			}
		}
	}

	if shouldRemove {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Remove(element)
		}
	} else if shouldReplace {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Replace(typedParent.IndexOf(element), replacementExpression)
		}
	}

	return nil
}

func StringNegationRewriter(stack *model.WalkStack, element model.Expression) error {
	var rewritten any

	switch negation := element.(type) {
	case *model.Negation:
		// If this is a negation then we should check to see if it's a comparison
		switch comparison := negation.Expression.(type) {
		case *model.Comparison:
			firstPartial := comparison.FirstPartial()

			// If the negated expression is a comparison check to see if it's a string comparison. This is done since
			// Neo4j comparison semantics for strings regarding `null` has edge cases that must be accounted for
			switch firstPartial.Operator {
			case model.OperatorStartsWith, model.OperatorEndsWith, model.OperatorContains:
				// Rewrite this comparison is a disjunction of the negation and a follow-on comparison to handle null
				// checks
				rewritten = &model.Parenthetical{
					Expression: model.NewDisjunction(
						negation,
						model.NewComparison(comparison.Left, model.OperatorIs, query.Literal(nil)),
					),
				}
			}
		}
	}

	// If we rewrote this element, replace it
	if rewritten != nil {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			for idx, expression := range typedParent.GetAll() {
				if expression == element {
					typedParent.Replace(idx, rewritten)
					break
				}
			}

		default:
			return fmt.Errorf("unable to replace rewritten string negation operation for parent type %T", stack.Trunk())
		}
	}

	return nil
}
