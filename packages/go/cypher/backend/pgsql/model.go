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

package pgsql

import (
	"fmt"
	cypherModel "github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/model/pg"

	"github.com/specterops/bloodhound/dawgs/graph"
)

const (
	OperatorJSONBFieldExists    cypherModel.Operator = "?"
	OperatorLike                cypherModel.Operator = "like"
	OperatorLikeCaseInsensitive cypherModel.Operator = "ilike"
)

type UpdatingClauseRewriter struct {
	kindMapper               KindMapper
	binder                   *Binder
	deletion                 *pg.Delete
	propertyReferenceSymbols map[string]struct{}
	propertyAdditions        map[string]map[string]any
	propertyRemovals         map[string][]string
	kindReferenceSymbols     map[string]struct{}
	kindRemovals             map[string][]graph.Kind
	kindAdditions            map[string][]graph.Kind
}

func NewUpdateClauseRewriter(binder *Binder, kindMapper KindMapper) *UpdatingClauseRewriter {
	return &UpdatingClauseRewriter{
		kindMapper:               kindMapper,
		binder:                   binder,
		deletion:                 pg.NewDelete(),
		propertyReferenceSymbols: map[string]struct{}{},
		propertyAdditions:        map[string]map[string]any{},
		propertyRemovals:         map[string][]string{},
		kindReferenceSymbols:     map[string]struct{}{},
		kindRemovals:             map[string][]graph.Kind{},
		kindAdditions:            map[string][]graph.Kind{},
	}
}

func (s *UpdatingClauseRewriter) newPropertyMutation(symbol string) (*pg.PropertyMutation, error) {
	if annotatedVariable, isBound := s.binder.LookupVariable(symbol); !isBound {
		return nil, fmt.Errorf("mutation variable reference %s is not bound", symbol)
	} else {
		return &pg.PropertyMutation{
			Reference: &pg.PropertiesReference{
				Reference: annotatedVariable,
			},
		}, nil
	}
}

func (s *UpdatingClauseRewriter) newKindMutation(symbol string) (*pg.KindMutation, error) {
	if annotatedVariable, isBound := s.binder.LookupVariable(symbol); !isBound {
		return nil, fmt.Errorf("mutation variable reference %s is not bound", symbol)
	} else {
		return &pg.KindMutation{
			Variable: annotatedVariable,
		}, nil
	}
}

func (s *UpdatingClauseRewriter) ToUpdatingClause() ([]cypherModel.Expression, error) {
	var updatingClauses []cypherModel.Expression

	if s.deletion.NodeDelete || s.deletion.EdgeDelete {
		updatingClauses = append(updatingClauses, s.deletion)
	}

	for referenceSymbol := range s.propertyReferenceSymbols {
		propertyMutation, err := s.newPropertyMutation(referenceSymbol)

		if err != nil {
			return nil, err
		}

		if propertyAdditions, hasPropertyAdditions := s.propertyAdditions[referenceSymbol]; hasPropertyAdditions {
			if propertyAdditionsJSONB, err := MapStringAnyToJSONB(propertyAdditions); err != nil {
				return nil, err
			} else if newParameter, err := s.binder.NewParameter(propertyAdditionsJSONB); err != nil {
				return nil, err
			} else {
				propertyMutation.Additions = newParameter
			}
		}

		if propertyRemovals, hasPropertyRemovals := s.propertyRemovals[referenceSymbol]; hasPropertyRemovals {
			if propertyRemovalsTextArray, err := StringSliceToTextArray(propertyRemovals); err != nil {
				return nil, err
			} else if newParameter, err := s.binder.NewParameter(propertyRemovalsTextArray); err != nil {
				return nil, err
			} else {
				propertyMutation.Removals = newParameter
			}
		}

		updatingClauses = append(updatingClauses, propertyMutation)
	}

	for referenceSymbol := range s.kindReferenceSymbols {
		kindMutation, err := s.newKindMutation(referenceSymbol)

		if err != nil {
			return nil, err
		}

		if kindAdditions, hasKindAdditions := s.kindAdditions[referenceSymbol]; hasKindAdditions {
			if kindInt2Array, missingKinds := s.kindMapper.MapKinds(kindAdditions); len(missingKinds) > 0 {
				return nil, fmt.Errorf("updating clause references the following unknown kinds: %v", missingKinds.Strings())
			} else if newParameter, err := s.binder.NewParameter(kindInt2Array); err != nil {
				return nil, err
			} else {
				kindMutation.Additions = newParameter
			}
		}

		if kindRemovals, hasKindRemovals := s.kindRemovals[referenceSymbol]; hasKindRemovals {
			if kindInt2Array, missingKinds := s.kindMapper.MapKinds(kindRemovals); len(missingKinds) > 0 {
				return nil, fmt.Errorf("updating clause references the following unknown kinds: %v", missingKinds.Strings())
			} else if newParameter, err := s.binder.NewParameter(kindInt2Array); err != nil {
				return nil, err
			} else {
				kindMutation.Removals = newParameter
			}
		}

		updatingClauses = append(updatingClauses, kindMutation)
	}

	return updatingClauses, nil
}

