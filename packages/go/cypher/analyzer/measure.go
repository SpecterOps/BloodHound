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

package analyzer

import (
	"fmt"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

// Weight constants aren't well named for right now. These are just dumb values to assign heuristic weight to certain
// query elements
const (
	weight1             int64 = 1
	weight2             int64 = 2
	weight3             int64 = 3
	weightHeavy         int64 = 7
	weightMaxComplexity int64 = 50
)

type ComplexityMeasure struct {
	Weight int64

	hasWhere             bool
	hasPatternProperties bool
	hasLimit             bool
	isCreate             bool

	numPatterns     int64
	numProjections  int64
	nodeLookupKinds map[string]graph.Kinds
}

func (s *ComplexityMeasure) onCreate(_ *model.WalkStack, _ *model.Create) error {
	// Let's add 1 per create
	s.Weight += weight1
	s.isCreate = true

	return nil
}

func (s *ComplexityMeasure) onDelete(_ *model.WalkStack, node *model.Delete) error {
	// Base weight for delete is 3, if detach is specified, we give a heavy weight on top to account
	// for the extra complexity of deleting relationships
	s.Weight += weight3
	if node.Detach {
		s.Weight += weightHeavy
	}

	return nil
}

func (s *ComplexityMeasure) onExit() {
	var hasKindMatcher bool

	for _, kindMatchers := range s.nodeLookupKinds {
		if len(kindMatchers) == 0 {
			// Unlabeled nodes will incur a lookup of all nodes in the graph
			s.Weight += weight2
		} else {
			hasKindMatcher = true
		}
	}

	// TODO: This is a little gross and needs to be refactored
	if !hasKindMatcher && !s.hasPatternProperties && !s.hasWhere && !s.hasLimit && !s.isCreate {
		s.Weight += weightMaxComplexity
	}
}

func (s *ComplexityMeasure) onFunctionInvocation(_ *model.WalkStack, node *model.FunctionInvocation) error {
	switch node.Name {
	case "collect":
		// Collect will force an eager aggregation
		s.Weight += weight2

	case "type":
		// Calling for a relationship's type is highly likely to be inefficient and should add weight
		s.Weight += weight2
	}

	return nil
}

func (s *ComplexityMeasure) onKindMatcher(_ *model.WalkStack, node *model.KindMatcher) error {
	switch typedReference := node.Reference.(type) {
	case *model.Variable:
		// This kind matcher narrows a node reference's kind and will result in an indexed lookup
		s.nodeLookupKinds[typedReference.Symbol] = s.nodeLookupKinds[typedReference.Symbol].Add(node.Kinds...)
	}

	return nil
}

func (s *ComplexityMeasure) onMerge(_ *model.WalkStack, node *model.Merge) error {
	// Let's add 1 per merge action
	s.Weight += weight1 * int64(len(node.MergeActions))

	return nil
}

func (s *ComplexityMeasure) onNodePattern(_ *model.WalkStack, node *model.NodePattern) error {
	if node.Binding == nil {
		if len(node.Kinds) == 0 {
			// Unlabeled, unbound nodes will incur a lookup of all nodes in the graph
			s.Weight += weight2
		}
	} else if nodePatternBinding, typeOK := node.Binding.(*model.Variable); !typeOK {
		return fmt.Errorf("expected variable for node pattern binding but got: %T", node.Binding)
	} else {
		nodeLookupKinds, hasBinding := s.nodeLookupKinds[nodePatternBinding.Symbol]

		if !hasBinding {
			nodeLookupKinds = node.Kinds
		} else {
			nodeLookupKinds = nodeLookupKinds.Add(node.Kinds...)
		}

		// Track this node pattern to see if any subsequent expressions will narrow its kind matchers
		s.nodeLookupKinds[nodePatternBinding.Symbol] = nodeLookupKinds
	}

	if node.Properties != nil {
		s.hasPatternProperties = true
	}

	return nil
}

func (s *ComplexityMeasure) onPartialComparison(_ *model.WalkStack, node *model.PartialComparison) error {
	switch node.Operator {
	case model.OperatorRegexMatch:
		// Regular expression matching incurs a weight since it can be far more involved than any of the other
		// string operators
		s.Weight += weight1
	}

	return nil
}

func (s *ComplexityMeasure) onPatternPart(_ *model.WalkStack, node *model.PatternPart) error {
	// All pattern parts incur a compounding weight
	s.numPatterns += 1
	s.Weight += s.numPatterns

	if node.ShortestPathPattern {
		// Rendering the shortest path, while cheaper than rendering all shortest paths, still could incur a large
		// search cost
		s.Weight += weight1
	}

	if node.AllShortestPathsPattern {
		// Rendering all shortest paths could result in a large search
		s.Weight += weight2
	}

	return nil
}

func (s *ComplexityMeasure) onProjection(_ *model.WalkStack, node *model.Projection) error {
	// We want to capture the cost of additional inline projections so ignore the first projection
	s.Weight += s.numProjections
	s.numProjections += 1

	if node.Distinct {
		// Distinct incurs a weight since it will change how the projection is materialized
		s.Weight += weight1
	}

	if node.Limit != nil {
		s.hasLimit = true
	}

	return nil
}

func (s *ComplexityMeasure) onQuantifier(_ *model.WalkStack, _ *model.Quantifier) error {
	// Quantifier expressions may increase the size of an inline projection to apply its contained filter and should
	// be weighted
	s.Weight += weight1
	return nil
}

func (s *ComplexityMeasure) onRelationshipPattern(_ *model.WalkStack, node *model.RelationshipPattern) error {
	numKindMatchers := len(node.Kinds)

	// All relationship lookups incur a weight
	s.Weight += weight1

	if node.Direction == graph.DirectionBoth {
		// Bidirectional searches add weight
		s.Weight += weight1
	}

	if numKindMatchers == 0 {
		// If user is expanding all relationship types add weight
		s.Weight += weight2
	}

	if node.Range != nil {
		if numKindMatchers > 2 {
			// If we're matching on more than two relationship types add weight
			s.Weight += weight1
		}

		if node.Range.StartIndex != nil && *node.Range.StartIndex > 1 {
			// Patterns that must have a floor greater than 1 may result in large expansions
			s.Weight += weight1
		}

		if node.Range.EndIndex == nil {
			// Unbounded range literals are likely to result in large expansions
			s.Weight += weight3
		} else if *node.Range.EndIndex > 1 {
			// Patterns that must have a ceiling greater than 1 may result in large expansions
			s.Weight += weight1
		}
	}

	if node.Properties != nil {
		s.hasPatternProperties = true
	}

	if len(node.Kinds) != 0 {
		s.hasPatternProperties = true
	}

	return nil
}

func (s *ComplexityMeasure) onRemove(_ *model.WalkStack, node *model.Remove) error {
	// Let's add 1 per remove
	s.Weight += weight1

	return nil
}

func (s *ComplexityMeasure) onSet(_ *model.WalkStack, node *model.Set) error {
	// Let's add 1 per set
	s.Weight += weight1

	return nil
}

func (s *ComplexityMeasure) onSortItem(_ *model.WalkStack, _ *model.SortItem) error {
	// Sorting incurs a weight since it will change how the projection is materialized
	s.Weight += weight1
	return nil
}

func (s *ComplexityMeasure) onWhere(_ *model.WalkStack, _ *model.Where) error {
	// Filters in the query plan may or may not take advantage of indexes and should be weighted accordingly
	s.Weight += weight1
	s.hasWhere = true
	return nil
}
