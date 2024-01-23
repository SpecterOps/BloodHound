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

func copySlice[T any, S []T](slice S) S {
	var valueCopy []T

	if slice != nil {
		valueCopy = make([]T, len(slice))

		for idx, item := range slice {
			valueCopy[idx] = Copy(item)
		}
	}

	return valueCopy
}

type CopyExtension[T any] func(value T) (T, bool)

func Copy[T any](value T, extensions ...CopyExtension[T]) T {
	var empty T

	switch typedValue := any(value).(type) {
	case *RegularQuery:
		return any(typedValue.copy()).(T)

	case *MultiPartQuery:
		return any(typedValue.copy()).(T)

	case *SingleQuery:
		return any(typedValue.copy()).(T)

	case *SinglePartQuery:
		return any(typedValue.copy()).(T)

	case *Match:
		return any(typedValue.copy()).(T)

	case *Projection:
		return any(typedValue.copy()).(T)

	case *IDInCollection:
		return any(typedValue.copy()).(T)

	case *FilterExpression:
		return any(typedValue.copy()).(T)

	case *Quantifier:
		return any(typedValue.copy()).(T)

	case *Where:
		return any(typedValue.copy()).(T)

	case *Return:
		return any(typedValue.copy()).(T)

	case *Remove:
		return any(typedValue.copy()).(T)

	case *SetItem:
		return any(typedValue.copy()).(T)

	case *Order:
		return any(typedValue.copy()).(T)

	case *Skip:
		return any(typedValue.copy()).(T)

	case *Limit:
		return any(typedValue.copy()).(T)

	case *RemoveItem:
		return any(typedValue.copy()).(T)

	case *Comparison:
		return any(typedValue.copy()).(T)

	case *ArithmeticExpression:
		return any(typedValue.copy()).(T)

	case *PartialArithmeticExpression:
		return any(typedValue.copy()).(T)

	case *PartialComparison:
		return any(typedValue.copy()).(T)

	case *Parenthetical:
		return any(typedValue.copy()).(T)

	case *Unwind:
		return any(typedValue.copy()).(T)

	case *FunctionInvocation:
		return any(typedValue.copy()).(T)

	case *Variable:
		return any(typedValue.copy()).(T)

	case *Parameter:
		return any(typedValue.copy()).(T)

	case *Literal:
		return any(typedValue.copy()).(T)

	case *ReadingClause:
		return any(typedValue.copy()).(T)

	case *PropertyLookup:
		return any(typedValue.copy()).(T)

	case *Set:
		return any(typedValue.copy()).(T)

	case *Delete:
		return any(typedValue.copy()).(T)

	case *Create:
		return any(typedValue.copy()).(T)

	case *KindMatcher:
		return any(typedValue.copy()).(T)

	case *Conjunction:
		return any(typedValue.copy()).(T)

	case *Disjunction:
		return any(typedValue.copy()).(T)

	case *ExclusiveDisjunction:
		return any(typedValue.copy()).(T)

	case expressionList:
		return any(typedValue.copy()).(T)

	case *PatternPart:
		return any(typedValue.copy()).(T)

	case *With:
		return any(typedValue.copy()).(T)

	case *Negation:
		return any(typedValue.copy()).(T)

	case *NodePattern:
		return any(typedValue.copy()).(T)

	case *SortItem:
		return any(typedValue.copy()).(T)

	case *ProjectionItem:
		return any(typedValue.copy()).(T)

	case *RelationshipPattern:
		return any(typedValue.copy()).(T)

	case *PatternRange:
		return any(typedValue.copy()).(T)

	case *PatternPredicate:
		return any(typedValue.copy()).(T)

	case *PatternElement:
		return any(typedValue.copy()).(T)

	case *UpdatingClause:
		return any(typedValue.copy()).(T)

	case *MultiPartQueryPart:
		return any(typedValue.copy()).(T)

	case []*MultiPartQueryPart:
		return any(copySlice(typedValue)).(T)

	case []*PartialArithmeticExpression:
		return any(copySlice(typedValue)).(T)

	case []*PartialComparison:
		return any(copySlice(typedValue)).(T)

	case []*PatternPart:
		return any(copySlice(typedValue)).(T)

	case []*UpdatingClause:
		return any(copySlice(typedValue)).(T)

	case []*ProjectionItem:
		return any(copySlice(typedValue)).(T)

	case []*SortItem:
		return any(copySlice(typedValue)).(T)

	case []*RemoveItem:
		return any(copySlice(typedValue)).(T)

	case []*SetItem:
		return any(copySlice(typedValue)).(T)

	case []*PatternElement:
		return any(copySlice(typedValue)).(T)

	case []*ReadingClause:
		return any(copySlice(typedValue)).(T)

	case []Expression:
		return any(copySlice(typedValue)).(T)

	case graph.Kinds:
		return any(typedValue.Copy()).(T)

	case *int64:
		if typedValue == nil {
			return empty
		}

		valueCopy := *typedValue
		return any(&valueCopy).(T)

	case []string:
		var valueCopy []string

		if typedValue != nil {
			valueCopy = make([]string, len(typedValue))
			copy(valueCopy, typedValue)
		}

		return any(valueCopy).(T)

	case nil:
		return empty

	default:
		for _, extension := range extensions {
			if valueCopy, handled := extension(value); handled {
				return valueCopy
			}
		}

		panic(fmt.Sprintf("unable to copy type %T", value))
	}
}
