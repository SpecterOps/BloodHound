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
	"strings"

	"github.com/specterops/bloodhound/cypher/models"
	cypher "github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func translateCypherAssignmentOperator(operator cypher.AssignmentOperator) (pgsql.Operator, error) {
	switch operator {
	case cypher.OperatorAssignment:
		return pgsql.OperatorAssignment, nil
	case cypher.OperatorLabelAssignment:
		return pgsql.OperatorLabelAssignment, nil
	default:
		return pgsql.UnsetOperator, fmt.Errorf("unsupported assignment operator %s", operator)
	}
}

func (s *Translator) translateRemoveItem(removeItem *cypher.RemoveItem) error {
	if removeItem.KindMatcher != nil {
		if variable, isVariable := removeItem.KindMatcher.Reference.(*cypher.Variable); !isVariable {
			return fmt.Errorf("expected variable for kind matcher reference but found type: %T", removeItem.KindMatcher.Reference)
		} else if binding, resolved := s.query.Scope.LookupString(variable.Symbol); !resolved {
			return fmt.Errorf("unable to find identifier %s", variable.Symbol)
		} else {
			return s.mutations.AddKindRemoval(s.query.Scope, binding.Identifier, removeItem.KindMatcher.Kinds)
		}
	}

	if removeItem.Property != nil {
		if propertyLookupExpression, err := s.treeTranslator.Pop(); err != nil {
			return err
		} else if propertyLookup, err := decomposePropertyLookup(propertyLookupExpression); err != nil {
			return err
		} else {
			return s.mutations.AddPropertyRemoval(s.query.Scope, propertyLookup)
		}
	}

	return nil
}

