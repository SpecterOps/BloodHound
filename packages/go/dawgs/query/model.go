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

	cypherModel "github.com/specterops/bloodhound/cypher/model"
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

func Update(clauses ...*cypherModel.UpdatingClause) []*cypherModel.UpdatingClause {
	return clauses
}

func Updatef(provider graph.CriteriaProvider) []*cypherModel.UpdatingClause {
	switch typedCriteria := provider().(type) {
	case []*cypherModel.UpdatingClause:
		return typedCriteria

	case []graph.Criteria:
		return convertCriteria[*cypherModel.UpdatingClause](typedCriteria...)

	case *cypherModel.UpdatingClause:
		return []*cypherModel.UpdatingClause{typedCriteria}

	default:
		return []*cypherModel.UpdatingClause{
			cypherModel.WithErrors(&cypherModel.UpdatingClause{}, fmt.Errorf("invalid type %T for update clause", typedCriteria)),
		}
	}
}

func AddKind(reference graph.Criteria, kind graph.Kind) *cypherModel.UpdatingClause {
	return cypherModel.NewUpdatingClause(&cypherModel.Set{
		Items: []*cypherModel.SetItem{{
			Left:     reference,
			Operator: cypherModel.OperatorLabelAssignment,
			Right:    graph.Kinds{kind},
		}},
	})
}

func AddKinds(reference graph.Criteria, kinds graph.Kinds) *cypherModel.UpdatingClause {
	return cypherModel.NewUpdatingClause(&cypherModel.Set{
		Items: []*cypherModel.SetItem{{
			Left:     reference,
			Operator: cypherModel.OperatorLabelAssignment,
			Right:    kinds,
		}},
	})
}

func DeleteKind(reference graph.Criteria, kind graph.Kind) *cypherModel.UpdatingClause {
	return cypherModel.NewUpdatingClause(&cypherModel.Remove{
		Items: []*cypherModel.RemoveItem{{
			KindMatcher: &cypherModel.KindMatcher{
				Reference: reference,
				Kinds:     graph.Kinds{kind},
			},
		}},
	})
}

func DeleteKinds(reference graph.Criteria, kinds graph.Kinds) *cypherModel.Remove {
	return &cypherModel.Remove{
		Items: []*cypherModel.RemoveItem{{
			KindMatcher: &cypherModel.KindMatcher{
				Reference: reference,
				Kinds:     kinds,
			},
		}},
	}
}

func SetProperty(reference graph.Criteria, value any) *cypherModel.UpdatingClause {
	return cypherModel.NewUpdatingClause(&cypherModel.Set{
		Items: []*cypherModel.SetItem{{
			Left:     reference,
			Operator: cypherModel.OperatorAssignment,
			Right:    Parameter(value),
		}},
	})
}

func SetProperties(reference graph.Criteria, properties map[string]any) *cypherModel.UpdatingClause {
	set := &cypherModel.Set{}

	for key, value := range properties {
		set.Items = append(set.Items, &cypherModel.SetItem{
			Left:     Property(reference, key),
			Operator: cypherModel.OperatorAssignment,
			Right:    Parameter(value),
		})
	}

	return cypherModel.NewUpdatingClause(set)
}

func DeleteProperty(reference *cypherModel.PropertyLookup) *cypherModel.UpdatingClause {
	return cypherModel.NewUpdatingClause(&cypherModel.Remove{
		Items: []*cypherModel.RemoveItem{{
			Property: reference,
		}},
	})
}

func DeleteProperties(reference graph.Criteria, propertyNames ...string) *cypherModel.UpdatingClause {
	removeClause := &cypherModel.Remove{}

	for _, propertyName := range propertyNames {
		removeClause.Items = append(removeClause.Items, &cypherModel.RemoveItem{
			Property: Property(reference, propertyName),
		})
	}

	return cypherModel.NewUpdatingClause(removeClause)
}

func Kind(reference graph.Criteria, kind graph.Kind) *cypherModel.KindMatcher {
	return &cypherModel.KindMatcher{
		Reference: reference,
		Kinds:     graph.Kinds{kind},
	}
}

func KindIn(reference graph.Criteria, kinds ...graph.Kind) *cypherModel.Parenthetical {
	expressions := make([]graph.Criteria, len(kinds))

	for idx, kind := range kinds {
		expressions[idx] = Kind(reference, kind)
	}

	return Or(expressions...)
}

func NodeProperty(name string) *cypherModel.PropertyLookup {
	return cypherModel.NewPropertyLookup(NodeSymbol, name)
}

func RelationshipProperty(name string) *cypherModel.PropertyLookup {
	return cypherModel.NewPropertyLookup(EdgeSymbol, name)
}

