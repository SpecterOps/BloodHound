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

package pgsql

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/cypher/model/walk"
)

func rewrite(stack *walk.WalkStack, original, rewritten cypher.Expression) error {
	switch typedTrunk := stack.Trunk().(type) {
	case cypher.ExpressionList:
		for idx, expression := range typedTrunk.GetAll() {
			if expression == original {
				typedTrunk.Replace(idx, rewritten)
			}
		}

	case *cypher.FunctionInvocation:
		for idx, expression := range typedTrunk.Arguments {
			if expression == original {
				typedTrunk.Arguments[idx] = rewritten
			}
		}

	case *cypher.ProjectionItem:
		typedTrunk.Expression = rewritten

	case *cypher.SetItem:
		if typedTrunk.Right == original {
			typedTrunk.Right = rewritten
		} else if typedTrunk.Left == original {
			typedTrunk.Left = rewritten
		} else {
			return fmt.Errorf("unable to match original expression against SetItem left and right operands")
		}

	case *cypher.PartialComparison:
		typedTrunk.Right = rewritten

	case *cypher.RemoveItem:
		switch typedRewritten := rewritten.(type) {
		case *cypher.KindMatcher:
			typedTrunk.KindMatcher = typedRewritten
		}

	case *cypher.Projection:
		for idx, projectionItem := range typedTrunk.Items {
			if projectionItem == original {
				typedTrunk.Items[idx] = rewritten
			}
		}

	case *cypher.Negation:
		typedTrunk.Expression = rewritten

	case *cypher.Comparison:
		if typedTrunk.Left == original {
			typedTrunk.Left = rewritten
		}

	case *cypher.Parenthetical:
		if typedTrunk.Expression == original {
			typedTrunk.Expression = rewritten
		}

	default:
		return fmt.Errorf("unable to replace expression for trunk type %T", stack.Trunk())
	}

	return nil
}
