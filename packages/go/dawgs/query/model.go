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
	"fmt"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func convertCriteria[T any](criteria ...graph.Criteria) []T {
	var (
		converted = make([]T, len(criteria))
	)

	for idx, nextCriteria := range criteria {
		converted[idx] = nextCriteria.(T)
	}

	return converted
}

func Update(clauses ...*model.UpdatingClause) []*model.UpdatingClause {
	return clauses
}

func Updatef(provider graph.CriteriaProvider) []*model.UpdatingClause {
	switch typedCriteria := provider().(type) {
	case []*model.UpdatingClause:
		return typedCriteria

	case []graph.Criteria:
		return convertCriteria[*model.UpdatingClause](typedCriteria...)

	case *model.UpdatingClause:
		return []*model.UpdatingClause{typedCriteria}

	default:
		return []*model.UpdatingClause{
			model.WithErrors(&model.UpdatingClause{}, fmt.Errorf("invalid type %T for update clause", typedCriteria)),
		}
	}
}

func AddKind(reference graph.Criteria, kind graph.Kind) *model.UpdatingClause {
	return model.NewUpdatingClause(&model.Set{
		Items: []*model.SetItem{{
			Left:     reference,
			Operator: model.OperatorLabelAssignment,
			Right:    graph.Kinds{kind},
		}},
	})
}

func AddKinds(reference graph.Criteria, kinds graph.Kinds) *model.UpdatingClause {
	return model.NewUpdatingClause(&model.Set{
		Items: []*model.SetItem{{
			Left:     reference,
			Operator: model.OperatorLabelAssignment,
			Right:    kinds,
		}},
	})
}

func DeleteKind(reference graph.Criteria, kind graph.Kind) *model.UpdatingClause {
	return model.NewUpdatingClause(&model.Remove{
		Items: []*model.RemoveItem{{
			KindMatcher: &model.KindMatcher{
				Reference: reference,
				Kinds:     graph.Kinds{kind},
			},
		}},
	})
}

func DeleteKinds(reference graph.Criteria, kinds graph.Kinds) *model.Remove {
	return &model.Remove{
		Items: []*model.RemoveItem{{
			KindMatcher: &model.KindMatcher{
				Reference: reference,
				Kinds:     kinds,
			},
		}},
	}
}

func SetProperty(reference graph.Criteria, value any) *model.UpdatingClause {
	return model.NewUpdatingClause(&model.Set{
		Items: []*model.SetItem{{
			Left:     reference,
			Operator: model.OperatorAssignment,
			Right:    Parameter(value),
		}},
	})
}

func SetProperties(reference graph.Criteria, properties map[string]any) *model.UpdatingClause {
	set := &model.Set{}

	for key, value := range properties {
		set.Items = append(set.Items, &model.SetItem{
			Left:     Property(reference, key),
			Operator: model.OperatorAssignment,
			Right:    Parameter(value),
		})
	}

	return model.NewUpdatingClause(set)
}

func DeleteProperty(reference *model.PropertyLookup) *model.UpdatingClause {
	return model.NewUpdatingClause(&model.Remove{
		Items: []*model.RemoveItem{{
			Property: reference,
		}},
	})
}

func DeleteProperties(reference graph.Criteria, propertyNames ...string) *model.UpdatingClause {
	removeClause := &model.Remove{}

	for _, propertyName := range propertyNames {
		removeClause.Items = append(removeClause.Items, &model.RemoveItem{
			Property: Property(reference, propertyName),
		})
	}

	return model.NewUpdatingClause(removeClause)
}

func Kind(reference graph.Criteria, kind graph.Kind) *model.KindMatcher {
	return &model.KindMatcher{
		Reference: reference,
		Kinds:     graph.Kinds{kind},
	}
}

func KindIn(reference graph.Criteria, kinds ...graph.Kind) *model.Parenthetical {
	expressions := make([]graph.Criteria, len(kinds))

	for idx, kind := range kinds {
		expressions[idx] = Kind(reference, kind)
	}

	return Or(expressions...)
}

func NodeProperty(name string) *model.PropertyLookup {
	return model.NewPropertyLookup(NodeSymbol, name)
}