func StartProperty(name string) *cypherModel.PropertyLookup {
	return cypherModel.NewPropertyLookup(EdgeStartSymbol, name)
}

func EndProperty(name string) *cypherModel.PropertyLookup {
	return cypherModel.NewPropertyLookup(EdgeEndSymbol, name)
}

func Property(qualifier graph.Criteria, name string) *cypherModel.PropertyLookup {
	return &cypherModel.PropertyLookup{
		Atom:    qualifier.(*cypherModel.Variable),
		Symbols: []string{name},
	}
}

func Count(reference graph.Criteria) *cypherModel.FunctionInvocation {
	return &cypherModel.FunctionInvocation{
		Name:      "count",
		Arguments: []cypherModel.Expression{reference},
	}
}

func CountDistinct(reference graph.Criteria) *cypherModel.FunctionInvocation {
	return &cypherModel.FunctionInvocation{
		Name:      "count",
		Distinct:  true,
		Arguments: []cypherModel.Expression{reference},
	}
}

func And(criteria ...graph.Criteria) *cypherModel.Conjunction {
	return cypherModel.NewConjunction(convertCriteria[cypherModel.Expression](criteria...)...)
}

func Or(criteria ...graph.Criteria) *cypherModel.Parenthetical {
	return &cypherModel.Parenthetical{
		Expression: cypherModel.NewDisjunction(convertCriteria[cypherModel.Expression](criteria...)...),
	}
}

func Xor(criteria ...graph.Criteria) *cypherModel.ExclusiveDisjunction {
	return cypherModel.NewExclusiveDisjunction(convertCriteria[cypherModel.Expression](criteria...)...)
}

func Parameter(value any) *cypherModel.Parameter {
	if parameter, isParameter := value.(*cypherModel.Parameter); isParameter {
		return parameter
	}

	return &cypherModel.Parameter{
		Value: value,
	}
}

func Literal(value any) *cypherModel.Literal {
	return &cypherModel.Literal{
		Value: value,
		Null:  value == nil,
	}
}

func KindsOf(ref graph.Criteria) *cypherModel.FunctionInvocation {
	switch typedRef := ref.(type) {
	case *cypherModel.Variable:
		switch typedRef.Symbol {
		case NodeSymbol, EdgeStartSymbol, EdgeEndSymbol:
			return &cypherModel.FunctionInvocation{
				Name:      "labels",
				Arguments: []cypherModel.Expression{ref},
			}

		case EdgeSymbol:
			return &cypherModel.FunctionInvocation{
				Name:      "type",
				Arguments: []cypherModel.Expression{ref},
			}

		default:
			return cypherModel.WithErrors(&cypherModel.FunctionInvocation{}, fmt.Errorf("invalid variable reference for KindsOf: %s", typedRef.Symbol))
		}

	default:
		return cypherModel.WithErrors(&cypherModel.FunctionInvocation{}, fmt.Errorf("invalid reference type for KindsOf: %T", ref))
	}
}

func Limit(limit int) *cypherModel.Limit {
	return &cypherModel.Limit{
		Value: Literal(limit),
	}
}

func Offset(offset int) *cypherModel.Skip {
	return &cypherModel.Skip{
		Value: Literal(offset),
	}
}

func StringContains(reference graph.Criteria, value string) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorContains, Parameter(value))
}

func StringStartsWith(reference graph.Criteria, value string) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorStartsWith, Parameter(value))
}

func StringEndsWith(reference graph.Criteria, value string) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorEndsWith, Parameter(value))
}

func CaseInsensitiveStringContains(reference graph.Criteria, value string) *cypherModel.Comparison {
	return cypherModel.NewComparison(
		cypherModel.NewSimpleFunctionInvocation("toLower", convertCriteria[cypherModel.Expression](reference)...),
		cypherModel.OperatorContains,
		Parameter(strings.ToLower(value)),
	)
}

func CaseInsensitiveStringStartsWith(reference graph.Criteria, value string) *cypherModel.Comparison {
	return cypherModel.NewComparison(
		cypherModel.NewSimpleFunctionInvocation("toLower", convertCriteria[cypherModel.Expression](reference)...),
		cypherModel.OperatorStartsWith,
		Parameter(strings.ToLower(value)),
	)
}

func CaseInsensitiveStringEndsWith(reference graph.Criteria, value string) *cypherModel.Comparison {
	return cypherModel.NewComparison(
		cypherModel.NewSimpleFunctionInvocation("toLower", convertCriteria[cypherModel.Expression](reference)...),
		cypherModel.OperatorEndsWith,
		Parameter(strings.ToLower(value)),
	)
}

func Equals(reference graph.Criteria, value any) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorEquals, Parameter(value))
}