func (s *Translator) translateSetItem(setItem *cypher.SetItem) error {
	if operator, err := translateCypherAssignmentOperator(setItem.Operator); err != nil {
		return err
	} else {
		switch operator {
		case pgsql.OperatorAssignment:
			if rightOperand, err := s.treeTranslator.Pop(); err != nil {
				return err
			} else if leftOperand, err := s.treeTranslator.Pop(); err != nil {
				return err
			} else if leftPropertyLookup, err := decomposePropertyLookup(leftOperand); err != nil {
				return err
			} else {
				return s.mutations.AddPropertyAssignment(s.query.Scope, leftPropertyLookup, operator, rightOperand)
			}

		case pgsql.OperatorLabelAssignment:
			if rightOperand, err := s.treeTranslator.Pop(); err != nil {
				return err
			} else if leftOperand, err := s.treeTranslator.Pop(); err != nil {
				return err
			} else if targetIdentifier, isIdentifier := leftOperand.(pgsql.Identifier); !isIdentifier {
				return fmt.Errorf("expected an identifier for kind assignment left operand but got: %T", leftOperand)
			} else if kindList, isKindListLiteral := rightOperand.(pgsql.KindListLiteral); !isKindListLiteral {
				return fmt.Errorf("expected an identifier for kind list right operand but got: %T", rightOperand)
			} else {
				return s.mutations.AddKindAssignment(s.query.Scope, targetIdentifier, kindList.Values)
			}

		default:
			return fmt.Errorf("unsupported set item operator: %s", operator)
		}
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
			s.treeTranslator.Push(pgsql.FunctionCall{
				Function: functionIdentifier,
				Bare:     true,
				CastType: dataType,
			})

		case pgsql.FunctionNow:
			s.treeTranslator.Push(pgsql.FunctionCall{
				Function: functionIdentifier,
				Bare:     false,
				CastType: dataType,
			})

		default:
			if precisionLiteral, err := pgsql.AsLiteral(defaultTimestampPrecision); err != nil {
				return err
			} else {
				s.treeTranslator.Push(pgsql.FunctionCall{
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
	} else if specArgument, err := s.treeTranslator.Pop(); err != nil {
		return err
	} else {
		s.treeTranslator.Push(pgsql.NewTypeCast(specArgument, dataType))
	}

	return nil
}

func (s *Translator) translateKindMatcher(kindMatcher *cypher.KindMatcher) error {
	if variable, isVariable := kindMatcher.Reference.(*cypher.Variable); !isVariable {
		return fmt.Errorf("expected variable for kind matcher reference but found type: %T", kindMatcher.Reference)
	} else if binding, resolved := s.query.Scope.LookupString(variable.Symbol); !resolved {
		return fmt.Errorf("unable to find identifier %s", variable.Symbol)
	} else if kindIDs, missingKinds := s.kindMapper.MapKinds(kindMatcher.Kinds); len(missingKinds) > 0 {
		return fmt.Errorf("unable to map kinds: %s", strings.Join(missingKinds.Strings(), ", "))
	} else if kindIDsLiteral, err := pgsql.AsLiteral(kindIDs); err != nil {
		return err
	} else {
		switch binding.DataType {
		case pgsql.NodeComposite:
			s.treeTranslator.Push(pgsql.CompoundIdentifier{binding.Identifier, pgsql.ColumnKindIDs})
			s.treeTranslator.Push(kindIDsLiteral)

			if err := s.treeTranslator.PopPushOperator(pgsql.OperatorPGArrayOverlap); err != nil {
				s.SetError(err)
			}

		case pgsql.EdgeComposite:
			s.treeTranslator.Push(pgsql.CompoundIdentifier{binding.Identifier, pgsql.ColumnKindID})
			s.treeTranslator.Push(pgsql.NewAnyExpression(kindIDsLiteral))

			if err := s.treeTranslator.PopPushOperator(pgsql.OperatorEquals); err != nil {
				s.SetError(err)
			}

		default:
			return fmt.Errorf("unexpected kind matcher reference data type: %s", binding.DataType)
		}
	}

	return nil
}

func (s *Translator) translateProjectionItem(scope *Scope, projectionItem *cypher.ProjectionItem) error {
	if alias, hasAlias, err := extractIdentifierFromCypherExpression(projectionItem); err != nil {
		return err
	} else if nextExpression, err := s.treeTranslator.Pop(); err != nil {
		return err
	} else if selectItem, isProjection := nextExpression.(pgsql.SelectItem); !isProjection {
		s.SetErrorf("invalid type for select item: %T", nextExpression)
	} else {
		if hasAlias {
			s.projections.CurrentProjection().SetAlias(alias)
		}

		switch typedSelectItem := selectItem.(type) {
		case pgsql.Identifier:
			// If this is an identifier then assume the identifier as the projection alias since the translator
			// rewrites all identifiers
			if !hasAlias {
				if boundSelectItem, bound := scope.Lookup(typedSelectItem); !bound {
					return fmt.Errorf("invalid identifier: %s", typedSelectItem)
				} else {
					s.projections.CurrentProjection().SetAlias(boundSelectItem.Aliased())
				}
			}

		case *pgsql.BinaryExpression:
			if typedSelectItem.Operator == pgsql.OperatorPropertyLookup {
				// TODO: This probably belongs somewhere else
				typedSelectItem.Operator = pgsql.OperatorJSONField
			}
		}

		s.projections.CurrentProjection().SelectItem = selectItem
	}

	return nil
}

func (s *Translator) translateProjection(projection *cypher.Projection) error {
	s.projections = NewProjectionClause()
	s.projections.Distinct = projection.Distinct

	if projection.Skip != nil {
		if cypherLiteral, isLiteral := projection.Skip.Value.(*cypher.Literal); !isLiteral {
			return fmt.Errorf("expected a literal skip value but received: %T", projection.Skip.Value)
		} else if pgLiteral, err := pgsql.AsLiteral(cypherLiteral.Value); err != nil {
			return err
		} else {
			s.query.Skip = models.ValueOptional[pgsql.Expression](pgLiteral)
		}
	}

	if projection.Limit != nil {
		if cypherLiteral, isLiteral := projection.Limit.Value.(*cypher.Literal); !isLiteral {
			return fmt.Errorf("expected a literal limit value but received: %T", projection.Limit.Value)
		} else if pgLiteral, err := pgsql.AsLiteral(cypherLiteral.Value); err != nil {
			return err
		} else {
			s.query.Limit = models.ValueOptional[pgsql.Expression](pgLiteral)
		}
	}

	return nil
}
