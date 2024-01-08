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
	"github.com/specterops/bloodhound/cypher/analyzer"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/query"
	"strconv"
	"strings"
)

func GetSymbol(expression model.Expression) (string, error) {
	switch typedExpression := expression.(type) {
	case *model.PatternElement:
		if nodePattern, isNodePattern := typedExpression.AsNodePattern(); isNodePattern {
			if nodePattern.Binding != nil {
				return GetSymbol(nodePattern.Binding)
			}
		} else if relationshipPattern, isRelationshipPattern := typedExpression.AsRelationshipPattern(); isRelationshipPattern {
			if relationshipPattern.Binding != nil {
				return GetSymbol(relationshipPattern.Binding)
			}
		}

	case *model.PatternPart:
		if typedExpression.Binding != nil {
			return GetSymbol(typedExpression.Binding)
		}

	case *model.Variable:
		return typedExpression.Symbol, nil

	case *AnnotatedVariable:
		return typedExpression.Symbol, nil

	default:
		return "", fmt.Errorf("unable to source symbol from expression type %T", expression)
	}

	return "", nil
}

type Binder struct {
	parameters          map[string]*AnnotatedParameter
	bindingTypeMappings map[string]DataType
	aliases             map[string]string
	patternBindings     map[string]struct{}
	syntheticBindings   map[string]struct{}
	nextParameterID     int
	nextBindingID       int
}

func NewBinder() *Binder {
	return &Binder{
		parameters:          map[string]*AnnotatedParameter{},
		bindingTypeMappings: map[string]DataType{},
		aliases:             map[string]string{},
		patternBindings:     map[string]struct{}{},
		syntheticBindings:   map[string]struct{}{},
		nextParameterID:     0,
		nextBindingID:       0,
	}
}

func (s *Binder) Parameters() map[string]any {
	parametersCopy := make(map[string]any, len(s.parameters))

	for _, parameter := range s.parameters {
		parametersCopy[parameter.Symbol] = parameter.Value
	}

	return parametersCopy
}

func (s *Binder) BindVariable(variable *model.Variable, bindingType DataType) *AnnotatedVariable {
	s.bindingTypeMappings[variable.Symbol] = bindingType
	return NewAnnotatedVariable(variable, bindingType)
}

func (s *Binder) BindPatternVariable(variable *model.Variable, bindingType DataType) *AnnotatedVariable {
	s.patternBindings[variable.Symbol] = struct{}{}
	return s.BindVariable(variable, bindingType)
}

func (s *Binder) BindingType(binding string) (DataType, bool) {
	if bindingType, isBound := s.bindingTypeMappings[binding]; isBound {
		return bindingType, isBound
	}

	return UnknownDataType, false
}

func (s *Binder) LookupVariable(symbol string) (*AnnotatedVariable, bool) {
	if dataType, isBound := s.BindingType(symbol); isBound {
		return NewAnnotatedVariable(model.NewVariableWithSymbol(symbol), dataType), true
	}

	return nil, false
}

func (s *Binder) IsSynthetic(binding string) bool {
	_, isSynthetic := s.syntheticBindings[binding]
	return isSynthetic
}

func (s *Binder) IsPatternBinding(binding string) bool {
	_, isPatternBinding := s.patternBindings[binding]
	return isPatternBinding
}

func (s *Binder) IsBound(binding string) bool {
	_, isBound := s.bindingTypeMappings[binding]
	return isBound
}

func (s *Binder) NewBinding(prefix string) string {
	// Spin to win
	for {
		binding := prefix + strconv.Itoa(s.nextBindingID)
		s.nextBindingID++

		if !s.IsBound(binding) {
			s.syntheticBindings[binding] = struct{}{}
			return binding
		}
	}
}

func (s *Binder) NewAnnotatedVariable(prefix string, bindingType DataType) *AnnotatedVariable {
	return s.BindVariable(s.NewVariable(prefix), bindingType)
}

func (s *Binder) NewVariable(prefix string) *model.Variable {
	return model.NewVariableWithSymbol(s.NewBinding(prefix))
}

func (s *Binder) NewParameterSymbol() string {
	nextParameterSymbol := "p" + strconv.Itoa(s.nextParameterID)
	s.nextParameterID++

	return nextParameterSymbol
}

func (s *Binder) NewParameter(value any) (*AnnotatedParameter, error) {
	var (
		parameterSymbol = s.NewParameterSymbol()
	)

	if parameterTypeAnnotation, err := NewSQLTypeAnnotationFromValue(value); err != nil {
		return nil, err
	} else {
		parameter := NewAnnotatedParameter(model.NewParameter(parameterSymbol, value), parameterTypeAnnotation.Type)

		// Record the parameter's value for mapping to the query later
		s.parameters[parameterSymbol] = parameter
		return parameter, nil
	}
}

