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
	"fmt"

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/walk"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type variableType int

const (
	typeNode variableType = iota
	typeEdge
	typePattern
	typeUnknown
)

var (
	highlySelectiveNodeProperties = map[string]struct{}{
		"objectid": {},
		"name":     {},
	}
)

func asVariable(expression cypher.Expression) (*cypher.Variable, bool) {
	if variableRef, typeOK := expression.(*cypher.Variable); typeOK && variableRef != nil {
		return variableRef, true
	}

	return nil, false
}

type trackedIdentifier struct {
	Identifier string
	Type       variableType
}

type identifierTracker struct {
	identifiers map[string]*trackedIdentifier
	aliases     map[string]*trackedIdentifier

	stashed *identifierTracker
}

func (s *identifierTracker) Pop() {
	lastStashed := s.stashed
	s.stashed = lastStashed.stashed

	s.identifiers = lastStashed.identifiers
	s.aliases = lastStashed.aliases
}

func (s *identifierTracker) Stash() {
	var (
		identifiersCopy = make(map[string]*trackedIdentifier, len(s.identifiers))
		aliasesCopy     = make(map[string]*trackedIdentifier, len(s.aliases))
	)

	for k, v := range s.identifiers {
		identifiersCopy[k] = v
	}

	for k, v := range s.aliases {
		aliasesCopy[k] = v
	}

	s.stashed = &identifierTracker{
		identifiers: identifiersCopy,
		aliases:     aliasesCopy,
		stashed:     s.stashed,
	}
}

func newIdentifierTracker() *identifierTracker {
	return &identifierTracker{
		identifiers: map[string]*trackedIdentifier{},
	}
}

func (s *identifierTracker) Clear() {
	for key := range s.identifiers {
		delete(s.identifiers, key)
	}
}

func (s *identifierTracker) Alias(alias, identifier string) {
	s.aliases[alias] = s.identifiers[identifier]
}

func (s *identifierTracker) Lookup(identifier string) (*trackedIdentifier, bool) {
	tracked, isTracked := s.identifiers[identifier]
	return tracked, isTracked
}

func (s *identifierTracker) Track(identifier string, identifierType variableType) {
	if _, exists := s.identifiers[identifier]; !exists {
		s.identifiers[identifier] = &trackedIdentifier{
			Identifier: identifier,
			Type:       identifierType,
		}
	}
}

type ComplexityMeasure struct {
	RelativeFitness        int64
	NumMatches             int64
	NumMultiPartQueryParts int64
}

type ComplexityVisitor struct {
	walk.Visitor[cypher.SyntaxNode]

	complexityMeasure ComplexityMeasure
	identifierTracker *identifierTracker
}

func NewComplexityVisitor() *ComplexityVisitor {
	return &ComplexityVisitor{
		Visitor:           walk.NewVisitor[cypher.SyntaxNode](),
		identifierTracker: newIdentifierTracker(),
	}
}

func (s *ComplexityVisitor) isHighlySelectivePropertyLookup(propertyLookup *cypher.PropertyLookup) (bool, error) {
	if variable, typeOK := propertyLookup.Atom.(*cypher.Variable); !typeOK {
		return false, fmt.Errorf("expected variable for property lookup atom but received %T", propertyLookup.Atom)
	} else if tracked, isTracked := s.identifierTracker.Lookup(variable.Symbol); !isTracked {
		return false, fmt.Errorf("unable to lookup identifier %s", variable.Symbol)
	} else if tracked.Type == typeNode {
		_, isHighlySelectiveProperty := highlySelectiveNodeProperties[propertyLookup.Symbol]
		return isHighlySelectiveProperty, nil
	}

	return false, nil
}

func calculatePatternPartFitness(untypedProperties cypher.Expression, kinds graph.Kinds) (int64, error) {
	fitness := int64(0)

	if untypedProperties != nil {
		if properties, typeOK := untypedProperties.(*cypher.Properties); !typeOK {
			return 0, fmt.Errorf("expected *cypher.Properties but received %T", properties)
		} else {
			// Each map item is a pattern matching assertion that behaves exactly like
			// an equality comparison
			for _, mapItem := range properties.Map.Items() {
				// validate if the key is highly selective
				if _, isHighlySelective := highlySelectiveNodeProperties[mapItem.Key]; isHighlySelective {
					fitness += 4
				} else {
					fitness += 3
				}
			}
		}
	}

	// Kind assertions usually increase selectivity
	if numKindAssertions := len(kinds); numKindAssertions > 0 {
		if numKindAssertions == 1 {
			fitness += 2
		} else if numKindAssertions <= 15 {
			// Multiple intersections are disjunctions which will narrow search space less effectively
			// than a single kind assertion
			fitness += 1
		}

		// Kind assertion disjunctions of greater than 15 do not receive a fitness score
	}

	return fitness, nil
}

