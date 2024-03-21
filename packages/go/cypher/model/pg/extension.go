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

package pg

import (
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/cypher/model/walk"
)

func Copy[T any](value T) T {
	return cypher.Copy(value, func(value T) (T, bool) {
		var valueCopy T

		switch typedValue := any(value).(type) {
		case *AnnotatedVariable:
			valueCopy = any(typedValue.copy()).(T)

		case *AnnotatedKindMatcher:
			valueCopy = any(typedValue.copy()).(T)

		default:
			return valueCopy, false
		}

		return valueCopy, true
	})
}

func CollectPGSQLTypes(nextCursor *walk.WalkCursor, expression cypher.Expression) bool {
	switch typedExpression := expression.(type) {
	case *PropertiesReference:
		walk.Collect(nextCursor, typedExpression.Reference)

	case *AnnotatedPropertyLookup:
		walk.CollectExpression(nextCursor, typedExpression.Atom)

	case *AnnotatedKindMatcher:
		walk.CollectExpression(nextCursor, typedExpression.Reference)

	case *Entity:
		walk.Collect(nextCursor, typedExpression.Binding)

	case *Subquery:
		walk.CollectSlice(nextCursor, typedExpression.PatternElements)
		walk.CollectExpression(nextCursor, typedExpression.Filter)

	case *PropertyMutation:
		walk.Collect(nextCursor, typedExpression.Reference)
		walk.Collect(nextCursor, typedExpression.Removals)
		walk.Collect(nextCursor, typedExpression.Additions)

	case *Delete:
		walk.Collect(nextCursor, typedExpression.Binding)

	case *KindMutation:
		walk.Collect(nextCursor, typedExpression.Variable)
		walk.Collect(nextCursor, typedExpression.Removals)
		walk.Collect(nextCursor, typedExpression.Additions)

	case *NodeKindsReference:
		walk.CollectExpression(nextCursor, typedExpression.Variable)

	case *EdgeKindReference:
		walk.CollectExpression(nextCursor, typedExpression.Variable)

	case *AnnotatedLiteral, *AnnotatedVariable, *AnnotatedParameter:
		// Valid types but no descent

	default:
		return false
	}

	return true
}