func RelationshipProperty(name string) *model.PropertyLookup {
	return model.NewPropertyLookup(RelationshipSymbol, name)
}

func StartProperty(name string) *model.PropertyLookup {
	return model.NewPropertyLookup(RelationshipStartSymbol, name)
}

func EndProperty(name string) *model.PropertyLookup {
	return model.NewPropertyLookup(RelationshipEndSymbol, name)
}

func Property(qualifier graph.Criteria, name string) *model.PropertyLookup {
	return &model.PropertyLookup{
		Atom:    qualifier.(*model.Variable),
		Symbols: []string{name},
	}
}

func Count(reference graph.Criteria) graph.Criteria {
	return &model.FunctionInvocation{
		Name:      "count",
		Arguments: []model.Expression{reference},
	}
}

func And(criteria ...graph.Criteria) *model.Conjunction {
	return model.NewConjunction(convertCriteria[model.Expression](criteria...)...)
}

func Or(criteria ...graph.Criteria) *model.Parenthetical {
	return &model.Parenthetical{
		Expression: model.NewDisjunction(convertCriteria[model.Expression](criteria...)...),
	}
}

func Xor(criteria ...graph.Criteria) *model.ExclusiveDisjunction {
	return model.NewExclusiveDisjunction(convertCriteria[model.Expression](criteria...)...)
}

func Parameter(value any) *model.Parameter {
	if parameter, isParameter := value.(*model.Parameter); isParameter {
		return parameter
	}

	return &model.Parameter{
		Value: value,
	}
}

func Literal(value any) *model.Literal {
	return &model.Literal{
		Value: value,
		Null:  value == nil,
	}
}

func KindsOf(ref graph.Criteria) *model.FunctionInvocation {
	switch typedRef := ref.(type) {
	case *model.Variable:
		switch typedRef.Symbol {
		case NodeSymbol, RelationshipStartSymbol, RelationshipEndSymbol:
			return &model.FunctionInvocation{
				Name:      "labels",
				Arguments: []model.Expression{ref},
			}

		case RelationshipSymbol:
			return &model.FunctionInvocation{
				Name:      "type",
				Arguments: []model.Expression{ref},
			}

		default:
			return model.WithErrors(&model.FunctionInvocation{}, fmt.Errorf("invalid variable reference for KindsOf: %s", typedRef.Symbol))
		}

	default:
		return model.WithErrors(&model.FunctionInvocation{}, fmt.Errorf("invalid reference type for KindsOf: %T", ref))
	}
}

func Limit(limit int) *model.Limit {
	return &model.Limit{
		Value: Literal(limit),
	}
}

func Offset(offset int) *model.Skip {
	return &model.Skip{
		Value: Literal(offset),
	}
}

func StringContains(reference graph.Criteria, value string) *model.Comparison {
	return model.NewComparison(reference, model.OperatorContains, Parameter(value))
}

func StringStartsWith(reference graph.Criteria, value string) *model.Comparison {
	return model.NewComparison(reference, model.OperatorStartsWith, Parameter(value))
}

func StringEndsWith(reference graph.Criteria, value string) *model.Comparison {
	return model.NewComparison(reference, model.OperatorEndsWith, Parameter(value))
}

func CaseInsensitiveStringContains(reference graph.Criteria, value string) *model.Comparison {
	return model.NewComparison(
		model.NewSimpleFunctionInvocation("toLower", convertCriteria[model.Expression](reference)...),
		model.OperatorContains,
		Parameter(strings.ToLower(value)),
	)
}

func CaseInsensitiveStringStartsWith(reference graph.Criteria, value string) *model.Comparison {
	return model.NewComparison(
		model.NewSimpleFunctionInvocation("toLower", convertCriteria[model.Expression](reference)...),
		model.OperatorStartsWith,
		Parameter(strings.ToLower(value)),
	)
}

func CaseInsensitiveStringEndsWith(reference graph.Criteria, value string) *model.Comparison {
	return model.NewComparison(
		model.NewSimpleFunctionInvocation("toLower", convertCriteria[model.Expression](reference)...),
		model.OperatorEndsWith,
		Parameter(strings.ToLower(value)),
	)
}