func GreaterThan(reference graph.Criteria, value any) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorGreaterThan, Parameter(value))
}

func After(reference graph.Criteria, value any) *cypherModel.Comparison {
	return GreaterThan(reference, value)
}

func GreaterThanOrEquals(reference graph.Criteria, value any) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorGreaterThanOrEqualTo, Parameter(value))
}

func LessThan(reference graph.Criteria, value any) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorLessThan, Parameter(value))
}

func Before(reference graph.Criteria, value time.Time) *cypherModel.Comparison {
	return LessThan(reference, value)
}

func LessThanOrEquals(reference graph.Criteria, value any) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorLessThanOrEqualTo, Parameter(value))
}

func Exists(reference graph.Criteria) *cypherModel.Comparison {
	return cypherModel.NewComparison(
		reference,
		cypherModel.OperatorIsNot,
		cypherModel.NewLiteral(nil, true),
	)
}

func HasRelationships(reference *cypherModel.Variable) *cypherModel.PatternPredicate {
	patternPredicate := cypherModel.NewPatternPredicate()

	patternPredicate.AddElement(&cypherModel.NodePattern{
		Binding: cypherModel.NewVariableWithSymbol(reference.Symbol),
	})

	patternPredicate.AddElement(&cypherModel.RelationshipPattern{
		Direction: graph.DirectionBoth,
	})

	patternPredicate.AddElement(&cypherModel.NodePattern{})

	return patternPredicate
}

func In(reference graph.Criteria, value any) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorIn, Parameter(value))
}

func InIDs[T *cypherModel.FunctionInvocation | *cypherModel.Variable](reference T, ids ...graph.ID) *cypherModel.Comparison {
	switch any(reference).(type) {
	case *cypherModel.FunctionInvocation:
		return cypherModel.NewComparison(reference, cypherModel.OperatorIn, Parameter(ids))

	default:
		return cypherModel.NewComparison(Identity(any(reference).(*cypherModel.Variable)), cypherModel.OperatorIn, Parameter(ids))
	}
}

func Where(expression graph.Criteria) *cypherModel.Where {
	whereClause := cypherModel.NewWhere()
	whereClause.AddSlice(convertCriteria[cypherModel.Expression](expression))

	return whereClause
}

func OrderBy(leaves ...graph.Criteria) *cypherModel.Order {
	return &cypherModel.Order{
		Items: convertCriteria[*cypherModel.SortItem](leaves...),
	}
}

func Order(reference, direction graph.Criteria) *cypherModel.SortItem {
	switch direction {
	case cypherModel.SortDescending:
		return &cypherModel.SortItem{
			Ascending:  false,
			Expression: reference,
		}

	default:
		return &cypherModel.SortItem{
			Ascending:  true,
			Expression: reference,
		}
	}
}

func Ascending() cypherModel.SortOrder {
	return cypherModel.SortAscending
}

func Descending() cypherModel.SortOrder {
	return cypherModel.SortDescending
}

func Delete(leaves ...graph.Criteria) *cypherModel.UpdatingClause {
	deleteClause := &cypherModel.Delete{
		Detach: true,
	}

	for _, leaf := range leaves {
		switch leaf.(*cypherModel.Variable).Symbol {
		case EdgeSymbol, EdgeStartSymbol, EdgeEndSymbol:
			deleteClause.Detach = false
		}

		deleteClause.Expressions = append(deleteClause.Expressions, leaf)
	}

	return cypherModel.NewUpdatingClause(deleteClause)
}

func NodePattern(kinds graph.Kinds, properties *cypherModel.Parameter) *cypherModel.NodePattern {
	return &cypherModel.NodePattern{
		Binding:    cypherModel.NewVariableWithSymbol(NodeSymbol),
		Kinds:      kinds,
		Properties: properties,
	}
}

func StartNodePattern(kinds graph.Kinds, properties *cypherModel.Parameter) *cypherModel.NodePattern {
	return &cypherModel.NodePattern{
		Binding:    cypherModel.NewVariableWithSymbol(EdgeStartSymbol),
		Kinds:      kinds,
		Properties: properties,
	}
}

func EndNodePattern(kinds graph.Kinds, properties *cypherModel.Parameter) *cypherModel.NodePattern {
	return &cypherModel.NodePattern{
		Binding:    cypherModel.NewVariableWithSymbol(EdgeEndSymbol),
		Kinds:      kinds,
		Properties: properties,
	}
}

func RelationshipPattern(kind graph.Kind, properties *cypherModel.Parameter, direction graph.Direction) *cypherModel.RelationshipPattern {
	return &cypherModel.RelationshipPattern{
		Binding:    cypherModel.NewVariableWithSymbol(EdgeSymbol),
		Kinds:      graph.Kinds{kind},
		Properties: properties,
		Direction:  direction,
	}
}