func anyToInt64(rawValue any) (int64, error) {
	switch typedValue := rawValue.(type) {

	case int8:
		return int64(typedValue), nil
	case int16:
		return int64(typedValue), nil
	case int32:
		return int64(typedValue), nil
	case int64:
		return typedValue, nil
	case int:
		return int64(typedValue), nil
	default:
		return 0, fmt.Errorf("unknown integer type: %T", rawValue)
	}
}

func (s *ComplexityVisitor) Enter(node cypher.SyntaxNode) {
	switch typedNode := node.(type) {
	case *cypher.PatternPart:
		if variable, hasVariable := asVariable(typedNode.Variable); hasVariable {
			s.identifierTracker.Track(variable.Symbol, typePattern)
		}

		// Shortest path and all shortest paths will reduce search space, often drastically
		if typedNode.ShortestPathPattern {
			// Shortest path reduces search space even further by limiting the results rendered down to one per
			// distinct path root.
			//
			// For example: match p = shortestPath((u:User)-[:MemberOf*..]->(g:Group)) return p
			//
			// The above query will render one shortest path per unique left-hand node `u`.
			s.complexityMeasure.RelativeFitness += 5
		}

		if typedNode.AllShortestPathsPattern {
			s.complexityMeasure.RelativeFitness += 4
		}

		// Proportionally penalize queries for long patterns that may contain variable expansion dynamics
		if len(typedNode.PatternElements) > 3 {
			s.complexityMeasure.RelativeFitness -= int64(len(typedNode.PatternElements))
		}

	case *cypher.Limit:
		// Limits typically reduce search space but only if the limit is reasonable
		if limitLiteral, typeOK := typedNode.Value.(*cypher.Literal); !typeOK {
			s.SetErrorf("expected literal but received %T", limitLiteral)
		} else if limitValue, err := anyToInt64(limitLiteral.Value); err != nil {
			s.SetError(err)
		} else if limitValue <= 300 {
			// 300 was chosen as the limit value purely out of prior experience and is subject to
			// modification if the heuristic proved to be ineffective
			s.complexityMeasure.RelativeFitness += 1
		}

	case *cypher.Projection:
		if typedNode.Distinct {
			// Distinct projections typically reduces result set size
			s.complexityMeasure.RelativeFitness += 1
		}

	case *cypher.Match:
		// Optional matches only expand the query's search space
		if typedNode.Optional {
			s.complexityMeasure.RelativeFitness -= 1
		}

		// Track this match
		s.complexityMeasure.NumMatches += 1

	case *cypher.MultiPartQueryPart:
		// Multipart queries typically mean some kind of pipeline, making this query less fit
		s.complexityMeasure.RelativeFitness -= 3

		// Track this part
		s.complexityMeasure.NumMultiPartQueryParts += 1

	case *cypher.NodePattern:
		if variable, hasVariable := asVariable(typedNode.Variable); hasVariable {
			s.identifierTracker.Track(variable.Symbol, typeNode)
		}

		if patternPartFitness, err := calculatePatternPartFitness(typedNode.Properties, typedNode.Kinds); err != nil {
			s.SetError(err)
		} else {
			s.complexityMeasure.RelativeFitness += patternPartFitness
		}

	case *cypher.RelationshipPattern:
		if variable, hasVariable := asVariable(typedNode.Variable); hasVariable {
			s.identifierTracker.Track(variable.Symbol, typeEdge)
		}

		if typedNode.Direction == graph.DirectionBoth {
			// Bidirectional expansions reduce search space similarly to a unidirectional expansion
			s.complexityMeasure.RelativeFitness -= 3
		}

		// Relationships represent some kind of association which often expands search space
		s.complexityMeasure.RelativeFitness -= 1

		// Variable expansions can either expand or contract path search space
		if typedNode.Range != nil {
			// Range min and max depth limits affect selectivity proportionally
			if typedNode.Range.StartIndex != nil {
				s.complexityMeasure.RelativeFitness += *typedNode.Range.StartIndex

				if typedNode.Range.EndIndex != nil {
					s.complexityMeasure.RelativeFitness -= *typedNode.Range.EndIndex
				}
			} else if typedNode.Range.EndIndex != nil {
				s.complexityMeasure.RelativeFitness -= *typedNode.Range.EndIndex
			} else {
				// Unbounded recursive expansion lowers selectivity by expanding the path search
				// space drastically
				s.complexityMeasure.RelativeFitness -= 3
			}
		}

		if patternPartFitness, err := calculatePatternPartFitness(typedNode.Properties, typedNode.Kinds); err != nil {
			s.SetError(err)
		} else {
			s.complexityMeasure.RelativeFitness += patternPartFitness
		}

	case *cypher.Negation:
		// Exclusions are not selective
		s.complexityMeasure.RelativeFitness -= 1

	case *cypher.Conjunction:
		// Conjoined expressions typically reduce search space
		s.complexityMeasure.RelativeFitness += 1

	case *cypher.Disjunction:
		// Conjoined expressions typically increase search space
		s.complexityMeasure.RelativeFitness -= 1

	case *cypher.Quantifier:
		// Quantifier statements requre some kind of in-query materialization or result walk
		s.complexityMeasure.RelativeFitness -= 1

		switch typedNode.Type {
		case cypher.QuantifierTypeAll, cypher.QuantifierTypeNone:
			// All and None quantifier statements are more selective than others
			s.complexityMeasure.RelativeFitness -= 1

		default:
			// All other quantifier statements are less selective
			s.complexityMeasure.RelativeFitness -= 2
		}

	case *cypher.Comparison:
		// If the left operand is a property lookup, check if the property being references is highly selective
		if propertyLookup, isPropertyLookup := typedNode.Left.(*cypher.PropertyLookup); isPropertyLookup {
			if isHighlySelective, err := s.isHighlySelectivePropertyLookup(propertyLookup); err != nil {
				s.SetError(err)
			} else if isHighlySelective {
				s.complexityMeasure.RelativeFitness += 4
			}
		}

		for _, partial := range typedNode.Partials {
			// If the next right operand is a property lookup, check if the property being references is highly selective
			if propertyLookup, isPropertyLookup := partial.Right.(*cypher.PropertyLookup); isPropertyLookup {
				if isHighlySelective, err := s.isHighlySelectivePropertyLookup(propertyLookup); err != nil {
					s.SetError(err)
				} else if isHighlySelective {
					s.complexityMeasure.RelativeFitness += 4
				}
			}

			// Weight values are given to each operator based on relative selectivity of each operator
			switch partial.Operator {
			case cypher.OperatorEquals, cypher.OperatorIs:
				s.complexityMeasure.RelativeFitness += 3

			case cypher.OperatorStartsWith, cypher.OperatorEndsWith, cypher.OperatorContains, cypher.OperatorRegexMatch:
				s.complexityMeasure.RelativeFitness += 2

			case cypher.OperatorGreaterThan, cypher.OperatorGreaterThanOrEqualTo, cypher.OperatorLessThan,
				cypher.OperatorLessThanOrEqualTo, cypher.OperatorIn:
				s.complexityMeasure.RelativeFitness += 1

			case cypher.OperatorNotEquals, cypher.OperatorIsNot:
				s.complexityMeasure.RelativeFitness -= 1
			}
		}

	case *cypher.KindMatcher:
		// Kind assertions usually increase selectivity
		if numKindAssertions := len(typedNode.Kinds); numKindAssertions == 1 {
			s.complexityMeasure.RelativeFitness += 2
		} else if numKindAssertions > 1 {
			// Multiple intersections are usually disjunctions which will narrow search space less effectively
			// than a single kind assertion
			s.complexityMeasure.RelativeFitness += 1
		}

	case *cypher.FilterExpression:
		s.identifierTracker.Stash()

	case *cypher.IDInCollection:
		if variable, isVariable := asVariable(typedNode.Variable); !isVariable {
			s.SetErrorf("expected *cypher.Variable but received %T", typedNode.Variable)
		} else {
			s.identifierTracker.Track(variable.Symbol, typeUnknown)
		}
	}
}