func Equals(reference graph.Criteria, value any) *model.Comparison {
	return model.NewComparison(reference, model.OperatorEquals, Parameter(value))
}

func GreaterThan(reference graph.Criteria, value any) *model.Comparison {
	return model.NewComparison(reference, model.OperatorGreaterThan, Parameter(value))
}

func After(reference graph.Criteria, value any) *model.Comparison {
	return GreaterThan(reference, value)
}

func GreaterThanOrEquals(reference graph.Criteria, value any) *model.Comparison {
	return model.NewComparison(reference, model.OperatorGreaterThanOrEqualTo, Parameter(value))
}

func LessThan(reference graph.Criteria, value any) *model.Comparison {
	return model.NewComparison(reference, model.OperatorLessThan, Parameter(value))
}

func Before(reference graph.Criteria, value time.Time) *model.Comparison {
	return LessThan(reference, value)
}

func LessThanOrEquals(reference graph.Criteria, value any) *model.Comparison {
	return model.NewComparison(reference, model.OperatorLessThanOrEqualTo, Parameter(value))
}

func Distinct(reference graph.Criteria) *model.FunctionInvocation {
	return model.NewSimpleFunctionInvocation("distinct", reference)
}

func Exists(reference graph.Criteria) *model.FunctionInvocation {
	return model.NewSimpleFunctionInvocation("exists", reference)
}

func HasRelationships(reference *model.Variable) graph.Criteria {
	return []*model.PatternPart{
		model.NewPatternPart().AddPatternElements(
			&model.NodePattern{
				Binding: reference.Symbol,
			},
			&model.RelationshipPattern{
				Direction: graph.DirectionBoth,
			},
			&model.NodePattern{},
		),
	}
}

func In(reference graph.Criteria, value any) *model.Comparison {
	return model.NewComparison(reference, model.OperatorIn, Parameter(value))
}

func InIDs[T *model.FunctionInvocation | *model.Variable](reference T, ids ...graph.ID) *model.Comparison {
	switch any(reference).(type) {
	case *model.FunctionInvocation:
		return model.NewComparison(reference, model.OperatorIn, Parameter(ids))

	default:
		return model.NewComparison(Identity(any(reference).(*model.Variable)), model.OperatorIn, Parameter(ids))
	}
}

func Where(expression graph.Criteria) *model.Where {
	return &model.Where{
		JoiningExpression: model.JoiningExpression{
			Expressions: convertCriteria[model.Expression](expression),
		},
	}
}

func OrderBy(leaves ...graph.Criteria) *model.Order {
	return &model.Order{
		Items: convertCriteria[*model.SortItem](leaves...),
	}
}

func Order(reference, direction graph.Criteria) *model.SortItem {
	switch direction {
	case model.SortDescending:
		return &model.SortItem{
			Ascending:  false,
			Expression: reference,
		}

	default:
		return &model.SortItem{
			Ascending:  true,
			Expression: reference,
		}
	}
}

func Ascending() model.SortOrder {
	return model.SortAscending
}

func Descending() model.SortOrder {
	return model.SortDescending
}

func Delete(leaves ...graph.Criteria) *model.UpdatingClause {
	deleteClause := &model.Delete{
		Detach: true,
	}

	for _, leaf := range leaves {
		switch leaf.(*model.Variable).Symbol {
		case RelationshipSymbol, RelationshipStartSymbol, RelationshipEndSymbol:
			deleteClause.Detach = false
		}

		deleteClause.Expressions = append(deleteClause.Expressions, leaf)
	}

	return model.NewUpdatingClause(deleteClause)
}

func NodePattern(kinds graph.Kinds, properties *model.Parameter) *model.NodePattern {
	return &model.NodePattern{
		Binding:    NodeSymbol,
		Kinds:      kinds,
		Properties: properties,
	}
}

func StartNodePattern(kinds graph.Kinds, properties *model.Parameter) *model.NodePattern {
	return &model.NodePattern{
		Binding:    RelationshipStartSymbol,
		Kinds:      kinds,
		Properties: properties,
	}
}

func EndNodePattern(kinds graph.Kinds, properties *model.Parameter) *model.NodePattern {
	return &model.NodePattern{
		Binding:    RelationshipEndSymbol,
		Kinds:      kinds,
		Properties: properties,
	}
}

