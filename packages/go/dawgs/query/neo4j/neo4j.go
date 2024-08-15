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

package neo4j

import (
	"bytes"
	"errors"
	"fmt"

	cypherBackend "github.com/specterops/bloodhound/cypher/models/cypher/format"

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

var (
	ErrAmbiguousQueryVariables = errors.New("query mixes node and relationship query variables")
)

type QueryBuilder struct {
	Parameters map[string]any

	query    *cypher.RegularQuery
	order    *cypher.Order
	prepared bool
}

func NewQueryBuilder(singleQuery *cypher.RegularQuery) *QueryBuilder {
	return &QueryBuilder{
		query: cypher.Copy(singleQuery),
	}
}

func NewEmptyQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: &cypher.RegularQuery{
			SingleQuery: &cypher.SingleQuery{
				SinglePartQuery: &cypher.SinglePartQuery{},
			},
		},
	}
}

func (s *QueryBuilder) liftRelationshipKindMatchers() cypher.Visitor {
	firstReadingClause := query.GetFirstReadingClause(s.query)

	return cypher.NewVisitor(func(stack *cypher.WalkStack, element cypher.Expression) error {
		if firstReadingClause == nil {
			return nil
		}

		if firstReadingClause.Match == nil {
			return fmt.Errorf("first reading clause of query has a nil match clause")
		}

		firstRelationshipPattern := firstReadingClause.Match.FirstRelationshipPattern()

		switch typedElement := element.(type) {
		case cypher.ExpressionList:
			var removeList []cypher.Expression

			for _, expression := range typedElement.GetAll() {
				switch typedExpression := expression.(type) {
				case *cypher.KindMatcher:
					switch variable := typedExpression.Reference.(type) {
					case *cypher.Variable:
						switch variable.Symbol {
						case query.EdgeSymbol:
							firstRelationshipPattern.Kinds = append(firstRelationshipPattern.Kinds, typedExpression.Kinds...)
							removeList = append(removeList, expression)
						}
					}
				}
			}

			for _, expression := range removeList {
				typedElement.Remove(expression)
			}
		}

		return nil
	}, nil)
}

func (s *QueryBuilder) rewriteParameters() error {
	parameterRewriter := query.NewParameterRewriter()

	if err := cypher.Walk(s.query, cypher.NewVisitor(parameterRewriter.Visit, nil)); err != nil {
		return err
	}

	s.Parameters = parameterRewriter.Parameters
	return nil
}

func (s *QueryBuilder) Apply(criteria graph.Criteria) {
	switch typedCriteria := criteria.(type) {
	case *cypher.Where:
		if query.GetFirstReadingClause(s.query) == nil {
			s.query.SingleQuery.SinglePartQuery.AddReadingClause(&cypher.ReadingClause{
				Match: cypher.NewMatch(false),
			})
		}

		query.GetFirstReadingClause(s.query).Match.Where = cypher.Copy(typedCriteria)

	case *cypher.Return:
		s.query.SingleQuery.SinglePartQuery.Return = cypher.Copy(typedCriteria)

	case *cypher.Limit:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Limit = cypher.Copy(typedCriteria)
		}

	case *cypher.Skip:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Skip = cypher.Copy(typedCriteria)
		}

	case *cypher.Order:
		s.order = cypher.Copy(typedCriteria)

	case []*cypher.UpdatingClause:
		for _, updatingClause := range typedCriteria {
			s.Apply(updatingClause)
		}

	case *cypher.UpdatingClause:
		s.query.SingleQuery.SinglePartQuery.AddUpdatingClause(cypher.Copy(typedCriteria))

	default:
		panic(fmt.Sprintf("invalid type for dawgs query: %T %+v", criteria, criteria))
	}
}

