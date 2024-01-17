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

package model

import (
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type Visitor func(parent, node any) error

func walkList[T any](enter, exit Visitor, parent any, nodeList []T) error {
	for i := 0; i < len(nodeList); i++ {
		if err := walkNodes(enter, exit, parent, nodeList[i]); err != nil {
			return err
		}
	}

	return nil
}

func walkNodes(enter, exit Visitor, parent any, nodes ...any) error {
	for _, node := range nodes {
		if enter != nil {
			if err := enter(parent, node); err != nil {
				return err
			}
		}

		switch typedNode := node.(type) {
		case ExpressionList:
			for idx := 0; idx < typedNode.Len(); idx++ {
				expression := typedNode.Get(idx)

				if err := walkNodes(enter, exit, node, expression); err != nil {
					return err
				}
			}

		case *RegularQuery:
			if err := walkNodes(enter, exit, node, typedNode.SingleQuery); err != nil {
				return err
			}

		case *SingleQuery:
			if typedNode.SinglePartQuery != nil {
				if err := walkNodes(enter, exit, node, typedNode.SinglePartQuery); err != nil {
					return err
				}
			} else if typedNode.MultiPartQuery != nil {
				if err := walkNodes(enter, exit, node, typedNode.MultiPartQuery); err != nil {
					return err
				}
			}

		case *MultiPartQuery:
			if err := walkList(enter, exit, typedNode, typedNode.Parts); err != nil {
				return err
			}

			if err := walkNodes(enter, exit, typedNode, typedNode.SinglePartQuery); err != nil {
				return err
			}

		case *MultiPartQueryPart:
			if err := walkList(enter, exit, typedNode, typedNode.ReadingClauses); err != nil {
				return err
			}

			if err := walkList(enter, exit, typedNode, typedNode.UpdatingClauses); err != nil {
				return err
			}

			if typedNode.With != nil {
				if err := walkNodes(enter, exit, typedNode, typedNode.With); err != nil {
					return err
				}
			}

		case *Quantifier:
			if err := walkNodes(enter, exit, typedNode, typedNode.Filter); err != nil {
				return err
			}

		case *FilterExpression:
			if err := walkNodes(enter, exit, typedNode, typedNode.Specifier); err != nil {
				return err
			}

			if typedNode.Where != nil {
				if err := walkNodes(enter, exit, typedNode, typedNode.Where); err != nil {
					return err
				}
			}

		case *IDInCollection:
			if err := walkNodes(enter, exit, typedNode, typedNode.Variable); err != nil {
				return err
			}

			if err := walkNodes(enter, exit, typedNode, typedNode.Expression); err != nil {
				return err
			}

		case *With:
			if err := walkNodes(enter, exit, node, typedNode.Projection); err != nil {
				return err
			}

			if typedNode.Where != nil {
				if err := walkNodes(enter, exit, node, typedNode.Where); err != nil {
					return err
				}
			}

		case *Unwind:
			if err := walkNodes(enter, exit, node, typedNode.Expression); err != nil {
				return err
			}

		case *ReadingClause:
			if typedNode.Match != nil {
				if err := walkNodes(enter, exit, node, typedNode.Match); err != nil {
					return err
				}
			}

			if typedNode.Unwind != nil {
				if err := walkNodes(enter, exit, node, typedNode.Unwind); err != nil {
					return err
				}
			}

		case *SinglePartQuery:
			if err := walkList(enter, exit, node, typedNode.ReadingClauses); err != nil {
				return err
			}

			if err := walkList(enter, exit, node, typedNode.UpdatingClauses); err != nil {
				return err
			}

			if typedNode.Return != nil {
				if err := walkNodes(enter, exit, node, typedNode.Return); err != nil {
					return err
				}
			}

		case *Remove:
			if err := walkList(enter, exit, node, typedNode.Items); err != nil {
				return err
			}

		case *Set:
			if err := walkList(enter, exit, node, typedNode.Items); err != nil {
				return err
			}

		case *SetItem:
			if err := walkNodes(enter, exit, node, typedNode.Right, typedNode.Left); err != nil {
				return err
			}

		case *Negation:
			if err := walkNodes(enter, exit, node, typedNode.Expression); err != nil {
				return err
			}

		case *PartialComparison:
			if err := walkNodes(enter, exit, node, typedNode.Right); err != nil {
				return err
			}

		case *Parenthetical:
			if err := walkNodes(enter, exit, node, typedNode.Expression); err != nil {
				return err
			}

		case *PatternElement:
			if err := walkNodes(enter, exit, typedNode, typedNode.Element); err != nil {
				return err
			}

		case *Match:
			if typedNode.Where != nil {
				if err := walkNodes(enter, exit, node, typedNode.Where); err != nil {
					return err
				}
			}

			if typedNode.Pattern != nil {
				if err := walkList(enter, exit, node, typedNode.Pattern); err != nil {
					return err
				}
			}

		case *Create:
			if err := walkList(enter, exit, node, typedNode.Pattern); err != nil {
				return err
			}

		case *Return:
			if err := walkNodes(enter, exit, node, typedNode.Projection); err != nil {
				return err
			}

		case *FunctionInvocation:
			if err := walkList(enter, exit, node, typedNode.Arguments); err != nil {
				return err
			}

		case *Comparison:
			if err := walkNodes(enter, exit, node, typedNode.Left); err != nil {
				return err
			}

			if err := walkList(enter, exit, node, typedNode.Partials); err != nil {
				return err
			}

		case []*PatternPart:
			if err := walkList(enter, exit, parent, typedNode); err != nil {
				return err
			}

		case *SortItem:
			if err := walkNodes(enter, exit, typedNode, typedNode.Expression); err != nil {
				return err
			}

		case *Order:
			if err := walkList(enter, exit, typedNode, typedNode.Items); err != nil {
				return err
			}

		case *Projection:
			if err := walkList(enter, exit, node, typedNode.Items); err != nil {
				return err
			}

			if typedNode.Order != nil {
				if err := walkNodes(enter, exit, typedNode, typedNode.Order); err != nil {
					return err
				}
			}

		case *ProjectionItem:
			if err := walkNodes(enter, exit, node, typedNode.Expression); err != nil {
				return err
			}

		case *ArithmeticExpression:
			if err := walkNodes(enter, exit, node, typedNode.Left); err != nil {
				return err
			}

			if err := walkList(enter, exit, node, typedNode.Partials); err != nil {
				return err
			}

		case *PartialArithmeticExpression:
			if err := walkNodes(enter, exit, node, typedNode.Right); err != nil {
				return err
			}

		case *Delete:
			for _, expression := range typedNode.Expressions {
				if err := walkNodes(enter, exit, node, expression); err != nil {
					return err
				}
			}

		case *KindMatcher:
			if err := walkNodes(enter, exit, node, typedNode.Reference); err != nil {
				return err
			}

		case *RemoveItem:
			if typedNode.KindMatcher != nil {
				if err := walkNodes(enter, exit, node, typedNode.KindMatcher); err != nil {
					return err
				}
			}

		case *PropertyLookup:
			if err := walkNodes(enter, exit, node, typedNode.Atom); err != nil {
				return err
			}

		case *UpdatingClause:
			if err := walkNodes(enter, exit, node, typedNode.Clause); err != nil {
				return err
			}

		case *NodePattern:
			if err := walkNodes(enter, exit, node, typedNode.Properties); err != nil {
				return err
			}

		case *PatternPart:
			if err := walkList(enter, exit, node, typedNode.PatternElements); err != nil {
				return err
			}

		case *RelationshipPattern:
			if err := walkNodes(enter, exit, node, typedNode.Properties); err != nil {
				return err
			}

		case *Properties:
			if err := walkNodes(enter, exit, node, typedNode.Parameter); err != nil {
				return err
			}

		case *Variable, *Literal, *Parameter, *RangeQuantifier, graph.Kinds:
			// Valid model elements but no further descent required

		case nil:
		default:
			return fmt.Errorf("unsupported type for model traversal %T(%T)", parent, node)
		}

		if exit != nil {
			if err := exit(parent, node); err != nil {
				return err
			}
		}
	}

	return nil
}

// Walk is a recursive, depth-first traversal implementation for the openCypher query model.
func Walk(element any, enter, exit Visitor) error {
	return walkNodes(enter, exit, nil, element)
}
