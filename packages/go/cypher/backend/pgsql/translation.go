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
	"github.com/specterops/bloodhound/cypher/analyzer"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/model/pg"
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

	case *pg.AnnotatedVariable:
		return typedExpression.Symbol, nil

	default:
		return "", fmt.Errorf("unable to source symbol from expression type %T", expression)
	}

	return "", nil
}

type Binder struct {
	parameters          map[string]*pg.AnnotatedParameter
	bindingTypeMappings map[string]pg.DataType
	aliases             map[string]string
	patternBindings     map[string]struct{}
	syntheticBindings   map[string]struct{}
	nextParameterID     int
	nextBindingID       int
}

func NewBinder() *Binder {
	return &Binder{
		parameters:          map[string]*pg.AnnotatedParameter{},
		bindingTypeMappings: map[string]pg.DataType{},
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

func (s *Binder) BindVariable(variable *model.Variable, bindingType pg.DataType) *pg.AnnotatedVariable {
	s.bindingTypeMappings[variable.Symbol] = bindingType
	return pg.NewAnnotatedVariable(variable, bindingType)
}

func (s *Binder) BindPatternVariable(variable *model.Variable, bindingType pg.DataType) *pg.AnnotatedVariable {
	s.patternBindings[variable.Symbol] = struct{}{}
	return s.BindVariable(variable, bindingType)
}

func (s *Binder) BindingType(binding string) (pg.DataType, bool) {
	if bindingType, isBound := s.bindingTypeMappings[binding]; isBound {
		return bindingType, isBound
	}

	return pg.UnknownDataType, false
}

func (s *Binder) LookupVariable(symbol string) (*pg.AnnotatedVariable, bool) {
	if dataType, isBound := s.BindingType(symbol); isBound {
		return pg.NewAnnotatedVariable(model.NewVariableWithSymbol(symbol), dataType), true
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

func (s *Binder) NewAnnotatedVariable(prefix string, bindingType pg.DataType) *pg.AnnotatedVariable {
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

func (s *Binder) NewParameter(value any) (*pg.AnnotatedParameter, error) {
	var (
		parameterSymbol = s.NewParameterSymbol()
	)

	if parameterTypeAnnotation, err := pg.NewSQLTypeAnnotationFromValue(value); err != nil {
		return nil, err
	} else {
		parameter := pg.NewAnnotatedParameter(model.NewParameter(parameterSymbol, value), parameterTypeAnnotation.Type)

		// Record the parameter's value for mapping to the query later
		s.parameters[parameterSymbol] = parameter
		return parameter, nil
	}
}

func (s *Binder) NewLiteral(literal *model.Literal) (*pg.AnnotatedLiteral, error) {
	if literalTypeAnnotation, err := pg.NewSQLTypeAnnotationFromValue(literal.Value); err != nil {
		return nil, err
	} else {
		return pg.NewAnnotatedLiteral(literal, literalTypeAnnotation.Type), nil
	}
}

func (s *Binder) NewAlias(originalSymbol string, alias *model.Variable) *pg.AnnotatedVariable {
	s.aliases[originalSymbol] = alias.Symbol

	if originalBindingType, isBound := s.bindingTypeMappings[originalSymbol]; isBound {
		return s.BindVariable(alias, originalBindingType)
	}

	return s.BindVariable(alias, pg.UnknownDataType)
}

func (s *Binder) Scan(regularQuery *model.RegularQuery) error {
	if err := analyzer.Analyze(regularQuery, func(analyzerInst *analyzer.Analyzer) {
		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, node *model.Parameter) error {
			// Rewrite all parameter symbols and collect their values
			if annotatedParameter, err := s.NewParameter(node.Value); err != nil {
				return err
			} else {
				return rewrite(stack, node, annotatedParameter)
			}
		})

		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, node *model.Literal) error {
			// Rewrite all parameter symbols and collect their values
			if annotatedLiteral, err := s.NewLiteral(node); err != nil {
				return err
			} else {
				return rewrite(stack, node, annotatedLiteral)
			}
		})

		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, patternPart *model.PatternPart) error {
			if patternPart.Binding != nil {
				if bindingVariable, typeOK := patternPart.Binding.(*model.Variable); !typeOK {
					return fmt.Errorf("expected variable for pattern part binding but got: %T", patternPart.Binding)
				} else {
					patternPart.Binding = s.BindPatternVariable(bindingVariable, pg.Path)
				}
			}
			return nil
		})

		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, patternElement *model.PatternElement) error {
			// Eagerly bind all ReadingClause pattern elements to simplify referencing when crafting SQL join statements
			if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
				if nodePattern.Binding == nil {
					nodePattern.Binding = s.NewAnnotatedVariable("n", pg.Node)
				} else if bindingVariable, typeOK := nodePattern.Binding.(*model.Variable); !typeOK {
					return fmt.Errorf("expected variable for node pattern binding but got: %T", nodePattern.Binding)
				} else if _, isPatternPredicate := stack.Trunk().(*model.PatternPredicate); isPatternPredicate {
					nodePattern.Binding = s.BindVariable(bindingVariable, pg.Node)
				} else {
					nodePattern.Binding = s.BindPatternVariable(bindingVariable, pg.Node)
				}
			} else {
				relationshipPattern, _ := patternElement.AsRelationshipPattern()

				if relationshipPattern.Binding == nil {
					relationshipPattern.Binding = s.NewAnnotatedVariable("e", pg.Edge)
				} else if bindingVariable, typeOK := relationshipPattern.Binding.(*model.Variable); !typeOK {
					return fmt.Errorf("expected variable for relationship pattern binding but got: %T", relationshipPattern.Binding)
				} else if _, isPatternPredicate := stack.Trunk().(*model.PatternPredicate); isPatternPredicate {
					relationshipPattern.Binding = s.BindVariable(bindingVariable, pg.Edge)
				} else {
					relationshipPattern.Binding = s.BindPatternVariable(bindingVariable, pg.Edge)
				}
			}

			return nil
		})

		analyzer.WithVisitor(analyzerInst, func(_ *model.WalkStack, node *model.ProjectionItem) error {
			if bindingVariable, isVariable := node.Binding.(*model.Variable); node.Binding != nil && isVariable {
				if projectionVariable, isVariable := node.Expression.(*model.Variable); isVariable {
					node.Binding = s.NewAlias(projectionVariable.Symbol, bindingVariable)
				}
			}

			return nil
		})

		analyzer.WithVisitor(analyzerInst, func(_ *model.WalkStack, node *model.Delete) error {
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
	}, pg.CollectPGSQLTypes); err != nil {
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

func (s *Translator) rewriteUpdatingClauses(_ *model.WalkStack, singlePartQuery *model.SinglePartQuery) error {
	return NewUpdateClauseRewriter(s.Bindings, s.kindMapper).RewriteUpdatingClauses(singlePartQuery)
}

func (s *Translator) liftNodePatternCriteria(_ *model.WalkStack, nodePattern *model.NodePattern) ([]model.Expression, error) {
	var criteria []model.Expression

	if nodePattern.Binding == nil {
		nodePattern.Binding = s.Bindings.NewVariable("n")
	}

	if len(nodePattern.Kinds) > 0 {
		kindMatcher := model.NewKindMatcher(nodePattern.Binding, nodePattern.Kinds)
		criteria = append(criteria, pg.NewAnnotatedKindMatcher(kindMatcher, pg.Node))
	}

	if nodePattern.Properties != nil {
		nodePropertyMatchers := nodePattern.Properties.(*model.Properties)

		if nodePropertyMatchers.Parameter != nil {
			return nil, fmt.Errorf("unable to translate property matcher paramter for node %s", nodePattern.Binding)
		}

		for propertyName, matcherValue := range nodePropertyMatchers.Map {
			if bindingVariable, typeOK := nodePattern.Binding.(*pg.AnnotatedVariable); !typeOK {
				return nil, fmt.Errorf("unexpected node pattern binding type for node pattern: %T", nodePattern.Binding)
			} else {
				propertyLookup := model.NewPropertyLookup(bindingVariable.Symbol, propertyName)

				if annotation, err := pg.NewSQLTypeAnnotationFromExpression(matcherValue); err != nil {
					return nil, err
				} else {
					criteria = append(criteria, model.NewComparison(
						pg.NewAnnotatedPropertyLookup(propertyLookup, annotation.Type),
						model.OperatorEquals,
						matcherValue,
					))
				}
			}
		}
	}

	return criteria, nil
}

func (s *Translator) liftRelationshipPatternCriteria(_ *model.WalkStack, relationshipPattern *model.RelationshipPattern) ([]model.Expression, error) {
	var criteria []model.Expression

	if relationshipPattern.Binding == nil {
		relationshipPattern.Binding = s.Bindings.NewVariable("e")
	}

	if len(relationshipPattern.Kinds) > 0 {
		kindMatcher := model.NewKindMatcher(relationshipPattern.Binding, relationshipPattern.Kinds)
		criteria = append(criteria, pg.NewAnnotatedKindMatcher(kindMatcher, pg.Edge))
	}

	if relationshipPattern.Properties != nil {
		edgePropertyMatchers := relationshipPattern.Properties.(*model.Properties)

		if edgePropertyMatchers.Parameter != nil {
			return nil, fmt.Errorf("unable to translate property matcher paramter for edge %s", relationshipPattern.Binding)
		}

		for propertyName, matcherValue := range edgePropertyMatchers.Map {
			if bindingVariable, typeOK := relationshipPattern.Binding.(*pg.AnnotatedVariable); !typeOK {
				return nil, fmt.Errorf("unexpected relationship pattern binding type: %T", relationshipPattern.Binding)
			} else {
				propertyLookup := model.NewPropertyLookup(bindingVariable.Symbol, propertyName)

				if annotation, err := pg.NewSQLTypeAnnotationFromExpression(matcherValue); err != nil {
					return nil, err
				} else {
					criteria = append(criteria, model.NewComparison(
						pg.NewAnnotatedPropertyLookup(propertyLookup, annotation.Type),
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
		subquery        = &pg.Subquery{
			PatternElements: patternPredicate.PatternElements,
		}
	)

	for _, patternElement := range subquery.PatternElements {
		if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
			// Is the node pattern bound to a variable and was that variable bound earlier in the AST?
			if bindingVariable, typeOK := nodePattern.Binding.(*pg.AnnotatedVariable); !typeOK {
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
			if bindingVariable, typeOK := relationshipPattern.Binding.(*pg.AnnotatedVariable); !typeOK {
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
	case *pg.AnnotatedVariable:
		return rewrite(stack, kindMatcher, pg.NewAnnotatedKindMatcher(kindMatcher, typedExpression.Type))

	case *model.Variable:
		if dataType, hasBindingType := s.Bindings.BindingType(typedExpression.Symbol); !hasBindingType {
			return fmt.Errorf("unable to locate a binding type for variable %s", typedExpression.Symbol)
		} else {
			return rewrite(stack, kindMatcher, pg.NewAnnotatedKindMatcher(kindMatcher, dataType))
		}

	default:
		return fmt.Errorf("unexpected kind matcher reference type %T", kindMatcher.Reference)
	}
}

func (s *Translator) rewriteComparison(stack *model.WalkStack, comparison *model.Comparison) (bool, error) {
	// Is this a property lookup comparison?
	switch typedLeftOperand := comparison.Left.(type) {
	case *model.PropertyLookup:
		// Try to suss out if this is a property existence check
		if len(comparison.Partials) == 1 {
			comparisonPartial := comparison.Partials[0]

			switch typedRightHand := comparisonPartial.Right.(type) {
			case *pg.AnnotatedLiteral:
				if typedRightHand.Null {
					// This is a null check for a property and must be rewritten for SQL
					switch comparisonPartial.Operator {
					case model.OperatorIsNot:
						if leftOperandVariable, isVariable := typedLeftOperand.Atom.(*model.Variable); !isVariable {
							return false, fmt.Errorf("unexpected expression as left operand %T", typedLeftOperand.Atom)
						} else if leftOperandTypedVariable, isBound := s.Bindings.LookupVariable(leftOperandVariable.Symbol); !isBound {
							return false, fmt.Errorf("left operand varaible %s is not bound", leftOperandTypedVariable.Symbol)
						} else if err := rewrite(stack, comparison, model.NewComparison(
							&pg.PropertiesReference{
								// TODO: Might need a copy?
								Reference: leftOperandTypedVariable,
							},
							OperatorJSONBFieldExists,
							pg.NewStringLiteral(typedLeftOperand.Symbols[0]),
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
								&pg.PropertiesReference{
									Reference: leftOperandTypedVariable,
								},
								OperatorJSONBFieldExists,
								pg.NewStringLiteral(typedLeftOperand.Symbols[0]),
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

func (s *Translator) rewritePartialComparison(_ *model.WalkStack, partial *model.PartialComparison) error {
	switch partial.Operator {
	case model.OperatorIn:
		switch partial.Right.(type) {
		case *model.Parameter, *pg.AnnotatedParameter:
			// When the "in" operator addresses right-hand parameter it must be rewritten as: "= any($param)"
			partial.Operator = model.OperatorEquals
			partial.Right = model.NewSimpleFunctionInvocation(pgsqlAnyFunction, partial.Right)
		}

	case model.OperatorStartsWith:
		// Replace this operator with the like operator
		partial.Operator = OperatorLike

		// If the right side isn't a string for any of these it's an error
		switch typedRightOperand := partial.Right.(type) {
		case *pg.AnnotatedLiteral:
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

		case *pg.AnnotatedParameter:
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
		case *pg.AnnotatedLiteral:
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

		case *pg.AnnotatedParameter:
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
		case *pg.AnnotatedLiteral:
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

		case *pg.AnnotatedParameter:
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

	case model.OperatorEquals:
		switch typedRightOperand := partial.Right.(type) {
		case *pg.AnnotatedLiteral:
			// If this is an array type then first wrap it in the `to_jsonb` function
			if typedRightOperand.Type.IsArrayType() {
				partial.Right = model.NewSimpleFunctionInvocation(pgsqlToJSONBFunction, partial.Right)
			}

		case *pg.AnnotatedParameter:
			// If this is an array type then rewrite it as a JSONB value
			if typedRightOperand.Type.IsArrayType() {
				newParameter := &pgtype.JSONB{}

				if err := newParameter.Set(typedRightOperand.Value); err != nil {
					return err
				}

				typedRightOperand.Value = newParameter
			}
		}
	}

	return nil
}

func (s *Translator) annotateComparisons(stack *model.WalkStack, comparison *model.Comparison) error {
	var (
		typeAnnotation *pg.SQLTypeAnnotation
		operator       model.Operator
	)

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
			// TODO: Overloading the operator means that we may miss partial comparison continuations
			operator = typedNode.Operator

			if err := s.rewritePartialComparison(stack, typedNode); err != nil {
				return err
			}

			comparisonWalkStack = append(comparisonWalkStack, typedNode.Right)

		case *pg.AnnotatedParameter:
			if typeAnnotation == nil {
				typeAnnotation = &pg.SQLTypeAnnotation{
					Type: typedNode.Type,
				}
			} else if typeAnnotation.Type != typedNode.Type {
				return fmt.Errorf("comparison contains mixed types: %s and %s", typeAnnotation.Type, typedNode.Type)
			}

		case *pg.AnnotatedLiteral:
			if typeAnnotation == nil {
				typeAnnotation = &pg.SQLTypeAnnotation{
					Type: typedNode.Type,
				}
			} else if typeAnnotation.Type != typedNode.Type {
				return fmt.Errorf("comparison contains mixed types: %s and %s", typeAnnotation.Type, typedNode.Type)
			}

		case *model.FunctionInvocation:
			var functionInvocationTypeAnnotation *pg.SQLTypeAnnotation

			switch typedNode.Name {
			case cypherDateFunction:
				functionInvocationTypeAnnotation = &pg.SQLTypeAnnotation{
					Type: pg.Date,
				}

			case cypherTimeFunction:
				functionInvocationTypeAnnotation = &pg.SQLTypeAnnotation{
					Type: pg.TimeWithTimeZone,
				}

			case cypherLocalTimeFunction:
				functionInvocationTypeAnnotation = &pg.SQLTypeAnnotation{
					Type: pg.TimeWithoutTimeZone,
				}

			case cypherDateTimeFunction:
				functionInvocationTypeAnnotation = &pg.SQLTypeAnnotation{
					Type: pg.TimestampWithTimeZone,
				}

			case cypherLocalDateTimeFunction:
				functionInvocationTypeAnnotation = &pg.SQLTypeAnnotation{
					Type: pg.TimestampWithoutTimeZone,
				}

			case cypherDurationFunction:
				functionInvocationTypeAnnotation = &pg.SQLTypeAnnotation{
					Type: pg.Interval,
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

			// If this is an array type we need to do some special rewriting negotiation for different operators
			if typeAnnotation.Type.IsArrayType() {
				switch operator {
				case model.OperatorIn:
					// If this is an operation such that <left> in <array-type> then we must wrap the right hand
					// operand using the pgsql any() function and type the left hand operand to the array's base type
					if baseType, err := typeAnnotation.Type.ArrayBaseType(); err != nil {
						return err
					} else {
						leftOperandType = baseType
					}

				default:
					// If this isn't a contains operator then rewrite the left hand operand type to jsonb and wrap
					// the right hand operand in the to_json function with the type annotation of jsonb
					leftOperandType = pg.JSONB
				}
			}

			// Rewrite the left operand so that the property lookup is correctly type annotated
			comparison.Left = pg.NewAnnotatedPropertyLookup(leftHandPropertyLookup, leftOperandType)

			for _, partialComparison := range comparison.Partials {
				switch typedPartialComparison := partialComparison.Right.(type) {
				case *model.PropertyLookup:
					// Make sure right hand operand property lookups are correctly type annotated
					annotatedPropertyLookup := pg.NewAnnotatedPropertyLookup(typedPartialComparison, typeAnnotation.Type)

					if err := rewrite(stack, partialComparison.Right, annotatedPropertyLookup); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *Translator) rewriteNegations(_ *model.WalkStack, negation *model.Negation) error {
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
					model.NewComparison(comparison.Left, model.OperatorIs, pg.NewAnnotatedLiteral(model.NewLiteral(nil, true), pg.Null)),
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

func (s *Translator) rewriteFunctionInvocations(stack *model.WalkStack, functionInvocation *model.FunctionInvocation) error {
	switch functionInvocation.Name {
	case cypherNodeLabelsFunction:
		switch typedArgument := functionInvocation.Arguments[0].(type) {
		case *model.Variable:
			return rewrite(stack, functionInvocation, pg.NewNodeKindsReference(pg.NewAnnotatedVariable(typedArgument, pg.Node)))

		case *pg.AnnotatedVariable:
			return rewrite(stack, functionInvocation, pg.NewNodeKindsReference(typedArgument))

		default:
			return fmt.Errorf("expected a variable as the first argument in %s function", functionInvocation.Name)
		}

	case cypherEdgeTypeFunction:
		switch typedArgument := functionInvocation.Arguments[0].(type) {
		case *model.Variable:
			return rewrite(stack, functionInvocation, pg.NewEdgeKindReference(pg.NewAnnotatedVariable(typedArgument, pg.Edge)))

		case *pg.AnnotatedVariable:
			return rewrite(stack, functionInvocation, pg.NewEdgeKindReference(typedArgument))

		default:
			return fmt.Errorf("expected a variable as the first argument in %s function", functionInvocation.Name)
		}

	case cypherToLowerFunction:
		switch typedArgument := functionInvocation.Arguments[0].(type) {
		case *model.PropertyLookup:
			functionInvocation.Arguments[0] = pg.NewAnnotatedPropertyLookup(typedArgument, pg.Text)
		}
	}

	return nil
}

func (s *Translator) annotateProjectionItems(_ *model.WalkStack, projectionItem *model.ProjectionItem) error {
	switch typedExpression := projectionItem.Expression.(type) {
	case *model.Variable:
		if bindingType, isBound := s.Bindings.BindingType(typedExpression.Symbol); !isBound {
			return fmt.Errorf("variable %s for projection item is not bound", typedExpression.Symbol)
		} else {
			projectionItem.Expression = pg.NewEntity(pg.NewAnnotatedVariable(typedExpression, bindingType))

			// Set projection item binding to the variable reference if there's no binding present
			if projectionItem.Binding == nil {
				projectionItem.Binding = pg.NewAnnotatedVariable(typedExpression, bindingType)
			}
		}
	}

	return nil
}

func (s *Translator) validatePropertyLookups(_ *model.WalkStack, propertyLookup *model.PropertyLookup) error {
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
		kindsByRef                = map[string]*pg.AnnotatedKindMatcher{}
		nonKindMatcherExpressions []model.Expression
	)

	for _, expression := range disjunction.GetAll() {
		switch typedExpression := expression.(type) {
		case *pg.AnnotatedKindMatcher:
			if binding, err := GetSymbol(typedExpression.Reference); err != nil {
				return err
			} else if kindMatcher, hasMatcher := kindsByRef[binding]; hasMatcher {
				kindMatcher.Kinds = append(kindMatcher.Kinds, typedExpression.Kinds...)
			} else {
				kindsByRef[binding] = pg.Copy(typedExpression)
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
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteStringNegations)
		analyzer.WithVisitor(analyzerInst, rewriter.annotateProjectionItems)
		analyzer.WithVisitor(analyzerInst, rewriter.validatePropertyLookups)
		analyzer.WithVisitor(analyzerInst, rewriter.annotateKindMatchers)
		analyzer.WithVisitor(analyzerInst, rewriter.liftMatchCriteria)
		analyzer.WithVisitor(analyzerInst, rewriter.annotateComparisons)
		analyzer.WithVisitor(analyzerInst, rewriter.translatePatternPredicates)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteFunctionInvocations)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteUpdatingClauses)
	}, pg.CollectPGSQLTypes); err != nil {
		return nil, err
	}

	// Optimization phase
	if err := analyzer.Analyze(regularQuery, func(analyzerInst *analyzer.Analyzer) {
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteNegations)
		analyzer.WithVisitor(analyzerInst, rewriter.rewriteKindFilters)
		analyzer.WithVisitor(analyzerInst, rewriter.removeEmptyExpressionLists)
	}, pg.CollectPGSQLTypes); err != nil {
		return nil, err
	}

	return bindings.Parameters(), nil
}