func Create(elements ...graph.Criteria) *cypherModel.UpdatingClause {
	var (
		pattern      = &cypherModel.PatternPart{}
		createClause = &cypherModel.Create{
			// Note: Unique is Neo4j specific and will not be supported here. Use of constraints for
			// uniqueness is expected instead.
			Unique:  false,
			Pattern: []*cypherModel.PatternPart{pattern},
		}
	)

	for _, element := range elements {
		switch typedElement := element.(type) {
		case *cypherModel.Variable:
			switch typedElement.Symbol {
			case NodeSymbol, EdgeStartSymbol, EdgeEndSymbol:
				pattern.AddPatternElements(&cypherModel.NodePattern{
					Binding: cypherModel.NewVariableWithSymbol(typedElement.Symbol),
				})

			default:
				createClause.AddError(fmt.Errorf("invalid variable reference create: %s", typedElement.Symbol))
			}

		case *cypherModel.NodePattern:
			pattern.AddPatternElements(typedElement)

		case *cypherModel.RelationshipPattern:
			pattern.AddPatternElements(typedElement)

		default:
			createClause.AddError(fmt.Errorf("invalid type for create: %T", element))
		}
	}

	return cypherModel.NewUpdatingClause(createClause)
}

func ReturningDistinct(elements ...graph.Criteria) *cypherModel.Return {
	returnCriteria := Returning(elements...)
	returnCriteria.Projection.Distinct = true

	return returnCriteria
}

func Returning(elements ...graph.Criteria) *cypherModel.Return {
	projection := &cypherModel.Projection{}

	for _, element := range elements {
		switch typedElement := element.(type) {
		case *cypherModel.Order:
			projection.Order = typedElement

		case *cypherModel.Limit:
			projection.Limit = typedElement

		case *cypherModel.Skip:
			projection.Skip = typedElement

		default:
			projection.Items = append(projection.Items, &cypherModel.ProjectionItem{
				Expression: element,
			})
		}
	}

	return &cypherModel.Return{
		Projection: projection,
	}
}

func Not(expression graph.Criteria) *cypherModel.Negation {
	return &cypherModel.Negation{
		Expression: expression,
	}
}

func IsNull(reference graph.Criteria) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorIs, Literal(nil))
}

func IsNotNull(reference graph.Criteria) *cypherModel.Comparison {
	return cypherModel.NewComparison(reference, cypherModel.OperatorIsNot, Literal(nil))
}

func GetFirstReadingClause(query *cypherModel.RegularQuery) *cypherModel.ReadingClause {
	if query.SingleQuery != nil && query.SingleQuery.SinglePartQuery != nil {
		readingClauses := query.SingleQuery.SinglePartQuery.ReadingClauses

		if len(readingClauses) > 0 {
			return readingClauses[0]
		}
	}

	return nil
}

func SinglePartQuery(expressions ...graph.Criteria) *cypherModel.RegularQuery {
	var (
		singlePartQuery = &cypherModel.SinglePartQuery{}
		query           = &cypherModel.RegularQuery{
			SingleQuery: &cypherModel.SingleQuery{
				SinglePartQuery: singlePartQuery,
			},
		}
	)

	for _, expression := range expressions {
		switch typedExpression := expression.(type) {
		case *cypherModel.Where:
			if firstReadingClause := GetFirstReadingClause(query); firstReadingClause != nil {
				firstReadingClause.Match.Where = typedExpression
			} else {
				singlePartQuery.AddReadingClause(&cypherModel.ReadingClause{
					Match: &cypherModel.Match{
						Where: typedExpression,
					},
					Unwind: nil,
				})
			}

		case *cypherModel.Return:
			singlePartQuery.Return = typedExpression

		case *cypherModel.Limit:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Limit = typedExpression
			}

		case *cypherModel.Skip:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Skip = typedExpression
			}

		case *cypherModel.Order:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Order = typedExpression
			}

		case *cypherModel.UpdatingClause:
			singlePartQuery.AddUpdatingClause(typedExpression)

		case []*cypherModel.UpdatingClause:
			for _, updatingClause := range typedExpression {
				singlePartQuery.AddUpdatingClause(updatingClause)
			}

		default:
			singlePartQuery.AddError(fmt.Errorf("invalid type for dawgs query: %T %+v", expression, expression))
		}
	}

	return query
}

func EmptySinglePartQuery() *cypherModel.RegularQuery {
	return &cypherModel.RegularQuery{
		SingleQuery: &cypherModel.SingleQuery{
			SinglePartQuery: &cypherModel.SinglePartQuery{},
		},
	}
}