func (s *Binder) NewAlias(originalSymbol string, alias *model.Variable) *AnnotatedVariable {
	s.aliases[originalSymbol] = alias.Symbol

	if originalBindingType, isBound := s.bindingTypeMappings[originalSymbol]; isBound {
		return s.BindVariable(alias, originalBindingType)
	}

	return s.BindVariable(alias, UnknownDataType)
}

func (s *Binder) Scan(regularQuery *model.RegularQuery) error {
	if err := analyzer.Analyze(regularQuery, func(analyzerInst *analyzer.Analyzer) {
		// TODO: auto parameterize all literals?
		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, node *model.Parameter) error {
			// Rewrite all parameter symbols and collect their values
			if rewrittenParameter, err := s.NewParameter(node.Value); err != nil {
				return err
			} else {
				return rewrite(stack, node, rewrittenParameter)
			}
		})

		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, patternElement *model.PatternElement) error {
			// Eagerly bind all ReadingClause pattern elements to simplify referencing when crafting SQL join statements
			if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
				if nodePattern.Binding == nil {
					nodePattern.Binding = s.NewAnnotatedVariable("n", Node)
				} else if bindingVariable, typeOK := nodePattern.Binding.(*model.Variable); !typeOK {
					return fmt.Errorf("expected variable for node pattern binding but got: %T", nodePattern.Binding)
				} else if _, isPatternPredicate := stack.Trunk().(*model.PatternPredicate); isPatternPredicate {
					nodePattern.Binding = s.BindVariable(bindingVariable, Node)
				} else {
					nodePattern.Binding = s.BindPatternVariable(bindingVariable, Node)
				}
			} else {
				relationshipPattern, _ := patternElement.AsRelationshipPattern()

				if relationshipPattern.Binding == nil {
					relationshipPattern.Binding = s.NewAnnotatedVariable("e", Edge)
				} else if bindingVariable, typeOK := relationshipPattern.Binding.(*model.Variable); !typeOK {
					return fmt.Errorf("expected variable for relationship pattern binding but got: %T", relationshipPattern.Binding)
				} else if _, isPatternPredicate := stack.Trunk().(*model.PatternPredicate); isPatternPredicate {
					relationshipPattern.Binding = s.BindVariable(bindingVariable, Edge)
				} else {
					relationshipPattern.Binding = s.BindPatternVariable(bindingVariable, Edge)
				}
			}

			return nil
		})

		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, node *model.ProjectionItem) error {
			if bindingVariable, isVariable := node.Binding.(*model.Variable); node.Binding != nil && isVariable {
				if projectionVariable, isVariable := node.Expression.(*model.Variable); isVariable {
					node.Binding = s.NewAlias(projectionVariable.Symbol, bindingVariable)
				}
			}

			return nil
		})

		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, node *model.Delete) error {
			for idx, expression := range node.Expressions {
				switch typedExpression := expression.(type) {
				case *model.Variable:
					if annotatedVariable, isAnnotated := s.LookupVariable(typedExpression.Symbol); !isAnnotated {
						return fmt.Errorf("unable to look up type annotation for variable reference: %s", typedExpression.Symbol)
					} else {
						node.Expressions[idx] = annotatedVariable
					}
				}
			}

			return nil
		})
	}, CollectPGSQLTypes); err != nil {
		return err
	}

	return nil
}

type Translator struct {
	builder      *strings.Builder
	Bindings     *Binder
	kindMapper   KindMapper
	regularQuery *model.RegularQuery
}

func NewTranslator(kindMapper KindMapper, bindings *Binder, regularQuery *model.RegularQuery) *Translator {
	return &Translator{
		builder:      &strings.Builder{},
		kindMapper:   kindMapper,
		Bindings:     bindings,
		regularQuery: regularQuery,
	}
}

func (s *Translator) rewriteUpdatingClauses(stack *model.WalkStack, singlePartQuery *model.SinglePartQuery) error {
	return NewUpdateClauseRewriter(s.Bindings, s.kindMapper).RewriteUpdatingClauses(singlePartQuery)
}

