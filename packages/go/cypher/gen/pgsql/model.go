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
	"github.com/jackc/pgtype"
	"github.com/specterops/bloodhound/cypher/model"
	pgModel "github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"time"
	"unsafe"
)

type annotationMapEntry struct {
	annotations []any
}

type AnnotationMap map[uintptr]*annotationMapEntry

func AnnotationMapPut[N any](annotationMap AnnotationMap, target *N, annotation any) {
	targetPtr := uintptr(unsafe.Pointer(target))

	if existingAnnotations, hasAnnotations := annotationMap[targetPtr]; hasAnnotations {
		existingAnnotations.annotations = append(existingAnnotations.annotations, annotation)
	} else {
		annotationMap[targetPtr] = &annotationMapEntry{
			annotations: []any{annotation},
		}
	}
}

func AnnotationMapGet[V, N any](annotationMap AnnotationMap, target *N) (V, bool) {
	var (
		empty     V
		targetPtr = uintptr(unsafe.Pointer(target))
	)

	if nodeAnnotations, hasNodeAnnotations := annotationMap[targetPtr]; hasNodeAnnotations {
		for _, annotation := range nodeAnnotations.annotations {
			if typedAnnotation, typeOK := annotation.(V); typeOK {
				return typedAnnotation, true
			}
		}
	}

	return empty, false
}

const (
	OperatorJSONBFieldExists    model.Operator = "?"
	OperatorLike                model.Operator = "like"
	OperatorLikeCaseInsensitive model.Operator = "ilike"
)

type NodeKindsReference struct {
	Variable model.Expression
}

func NewNodeKindsReference(ref *AnnotatedVariable) *NodeKindsReference {
	return &NodeKindsReference{
		Variable: ref,
	}
}

type EdgeKindReference struct {
	Variable model.Expression
}

func NewEdgeKindReference(ref *AnnotatedVariable) *EdgeKindReference {
	return &EdgeKindReference{
		Variable: ref,
	}
}

type Delete struct {
	Binding    *AnnotatedVariable
	NodeDelete bool
	EdgeDelete bool
}

func NewDelete() *Delete {
	return &Delete{
		NodeDelete: false,
		EdgeDelete: false,
	}
}

func (s *Delete) IsMixed() bool {
	return s.NodeDelete && s.EdgeDelete
}

func (s *Delete) Table() string {
	if s.NodeDelete {
		return pgModel.NodeTable
	}

	if s.EdgeDelete {
		return pgModel.EdgeTable
	}

	return ""
}

type PropertyMutation struct {
	Reference *PropertiesReference
	Additions *AnnotatedParameter
	Removals  *AnnotatedParameter
}

type KindMutation struct {
	Variable  *AnnotatedVariable
	Additions *AnnotatedParameter
	Removals  *AnnotatedParameter
}

type UpdatingClauseRewriter struct {
	kindMapper               KindMapper
	binder                   *Binder
	deletion                 *Delete
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
		deletion:                 NewDelete(),
		propertyReferenceSymbols: map[string]struct{}{},
		propertyAdditions:        map[string]map[string]any{},
		propertyRemovals:         map[string][]string{},
		kindReferenceSymbols:     map[string]struct{}{},
		kindRemovals:             map[string][]graph.Kind{},
		kindAdditions:            map[string][]graph.Kind{},
	}
}

func (s *UpdatingClauseRewriter) newPropertyMutation(symbol string) (*PropertyMutation, error) {
	if annotatedVariable, isBound := s.binder.LookupVariable(symbol); !isBound {
		return nil, fmt.Errorf("mutation variable reference %s is not bound", symbol)
	} else {
		return &PropertyMutation{
			Reference: &PropertiesReference{
				Reference: annotatedVariable,
			},
		}, nil
	}
}

func (s *UpdatingClauseRewriter) newKindMutation(symbol string) (*KindMutation, error) {
	if annotatedVariable, isBound := s.binder.LookupVariable(symbol); !isBound {
		return nil, fmt.Errorf("mutation variable reference %s is not bound", symbol)
	} else {
		return &KindMutation{
			Variable: annotatedVariable,
		}, nil
	}
}