func (s *UpdatingClauseRewriter) rewriteDeleteClause(singlePartQuery *cypherModel.SinglePartQuery, deleteClause *cypherModel.Delete) error {
	for _, deleteStatementExpression := range deleteClause.Expressions {
		switch typedExpression := deleteStatementExpression.(type) {
		case *pg.AnnotatedVariable:
			switch typedExpression.Type {
			case pg.Node:
				if s.deletion.NodeDelete {
					return fmt.Errorf("multiple node delete statements are not supported")
				}

				s.deletion.Binding = typedExpression
				s.deletion.NodeDelete = true

			case pg.Edge:
				if s.deletion.EdgeDelete {
					return fmt.Errorf("multiple edge delete statements are not supported")
				}

				s.deletion.Binding = typedExpression
				s.deletion.EdgeDelete = true

			default:
				return fmt.Errorf("unexpected variable type: %s", typedExpression.Type.String())
			}

		default:
			return fmt.Errorf("unexpected expression for delete: %T", deleteStatementExpression)
		}
	}

	if s.deletion.IsMixed() {
		return fmt.Errorf("mixed deletions are not supported")
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if matchClause := readingClause.Match; matchClause != nil {
			var additionalWhereClauses []cypherModel.Expression

			for _, pattern := range matchClause.Pattern {
				if len(pattern.PatternElements) <= 1 {
					// This pattern does not have a relationship and therefore no joining criteria is required
					continue
				}

				for idx, patternElement := range pattern.PatternElements {
					if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
						var (
							lastNode   = idx+1 >= len(pattern.PatternElements)
							relBinding *pg.AnnotatedVariable
							direction  graph.Direction
						)

						if !lastNode {
							// Look forward to the next relationship pattern
							relPattern, _ := pattern.PatternElements[idx+1].AsRelationshipPattern()
							direction = relPattern.Direction

							switch typedBinding := relPattern.Binding.(type) {
							case *pg.AnnotatedVariable:
								relBinding = typedBinding
							default:
								return fmt.Errorf("unexpected variable for relationship pattern binding: %T", relPattern.Binding)
							}
						} else {
							// Look backward to the last relationship pattern
							relPattern, _ := pattern.PatternElements[idx-1].AsRelationshipPattern()
							direction, _ = relPattern.Direction.Reverse()

							switch typedBinding := relPattern.Binding.(type) {
							case *pg.AnnotatedVariable:
								relBinding = typedBinding
							default:
								return fmt.Errorf("unexpected variable for relationship pattern binding: %T", relPattern.Binding)
							}
						}

						switch direction {
						case graph.DirectionInbound:
							bindingCopy := pg.Copy(relBinding)
							bindingCopy.Symbol += ".end_id"

							additionalWhereClauses = append(additionalWhereClauses, cypherModel.NewComparison(
								cypherModel.NewSimpleFunctionInvocation(cypherIdentityFunction, nodePattern.Binding),
								cypherModel.OperatorEquals,
								bindingCopy,
							))

						case graph.DirectionOutbound:
							bindingCopy := pg.Copy(relBinding)
							bindingCopy.Symbol += ".start_id"

							additionalWhereClauses = append(additionalWhereClauses, cypherModel.NewComparison(
								cypherModel.NewSimpleFunctionInvocation(cypherIdentityFunction, nodePattern.Binding),
								cypherModel.OperatorEquals,
								bindingCopy,
							))

						default:
							return fmt.Errorf("invalid pattern direction: %d", direction)
						}
					}
				}
			}

			if len(additionalWhereClauses) > 0 {
				additionalWhereClause := cypherModel.NewConjunction(additionalWhereClauses...)

				if matchClause.Where == nil {
					matchClause.Where = cypherModel.NewWhere()
				}

				if len(matchClause.Where.Expressions) > 0 {
					matchClause.Where.Expressions = []cypherModel.Expression{
						cypherModel.NewConjunction(append(matchClause.Where.Expressions, additionalWhereClause)...),
					}
				} else {
					matchClause.Where.Add(additionalWhereClause)
				}
			}
		}
	}

	return nil
}

