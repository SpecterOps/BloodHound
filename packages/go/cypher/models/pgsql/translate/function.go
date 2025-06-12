// Copyright 2024 Specter Ops, Inc.
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

package translate

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/models/walk"
	"strings"

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func SymbolsFor(node pgsql.SyntaxNode) (*pgsql.SymbolTable, error) {
	instance := pgsql.NewSymbolTable()

	return instance, walk.PgSQL(node, walk.NewSimpleVisitor[pgsql.SyntaxNode](func(node pgsql.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
		switch typedNode := node.(type) {
		case pgsql.Identifier:
			instance.AddIdentifier(typedNode)

		case pgsql.CompoundIdentifier:
			instance.AddCompoundIdentifier(typedNode)
		}
	}))
}

func GetAggregatedFunctionParameterSymbols(call pgsql.FunctionCall) (*pgsql.SymbolTable, error) {
	var (
		symbolTable = pgsql.NewSymbolTable()
		callStack   = []pgsql.FunctionCall{call}
	)

	for len(callStack) > 0 {
		nextCall := callStack[len(callStack)-1]
		callStack = callStack[:len(callStack)-1]

		if pgsql.IsAggregateFunction(nextCall.Function) {
			if functionParameterSymbols, err := SymbolsFor(nextCall); err != nil {
				return nil, err
			} else {
				symbolTable.AddTable(functionParameterSymbols)
			}
		} else {
			for _, parameter := range nextCall.Parameters {
				switch typedParameter := parameter.(type) {
				case pgsql.FunctionCall:
					callStack = append(callStack, typedParameter)
				}
			}
		}
	}

	return symbolTable, nil
}

