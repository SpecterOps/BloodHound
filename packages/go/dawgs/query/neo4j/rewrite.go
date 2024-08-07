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

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/dawgs/query"
)

func RemoveEmptyExpressionLists(stack *cypher.WalkStack, element cypher.Expression) error {
	var (
		shouldRemove  = false
		shouldReplace = false

		replacementExpression cypher.Expression
	)

	switch typedElement := element.(type) {
	case cypher.ExpressionList:
		shouldRemove = typedElement.Len() == 0

	case *cypher.Parenthetical:
		switch typedParentheticalElement := typedElement.Expression.(type) {
		case cypher.ExpressionList:
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
		case cypher.ExpressionList:
			typedParent.Remove(element)
		}
	} else if shouldReplace {
		switch typedParent := stack.Trunk().(type) {
		case cypher.ExpressionList:
			typedParent.Replace(typedParent.IndexOf(element), replacementExpression)
		}
	}

	return nil
}

func StringNegationRewriter(stack *cypher.WalkStack, element cypher.Expression) error {
	var rewritten any

	switch negation := element.(type) {
	case *cypher.Negation:
		// If this is a negation then we should check to see if it's a comparison
		for cursor := negation.Expression; element != nil; {
			switch typedCursor := cursor.(type) {
			case *cypher.Parenthetical:
				cursor = typedCursor.Expression
				continue

			case *cypher.Comparison:
				firstPartial := typedCursor.FirstPartial()

				// If the negated expression is a comparison check to see if it's a string comparison. This is done since
				// Neo4j comparison semantics for strings regarding `null` has edge cases that must be accounted for
				switch firstPartial.Operator {
				case cypher.OperatorStartsWith, cypher.OperatorEndsWith, cypher.OperatorContains:
					// Rewrite this comparison is a disjunction of the negation and a follow-on comparison to handle null
					// checks
					rewritten = &cypher.Parenthetical{
						Expression: cypher.NewDisjunction(
							negation,
							cypher.NewComparison(typedCursor.Left, cypher.OperatorIs, query.Literal(nil)),
						),
					}
				}
			}

			break
		}
	}

	// If we rewrote this element, replace it
	if rewritten != nil {
		switch typedParent := stack.Trunk().(type) {
		case cypher.ExpressionList:
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
