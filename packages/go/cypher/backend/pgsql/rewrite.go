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
	"github.com/specterops/bloodhound/cypher/model"
)

func rewrite(stack *model.WalkStack, original, rewritten model.Expression) error {
	switch typedTrunk := stack.Trunk().(type) {
	case model.ExpressionList:
		for idx, expression := range typedTrunk.GetAll() {
			if expression == original {
				typedTrunk.Replace(idx, rewritten)
			}
		}

	case *model.FunctionInvocation:
		for idx, expression := range typedTrunk.Arguments {
			if expression == original {
				typedTrunk.Arguments[idx] = rewritten
			}
		}
		
	case *model.ProjectionItem:
		typedTrunk.Expression = rewritten

	case *model.SetItem:
		if typedTrunk.Right == original {
			typedTrunk.Right = rewritten
		} else if typedTrunk.Left == original {
			typedTrunk.Left = rewritten
		} else {
			return fmt.Errorf("unable to match original expression against SetItem left and right operands")
		}

	case *model.PartialComparison:
		typedTrunk.Right = rewritten

	case *model.RemoveItem:
		switch typedRewritten := rewritten.(type) {
		case *model.KindMatcher:
			typedTrunk.KindMatcher = typedRewritten
		}

	case *model.Projection:
		for idx, projectionItem := range typedTrunk.Items {
			if projectionItem == original {
				typedTrunk.Items[idx] = rewritten
			}
		}

	case *model.Negation:
		typedTrunk.Expression = rewritten

	case *model.Comparison:
		if typedTrunk.Left == original {
			typedTrunk.Left = rewritten
		}

	case *model.Parenthetical:
		if typedTrunk.Expression == original {
			typedTrunk.Expression = rewritten
		}

	default:
		return fmt.Errorf("unable to replace expression for trunk type %T", stack.Trunk())
	}

	return nil
}