func (s *Translator) liftNodePatternCriteria(stack *model.WalkStack, nodePattern *model.NodePattern) ([]model.Expression, error) {
	var criteria []model.Expression

	if nodePattern.Binding == nil {
		nodePattern.Binding = s.Bindings.NewVariable("n")
	}

	if len(nodePattern.Kinds) > 0 {
		kindMatcher := model.NewKindMatcher(nodePattern.Binding, nodePattern.Kinds)
		criteria = append(criteria, NewAnnotatedKindMatcher(kindMatcher, Node))
	}

	if nodePattern.Properties != nil {
		nodePropertyMatchers := nodePattern.Properties.(*model.Properties)

		if nodePropertyMatchers.Parameter != nil {
			return nil, fmt.Errorf("unable to translate property matcher paramter for node %s", nodePattern.Binding)
		}

		for propertyName, matcherValue := range nodePropertyMatchers.Map {
			if bindingVariable, typeOK := nodePattern.Binding.(*AnnotatedVariable); !typeOK {
				return nil, fmt.Errorf("unexpected node pattern binding type for node pattern: %T", nodePattern.Binding)
			} else {
				propertyLookup := model.NewPropertyLookup(bindingVariable.Symbol, propertyName)

				if annotation, err := NewSQLTypeAnnotationFromExpression(matcherValue); err != nil {
					return nil, err
				} else {
					criteria = append(criteria, model.NewComparison(
						NewAnnotatedPropertyLookup(propertyLookup, annotation.Type),
						model.OperatorEquals,
						matcherValue,
					))
				}
			}
		}
	}

	return criteria, nil
}

func (s *Translator) liftRelationshipPatternCriteria(stack *model.WalkStack, relationshipPattern *model.RelationshipPattern) ([]model.Expression, error) {
	var criteria []model.Expression

	if relationshipPattern.Binding == nil {
		relationshipPattern.Binding = s.Bindings.NewVariable("e")
	}

	if len(relationshipPattern.Kinds) > 0 {
		kindMatcher := model.NewKindMatcher(relationshipPattern.Binding, relationshipPattern.Kinds)
		criteria = append(criteria, NewAnnotatedKindMatcher(kindMatcher, Edge))
	}

	if relationshipPattern.Properties != nil {
		edgePropertyMatchers := relationshipPattern.Properties.(*model.Properties)

		if edgePropertyMatchers.Parameter != nil {
			return nil, fmt.Errorf("unable to translate property matcher paramter for edge %s", relationshipPattern.Binding)
		}

		for propertyName, matcherValue := range edgePropertyMatchers.Map {
			if bindingVariable, typeOK := relationshipPattern.Binding.(*AnnotatedVariable); !typeOK {
				return nil, fmt.Errorf("unexpected relationship pattern binding type: %T", relationshipPattern.Binding)
			} else {
				propertyLookup := model.NewPropertyLookup(bindingVariable.Symbol, propertyName)

				if annotation, err := NewSQLTypeAnnotationFromExpression(matcherValue); err != nil {
					return nil, err
				} else {
					criteria = append(criteria, model.NewComparison(
						NewAnnotatedPropertyLookup(propertyLookup, annotation.Type),
						model.OperatorEquals,
						matcherValue,
					))
				}
			}
		}
	}

	return criteria, nil
}

func (s *Translator) liftPatternElementCriteria(stack *model.WalkStack, patternElement *model.PatternElement) ([]model.Expression, error) {
	if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
		return s.liftNodePatternCriteria(stack, nodePattern)
	}

	relationshipPattern, _ := patternElement.AsRelationshipPattern()
	return s.liftRelationshipPatternCriteria(stack, relationshipPattern)
}