func (s *UpdatingClauseRewriter) ToUpdatingClause() ([]model.Expression, error) {
	var updatingClauses []model.Expression

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

func (s *UpdatingClauseRewriter) rewriteDeleteClause(singlePartQuery *model.SinglePartQuery, deleteClause *model.Delete) error {
	for _, deleteStatementExpression := range deleteClause.Expressions {
		switch typedExpression := deleteStatementExpression.(type) {
		case *AnnotatedVariable:
			switch typedExpression.Type {
			case Node:
				if s.deletion.NodeDelete {
					return fmt.Errorf("multiple node delete statements are not supported")
				}

				s.deletion.Binding = typedExpression
				s.deletion.NodeDelete = true

			case Edge:
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
			var additionalWhereClauses []model.Expression

			for _, pattern := range matchClause.Pattern {
				if len(pattern.PatternElements) <= 1 {
					// This pattern does not have a relationship and therefore no joining criteria is required
					continue
				}

				for idx, patternElement := range pattern.PatternElements {
					if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
						var (
							lastNode   = idx+1 >= len(pattern.PatternElements)
							relBinding *AnnotatedVariable
							direction  graph.Direction
						)

						if !lastNode {
							// Look forward to the next relationship pattern
							relPattern, _ := pattern.PatternElements[idx+1].AsRelationshipPattern()
							direction = relPattern.Direction

							switch typedBinding := relPattern.Binding.(type) {
							case *AnnotatedVariable:
								relBinding = typedBinding
							default:
								return fmt.Errorf("unexpected variable for relationship pattern binding: %T", relPattern.Binding)
							}
						} else {
							// Look backward to the last relationship pattern
							relPattern, _ := pattern.PatternElements[idx-1].AsRelationshipPattern()
							direction, _ = relPattern.Direction.Reverse()

							switch typedBinding := relPattern.Binding.(type) {
							case *AnnotatedVariable:
								relBinding = typedBinding
							default:
								return fmt.Errorf("unexpected variable for relationship pattern binding: %T", relPattern.Binding)
							}
						}

						switch direction {
						case graph.DirectionInbound:
							bindingCopy := Copy(relBinding)
							bindingCopy.Symbol += ".end_id"

							additionalWhereClauses = append(additionalWhereClauses, model.NewComparison(
								model.NewSimpleFunctionInvocation(cypherIdentityFunction, nodePattern.Binding),
								model.OperatorEquals,
								bindingCopy,
							))

						case graph.DirectionOutbound:
							bindingCopy := Copy(relBinding)
							bindingCopy.Symbol += ".start_id"

							additionalWhereClauses = append(additionalWhereClauses, model.NewComparison(
								model.NewSimpleFunctionInvocation(cypherIdentityFunction, nodePattern.Binding),
								model.OperatorEquals,
								bindingCopy,
							))

						default:
							return fmt.Errorf("invalid pattern direction: %d", direction)
						}
					}
				}
			}

			if len(additionalWhereClauses) > 0 {
				additionalWhereClause := model.NewConjunction(additionalWhereClauses...)

				if matchClause.Where == nil {
					matchClause.Where = model.NewWhere()
				}

				if len(matchClause.Where.Expressions) > 0 {
					matchClause.Where.Expressions = []model.Expression{
						model.NewConjunction(append(matchClause.Where.Expressions, additionalWhereClause)...),
					}
				} else {
					matchClause.Where.Add(additionalWhereClause)
				}
			}
		}
	}

	return nil
}

func (s *UpdatingClauseRewriter) RewriteUpdatingClauses(singlePartQuery *model.SinglePartQuery) error {
	for _, updatingClause := range singlePartQuery.UpdatingClauses {
		typedUpdatingClause, isUpdatingClause := updatingClause.(*model.UpdatingClause)

		if !isUpdatingClause {
			return fmt.Errorf("unexpected type for updating clause: %T", updatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *model.Create:
			return fmt.Errorf("create unsupported")

		case *model.Delete:
			if err := s.rewriteDeleteClause(singlePartQuery, typedClause); err != nil {
				return err
			}

		case *model.Set:
			for _, setItem := range typedClause.Items {
				switch leftHandOperand := setItem.Left.(type) {
				case *model.Variable:
					switch rightHandOperand := setItem.Right.(type) {
					case graph.Kinds:
						s.TrackKindAddition(leftHandOperand.Symbol, rightHandOperand...)

					default:
						return fmt.Errorf("unexpected right side operand type %T for kind setter", setItem.Right)
					}

				case *model.PropertyLookup:
					switch setItem.Operator {
					case model.OperatorAssignment:
						var (
							// TODO: Type negotiation
							referenceSymbol = leftHandOperand.Atom.(*model.Variable).Symbol
							propertyName    = leftHandOperand.Symbols[0]
						)

						switch rightHandOperand := setItem.Right.(type) {
						case *model.Literal:
							// TODO: Negotiate null literals
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						case *AnnotatedLiteral:
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						case *model.Parameter:
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						case *AnnotatedParameter:
							s.TrackPropertyAddition(referenceSymbol, propertyName, rightHandOperand.Value)

						default:
							return fmt.Errorf("unexpected right side operand type %T for property setter", setItem.Right)
						}

					default:
						return fmt.Errorf("unsupported assignment operator: %s", setItem.Operator)
					}
				}
			}

		case *model.Remove:
			for _, removeItem := range typedClause.Items {
				if removeItem.KindMatcher != nil {
					if kindMatcher, typeOK := removeItem.KindMatcher.(*model.KindMatcher); !typeOK {
						return fmt.Errorf("unexpected remove item kind matcher expression: %T", removeItem.KindMatcher)
					} else if kindMatcherReference, typeOK := kindMatcher.Reference.(*model.Variable); !typeOK {
						return fmt.Errorf("unexpected remove matcher reference expression: %T", kindMatcher.Reference)
					} else {
						s.TrackKindRemoval(kindMatcherReference.Symbol, kindMatcher.Kinds...)
					}
				}

				if removeItem.Property != nil {
					var (
						// TODO: Type negotiation
						referenceSymbol = removeItem.Property.Atom.(*model.Variable).Symbol
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

type AnnotatedKindMatcher struct {
	model.KindMatcher
	Type DataType
}

func NewAnnotatedKindMatcher(kindMatcher *model.KindMatcher, dataType DataType) *AnnotatedKindMatcher {
	return &AnnotatedKindMatcher{
		KindMatcher: *kindMatcher,
		Type:        dataType,
	}
}

func (s *AnnotatedKindMatcher) copy() *AnnotatedKindMatcher {
	return &AnnotatedKindMatcher{
		KindMatcher: model.KindMatcher{
			Reference: s.Reference,
			Kinds:     s.Kinds,
		},
		Type: s.Type,
	}
}

type AnnotatedParameter struct {
	model.Parameter
	Type DataType
}

func NewAnnotatedParameter(parameter *model.Parameter, dataType DataType) *AnnotatedParameter {
	return &AnnotatedParameter{
		Parameter: *parameter,
		Type:      dataType,
	}
}

type Entity struct {
	Binding *AnnotatedVariable
}

func NewEntity(variable *AnnotatedVariable) *Entity {
	return &Entity{
		Binding: variable,
	}
}

type AnnotatedVariable struct {
	model.Variable
	Type DataType
}

func NewAnnotatedVariable(variable *model.Variable, dataType DataType) *AnnotatedVariable {
	return &AnnotatedVariable{
		Variable: *variable,
		Type:     dataType,
	}
}

func (s *AnnotatedVariable) copy() *AnnotatedVariable {
	if s == nil {
		return nil
	}

	return &AnnotatedVariable{
		Variable: model.Variable{
			Symbol: s.Symbol,
		},
		Type: s.Type,
	}
}

type AnnotatedPropertyLookup struct {
	model.PropertyLookup
	Type DataType
}

func NewAnnotatedPropertyLookup(propertyLookup *model.PropertyLookup, dataType DataType) *AnnotatedPropertyLookup {
	return &AnnotatedPropertyLookup{
		PropertyLookup: *propertyLookup,
		Type:           dataType,
	}
}

type AnnotatedLiteral struct {
	model.Literal
	Type DataType
}

func NewAnnotatedLiteral(literal *model.Literal, dataType DataType) *AnnotatedLiteral {
	return &AnnotatedLiteral{
		Literal: *literal,
		Type:    dataType,
	}
}

func NewStringLiteral(value string) *AnnotatedLiteral {
	return NewAnnotatedLiteral(model.NewStringLiteral(value), Text)
}

type PropertiesReference struct {
	Reference *AnnotatedVariable
}

type Subquery struct {
	PatternElements []*model.PatternElement
	Filter          model.Expression
}

type SubQueryAnnotation struct {
	FilterExpression model.Expression
}

type SQLTypeAnnotation struct {
	Type DataType
}

func NewSQLTypeAnnotationFromExpression(expression model.Expression) (*SQLTypeAnnotation, error) {
	switch typedExpression := expression.(type) {
	case *model.Parameter:
		return NewSQLTypeAnnotationFromValue(typedExpression.Value)

	case *model.Literal:
		return NewSQLTypeAnnotationFromLiteral(typedExpression)

	case *model.ListLiteral:
		var expectedTypeAnnotation *SQLTypeAnnotation

		for _, listExpressionItem := range *typedExpression {
			if listExpressionItemLiteral, isLiteral := listExpressionItem.(*model.Literal); isLiteral {
				if literalTypeAnnotation, err := NewSQLTypeAnnotationFromLiteral(listExpressionItemLiteral); err != nil {
					return nil, err
				} else if expectedTypeAnnotation != nil && expectedTypeAnnotation.Type != literalTypeAnnotation.Type {
					return nil, fmt.Errorf("list literal contains mixed types")
				} else {
					expectedTypeAnnotation = literalTypeAnnotation
				}
			}
		}

		return expectedTypeAnnotation, nil

	default:
		return nil, fmt.Errorf("unsupported expression type %T for SQL type annotation", expression)
	}
}

func NewSQLTypeAnnotationFromLiteral(literal *model.Literal) (*SQLTypeAnnotation, error) {
	if literal.Null {
		return &SQLTypeAnnotation{
			Type: Null,
		}, nil
	}

	return NewSQLTypeAnnotationFromValue(literal.Value)
}

func NewSQLTypeAnnotationFromValue(value any) (*SQLTypeAnnotation, error) {
	switch typedValue := value.(type) {
	case []uint16, []int16, pgtype.Int2Array:
		return &SQLTypeAnnotation{
			Type: Int2Array,
		}, nil

	case []uint32, []int32, []graph.ID, pgtype.Int4Array:
		return &SQLTypeAnnotation{
			Type: Int4Array,
		}, nil

	case []uint64, []int64, pgtype.Int8Array:
		return &SQLTypeAnnotation{
			Type: Int8Array,
		}, nil

	case uint16, int16:
		return &SQLTypeAnnotation{
			Type: Int2,
		}, nil

	case uint32, int32, graph.ID:
		return &SQLTypeAnnotation{
			Type: Int4,
		}, nil

	case uint, int, uint64, int64:
		return &SQLTypeAnnotation{
			Type: Int8,
		}, nil

	case float32:
		return &SQLTypeAnnotation{
			Type: Float4,
		}, nil

	case []float32:
		return &SQLTypeAnnotation{
			Type: Float4Array,
		}, nil

	case float64:
		return &SQLTypeAnnotation{
			Type: Float8,
		}, nil

	case []float64:
		return &SQLTypeAnnotation{
			Type: Float8Array,
		}, nil

	case bool:
		return &SQLTypeAnnotation{
			Type: Boolean,
		}, nil

	case string:
		return &SQLTypeAnnotation{
			Type: Text,
		}, nil

	case time.Time:
		return &SQLTypeAnnotation{
			Type: TimestampWithTimeZone,
		}, nil

	case pgtype.JSONB:
		return &SQLTypeAnnotation{
			Type: JSONB,
		}, nil

	case []string, pgtype.TextArray:
		return &SQLTypeAnnotation{
			Type: TextArray,
		}, nil

	case *model.ListLiteral:
		return NewSQLTypeAnnotationFromExpression(typedValue)

	default:
		return nil, fmt.Errorf("literal type %T is not supported", value)
	}
}