func (s *Translator) translateFunction(typedExpression *cypher.FunctionInvocation) {
	switch formattedName := strings.ToLower(typedExpression.Name); formattedName {
	case cypher.DurationFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if referenceArgument, typeOK := argument.(pgsql.Literal); !typeOK {
			s.SetErrorf("expected a string literal for the cypher function: %s but received %T", typedExpression.Name, argument)
		} else if _, typeOK := referenceArgument.Value.(string); !typeOK {
			s.SetErrorf("expected a string literal for the cypher function: %s but received %T", typedExpression.Name, argument)
		} else {
			// Rewrite the cast type of the literal to the PgSQL interval type. The equivalent PgSQL form does not
			// require a function call
			referenceArgument.CastType = pgsql.Interval
			s.treeTranslator.PushOperand(referenceArgument)
		}

	case cypher.IdentityFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if referenceArgument, typeOK := argument.(pgsql.Identifier); !typeOK {
			s.SetErrorf("expected an identifier for the cypher function: %s but received %T", typedExpression.Name, argument)
		} else {
			s.treeTranslator.PushOperand(pgsql.CompoundIdentifier{referenceArgument, pgsql.ColumnID})
		}

	case cypher.LocalTimeFunction:
		if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.TimeWithoutTimeZone); err != nil {
			s.SetError(err)
		}

	case cypher.LocalDateTimeFunction:
		if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.TimestampWithoutTimeZone); err != nil {
			s.SetError(err)
		}

	case cypher.DateFunction:
		if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.Date); err != nil {
			s.SetError(err)
		}

	case cypher.DateTimeFunction:
		if err := s.translateDateTimeFunctionCall(typedExpression, pgsql.TimestampWithTimeZone); err != nil {
			s.SetError(err)
		}

	case cypher.EdgeTypeFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if identifier, isIdentifier := argument.(pgsql.Identifier); !isIdentifier {
			s.SetErrorf("expected an identifier for the cypher function: %s but received %T", typedExpression.Name, argument)
		} else {
			s.treeTranslator.PushOperand(pgsql.CompoundIdentifier{identifier, pgsql.ColumnKindID})
		}

	case cypher.NodeLabelsFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if identifier, isIdentifier := argument.(pgsql.Identifier); !isIdentifier {
			s.SetErrorf("expected an identifier for the cypher function: %s but received %T", typedExpression.Name, argument)
		} else {
			s.treeTranslator.PushOperand(pgsql.CompoundIdentifier{identifier, pgsql.ColumnKindIDs})
		}

	case cypher.CountFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			s.treeTranslator.PushOperand(pgsql.FunctionCall{
				Function:   pgsql.FunctionCount,
				Parameters: []pgsql.Expression{argument},
				CastType:   pgsql.Int8,
			})
		}

	case cypher.StringSplitToArrayFunction:
		if typedExpression.NumArguments() != 2 {
			s.SetError(fmt.Errorf("expected two arguments for cypher function %s", typedExpression.Name))
		} else if delimiter, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else if splitReference, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			if _, hasHint := GetTypeHint(splitReference); !hasHint {
				// Do our best to coerce the type into text
				if typedSplitRef, err := TypeCastExpression(splitReference, pgsql.Text); err != nil {
					s.SetError(err)
				} else {
					s.treeTranslator.PushOperand(pgsql.FunctionCall{
						Function:   pgsql.FunctionStringToArray,
						Parameters: []pgsql.Expression{typedSplitRef, delimiter},
						CastType:   pgsql.TextArray,
					})
				}
			} else {
				s.treeTranslator.PushOperand(pgsql.FunctionCall{
					Function:   pgsql.FunctionStringToArray,
					Parameters: []pgsql.Expression{splitReference, delimiter},
					CastType:   pgsql.TextArray,
				})
			}
		}

	case cypher.ToLowerFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(argument); isPropertyLookup {
				// Rewrite the property lookup operator with a JSON text field lookup
				propertyLookup.Operator = pgsql.OperatorJSONTextField
			}

			s.treeTranslator.PushOperand(pgsql.FunctionCall{
				Function:   pgsql.FunctionToLower,
				Parameters: []pgsql.Expression{argument},
				CastType:   pgsql.Text,
			})
		}

	case cypher.ListSizeFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			var functionCall pgsql.FunctionCall

			if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(argument); isPropertyLookup {
				// Ensure that the JSONB array length function receives the JSONB type
				propertyLookup.Operator = pgsql.OperatorJSONField

				functionCall = pgsql.FunctionCall{
					Function:   pgsql.FunctionJSONBArrayLength,
					Parameters: []pgsql.Expression{argument},
					CastType:   pgsql.Int,
				}
			} else {
				functionCall = pgsql.FunctionCall{
					Function:   pgsql.FunctionArrayLength,
					Parameters: []pgsql.Expression{argument, pgsql.NewLiteral(1, pgsql.Int)},
					CastType:   pgsql.Int,
				}
			}

			s.treeTranslator.PushOperand(functionCall)
		}

	case cypher.ToUpperFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(argument); isPropertyLookup {
				// Rewrite the property lookup operator with a JSON text field lookup
				propertyLookup.Operator = pgsql.OperatorJSONTextField
			}

			s.treeTranslator.PushOperand(pgsql.FunctionCall{
				Function:   pgsql.FunctionToUpper,
				Parameters: []pgsql.Expression{argument},
				CastType:   pgsql.Text,
			})
		}

	case cypher.ToStringFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			s.treeTranslator.PushOperand(pgsql.NewTypeCast(argument, pgsql.Text))
		}

	case cypher.ToIntegerFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			s.treeTranslator.PushOperand(pgsql.NewTypeCast(argument, pgsql.Int8))
		}

	case cypher.CoalesceFunction:
		if err := s.translateCoalesceFunction(typedExpression); err != nil {
			s.SetError(err)
		}

	case cypher.CollectFunction:
		if typedExpression.NumArguments() != 1 {
			s.SetError(fmt.Errorf("expected only one argument for cypher function: %s", typedExpression.Name))
		} else if argument, err := s.treeTranslator.PopOperand(); err != nil {
			s.SetError(err)
		} else {
			castType := pgsql.AnyArray

			switch typedArgument := unwrapParenthetical(argument).(type) {
			case pgsql.Identifier:
				if binding, bound := s.scope.Lookup(typedArgument); !bound {
					s.SetError(fmt.Errorf("binding not found for collect function argument %s", typedExpression.Name))
				} else if bindingArrayType, err := binding.DataType.ToArrayType(); err != nil {
					s.SetError(err)
				} else {
					castType = bindingArrayType
				}
			}

			s.treeTranslator.PushOperand(
				pgsql.FunctionCall{
					Function: pgsql.FunctionArrayRemove,
					Parameters: []pgsql.Expression{
						pgsql.FunctionCall{
							Function:   pgsql.FunctionArrayAggregate,
							Parameters: []pgsql.Expression{argument},
							Distinct:   typedExpression.Distinct,
							CastType:   castType,
						},
						pgsql.NullLiteral(),
					},
					CastType: castType,
				},
			)
		}

	default:
		s.SetErrorf("unknown cypher function: %s", typedExpression.Name)
	}
}

