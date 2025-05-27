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

package query

import (
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cypher/models/walk"

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	ErrAmbiguousQueryVariables = errors.New("query mixes node and relationship query variables")
)

type Cache struct {
}

type Builder struct {
	regularQuery *cypher.RegularQuery
	cache        *Cache
}

func NewBuilder(cache *Cache) *Builder {
	return &Builder{
		regularQuery: EmptySinglePartQuery(),
		cache:        cache,
	}
}

func NewBuilderWithCriteria(criteria ...graph.Criteria) *Builder {
	builder := NewBuilder(nil)
	builder.Apply(criteria...)

	return builder
}

func (s *Builder) RegularQuery() *cypher.RegularQuery {
	return s.regularQuery
}

func (s *Builder) Build(allShortestPaths bool) (*cypher.RegularQuery, error) {
	if err := s.prepareMatch(allShortestPaths); err != nil {
		return nil, err
	}

	return s.regularQuery, nil
}

func (s *Builder) prepareMatch(allShortestPaths bool) error {
	var (
		patternPart = &cypher.PatternPart{}

		singleNodeBound    = false
		creatingSingleNode = false

		startNodeBound    = false
		creatingStartNode = false
		endNodeBound      = false
		creatingEndNode   = false
		edgeBound         = false
		creatingEdge      = false

		isRelationshipQuery = false

		bindWalk = walk.NewSimpleVisitor[cypher.SyntaxNode](func(node cypher.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
			switch typedNode := node.(type) {
			case *cypher.Variable:
				switch typedNode.Symbol {
				case NodeSymbol:
					singleNodeBound = true

				case EdgeStartSymbol:
					startNodeBound = true
					isRelationshipQuery = true

				case EdgeEndSymbol:
					endNodeBound = true
					isRelationshipQuery = true

				case EdgeSymbol:
					edgeBound = true
					isRelationshipQuery = true
				}
			}
		})
	)

	// Zip through updating clauses first
	for _, updatingClause := range s.regularQuery.SingleQuery.SinglePartQuery.UpdatingClauses {
		typedUpdatingClause, isUpdatingClause := updatingClause.(*cypher.UpdatingClause)

		if !isUpdatingClause {
			return fmt.Errorf("unexpected type for updating clause: %T", updatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *cypher.Create:
			if err := walk.Cypher(typedClause, walk.NewSimpleVisitor[cypher.SyntaxNode](func(node cypher.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
				switch typedNode := node.(type) {
				case *cypher.NodePattern:
					if patternBinding, typeOK := typedNode.Variable.(*cypher.Variable); !typeOK {
						errorHandler.SetErrorf("expected variable for pattern binding but got: %T", typedNode.Variable)
					} else {
						switch patternBinding.Symbol {
						case NodeSymbol:
							creatingSingleNode = true

						case EdgeStartSymbol:
							creatingStartNode = true

						case EdgeEndSymbol:
							creatingEndNode = true
						}
					}

				case *cypher.RelationshipPattern:
					if patternBinding, typeOK := typedNode.Variable.(*cypher.Variable); !typeOK {
						errorHandler.SetErrorf("expected variable for pattern binding but got: %T", typedNode.Variable)
					} else {
						switch patternBinding.Symbol {
						case EdgeSymbol:
							creatingEdge = true
						}
					}
				}
			})); err != nil {
				return err
			}

		case *cypher.Delete:
			if err := walk.Cypher(typedClause, bindWalk); err != nil {
				return err
			}
		}
	}

	// Is there a where clause?
	if firstReadingClause := GetFirstReadingClause(s.regularQuery); firstReadingClause != nil && firstReadingClause.Match.Where != nil {
		if err := walk.Cypher(firstReadingClause.Match.Where, bindWalk); err != nil {
			return err
		}
	}

	// Is there a return clause
	if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
		if err := walk.Cypher(s.regularQuery.SingleQuery.SinglePartQuery.Return, bindWalk); err != nil {
			return err
		}
	}

	// Validate we're not mixing references
	if isRelationshipQuery && singleNodeBound {
		return ErrAmbiguousQueryVariables
	}

	if singleNodeBound && !creatingSingleNode {
		// Bind the single-node variable
		patternPart.AddPatternElements(&cypher.NodePattern{
			Variable: cypher.NewVariableWithSymbol(NodeSymbol),
		})
	}

	if startNodeBound {
		// Bind the start-node variable
		patternPart.AddPatternElements(&cypher.NodePattern{
			Variable: cypher.NewVariableWithSymbol(EdgeStartSymbol),
		})
	}

	if isRelationshipQuery {
		if !startNodeBound && !creatingStartNode {
			// Add an empty node pattern if the start node isn't bound, and we aren't creating it
			patternPart.AddPatternElements(&cypher.NodePattern{})
		}

		if !creatingEdge {
			if edgeBound {
				// Bind the edge variable
				patternPart.AddPatternElements(&cypher.RelationshipPattern{
					Variable:  cypher.NewVariableWithSymbol(EdgeSymbol),
					Direction: graph.DirectionOutbound,
				})
			} else {
				patternPart.AddPatternElements(&cypher.RelationshipPattern{
					Direction: graph.DirectionOutbound,
				})
			}
		}

		if !endNodeBound && !creatingEndNode {
			patternPart.AddPatternElements(&cypher.NodePattern{})
		}
	}

	if endNodeBound {
		// Add an empty node pattern if the end node isn't bound, and we aren't creating it
		patternPart.AddPatternElements(&cypher.NodePattern{
			Variable: cypher.NewVariableWithSymbol(EdgeEndSymbol),
		})
	}

	if allShortestPaths {
		patternPart.AllShortestPathsPattern = true
		patternPart.Variable = cypher.NewVariableWithSymbol(PathSymbol)

		// Update all relationship PatternElements to expand fully (*..)
		for _, patternElement := range patternPart.PatternElements {
			if relationshipPattern, isRelationshipPattern := patternElement.AsRelationshipPattern(); isRelationshipPattern {
				relationshipPattern.Range = &cypher.PatternRange{}
			}
		}
	}

	if firstReadingClause := GetFirstReadingClause(s.regularQuery); firstReadingClause != nil {
		firstReadingClause.Match.Pattern = []*cypher.PatternPart{patternPart}
	} else if len(patternPart.PatternElements) > 0 {
		s.regularQuery.SingleQuery.SinglePartQuery.AddReadingClause(&cypher.ReadingClause{
			Match: &cypher.Match{
				Pattern: []*cypher.PatternPart{
					patternPart,
				},
			},
		})
	}

	return nil
}