func (s *QueryBuilder) prepareMatch() error {
	var (
		patternPart = &cypher.PatternPart{}

		singleNodeBound    = false
		creatingSingleNode = false

		startNodeBound       = false
		creatingStartNode    = false
		endNodeBound         = false
		creatingEndNode      = false
		relationshipBound    = false
		creatingRelationship = false

		isRelationshipQuery = false

		bindWalk = cypher.NewVisitor(func(stack *cypher.WalkStack, element cypher.Expression) error {
			switch typedElement := element.(type) {
			case *cypher.Variable:
				switch typedElement.Symbol {
				case query.NodeSymbol:
					singleNodeBound = true

				case query.EdgeStartSymbol:
					startNodeBound = true
					isRelationshipQuery = true

				case query.EdgeEndSymbol:
					endNodeBound = true
					isRelationshipQuery = true

				case query.EdgeSymbol:
					relationshipBound = true
					isRelationshipQuery = true
				}
			}

			return nil
		}, nil)
	)

	// Zip through updating clauses first
	for _, updatingClause := range s.query.SingleQuery.SinglePartQuery.UpdatingClauses {
		typedUpdatingClause, typeOK := updatingClause.(*cypher.UpdatingClause)

		if !typeOK {
			return fmt.Errorf("unexpected updating clause type %T", typedUpdatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *cypher.Create:
			if err := cypher.Walk(typedClause, cypher.NewVisitor(func(stack *cypher.WalkStack, element cypher.Expression) error {
				switch typedElement := element.(type) {
				case *cypher.NodePattern:
					if typedBinding, isVariable := typedElement.Binding.(*cypher.Variable); !isVariable {
						return fmt.Errorf("expected variable but got %T", typedElement.Binding)
					} else {
						switch typedBinding.Symbol {
						case query.NodeSymbol:
							creatingSingleNode = true

						case query.EdgeStartSymbol:
							creatingStartNode = true

						case query.EdgeEndSymbol:
							creatingEndNode = true
						}
					}

				case *cypher.RelationshipPattern:
					if typedBinding, isVariable := typedElement.Binding.(*cypher.Variable); !isVariable {
						return fmt.Errorf("expected variable but got %T", typedElement.Binding)
					} else {
						switch typedBinding.Symbol {
						case query.EdgeSymbol:
							creatingRelationship = true
						}
					}
				}

				return nil
			}, nil)); err != nil {
				return err
			}

		case *cypher.Delete:
			if err := cypher.Walk(typedClause, bindWalk, nil); err != nil {
				return err
			}
		}
	}

	// Is there a where clause?
	if firstReadingClause := query.GetFirstReadingClause(s.query); firstReadingClause != nil && firstReadingClause.Match.Where != nil {
		if err := cypher.Walk(firstReadingClause.Match.Where, bindWalk, nil); err != nil {
			return err
		}
	}

	// Is there a return clause
	if spqReturn := s.query.SingleQuery.SinglePartQuery.Return; spqReturn != nil && spqReturn.Projection != nil {
		// Did we have an order specified?
		if s.order != nil {
			if spqReturn.Projection.Order != nil {
				return fmt.Errorf("order specified twice")
			}

			s.query.SingleQuery.SinglePartQuery.Return.Projection.Order = s.order
		}

		if err := cypher.Walk(s.query.SingleQuery.SinglePartQuery.Return, bindWalk, nil); err != nil {
			return err
		}
	}

	// Validate we're not mixing references
	if isRelationshipQuery && singleNodeBound {
		return ErrAmbiguousQueryVariables
	}

	if singleNodeBound && !creatingSingleNode {
		patternPart.AddPatternElements(&cypher.NodePattern{
			Binding: cypher.NewVariableWithSymbol(query.NodeSymbol),
		})
	}

	if startNodeBound {
		patternPart.AddPatternElements(&cypher.NodePattern{
			Binding: cypher.NewVariableWithSymbol(query.EdgeStartSymbol),
		})
	}

	if isRelationshipQuery {
		if !startNodeBound && !creatingStartNode {
			patternPart.AddPatternElements(&cypher.NodePattern{})
		}

		if !creatingRelationship {
			if relationshipBound {
				patternPart.AddPatternElements(&cypher.RelationshipPattern{
					Binding:   cypher.NewVariableWithSymbol(query.EdgeSymbol),
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
		patternPart.AddPatternElements(&cypher.NodePattern{
			Binding: cypher.NewVariableWithSymbol(query.EdgeEndSymbol),
		})
	}

	if firstReadingClause := query.GetFirstReadingClause(s.query); firstReadingClause != nil {
		firstReadingClause.Match.Pattern = []*cypher.PatternPart{patternPart}
	} else if len(patternPart.PatternElements) > 0 {
		s.query.SingleQuery.SinglePartQuery.AddReadingClause(&cypher.ReadingClause{
			Match: &cypher.Match{
				Pattern: []*cypher.PatternPart{
					patternPart,
				},
			},
		})
	}

	return nil
}

func (s *QueryBuilder) compilationErrors() error {
	var modelErrors []error

	cypher.Walk(s.query, cypher.NewVisitor(func(stack *cypher.WalkStack, element cypher.Expression) error {
		if errorNode, typeOK := element.(cypher.Fallible); typeOK {
			if len(errorNode.Errors()) > 0 {
				modelErrors = append(modelErrors, errorNode.Errors()...)
			}
		}

		return nil
	}, nil))

	return errors.Join(modelErrors...)
}

func (s *QueryBuilder) Prepare() error {
	if s.prepared {
		return nil
	}

	s.prepared = true

	if s.query.SingleQuery.SinglePartQuery == nil {
		return fmt.Errorf("single part query is nil")
	}

	if err := s.compilationErrors(); err != nil {
		return err
	}

	if err := s.prepareMatch(); err != nil {
		return err
	}

	if err := s.rewriteParameters(); err != nil {
		return err
	}

	if err := cypher.Walk(s.query, cypher.NewVisitor(StringNegationRewriter, nil)); err != nil {
		return err
	}

	if err := cypher.Walk(s.query, s.liftRelationshipKindMatchers()); err != nil {
		return err
	}

	return cypher.Walk(s.query, cypher.NewVisitor(nil, RemoveEmptyExpressionLists))
}

func (s *QueryBuilder) PrepareAllShortestPaths() error {
	if err := s.Prepare(); err != nil {
		return err
	} else {
		firstReadingClause := query.GetFirstReadingClause(s.query)

		// Set all pattern parts to search for the shortest paths and bind them
		if len(firstReadingClause.Match.Pattern) > 1 {
			return fmt.Errorf("only expected one pattern")
		}

		// Grab the first pattern part
		patternPart := firstReadingClause.Match.Pattern[0]

		// Bind the path
		patternPart.Binding = cypher.NewVariableWithSymbol(query.PathSymbol)

		// Set the pattern to search for all shortest paths
		patternPart.AllShortestPathsPattern = true

		// Update all relationship PatternElements to expand fully (*..)
		for _, patternElement := range patternPart.PatternElements {
			if relationshipPattern, isRelationshipPattern := patternElement.AsRelationshipPattern(); isRelationshipPattern {
				relationshipPattern.Range = &cypher.PatternRange{}
			}
		}

		return nil
	}
}

func (s *QueryBuilder) Render() (string, error) {
	buffer := &bytes.Buffer{}

	if err := cypherBackend.NewCypherEmitter(false).Write(s.query, buffer); err != nil {
		return "", err
	} else {
		return buffer.String(), nil
	}
}