func (s *ComplexityVisitor) onWithStatementProjectionItem(expr cypher.Expression) error {
	if projectionItem, typeOK := expr.(*cypher.ProjectionItem); !typeOK {
		return fmt.Errorf("expected projection item but received %T", expr)
	} else if projectionItem.Alias != nil {
		if variable, isVariable := asVariable(projectionItem.Alias); !isVariable {
			return fmt.Errorf("expected variable for projection alias but received %T", projectionItem.Alias)
		} else {
			// Future effort here would be to plumb type inference by inspecting the projection item's expression
			s.identifierTracker.Track(variable.Symbol, typeUnknown)
		}
	}

	return nil
}

func (s *ComplexityVisitor) Exit(node cypher.SyntaxNode) {
	switch typedNode := node.(type) {
	case *cypher.With:
		s.identifierTracker.Clear()

		if typedNode.Projection != nil {
			for _, projectionItem := range typedNode.Projection.Items {
				if err := s.onWithStatementProjectionItem(projectionItem); err != nil {
					s.SetError(err)
				}
			}
		}

	case *cypher.SinglePartQuery:
		s.identifierTracker.Clear()

		// Attempt to capture the impact of the length of a query's potential pipeline
		if s.complexityMeasure.NumMultiPartQueryParts > 2 {
			// Proportionally penalize multipart queries that have a pipeline depth of greater than 2
			s.complexityMeasure.RelativeFitness -= s.complexityMeasure.NumMultiPartQueryParts
		}

		if s.complexityMeasure.NumMatches > 2 {
			// Proportionally penalize queries that have a significant number of matches present
			s.complexityMeasure.RelativeFitness -= s.complexityMeasure.NumMatches
		}

	case *cypher.FilterExpression:
		s.identifierTracker.Pop()
	}
}

func QueryComplexity(query *cypher.RegularQuery) (ComplexityMeasure, error) {
	visitor := NewComplexityVisitor()
	return visitor.complexityMeasure, walk.Cypher(query, visitor)
}
