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
	"errors"
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/cypher/model/walk"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type Analyzer struct {
	handlers []func(stack *walk.WalkStack, node cypher.Expression) error
}

func (s *Analyzer) walkFunc(stack *walk.WalkStack, expression cypher.Expression) error {
	var errs []error

	for _, handler := range s.handlers {
		if err := handler(stack, expression); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (s *Analyzer) Analyze(query any, extensions ...walk.CollectorFunc) error {
	return walk.Walk(query, walk.NewVisitor(s.walkFunc, nil), extensions...)
}

func Analyze(query any, registrationFunc func(analyzerInst *Analyzer), extensions ...walk.CollectorFunc) error {
	analyzer := &Analyzer{}
	registrationFunc(analyzer)

	return analyzer.Analyze(query, extensions...)
}

type TypedVisitor[T cypher.Expression] func(stack *walk.WalkStack, node T) error

func WithVisitor[T cypher.Expression](analyzer *Analyzer, visitorFunc TypedVisitor[T]) {
	analyzer.handlers = append(analyzer.handlers, func(walkStack *walk.WalkStack, node cypher.Expression) error {
		if typedNode, typeOK := node.(T); typeOK {
			if err := visitorFunc(walkStack, typedNode); err != nil {
				return err
			}
		}

		return nil
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

func (s *ComplexityMeasure) onFunctionInvocation(_ *walk.WalkStack, node *cypher.FunctionInvocation) error {
	switch node.Name {
	case "collect":
		// Collect will force an eager aggregation
		s.Weight += Weight2

	case "type":
		// Calling for a relationship's type is highly likely to be inefficient and should add weight
		s.Weight += Weight2
	}

	return nil
}

func (s *ComplexityMeasure) onQuantifier(_ *walk.WalkStack, _ *cypher.Quantifier) error {
	// Quantifier expressions may increase the size of an inline projection to apply its contained filter and should
	// be weighted
	s.Weight += Weight1
	return nil
}

func (s *ComplexityMeasure) onFilterExpression(_ *walk.WalkStack, _ *cypher.FilterExpression) error {
	// Filter expressions convert directly into a filter in the query plan which may or may not take advantage
	// of indexes and should be weighted accordingly
	s.Weight += Weight1
	return nil
}

func (s *ComplexityMeasure) onKindMatcher(_ *walk.WalkStack, node *cypher.KindMatcher) error {
	switch typedReference := node.Reference.(type) {
	case *cypher.Variable:
		// This kind matcher narrows a node reference's kind and will result in an indexed lookup
		s.nodeLookupKinds[typedReference.Symbol] = s.nodeLookupKinds[typedReference.Symbol].Add(node.Kinds...)
	}

	return nil
}

func (s *ComplexityMeasure) onPatternPart(_ *walk.WalkStack, node *cypher.PatternPart) error {
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

	return nil
}

func (s *ComplexityMeasure) onSortItem(_ *walk.WalkStack, _ *cypher.SortItem) error {
	// Sorting incurs a weight since it will change how the projection is materialized
	s.Weight += Weight1
	return nil
}

func (s *ComplexityMeasure) onProjection(_ *walk.WalkStack, node *cypher.Projection) error {
	// We want to capture the cost of additional inline projections so ignore the first projection
	s.Weight += s.numProjections
	s.numProjections += 1

	if node.Distinct {
		// Distinct incurs a weight since it will change how the projection is materialized
		s.Weight += Weight1
	}

	return nil
}

func (s *ComplexityMeasure) onPartialComparison(_ *walk.WalkStack, node *cypher.PartialComparison) error {
	switch node.Operator {
	case cypher.OperatorRegexMatch:
		// Regular expression matching incurs a weight since it can be far more involved than any of the other
		// string operators
		s.Weight += Weight1
	}

	return nil
}

func (s *ComplexityMeasure) onNodePattern(_ *walk.WalkStack, node *cypher.NodePattern) error {
	if node.Binding == nil {
		if len(node.Kinds) == 0 {
			// Unlabeled, unbound nodes will incur a lookup of all nodes in the graph
			s.Weight += Weight2
		}
	} else if nodePatternBinding, typeOK := node.Binding.(*cypher.Variable); !typeOK {
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

	return nil
}

func (s *ComplexityMeasure) onRelationshipPattern(_ *walk.WalkStack, node *cypher.RelationshipPattern) error {
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

	return nil
}

func (s *ComplexityMeasure) onExit() {
	for _, kindMatchers := range s.nodeLookupKinds {
		if len(kindMatchers) == 0 {
			// Unlabeled nodes will incur a lookup of all nodes in the graph
			s.Weight += Weight2
		}
	}
}

func QueryComplexity(query *cypher.RegularQuery) (*ComplexityMeasure, error) {
	var (
		analyzer = &Analyzer{}
		measure  = &ComplexityMeasure{
			nodeLookupKinds: map[string]graph.Kinds{},
		}
	)

	WithVisitor(analyzer, measure.onPatternPart)
	WithVisitor(analyzer, measure.onNodePattern)
	WithVisitor(analyzer, measure.onProjection)
	WithVisitor(analyzer, measure.onRelationshipPattern)
	WithVisitor(analyzer, measure.onFunctionInvocation)
	WithVisitor(analyzer, measure.onKindMatcher)
	WithVisitor(analyzer, measure.onQuantifier)
	WithVisitor(analyzer, measure.onFilterExpression)
	WithVisitor(analyzer, measure.onSortItem)
	WithVisitor(analyzer, measure.onPartialComparison)

	if err := analyzer.Analyze(query); err != nil {
		return nil, err
	}

	// Wrap up with a call to the exit function
	measure.onExit()

	return measure, nil
}