func (s *Translator) translateDateTimeFunctionCall(cypherFunc *cypher.FunctionInvocation, dataType pgsql.DataType) error {
	// Ensure the local date time function uses the default precision
	const defaultTimestampPrecision = 6

	var functionIdentifier pgsql.Identifier

	switch dataType {
	case pgsql.Date:
		functionIdentifier = pgsql.FunctionCurrentDate

	case pgsql.TimeWithoutTimeZone:
		functionIdentifier = pgsql.FunctionLocalTime

	case pgsql.TimeWithTimeZone:
		functionIdentifier = pgsql.FunctionCurrentTime

	case pgsql.TimestampWithoutTimeZone:
		functionIdentifier = pgsql.FunctionLocalTimestamp

	case pgsql.TimestampWithTimeZone:
		functionIdentifier = pgsql.FunctionNow

	default:
		return fmt.Errorf("unable to convert date function with data type: %s", dataType)
	}

	// Apply defaults for this function
	if !cypherFunc.HasArguments() {
		switch functionIdentifier {
		case pgsql.FunctionCurrentDate:
			s.treeTranslator.PushOperand(pgsql.FunctionCall{
				Function: functionIdentifier,
				Bare:     true,
				CastType: dataType,
			})

		case pgsql.FunctionNow:
			s.treeTranslator.PushOperand(pgsql.FunctionCall{
				Function: functionIdentifier,
				Bare:     false,
				CastType: dataType,
			})

		default:
			if precisionLiteral, err := pgsql.AsLiteral(defaultTimestampPrecision); err != nil {
				return err
			} else {
				s.treeTranslator.PushOperand(pgsql.FunctionCall{
					Function: functionIdentifier,
					Parameters: []pgsql.Expression{
						precisionLiteral,
					},
					CastType: dataType,
				})
			}
		}
	} else if cypherFunc.NumArguments() > 1 {
		return fmt.Errorf("expected only one text argument for cypher function: %s", cypherFunc.Name)
	} else if specArgument, err := s.treeTranslator.PopOperand(); err != nil {
		return err
	} else {
		s.treeTranslator.PushOperand(pgsql.NewTypeCast(specArgument, dataType))
	}

	return nil
}

func (s *Translator) translateCoalesceFunction(functionInvocation *cypher.FunctionInvocation) error {
	if numArgs := functionInvocation.NumArguments(); numArgs == 0 {
		s.SetError(fmt.Errorf("expected at least one argument for cypher function: %s", functionInvocation.Name))
	} else {
		var (
			arguments    = make([]pgsql.Expression, numArgs)
			expectedType = pgsql.UnsetDataType
		)

		// This loop is used to pop off the coalesce function arguments in the intended order (since they're
		// pushed onto the translator stack).
		for idx := range functionInvocation.Arguments {
			if argument, err := s.treeTranslator.PopOperand(); err != nil {
				return err
			} else {
				arguments[numArgs-idx-1] = argument
			}
		}

		// Find and validate types of the arguments
		for _, argument := range arguments {
			// Properties have no type information and should be skipped
			if argumentType, err := InferExpressionType(argument); err != nil {
				return err
			} else if argumentType.IsKnown() {
				// If the expected type isn't known yet then assign the known inferred type to it
				if !expectedType.IsKnown() {
					expectedType = argumentType
				} else if expectedType != argumentType {
					// All other inferrable argument types must match the first inferred type encountered
					return fmt.Errorf("types in coalesce function must match %s but got %s", expectedType, argumentType)
				}
			}
		}

		if expectedType.IsKnown() {
			// Rewrite any property lookup operators now that we have some type information
			for idx, argument := range arguments {
				if propertyLookup, isPropertyLookup := expressionToPropertyLookupBinaryExpression(argument); isPropertyLookup {
					arguments[idx] = rewritePropertyLookupOperator(propertyLookup, expectedType)
				}
			}
		}

		// Translate the function call to the expected SQL form
		s.treeTranslator.PushOperand(pgsql.FunctionCall{
			Function:   pgsql.FunctionCoalesce,
			Parameters: arguments,
			CastType:   expectedType,
		})
	}

	return nil
}