func (s *UpdatingClauseRewriter) RewriteUpdatingClauses(singlePartQuery *cypherModel.SinglePartQuery) error {
	for _, updatingClause := range singlePartQuery.UpdatingClauses {
		typedUpdatingClause, isUpdatingClause := updatingClause.(*cypherModel.UpdatingClause)

		if !isUpdatingClause {
			return fmt.Errorf("unexpected type for updating clause: %T", updatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *cypherModel.Create:
			return fmt.Errorf("create unsupported")

		case *cypherModel.Delete:
			if err := s.rewriteDeleteClause(singlePartQuery, typedClause); err != nil {
				return err
			}

		case *cypherModel.Set:
			for _, setItem := range typedClause.Items {
				switch leftHandOperand := setItem.Left.(type) {
				case *cypherModel.Variable:
					switch rightHandOperand := setItem.Right.(type) {
					case graph.Kinds:
						s.TrackKindAddition(leftHandOperand.Symbol, rightHandOperand...)

					default:
						return fmt.Errorf("unexpected right side operand type %T for kind setter", setItem.Right)
					}

				case *cypherModel.PropertyLookup:
					switch setItem.Operator {
					case cypherModel.OperatorAssignment:
						var (
							// TODO: Type negotiation
							referenceSymbol = leftHandOperand.Atom.(*cypherModel.Variable).Symbol
							propertyName    = leftHandOperand.Symbols[0]
						)

						switch rightHandOperand := setItem.Right.(type) {
						case *cypherModel.Literal:
							// TODO: Negotiate null literals
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						case *pg.AnnotatedLiteral:
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						case *cypherModel.Parameter:
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						case *pg.AnnotatedParameter:
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						default:
							return fmt.Errorf("unexpected right side operand type %T for property setter", setItem.Right)
						}

					default:
						return fmt.Errorf("unsupported assignment operator: %s", setItem.Operator)
					}
				}
			}

		case *cypherModel.Remove:
			for _, removeItem := range typedClause.Items {
				if removeItem.KindMatcher != nil {
					if kindMatcher, typeOK := removeItem.KindMatcher.(*cypherModel.KindMatcher); !typeOK {
						return fmt.Errorf("unexpected remove item kind matcher expression: %T", removeItem.KindMatcher)
					} else if kindMatcherReference, typeOK := kindMatcher.Reference.(*cypherModel.Variable); !typeOK {
						return fmt.Errorf("unexpected remove matcher reference expression: %T", kindMatcher.Reference)
					} else {
						s.TrackKindRemoval(kindMatcherReference.Symbol, kindMatcher.Kinds...)
					}
				}

				if removeItem.Property != nil {
					var (
						// TODO: Type negotiation
						referenceSymbol = removeItem.Property.Atom.(*cypherModel.Variable).Symbol
						propertyName    = removeItem.Property.Symbols[0]
					)

					s.TrackPropertyRemoval(referenceSymbol, propertyName)
				}
			}
		}
	}

	if updatingClauses, err := s.ToUpdatingClause(); err != nil {
		return err
	} else {
		singlePartQuery.UpdatingClauses = updatingClauses
	}

	return nil
}

func (s *UpdatingClauseRewriter) HasAdditions() bool {
	return len(s.propertyAdditions) > 0 || len(s.kindAdditions) > 0
}

func (s *UpdatingClauseRewriter) HasRemovals() bool {
	return len(s.propertyRemovals) > 0 || len(s.kindRemovals) > 0
}

func (s *UpdatingClauseRewriter) HasChanges() bool {
	return s.HasAdditions() || s.HasRemovals()
}

func (s *UpdatingClauseRewriter) TrackKindAddition(referenceSymbol string, kinds ...graph.Kind) {
	s.kindReferenceSymbols[referenceSymbol] = struct{}{}

	if existingAdditions, hasAdditions := s.kindAdditions[referenceSymbol]; hasAdditions {
		s.kindAdditions[referenceSymbol] = append(existingAdditions, kinds...)
	} else {
		s.kindAdditions[referenceSymbol] = kinds
	}
}

func (s *UpdatingClauseRewriter) TrackKindRemoval(referenceSymbol string, kinds ...graph.Kind) {
	s.kindReferenceSymbols[referenceSymbol] = struct{}{}

	if existingRemovals, hasRemovals := s.kindRemovals[referenceSymbol]; hasRemovals {
		s.kindRemovals[referenceSymbol] = append(existingRemovals, kinds...)
	} else {
		s.kindRemovals[referenceSymbol] = kinds
	}
}

func (s *UpdatingClauseRewriter) TrackPropertyAddition(referenceSymbol, propertyName string, value any) {
	s.propertyReferenceSymbols[referenceSymbol] = struct{}{}

	if existingAdditions, hasAdditions := s.propertyAdditions[referenceSymbol]; hasAdditions {
		existingAdditions[propertyName] = value
	} else {
		s.propertyAdditions[referenceSymbol] = map[string]any{
			propertyName: value,
		}
	}
}

func (s *UpdatingClauseRewriter) TrackPropertyRemoval(referenceSymbol, propertyName string) {
	s.propertyReferenceSymbols[referenceSymbol] = struct{}{}

	if existingRemovals, hasRemovals := s.propertyRemovals[referenceSymbol]; hasRemovals {
		s.propertyRemovals[referenceSymbol] = append(existingRemovals, propertyName)
	} else {
		s.propertyRemovals[referenceSymbol] = []string{propertyName}
	}
}
