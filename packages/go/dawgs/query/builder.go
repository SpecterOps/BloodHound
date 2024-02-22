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
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	ErrAmbiguousQueryVariables = errors.New("query mixes node and relationship query variables")
)

type Cache struct {
}

type Builder struct {
	regularQuery *model.RegularQuery
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

func (s *Builder) RegularQuery() *model.RegularQuery {
	return s.regularQuery
}

func (s *Builder) Build() (*model.RegularQuery, error) {
	if err := s.prepareMatch(); err != nil {
		return nil, err
	}

	return s.regularQuery, nil
}

func (s *Builder) prepareMatch() error {
	var (
		patternPart = &model.PatternPart{}

		singleNodeBound    = false
		creatingSingleNode = false

		startNodeBound    = false
		creatingStartNode = false
		endNodeBound      = false
		creatingEndNode   = false
		edgeBound         = false
		creatingEdge      = false

		isRelationshipQuery = false

		bindWalk = model.NewVisitor(func(stack *model.WalkStack, branch model.Expression) error {
			switch typedElement := branch.(type) {
			case *model.Variable:
				switch typedElement.Symbol {
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

			return nil
		}, nil)
	)

	// Zip through updating clauses first
	for _, updatingClause := range s.regularQuery.SingleQuery.SinglePartQuery.UpdatingClauses {
		typedUpdatingClause, isUpdatingClause := updatingClause.(*model.UpdatingClause)

		if !isUpdatingClause {
			return fmt.Errorf("unexpected type for updating clause: %T", updatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *model.Create:
			if err := model.Walk(typedClause, model.NewVisitor(func(stack *model.WalkStack, element model.Expression) error {
				switch typedElement := element.(type) {
				case *model.NodePattern:
					if patternBinding, typeOK := typedElement.Binding.(*model.Variable); !typeOK {
						return fmt.Errorf("expected variable for pattern binding but got: %T", typedElement.Binding)
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

				case *model.RelationshipPattern:
					if patternBinding, typeOK := typedElement.Binding.(*model.Variable); !typeOK {
						return fmt.Errorf("expected variable for pattern binding but got: %T", typedElement.Binding)
					} else {
						switch patternBinding.Symbol {
						case EdgeSymbol:
							creatingEdge = true
						}
					}
				}

				return nil
			}, nil)); err != nil {
				return err
			}

		case *model.Delete:
			if err := model.Walk(typedClause, bindWalk); err != nil {
				return err
			}
		}
	}

	// Is there a where clause?
	if firstReadingClause := GetFirstReadingClause(s.regularQuery); firstReadingClause != nil && firstReadingClause.Match.Where != nil {
		if err := model.Walk(firstReadingClause.Match.Where, bindWalk); err != nil {
			return err
		}
	}

	// Is there a return clause
	if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
		if err := model.Walk(s.regularQuery.SingleQuery.SinglePartQuery.Return, bindWalk); err != nil {
			return err
		}
	}

	// Validate we're not mixing references
	if isRelationshipQuery && singleNodeBound {
		return ErrAmbiguousQueryVariables
	}

	if singleNodeBound && !creatingSingleNode {
		// Bind the single-node variable
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(NodeSymbol),
		})
	}

	if startNodeBound {
		// Bind the start-node variable
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(EdgeStartSymbol),
		})
	}

	if isRelationshipQuery {
		if !startNodeBound && !creatingStartNode {
			// Add an empty node pattern if the start node isn't bound, and we aren't creating it
			patternPart.AddPatternElements(&model.NodePattern{})
		}

		if !creatingEdge {
			if edgeBound {
				// Bind the edge variable
				patternPart.AddPatternElements(&model.RelationshipPattern{
					Binding:   model.NewVariableWithSymbol(EdgeSymbol),
					Direction: graph.DirectionOutbound,
				})
			} else {
				patternPart.AddPatternElements(&model.RelationshipPattern{
					Direction: graph.DirectionOutbound,
				})
			}
		}

		if !endNodeBound && !creatingEndNode {
			patternPart.AddPatternElements(&model.NodePattern{})
		}
	}

	if endNodeBound {
		// Add an empty node pattern if the end node isn't bound, and we aren't creating it
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(EdgeEndSymbol),
		})
	}

	if firstReadingClause := GetFirstReadingClause(s.regularQuery); firstReadingClause != nil {
		firstReadingClause.Match.Pattern = []*model.PatternPart{patternPart}
	} else if len(patternPart.PatternElements) > 0 {
		s.regularQuery.SingleQuery.SinglePartQuery.AddReadingClause(&model.ReadingClause{
			Match: &model.Match{
				Pattern: []*model.PatternPart{
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

		case *model.Where:
			firstReadingClause := GetFirstReadingClause(s.regularQuery)

			if firstReadingClause == nil {
				firstReadingClause = &model.ReadingClause{
					Match: model.NewMatch(false),
				}

				s.regularQuery.SingleQuery.SinglePartQuery.AddReadingClause(firstReadingClause)
			}

			firstReadingClause.Match.Where = model.Copy(typedCriteria)

		case *model.Return:
			s.regularQuery.SingleQuery.SinglePartQuery.Return = typedCriteria

		case *model.Limit:
			if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
				s.regularQuery.SingleQuery.SinglePartQuery.Return.Projection.Limit = model.Copy(typedCriteria)
			}

		case *model.Skip:
			if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
				s.regularQuery.SingleQuery.SinglePartQuery.Return.Projection.Skip = model.Copy(typedCriteria)
			}

		case *model.Order:
			if s.regularQuery.SingleQuery.SinglePartQuery.Return != nil {
				s.regularQuery.SingleQuery.SinglePartQuery.Return.Projection.Order = model.Copy(typedCriteria)
			}

		case []*model.UpdatingClause:
			for _, updatingClause := range typedCriteria {
				s.Apply(updatingClause)
			}

		case *model.UpdatingClause:
			s.regularQuery.SingleQuery.SinglePartQuery.AddUpdatingClause(model.Copy(typedCriteria))

		default:
			panic(fmt.Errorf("invalid type for dawgs query: %T %+v", criteria, criteria))
		}
	}
}
