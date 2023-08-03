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

package analyzer

import (
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type Analyzer struct {
	handlers []func(node any)
}

func (s *Analyzer) Analyze(query *model.RegularQuery) error {
	return model.Walk(
		query, func(parent, node any) error {
			for _, handler := range s.handlers {
				handler(node)
			}

			return nil
		},
		nil,
	)
}

func WithVisitor[T any](analyzer *Analyzer, visitorFunc func(node T)) {
	analyzer.handlers = append(analyzer.handlers, func(node any) {
		if typedNode, typeOK := node.(T); typeOK {
			visitorFunc(typedNode)
		}
	})
}

// Weight constants aren't well named for right now. These are just dumb values to assign heuristic weight to certain
// query elements
const (
	Weight1 float64 = iota + 1
	Weight2
	Weight3
)

type ComplexityMeasure struct {
	Weight float64

	numPatterns     float64
	numProjections  float64
	nodeLookupKinds map[string]graph.Kinds
}

func (s *ComplexityMeasure) onFunctionInvocation(node *model.FunctionInvocation) {
	switch node.Name {
	case "collect":
		// Collect will force an eager aggregation
		s.Weight += Weight2

	case "type":
		// Calling for a relationship's type is highly likely to be inefficient and should add weight
		s.Weight += Weight2
	}
}

func (s *ComplexityMeasure) onQuantifier(node *model.Quantifier) {
	// Quantifier expressions may increase the size of an inline projection to apply its contained filter and should
	// be weighted
	s.Weight += Weight1
}

func (s *ComplexityMeasure) onFilterExpression(node *model.FilterExpression) {
	// Filter expressions convert directly into a filter in the query plan which may or may not take advantage
	// of indexes and should be weighted accordingly
	s.Weight += Weight1
}

func (s *ComplexityMeasure) onKindMatcher(node *model.KindMatcher) {
	switch typedReference := node.Reference.(type) {
	case *model.Variable:
		// This kind matcher narrows a node reference's kind and will result in an indexed lookup
		s.nodeLookupKinds[typedReference.Symbol] = s.nodeLookupKinds[typedReference.Symbol].Add(node.Kinds...)
	}
}

func (s *ComplexityMeasure) onPatternPart(node *model.PatternPart) {
	// All pattern parts incur a compounding weight
	s.numPatterns += 1
	s.Weight += s.numPatterns

	if node.ShortestPathPattern {
		// Rendering the shortest path, while cheaper than rendering all shortest paths, still could incur a large
		// search cost
		s.Weight += Weight1
	}

	if node.AllShortestPathsPattern {
		// Rendering all shortest paths could result in a large search
		s.Weight += Weight2
	}
}

func (s *ComplexityMeasure) onSortItem(node *model.SortItem) {
	// Sorting incurs a weight since it will change how the projection is materialized
	s.Weight += Weight1
}

func (s *ComplexityMeasure) onProjection(node *model.Projection) {
	// We want to capture the cost of additional inline projections so ignore the first projection
	s.Weight += s.numProjections
	s.numProjections += 1

	if node.Distinct {
		// Distinct incurs a weight since it will change how the projection is materialized
		s.Weight += Weight1
	}
}

func (s *ComplexityMeasure) onPartialComparison(node *model.PartialComparison) {
	switch node.Operator {
	case model.OperatorRegexMatch:
		// Regular expression matching incurs a weight since it can be far more involved than any of the other
		// string operators
		s.Weight += Weight1
	}
}

func (s *ComplexityMeasure) onNodePattern(node *model.NodePattern) {
	if node.Binding == "" {
		if len(node.Kinds) == 0 {
			// Unlabeled, unbound nodes will incur a lookup of all nodes in the graph
			s.Weight += Weight2
		}
	} else {
		nodeLookupKinds, hasBinding := s.nodeLookupKinds[node.Binding]

		if !hasBinding {
			nodeLookupKinds = node.Kinds
		} else {
			nodeLookupKinds = nodeLookupKinds.Add(node.Kinds...)
		}

		// Track this node pattern to see if any subsequent expressions will narrow its kind matchers
		s.nodeLookupKinds[node.Binding] = nodeLookupKinds
	}
}

func (s *ComplexityMeasure) onRelationshipPattern(node *model.RelationshipPattern) {
	numKindMatchers := len(node.Kinds)

	// All relationship lookups incur a weight
	s.Weight += Weight1

	if node.Direction == graph.DirectionBoth {
		// Bidirectional searches add weight
		s.Weight += Weight1
	}

	if numKindMatchers == 0 {
		// If user is expanding all relationship types add weight
		s.Weight += Weight2
	}

	if node.Range != nil {
		if numKindMatchers > 2 {
			// If we're matching on more than two relationship types add weight
			s.Weight += Weight1
		}

		if node.Range.StartIndex != nil && *node.Range.StartIndex > 1 {
			// Patterns that must have a floor greater than 1 may result in large expansions
			s.Weight += Weight1
		}

		if node.Range.EndIndex == nil {
			// Unbounded range literals are likely to result in large expansions
			s.Weight += Weight3
		} else if *node.Range.EndIndex > 1 {
			// Patterns that must have a ceiling greater than 1 may result in large expansions
			s.Weight += Weight1
		}
	}
}

func (s *ComplexityMeasure) onExit() {
	for _, kindMatchers := range s.nodeLookupKinds {
		if len(kindMatchers) == 0 {
			// Unlabeled nodes will incur a lookup of all nodes in the graph
			s.Weight += Weight2
		}
	}
}

func QueryComplexity(query *model.RegularQuery) (*ComplexityMeasure, error) {
	var (
		analyzer = &Analyzer{}
		measure  = &ComplexityMeasure{
			nodeLookupKinds: map[string]graph.Kinds{},
		}
	)

	WithVisitor[*model.PatternPart](analyzer, measure.onPatternPart)
	WithVisitor[*model.NodePattern](analyzer, measure.onNodePattern)
	WithVisitor[*model.Projection](analyzer, measure.onProjection)
	WithVisitor[*model.RelationshipPattern](analyzer, measure.onRelationshipPattern)
	WithVisitor[*model.FunctionInvocation](analyzer, measure.onFunctionInvocation)
	WithVisitor[*model.KindMatcher](analyzer, measure.onKindMatcher)
	WithVisitor[*model.Quantifier](analyzer, measure.onQuantifier)
	WithVisitor[*model.FilterExpression](analyzer, measure.onFilterExpression)
	WithVisitor[*model.SortItem](analyzer, measure.onSortItem)
	WithVisitor[*model.PartialComparison](analyzer, measure.onPartialComparison)

	if err := analyzer.Analyze(query); err != nil {
		return nil, err
	}

	// Wrap up with a call to the exit function
	measure.onExit()

	return measure, nil
}