func (s *Builder) Apply(criteria ...graph.Criteria) {
	for _, nextCriteria := range criteria {
		switch typedCriteria := nextCriteria.(type) {
		case []graph.Criteria:
			s.Apply(typedCriteria...)

		case *cypher.Where:
			firstReadingClause := GetFirstReadingClause(s.regularQuery)

			if firstReadingClause == nil {
				firstReadingClause = &cypher.ReadingClause{
					Match: cypher.NewMatch(false),
				}

				s.regularQuery.SingleQuery.SinglePartQuery.AddReadingClause(firstReadingClause)
			}

			firstReadingClause.Match.Where = cypher.Copy(typedCriteria)

		case *cypher.Return:
			s.regularQuery.SingleQuery.SinglePartQuery.Return = typedCriteria

		case *cypher.Limit:
			if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
				s.regularQuery.SingleQuery.SinglePartQuery.Return.Projection.Limit = cypher.Copy(typedCriteria)
			}

		case *cypher.Skip:
			if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
				s.regularQuery.SingleQuery.SinglePartQuery.Return.Projection.Skip = cypher.Copy(typedCriteria)
			}

		case *cypher.Order:
			if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
				s.regularQuery.SingleQuery.SinglePartQuery.Return.Projection.Order = cypher.Copy(typedCriteria)
			}

		case []*cypher.UpdatingClause:
			for _, updatingClause := range typedCriteria {
				s.Apply(updatingClause)
			}

		case *cypher.UpdatingClause:
			s.regularQuery.SingleQuery.SinglePartQuery.AddUpdatingClause(cypher.Copy(typedCriteria))

		default:
			panic(fmt.Errorf("invalid type for dawgs query: %T %+v", typedCriteria, typedCriteria))
		}
	}
}