func (s *Translator) translatePatternPredicates(stack *model.WalkStack, patternPredicate *model.PatternPredicate) error {
	var (
		subqueryFilters []model.Expression
		subquery        = &Subquery{
			PatternElements: patternPredicate.PatternElements,
		}
	)

	for _, patternElement := range subquery.PatternElements {
		if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
			// Is the node pattern bound to a variable and was that variable bound earlier in the AST?
			if bindingVariable, typeOK := nodePattern.Binding.(*AnnotatedVariable); !typeOK {
				return fmt.Errorf("unexpected node pattern binding type for pattern predicate: %T", nodePattern.Binding)
			} else if nodePattern.Binding != nil && !s.Bindings.IsSynthetic(bindingVariable.Symbol) && s.Bindings.IsPatternBinding(bindingVariable.Symbol) {
				// Since this pattern element is bound to a pre-existing referenced pattern element we have to match
				// against it by its identity
				var (
					oldBinding = nodePattern.Binding
					newBinding = s.Bindings.NewAnnotatedVariable("n", bindingVariable.Type)
				)

				nodePattern.Binding = newBinding
				subqueryFilters = append(subqueryFilters, model.NewComparison(
					model.NewSimpleFunctionInvocation(
						cypherIdentityFunction,
						oldBinding,
					),
					model.OperatorEquals,
					model.NewSimpleFunctionInvocation(
						cypherIdentityFunction,
						newBinding,
					),
				))
			}

			if criteria, err := s.liftNodePatternCriteria(stack, nodePattern); err != nil {
				return err
			} else {
				subqueryFilters = append(subqueryFilters, criteria...)
			}
		} else {
			relationshipPattern, _ := patternElement.AsRelationshipPattern()

			// Is the relationship pattern bound to a variable and was that variable bound earlier in the AST?
			if bindingVariable, typeOK := relationshipPattern.Binding.(*AnnotatedVariable); !typeOK {
				return fmt.Errorf("unexpected relationship pattern binding type: %T", relationshipPattern.Binding)
			} else if relationshipPattern.Binding != nil && !s.Bindings.IsSynthetic(bindingVariable.Symbol) && s.Bindings.IsPatternBinding(bindingVariable.Symbol) {
				// Since this pattern element is bound to a pre-existing referenced pattern element we have to match
				// against it by its identity
				var (
					oldBinding = relationshipPattern.Binding
					newBinding = s.Bindings.NewAnnotatedVariable("e", bindingVariable.Type)
				)

				relationshipPattern.Binding = newBinding
				subqueryFilters = append(subqueryFilters, model.NewComparison(
					model.NewSimpleFunctionInvocation(
						cypherIdentityFunction,
						oldBinding,
					),
					model.OperatorEquals,
					model.NewSimpleFunctionInvocation(
						cypherIdentityFunction,
						newBinding,
					),
				))
			}

			if criteria, err := s.liftRelationshipPatternCriteria(stack, relationshipPattern); err != nil {
				return err
			} else {
				subqueryFilters = append(subqueryFilters, criteria...)
			}
		}

	}

	if len(subqueryFilters) > 0 {
		subquery.Filter = model.NewConjunction(subqueryFilters...)

		return rewrite(stack, patternPredicate, subquery)
	}

	return nil
}

func (s *Translator) liftMatchCriteria(stack *model.WalkStack, match *model.Match) error {
	var additionalCriteria []model.Expression

	for _, patternPart := range match.Pattern {
		for _, patternElement := range patternPart.PatternElements {
			if patternElementCriteria, err := s.liftPatternElementCriteria(stack, patternElement); err != nil {
				return err
			} else {
				additionalCriteria = append(additionalCriteria, patternElementCriteria...)
			}
		}
	}

	if len(additionalCriteria) > 0 {
		if match.Where == nil {
			match.Where = model.NewWhere()
		}

		match.Where.Expressions = []model.Expression{
			model.NewConjunction(append(additionalCriteria, match.Where.Expressions...)...),
		}
	}

	return nil
}

func (s *Translator) annotateKindMatchers(stack *model.WalkStack, kindMatcher *model.KindMatcher) error {
	switch typedExpression := kindMatcher.Reference.(type) {
	case *AnnotatedVariable:
		return rewrite(stack, kindMatcher, NewAnnotatedKindMatcher(kindMatcher, typedExpression.Type))

	case *model.Variable:
		if dataType, hasBindingType := s.Bindings.BindingType(typedExpression.Symbol); !hasBindingType {
			return fmt.Errorf("unable to locate a binding type for variable %s", typedExpression.Symbol)
		} else {
			return rewrite(stack, kindMatcher, NewAnnotatedKindMatcher(kindMatcher, dataType))
		}

	default:
		return fmt.Errorf("unexpected kind matcher reference type %T", kindMatcher.Reference)
	}

	return nil
}