func RelationshipPattern(kind graph.Kind, properties *model.Parameter, direction graph.Direction) *model.RelationshipPattern {
	return &model.RelationshipPattern{
		Binding:    RelationshipSymbol,
		Kinds:      graph.Kinds{kind},
		Properties: properties,
		Direction:  direction,
	}
}

func Create(elements ...graph.Criteria) *model.UpdatingClause {
	var (
		pattern      = &model.PatternPart{}
		createClause = &model.Create{
			// Note: Unique is Neo4j specific and will not be supported here. Use of constraints for
			// uniqueness is expected instead.
			Unique:  false,
			Pattern: []*model.PatternPart{pattern},
		}
	)

	for _, element := range elements {
		switch typedElement := element.(type) {
		case *model.Variable:
			switch typedElement.Symbol {
			case NodeSymbol, RelationshipStartSymbol, RelationshipEndSymbol:
				pattern.AddPatternElements(&model.NodePattern{
					Binding: typedElement.Symbol,
				})

			default:
				createClause.AddError(fmt.Errorf("invalid variable reference create: %s", typedElement.Symbol))
			}

		case *model.NodePattern:
			pattern.AddPatternElements(typedElement)

		case *model.RelationshipPattern:
			pattern.AddPatternElements(typedElement)

		default:
			createClause.AddError(fmt.Errorf("invalid type for create: %T", element))
		}
	}

	return model.NewUpdatingClause(createClause)
}

func Returning(elements ...graph.Criteria) *model.Return {
	projection := &model.Projection{}

	for _, element := range elements {
		switch typedElement := element.(type) {
		case *model.Order:
			projection.Order = typedElement

		case *model.Limit:
			projection.Limit = typedElement

		case *model.Skip:
			projection.Skip = typedElement

		default:
			projection.Items = append(projection.Items, &model.ProjectionItem{
				Expression: element,
			})
		}
	}

	return &model.Return{
		Projection: projection,
	}
}

func Not(expression graph.Criteria) *model.Negation {
	return &model.Negation{
		Expression: expression,
	}
}

func IsNull(reference graph.Criteria) *model.Comparison {
	return model.NewComparison(reference, model.OperatorIs, Literal(nil))
}

func IsNotNull(reference graph.Criteria) *model.Comparison {
	return model.NewComparison(reference, model.OperatorIsNot, Literal(nil))
}

func GetFirstReadingClause(query *model.RegularQuery) *model.ReadingClause {
	if query.SingleQuery != nil && query.SingleQuery.SinglePartQuery != nil {
		readingClauses := query.SingleQuery.SinglePartQuery.ReadingClauses

		if len(readingClauses) > 0 {
			return readingClauses[0]
		}
	}

	return nil
}

func SinglePartQuery(expressions ...graph.Criteria) *model.RegularQuery {
	var (
		singlePartQuery = &model.SinglePartQuery{}
		query           = &model.RegularQuery{
			SingleQuery: &model.SingleQuery{
				SinglePartQuery: singlePartQuery,
			},
		}
	)

	for _, expression := range expressions {
		switch typedExpression := expression.(type) {
		case *model.Where:
			if firstReadingClause := GetFirstReadingClause(query); firstReadingClause != nil {
				firstReadingClause.Match.Where = typedExpression
			} else {
				singlePartQuery.AddReadingClause(&model.ReadingClause{
					Match: &model.Match{
						Where: typedExpression,
					},
					Unwind: nil,
				})
			}

		case *model.Return:
			singlePartQuery.Return = typedExpression

		case *model.Limit:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Limit = typedExpression
			}

		case *model.Skip:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Skip = typedExpression
			}

		case *model.Order:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Order = typedExpression
			}

		case *model.UpdatingClause:
			singlePartQuery.AddUpdatingClause(typedExpression)

		case []*model.UpdatingClause:
			for _, updatingClause := range typedExpression {
				singlePartQuery.AddUpdatingClause(updatingClause)
			}

		default:
			singlePartQuery.AddError(fmt.Errorf("invalid type for dawgs query: %T %+v", expression, expression))
		}
	}

	return query
}
