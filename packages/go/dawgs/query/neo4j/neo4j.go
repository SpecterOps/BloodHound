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

	"github.com/specterops/bloodhound/cypher/backend/cypher"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

var (
	ErrAmbiguousQueryVariables = errors.New("query mixes node and relationship query variables")
)

type QueryBuilder struct {
	Parameters map[string]any

	query    *model.RegularQuery
	order    *model.Order
	prepared bool
}

func NewQueryBuilder(singleQuery *model.RegularQuery) *QueryBuilder {
	return &QueryBuilder{
		query: model.Copy(singleQuery),
	}
}

func NewEmptyQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: &model.RegularQuery{
			SingleQuery: &model.SingleQuery{
				SinglePartQuery: &model.SinglePartQuery{},
			},
		},
	}
}

func (s *QueryBuilder) liftRelationshipKindMatchers() model.Visitor {
	firstReadingClause := query.GetFirstReadingClause(s.query)

	return model.NewVisitor(func(stack *model.WalkStack, element model.Expression) error {
		if firstReadingClause == nil {
			return nil
		}

		if firstReadingClause.Match == nil {
			return fmt.Errorf("first reading clause of query has a nil match clause")
		}

		firstRelationshipPattern := firstReadingClause.Match.FirstRelationshipPattern()

		switch typedElement := element.(type) {
		case model.ExpressionList:
			var removeList []model.Expression

			for _, expression := range typedElement.GetAll() {
				switch typedExpression := expression.(type) {
				case *model.KindMatcher:
					switch variable := typedExpression.Reference.(type) {
					case *model.Variable:
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

	if err := model.Walk(s.query, model.NewVisitor(parameterRewriter.Visit, nil)); err != nil {
		return err
	}

	s.Parameters = parameterRewriter.Parameters
	return nil
}

func (s *QueryBuilder) Apply(criteria graph.Criteria) {
	switch typedCriteria := criteria.(type) {
	case *model.Where:
		if query.GetFirstReadingClause(s.query) == nil {
			s.query.SingleQuery.SinglePartQuery.AddReadingClause(&model.ReadingClause{
				Match: model.NewMatch(false),
			})
		}

		query.GetFirstReadingClause(s.query).Match.Where = model.Copy(typedCriteria)

	case *model.Return:
		s.query.SingleQuery.SinglePartQuery.Return = model.Copy(typedCriteria)

	case *model.Limit:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Limit = model.Copy(typedCriteria)
		}

	case *model.Skip:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Skip = model.Copy(typedCriteria)
		}

	case *model.Order:
		s.order = model.Copy(typedCriteria)

	case []*model.UpdatingClause:
		for _, updatingClause := range typedCriteria {
			s.Apply(updatingClause)
		}

	case *model.UpdatingClause:
		s.query.SingleQuery.SinglePartQuery.AddUpdatingClause(model.Copy(typedCriteria))

	default:
		panic(fmt.Sprintf("invalid type for dawgs query: %T %+v", criteria, criteria))
	}
}

func (s *QueryBuilder) prepareMatch() error {
	var (
		patternPart = &model.PatternPart{}

		singleNodeBound    = false
		creatingSingleNode = false

		startNodeBound       = false
		creatingStartNode    = false
		endNodeBound         = false
		creatingEndNode      = false
		relationshipBound    = false
		creatingRelationship = false

		isRelationshipQuery = false

		bindWalk = model.NewVisitor(func(stack *model.WalkStack, element model.Expression) error {
			switch typedElement := element.(type) {
			case *model.Variable:
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
		typedUpdatingClause, typeOK := updatingClause.(*model.UpdatingClause)

		if !typeOK {
			return fmt.Errorf("unexpected updating clause type %T", typedUpdatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *model.Create:
			if err := model.Walk(typedClause, model.NewVisitor(func(stack *model.WalkStack, element model.Expression) error {
				switch typedElement := element.(type) {
				case *model.NodePattern:
					if typedBinding, isVariable := typedElement.Binding.(*model.Variable); !isVariable {
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

				case *model.RelationshipPattern:
					if typedBinding, isVariable := typedElement.Binding.(*model.Variable); !isVariable {
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

		case *model.Delete:
			if err := model.Walk(typedClause, bindWalk, nil); err != nil {
				return err
			}
		}
	}

	// Is there a where clause?
	if firstReadingClause := query.GetFirstReadingClause(s.query); firstReadingClause != nil && firstReadingClause.Match.Where != nil {
		if err := model.Walk(firstReadingClause.Match.Where, bindWalk, nil); err != nil {
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

		if err := model.Walk(s.query.SingleQuery.SinglePartQuery.Return, bindWalk, nil); err != nil {
			return err
		}
	}

	// Validate we're not mixing references
	if isRelationshipQuery && singleNodeBound {
		return ErrAmbiguousQueryVariables
	}

	if singleNodeBound && !creatingSingleNode {
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(query.NodeSymbol),
		})
	}

	if startNodeBound {
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(query.EdgeStartSymbol),
		})
	}

	if isRelationshipQuery {
		if !startNodeBound && !creatingStartNode {
			patternPart.AddPatternElements(&model.NodePattern{})
		}

		if !creatingRelationship {
			if relationshipBound {
				patternPart.AddPatternElements(&model.RelationshipPattern{
					Binding:   model.NewVariableWithSymbol(query.EdgeSymbol),
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
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(query.EdgeEndSymbol),
		})
	}

	if firstReadingClause := query.GetFirstReadingClause(s.query); firstReadingClause != nil {
		firstReadingClause.Match.Pattern = []*model.PatternPart{patternPart}
	} else if len(patternPart.PatternElements) > 0 {
		s.query.SingleQuery.SinglePartQuery.AddReadingClause(&model.ReadingClause{
			Match: &model.Match{
				Pattern: []*model.PatternPart{
					patternPart,
				},
			},
		})
	}

	return nil
}

func (s *QueryBuilder) compilationErrors() error {
	var modelErrors []error

	model.Walk(s.query, model.NewVisitor(func(stack *model.WalkStack, element model.Expression) error {
		if errorNode, typeOK := element.(model.Fallible); typeOK {
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

	if err := model.Walk(s.query, model.NewVisitor(StringNegationRewriter, nil)); err != nil {
		return err
	}

	if err := model.Walk(s.query, s.liftRelationshipKindMatchers()); err != nil {
		return err
	}

	return model.Walk(s.query, model.NewVisitor(nil, RemoveEmptyExpressionLists))
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
		patternPart.Binding = model.NewVariableWithSymbol(query.PathSymbol)

		// Set the pattern to search for all shortest paths
		patternPart.AllShortestPathsPattern = true

		// Update all relationship PatternElements to expand fully (*..)
		for _, patternElement := range patternPart.PatternElements {
			if relationshipPattern, isRelationshipPattern := patternElement.AsRelationshipPattern(); isRelationshipPattern {
				relationshipPattern.Range = &model.PatternRange{}
			}
		}

		return nil
	}
}

func (s *QueryBuilder) Render() (string, error) {
	buffer := &bytes.Buffer{}

	if err := cypher.NewCypherEmitter(false).Write(s.query, buffer); err != nil {
		return "", err
	} else {
		return buffer.String(), nil
	}
}