func (s *Translator) rewriteComparison(stack *model.WalkStack, comparison *model.Comparison) (bool, error) {
	// Is this a property lookup comparison?
	switch typedLeftOperand := comparison.Left.(type) {
	case *model.PropertyLookup:
		// Try to suss out if this is a property existence check
		if len(comparison.Partials) == 1 {
			comparisonPartial := comparison.Partials[0]

			switch typedRightHand := comparisonPartial.Right.(type) {
			case *model.Literal:
				if typedRightHand.Null {
					// This is a null check for a property and must be rewritten for SQL
					switch comparisonPartial.Operator {
					case model.OperatorIsNot:
						if leftOperandVariable, isVariable := typedLeftOperand.Atom.(*model.Variable); !isVariable {
							return false, fmt.Errorf("unexpected expression as left operand %T", typedLeftOperand.Atom)
						} else if leftOperandTypedVariable, isBound := s.Bindings.LookupVariable(leftOperandVariable.Symbol); !isBound {
							return false, fmt.Errorf("left operand varaible %s is not bound", leftOperandTypedVariable.Symbol)
						} else if err := rewrite(stack, comparison, model.NewComparison(
							&PropertiesReference{
								// TODO: Might need a copy?
								Reference: leftOperandTypedVariable,
							},
							OperatorJSONBFieldExists,
							NewStringLiteral(typedLeftOperand.Symbols[0]),
						)); err != nil {
							return false, err
						}

					case model.OperatorIs:
						if leftOperandVariable, isVariable := typedLeftOperand.Atom.(*model.Variable); !isVariable {
							return false, fmt.Errorf("unexpected expression as left operand %T", typedLeftOperand.Atom)
						} else if leftOperandTypedVariable, isBound := s.Bindings.LookupVariable(leftOperandVariable.Symbol); !isBound {
							return false, fmt.Errorf("left operand varaible %s is not bound", leftOperandTypedVariable.Symbol)
						} else if err := rewrite(stack, comparison, model.NewNegation(
							model.NewComparison(
								&PropertiesReference{
									Reference: leftOperandTypedVariable,
								},
								OperatorJSONBFieldExists,
								NewStringLiteral(typedLeftOperand.Symbols[0]),
							)),
						); err != nil {
							return false, err
						}
					}

					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (s *Translator) rewritePartialComparison(stack *model.WalkStack, partial *model.PartialComparison) error {
	switch partial.Operator {
	case model.OperatorIn:
		switch partial.Right.(type) {
		case *model.Parameter, *AnnotatedParameter:
			// When the "in" operator addresses right-hand parameter it must be rewritten as: "= any($param)"
			partial.Operator = model.OperatorEquals
			partial.Right = model.NewSimpleFunctionInvocation(pgsqlAnyFunction, partial.Right)
		}

	case model.OperatorStartsWith:
		// Replace this operator with the like operator
		partial.Operator = OperatorLike

		// If the right side isn't a string for any of these it's an error
		switch typedRightOperand := partial.Right.(type) {
		case *model.Literal:
			if stringValue, isString := typedRightOperand.Value.(string); !isString {
				return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
			} else {
				// Strip the wrapping single quotes first
				s.builder.Reset()
				s.builder.WriteString("'")
				s.builder.WriteString(stringValue[1 : len(stringValue)-1])
				s.builder.WriteString("%'")

				typedRightOperand.Value = s.builder.String()
			}

		case *AnnotatedParameter:
			if stringValue, isString := typedRightOperand.Value.(string); !isString {
				return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
			} else {
				// Parameters are raw values and have no quotes
				s.builder.Reset()
				s.builder.WriteString(stringValue)
				s.builder.WriteString("%")

				typedRightOperand.Value = s.builder.String()
			}

		default:
			return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
		}

	case model.OperatorContains:
		// Replace this operator with the like operator
		partial.Operator = OperatorLike

		// If the right side isn't a string for any of these it's an error
		switch typedRightOperand := partial.Right.(type) {
		case *model.Literal:
			if stringValue, isString := typedRightOperand.Value.(string); !isString {
				return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
			} else {
				// Strip the wrapping single quotes first
				s.builder.Reset()
				s.builder.WriteString("'%")
				s.builder.WriteString(stringValue[1 : len(stringValue)-1])
				s.builder.WriteString("%'")

				typedRightOperand.Value = s.builder.String()
			}

		case *AnnotatedParameter:
			if stringValue, isString := typedRightOperand.Value.(string); !isString {
				return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
			} else {
				// Parameters are raw values and have no quotes
				s.builder.Reset()
				s.builder.WriteString("%")
				s.builder.WriteString(stringValue)
				s.builder.WriteString("%")

				typedRightOperand.Value = s.builder.String()
			}

		default:
			return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
		}

	case model.OperatorEndsWith:
		// Replace this operator with the like operator
		partial.Operator = OperatorLike

		// If the right side isn't a string for any of these it's an error
		switch typedRightOperand := partial.Right.(type) {
		case *model.Literal:
			if stringValue, isString := typedRightOperand.Value.(string); !isString {
				return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
			} else {
				// Strip the wrapping single quotes first
				s.builder.Reset()
				s.builder.WriteString("'%")
				s.builder.WriteString(stringValue[1 : len(stringValue)-1])
				s.builder.WriteString("'")

				typedRightOperand.Value = s.builder.String()
			}

		case *AnnotatedParameter:
			if stringValue, isString := typedRightOperand.Value.(string); !isString {
				return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
			} else {
				// Parameters are raw values and have no quotes
				s.builder.Reset()
				s.builder.WriteString("%")
				s.builder.WriteString(stringValue)

				typedRightOperand.Value = s.builder.String()
			}

		default:
			return fmt.Errorf("string operator \"%s\" expects a string literal or parameter as its right opperand", partial.Operator.String())
		}
	}

	return nil
}

func (s *Translator) annotateComparisons(stack *model.WalkStack, comparison *model.Comparison) error {
	var typeAnnotation *SQLTypeAnnotation

	if rewritten, err := s.rewriteComparison(stack, comparison); err != nil {
		return err
	} else if rewritten {
		return nil
	}

	for comparisonWalkStack := []model.Expression{comparison}; len(comparisonWalkStack) > 0; {
		next := comparisonWalkStack[len(comparisonWalkStack)-1]
		comparisonWalkStack = comparisonWalkStack[:len(comparisonWalkStack)-1]

		switch typedNode := next.(type) {
		case *model.Comparison:
			comparisonWalkStack = append(comparisonWalkStack, typedNode.Left)

			for _, partial := range typedNode.Partials {
				comparisonWalkStack = append(comparisonWalkStack, partial)
			}

		case *model.PartialComparison:
			if err := s.rewritePartialComparison(stack, typedNode); err != nil {
				return err
			}

			comparisonWalkStack = append(comparisonWalkStack, typedNode.Right)

		case *AnnotatedParameter:
			if typeAnnotation == nil {
				typeAnnotation = &SQLTypeAnnotation{
					Type: typedNode.Type,
				}
			} else if typeAnnotation.Type != typedNode.Type {
				return fmt.Errorf("comparison contains mixed types: %s and %s", typeAnnotation.Type, typedNode.Type)
			}

		case *model.Literal:
			if literalTypeAnnotation, err := NewSQLTypeAnnotationFromExpression(typedNode); err != nil {
				return err
			} else if typeAnnotation == nil {
				typeAnnotation = literalTypeAnnotation
			} else if typeAnnotation.Type != literalTypeAnnotation.Type {
				return fmt.Errorf("comparison contains mixed types: %s and %s", typeAnnotation.Type, literalTypeAnnotation.Type)
			}

		case *model.FunctionInvocation:
			var functionInvocationTypeAnnotation *SQLTypeAnnotation

			switch typedNode.Name {
			case cypherDateFunction:
				functionInvocationTypeAnnotation = &SQLTypeAnnotation{
					Type: Date,
				}

			case cypherTimeFunction:
				functionInvocationTypeAnnotation = &SQLTypeAnnotation{
					Type: TimeWithTimeZone,
				}

			case cypherLocalTimeFunction:
				functionInvocationTypeAnnotation = &SQLTypeAnnotation{
					Type: TimeWithoutTimeZone,
				}

			case cypherDateTimeFunction:
				functionInvocationTypeAnnotation = &SQLTypeAnnotation{
					Type: TimestampWithTimeZone,
				}

			case cypherLocalDateTimeFunction:
				functionInvocationTypeAnnotation = &SQLTypeAnnotation{
					Type: TimestampWithoutTimeZone,
				}

			case cypherDurationFunction:
				functionInvocationTypeAnnotation = &SQLTypeAnnotation{
					Type: Interval,
				}

			default:
				// If we couldn't figure out a type from the function name then inspect the function's argument list
				comparisonWalkStack = append(comparisonWalkStack, typedNode.Arguments...)
			}

			// If there was a function invocation type, check to validate that we're not producing mixed type
			// annotations for the comparison
			if functionInvocationTypeAnnotation != nil {
				if typeAnnotation == nil {
					typeAnnotation = functionInvocationTypeAnnotation
				} else if typeAnnotation.Type != functionInvocationTypeAnnotation.Type {
					return fmt.Errorf("comparison contains mixed types: %s and %s", typeAnnotation.Type, functionInvocationTypeAnnotation.Type)
				}
			}
		}
	}

	if typeAnnotation != nil {
		if leftHandPropertyLookup, typeOK := comparison.Left.(*model.PropertyLookup); typeOK {
			leftOperandType := typeAnnotation.Type

			// When adding type annotations to property lookups assume that comparisons against arrays are
			// performing a contains operation. This is probably a bad assumption, but I can't think of a
			// better heuristic at the moment.
			if typeAnnotation.Type.IsArrayType() {
				if baseType, err := typeAnnotation.Type.ArrayBaseType(); err != nil {
					return err
				} else {
					leftOperandType = baseType
				}
			}

			// Rewrite the left operand so that the property lookup is correctly type annotated
			comparison.Left = NewAnnotatedPropertyLookup(leftHandPropertyLookup, leftOperandType)

			for _, partialComparison := range comparison.Partials {
				if rightHandPropertyLookup, typeOK := partialComparison.Right.(*model.PropertyLookup); typeOK {
					annotatedPropertyLookup := NewAnnotatedPropertyLookup(rightHandPropertyLookup, typeAnnotation.Type)

					if err := rewrite(stack, partialComparison.Right, annotatedPropertyLookup); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *Translator) rewriteNegations(stack *model.WalkStack, negation *model.Negation) error {
	// Wrap negations that contain a list of expressions in a parenthetical expression to ensure that evaluation
	// happens as intended by the author of the query
	if _, isExpressionList := negation.Expression.(model.ExpressionList); isExpressionList {
		negation.Expression = model.NewParenthetical(negation.Expression)
	}

	return nil
}

func (s *Translator) rewriteStringNegations(stack *model.WalkStack, negation *model.Negation) error {
	var rewritten any

	// If this is a negation then we should check to see if it's a comparison
	switch comparison := negation.Expression.(type) {
	case *model.Comparison:
		firstPartial := comparison.FirstPartial()

		// If the negated expression is a comparison check to see if it's a string comparison. This is done since
		// comparison semantics for strings regarding `null` has edge cases that must be accounted for
		switch firstPartial.Operator {
		case model.OperatorStartsWith, model.OperatorEndsWith, model.OperatorContains:
			// Rewrite this comparison is a disjunction of the negation and a follow-on comparison to handle null
			// checks
			rewritten = &model.Parenthetical{
				Expression: model.NewDisjunction(
					negation,
					model.NewComparison(comparison.Left, model.OperatorIs, query.Literal(nil)),
				),
			}
		}
	}

	// If we rewrote this element, replace it
	if rewritten != nil {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			for idx, expression := range typedParent.GetAll() {
				if expression == negation {
					typedParent.Replace(idx, rewritten)
					break
				}
			}

		default:
			return fmt.Errorf("unable to replace rewritten string negation operation for parent type %T", stack.Trunk())
		}
	}

	return nil
}

func (s *Translator) annotatePatternBindings(stack *model.WalkStack, patternPart *model.PatternPart) error {
	// Binding of pattern parts is optional since we do not need them to perform joins but if there is a
	// binding present we need to annotate its type
	if patternPart.Binding != nil {
		if bindingVariable, typeOK := patternPart.Binding.(*model.Variable); !typeOK {
			return fmt.Errorf("expected variable for pattern part binding but got: %T", patternPart.Binding)
		} else {
			patternPart.Binding = NewAnnotatedVariable(bindingVariable, Path)
		}
	}

	return nil
}

func (s *Translator) rewriteFunctionInvocations(stack *model.WalkStack, functionInvocation *model.FunctionInvocation) error {
	switch functionInvocation.Name {
	case cypherNodeLabelsFunction:
		switch typedArgument := functionInvocation.Arguments[0].(type) {
		case *model.Variable:
			return rewrite(stack, functionInvocation, NewNodeKindsReference(NewAnnotatedVariable(typedArgument, Node)))

		case *AnnotatedVariable:
			return rewrite(stack, functionInvocation, NewNodeKindsReference(typedArgument))

		default:
			return fmt.Errorf("expected a variable as the first argument in %s function", functionInvocation.Name)
		}

	case cypherEdgeTypeFunction:
		switch typedArgument := functionInvocation.Arguments[0].(type) {
		case *model.Variable:
			return rewrite(stack, functionInvocation, NewEdgeKindReference(NewAnnotatedVariable(typedArgument, Edge)))

		case *AnnotatedVariable:
			return rewrite(stack, functionInvocation, NewEdgeKindReference(typedArgument))

		default:
			return fmt.Errorf("expected a variable as the first argument in %s function", functionInvocation.Name)
		}

	case cypherToLowerFunction:
		switch typedArgument := functionInvocation.Arguments[0].(type) {
		case *model.PropertyLookup:
			functionInvocation.Arguments[0] = NewAnnotatedPropertyLookup(typedArgument, Text)
		}
	}

	return nil
}

func (s *Translator) annotateProjectionItems(stack *model.WalkStack, projectionItem *model.ProjectionItem) error {
	switch typedExpression := projectionItem.Expression.(type) {
	case *model.Variable:
		if bindingType, isBound := s.Bindings.BindingType(typedExpression.Symbol); !isBound {
			return fmt.Errorf("variable %s for projection item is not bound", typedExpression.Symbol)
		} else {
			projectionItem.Expression = NewEntity(NewAnnotatedVariable(typedExpression, bindingType))

			// Set projection item binding to the variable reference if there's no binding present
			if projectionItem.Binding == nil {
				projectionItem.Binding = NewAnnotatedVariable(typedExpression, bindingType)
			}
		}
	}

	return nil
}

func (s *Translator) validatePropertyLookups(stack *model.WalkStack, propertyLookup *model.PropertyLookup) error {
	if len(propertyLookup.Symbols) != 1 {
		return fmt.Errorf("expected a single-depth propertly lookup")
	}

	return nil
}
func (s *Translator) removeEmptyExpressionLists(stack *model.WalkStack, element model.Expression) error {
	var (
		shouldRemove  = false
		shouldReplace = false

		replacementExpression model.Expression
	)

	switch typedElement := element.(type) {
	case model.ExpressionList:
		shouldRemove = typedElement.Len() == 0

	case *model.Parenthetical:
		switch typedParentheticalElement := typedElement.Expression.(type) {
		case model.ExpressionList:
			numExpressions := typedParentheticalElement.Len()

			shouldRemove = numExpressions == 0
			shouldReplace = numExpressions == 1

			if shouldReplace {
				// Dump the parenthetical and the joined expression by grabbing the only element in the joined
				// expression for replacement
				replacementExpression = typedParentheticalElement.Get(0)
			}
		}
	}

	if shouldRemove {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Remove(element)
		}
	} else if shouldReplace {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Replace(typedParent.IndexOf(element), replacementExpression)
		}
	}

	return nil
}
func (s *Translator) rewriteKindFilters(stack *model.WalkStack, disjunction *model.Disjunction) error {
	var (
		kindsByRef                = map[string]*AnnotatedKindMatcher{}
		nonKindMatcherExpressions []model.Expression
	)

	for _, expression := range disjunction.GetAll() {
		switch typedExpression := expression.(type) {
		case *AnnotatedKindMatcher:
			if binding, err := GetSymbol(typedExpression.Reference); err != nil {
				return err
			} else if kindMatcher, hasMatcher := kindsByRef[binding]; hasMatcher {
				kindMatcher.Kinds = append(kindMatcher.Kinds, typedExpression.Kinds...)
			} else {
				kindsByRef[binding] = Copy(typedExpression)
			}

		default:
			nonKindMatcherExpressions = append(nonKindMatcherExpressions, typedExpression)
		}
	}

	kindMatchers := make([]model.Expression, 0, len(kindsByRef))

	for _, kindMatcher := range kindsByRef {
		kindMatchers = append(kindMatchers, kindMatcher)
	}

	if len(nonKindMatcherExpressions) == 0 {
		if len(kindMatchers) == 1 {
			return rewrite(stack, disjunction, kindMatchers[0])
		} else {
			return rewrite(stack, disjunction, model.NewDisjunction(kindMatchers...))
		}
	} else if len(kindMatchers) > 0 {
		return rewrite(stack, disjunction, model.NewDisjunction(append(nonKindMatcherExpressions, kindMatchers...)...))
	}

	return nil
}

func Translate(regularQuery *model.RegularQuery, kindMapper KindMapper) (map[string]any, error) {
	var (
		bindings = NewBinder()
		rewriter = NewTranslator(kindMapper, bindings, regularQuery)
	)

	if err := bindings.Scan(regularQuery); err != nil {
		return nil, err
	}

	// Rewrite phase
	if err := analyzer.Analyze(regularQuery, func(analyzerInst *analyzer.Analyzer) {
		analyzer.WithVisitor(analyzerInst, rewriter.annotatePatternBindings)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteStringNegations)
		analyzer.WithVisitor(analyzerInst, rewriter.annotateProjectionItems)
		analyzer.WithVisitor(analyzerInst, rewriter.validatePropertyLookups)
		analyzer.WithVisitor(analyzerInst, rewriter.annotateKindMatchers)
		analyzer.WithVisitor(analyzerInst, rewriter.liftMatchCriteria)
		analyzer.WithVisitor(analyzerInst, rewriter.annotateComparisons)
		analyzer.WithVisitor(analyzerInst, rewriter.translatePatternPredicates)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteFunctionInvocations)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteUpdatingClauses)
	}, CollectPGSQLTypes); err != nil {
		return nil, err
	}

	// Optimization phase
	if err := analyzer.Analyze(regularQuery, func(analyzerInst *analyzer.Analyzer) {
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteNegations)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteKindFilters)
		analyzer.WithVisitor(analyzerInst, rewriter.removeEmptyExpressionLists)
	}, CollectPGSQLTypes); err != nil {
		return nil, err
	}

	return bindings.Parameters(), nil
}
