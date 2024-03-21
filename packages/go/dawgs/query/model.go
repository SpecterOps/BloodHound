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
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"strings"
	"time"

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

func Update(clauses ...*cypher.UpdatingClause) []*cypher.UpdatingClause {
	return clauses
}

func Updatef(provider graph.CriteriaProvider) []*cypher.UpdatingClause {
	switch typedCriteria := provider().(type) {
	case []*cypher.UpdatingClause:
		return typedCriteria

	case []graph.Criteria:
		return convertCriteria[*cypher.UpdatingClause](typedCriteria...)

	case *cypher.UpdatingClause:
		return []*cypher.UpdatingClause{typedCriteria}

	default:
		return []*cypher.UpdatingClause{
			cypher.WithErrors(&cypher.UpdatingClause{}, fmt.Errorf("invalid type %T for update clause", typedCriteria)),
		}
	}
}

func AddKind(reference graph.Criteria, kind graph.Kind) *cypher.UpdatingClause {
	return cypher.NewUpdatingClause(&cypher.Set{
		Items: []*cypher.SetItem{{
			Left:     reference,
			Operator: cypher.OperatorLabelAssignment,
			Right:    graph.Kinds{kind},
		}},
	})
}

func AddKinds(reference graph.Criteria, kinds graph.Kinds) *cypher.UpdatingClause {
	return cypher.NewUpdatingClause(&cypher.Set{
		Items: []*cypher.SetItem{{
			Left:     reference,
			Operator: cypher.OperatorLabelAssignment,
			Right:    kinds,
		}},
	})
}

func DeleteKind(reference graph.Criteria, kind graph.Kind) *cypher.UpdatingClause {
	return cypher.NewUpdatingClause(&cypher.Remove{
		Items: []*cypher.RemoveItem{{
			KindMatcher: &cypher.KindMatcher{
				Reference: reference,
				Kinds:     graph.Kinds{kind},
			},
		}},
	})
}

func DeleteKinds(reference graph.Criteria, kinds graph.Kinds) *cypher.Remove {
	return &cypher.Remove{
		Items: []*cypher.RemoveItem{{
			KindMatcher: &cypher.KindMatcher{
				Reference: reference,
				Kinds:     kinds,
			},
		}},
	}
}

func SetProperty(reference graph.Criteria, value any) *cypher.UpdatingClause {
	return cypher.NewUpdatingClause(&cypher.Set{
		Items: []*cypher.SetItem{{
			Left:     reference,
			Operator: cypher.OperatorAssignment,
			Right:    Parameter(value),
		}},
	})
}

func SetProperties(reference graph.Criteria, properties map[string]any) *cypher.UpdatingClause {
	set := &cypher.Set{}

	for key, value := range properties {
		set.Items = append(set.Items, &cypher.SetItem{
			Left:     Property(reference, key),
			Operator: cypher.OperatorAssignment,
			Right:    Parameter(value),
		})
	}

	return cypher.NewUpdatingClause(set)
}

func DeleteProperty(reference *cypher.PropertyLookup) *cypher.UpdatingClause {
	return cypher.NewUpdatingClause(&cypher.Remove{
		Items: []*cypher.RemoveItem{{
			Property: reference,
		}},
	})
}

func DeleteProperties(reference graph.Criteria, propertyNames ...string) *cypher.UpdatingClause {
	removeClause := &cypher.Remove{}

	for _, propertyName := range propertyNames {
		removeClause.Items = append(removeClause.Items, &cypher.RemoveItem{
			Property: Property(reference, propertyName),
		})
	}

	return cypher.NewUpdatingClause(removeClause)
}

func Kind(reference graph.Criteria, kind graph.Kind) *cypher.KindMatcher {
	return &cypher.KindMatcher{
		Reference: reference,
		Kinds:     graph.Kinds{kind},
	}
}

func KindIn(reference graph.Criteria, kinds ...graph.Kind) *cypher.Parenthetical {
	expressions := make([]graph.Criteria, len(kinds))

	for idx, kind := range kinds {
		expressions[idx] = Kind(reference, kind)
	}

	return Or(expressions...)
}

func NodeProperty(name string) *cypher.PropertyLookup {
	return cypher.NewPropertyLookup(NodeSymbol, name)
}

func RelationshipProperty(name string) *cypher.PropertyLookup {
	return cypher.NewPropertyLookup(EdgeSymbol, name)
}

func StartProperty(name string) *cypher.PropertyLookup {
	return cypher.NewPropertyLookup(EdgeStartSymbol, name)
}

func EndProperty(name string) *cypher.PropertyLookup {
	return cypher.NewPropertyLookup(EdgeEndSymbol, name)
}

func Property(qualifier graph.Criteria, name string) *cypher.PropertyLookup {
	return &cypher.PropertyLookup{
		Atom:    qualifier.(*cypher.Variable),
		Symbols: []string{name},
	}
}

func Count(reference graph.Criteria) *cypher.FunctionInvocation {
	return &cypher.FunctionInvocation{
		Name:      "count",
		Arguments: []cypher.Expression{reference},
	}
}

func CountDistinct(reference graph.Criteria) *cypher.FunctionInvocation {
	return &cypher.FunctionInvocation{
		Name:      "count",
		Distinct:  true,
		Arguments: []cypher.Expression{reference},
	}
}

func And(criteria ...graph.Criteria) *cypher.Conjunction {
	return cypher.NewConjunction(convertCriteria[cypher.Expression](criteria...)...)
}

func Or(criteria ...graph.Criteria) *cypher.Parenthetical {
	return &cypher.Parenthetical{
		Expression: cypher.NewDisjunction(convertCriteria[cypher.Expression](criteria...)...),
	}
}

func Xor(criteria ...graph.Criteria) *cypher.ExclusiveDisjunction {
	return cypher.NewExclusiveDisjunction(convertCriteria[cypher.Expression](criteria...)...)
}

func Parameter(value any) *cypher.Parameter {
	if parameter, isParameter := value.(*cypher.Parameter); isParameter {
		return parameter
	}

	return &cypher.Parameter{
		Value: value,
	}
}

func Literal(value any) *cypher.Literal {
	return &cypher.Literal{
		Value: value,
		Null:  value == nil,
	}
}

func KindsOf(ref graph.Criteria) *cypher.FunctionInvocation {
	switch typedRef := ref.(type) {
	case *cypher.Variable:
		switch typedRef.Symbol {
		case NodeSymbol, EdgeStartSymbol, EdgeEndSymbol:
			return &cypher.FunctionInvocation{
				Name:      "labels",
				Arguments: []cypher.Expression{ref},
			}

		case EdgeSymbol:
			return &cypher.FunctionInvocation{
				Name:      "type",
				Arguments: []cypher.Expression{ref},
			}

		default:
			return cypher.WithErrors(&cypher.FunctionInvocation{}, fmt.Errorf("invalid variable reference for KindsOf: %s", typedRef.Symbol))
		}

	default:
		return cypher.WithErrors(&cypher.FunctionInvocation{}, fmt.Errorf("invalid reference type for KindsOf: %T", ref))
	}
}

func Limit(limit int) *cypher.Limit {
	return &cypher.Limit{
		Value: Literal(limit),
	}
}

func Offset(offset int) *cypher.Skip {
	return &cypher.Skip{
		Value: Literal(offset),
	}
}

func StringContains(reference graph.Criteria, value string) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorContains, Parameter(value))
}

func StringStartsWith(reference graph.Criteria, value string) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorStartsWith, Parameter(value))
}

func StringEndsWith(reference graph.Criteria, value string) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorEndsWith, Parameter(value))
}

func CaseInsensitiveStringContains(reference graph.Criteria, value string) *cypher.Comparison {
	return cypher.NewComparison(
		cypher.NewSimpleFunctionInvocation("toLower", convertCriteria[cypher.Expression](reference)...),
		cypher.OperatorContains,
		Parameter(strings.ToLower(value)),
	)
}

func CaseInsensitiveStringStartsWith(reference graph.Criteria, value string) *cypher.Comparison {
	return cypher.NewComparison(
		cypher.NewSimpleFunctionInvocation("toLower", convertCriteria[cypher.Expression](reference)...),
		cypher.OperatorStartsWith,
		Parameter(strings.ToLower(value)),
	)
}

func CaseInsensitiveStringEndsWith(reference graph.Criteria, value string) *cypher.Comparison {
	return cypher.NewComparison(
		cypher.NewSimpleFunctionInvocation("toLower", convertCriteria[cypher.Expression](reference)...),
		cypher.OperatorEndsWith,
		Parameter(strings.ToLower(value)),
	)
}

func Equals(reference graph.Criteria, value any) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorEquals, Parameter(value))
}

func GreaterThan(reference graph.Criteria, value any) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorGreaterThan, Parameter(value))
}

func After(reference graph.Criteria, value any) *cypher.Comparison {
	return GreaterThan(reference, value)
}

func GreaterThanOrEquals(reference graph.Criteria, value any) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorGreaterThanOrEqualTo, Parameter(value))
}

func LessThan(reference graph.Criteria, value any) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorLessThan, Parameter(value))
}

func Before(reference graph.Criteria, value time.Time) *cypher.Comparison {
	return LessThan(reference, value)
}

func LessThanOrEquals(reference graph.Criteria, value any) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorLessThanOrEqualTo, Parameter(value))
}

func Exists(reference graph.Criteria) *cypher.Comparison {
	return cypher.NewComparison(
		reference,
		cypher.OperatorIsNot,
		cypher.NewLiteral(nil, true),
	)
}

func HasRelationships(reference *cypher.Variable) *cypher.PatternPredicate {
	patternPredicate := cypher.NewPatternPredicate()

	patternPredicate.AddElement(&cypher.NodePattern{
		Binding: cypher.NewVariableWithSymbol(reference.Symbol),
	})

	patternPredicate.AddElement(&cypher.RelationshipPattern{
		Direction: graph.DirectionBoth,
	})

	patternPredicate.AddElement(&cypher.NodePattern{})

	return patternPredicate
}

func In(reference graph.Criteria, value any) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorIn, Parameter(value))
}

func InIDs[T *cypher.FunctionInvocation | *cypher.Variable](reference T, ids ...graph.ID) *cypher.Comparison {
	switch any(reference).(type) {
	case *cypher.FunctionInvocation:
		return cypher.NewComparison(reference, cypher.OperatorIn, Parameter(ids))

	default:
		return cypher.NewComparison(Identity(any(reference).(*cypher.Variable)), cypher.OperatorIn, Parameter(ids))
	}
}

func Where(expression graph.Criteria) *cypher.Where {
	whereClause := cypher.NewWhere()
	whereClause.AddSlice(convertCriteria[cypher.Expression](expression))

	return whereClause
}

func OrderBy(leaves ...graph.Criteria) *cypher.Order {
	return &cypher.Order{
		Items: convertCriteria[*cypher.SortItem](leaves...),
	}
}

func Order(reference, direction graph.Criteria) *cypher.SortItem {
	switch direction {
	case cypher.SortDescending:
		return &cypher.SortItem{
			Ascending:  false,
			Expression: reference,
		}

	default:
		return &cypher.SortItem{
			Ascending:  true,
			Expression: reference,
		}
	}
}

func Ascending() cypher.SortOrder {
	return cypher.SortAscending
}

func Descending() cypher.SortOrder {
	return cypher.SortDescending
}

func Delete(leaves ...graph.Criteria) *cypher.UpdatingClause {
	deleteClause := &cypher.Delete{
		Detach: true,
	}

	for _, leaf := range leaves {
		switch leaf.(*cypher.Variable).Symbol {
		case EdgeSymbol, EdgeStartSymbol, EdgeEndSymbol:
			deleteClause.Detach = false
		}

		deleteClause.Expressions = append(deleteClause.Expressions, leaf)
	}

	return cypher.NewUpdatingClause(deleteClause)
}

func NodePattern(kinds graph.Kinds, properties *cypher.Parameter) *cypher.NodePattern {
	return &cypher.NodePattern{
		Binding:    cypher.NewVariableWithSymbol(NodeSymbol),
		Kinds:      kinds,
		Properties: properties,
	}
}

func StartNodePattern(kinds graph.Kinds, properties *cypher.Parameter) *cypher.NodePattern {
	return &cypher.NodePattern{
		Binding:    cypher.NewVariableWithSymbol(EdgeStartSymbol),
		Kinds:      kinds,
		Properties: properties,
	}
}

func EndNodePattern(kinds graph.Kinds, properties *cypher.Parameter) *cypher.NodePattern {
	return &cypher.NodePattern{
		Binding:    cypher.NewVariableWithSymbol(EdgeEndSymbol),
		Kinds:      kinds,
		Properties: properties,
	}
}

func RelationshipPattern(kind graph.Kind, properties *cypher.Parameter, direction graph.Direction) *cypher.RelationshipPattern {
	return &cypher.RelationshipPattern{
		Binding:    cypher.NewVariableWithSymbol(EdgeSymbol),
		Kinds:      graph.Kinds{kind},
		Properties: properties,
		Direction:  direction,
	}
}

func Create(elements ...graph.Criteria) *cypher.UpdatingClause {
	var (
		pattern      = &cypher.PatternPart{}
		createClause = &cypher.Create{
			// Note: Unique is Neo4j specific and will not be supported here. Use of constraints for
			// uniqueness is expected instead.
			Unique:  false,
			Pattern: []*cypher.PatternPart{pattern},
		}
	)

	for _, element := range elements {
		switch typedElement := element.(type) {
		case *cypher.Variable:
			switch typedElement.Symbol {
			case NodeSymbol, EdgeStartSymbol, EdgeEndSymbol:
				pattern.AddPatternElements(&cypher.NodePattern{
					Binding: cypher.NewVariableWithSymbol(typedElement.Symbol),
				})

			default:
				createClause.AddError(fmt.Errorf("invalid variable reference create: %s", typedElement.Symbol))
			}

		case *cypher.NodePattern:
			pattern.AddPatternElements(typedElement)

		case *cypher.RelationshipPattern:
			pattern.AddPatternElements(typedElement)

		default:
			createClause.AddError(fmt.Errorf("invalid type for create: %T", element))
		}
	}

	return cypher.NewUpdatingClause(createClause)
}

func ReturningDistinct(elements ...graph.Criteria) *cypher.Return {
	returnCriteria := Returning(elements...)
	returnCriteria.Projection.Distinct = true

	return returnCriteria
}

func Returning(elements ...graph.Criteria) *cypher.Return {
	projection := &cypher.Projection{}

	for _, element := range elements {
		switch typedElement := element.(type) {
		case *cypher.Order:
			projection.Order = typedElement

		case *cypher.Limit:
			projection.Limit = typedElement

		case *cypher.Skip:
			projection.Skip = typedElement

		default:
			projection.Items = append(projection.Items, &cypher.ProjectionItem{
				Expression: element,
			})
		}
	}

	return &cypher.Return{
		Projection: projection,
	}
}

func Not(expression graph.Criteria) *cypher.Negation {
	return &cypher.Negation{
		Expression: expression,
	}
}

func IsNull(reference graph.Criteria) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorIs, Literal(nil))
}

func IsNotNull(reference graph.Criteria) *cypher.Comparison {
	return cypher.NewComparison(reference, cypher.OperatorIsNot, Literal(nil))
}

func GetFirstReadingClause(query *cypher.RegularQuery) *cypher.ReadingClause {
	if query.SingleQuery != nil && query.SingleQuery.SinglePartQuery != nil {
		readingClauses := query.SingleQuery.SinglePartQuery.ReadingClauses

		if len(readingClauses) > 0 {
			return readingClauses[0]
		}
	}

	return nil
}

func SinglePartQuery(expressions ...graph.Criteria) *cypher.RegularQuery {
	var (
		singlePartQuery = &cypher.SinglePartQuery{}
		query           = &cypher.RegularQuery{
			SingleQuery: &cypher.SingleQuery{
				SinglePartQuery: singlePartQuery,
			},
		}
	)

	for _, expression := range expressions {
		switch typedExpression := expression.(type) {
		case *cypher.Where:
			if firstReadingClause := GetFirstReadingClause(query); firstReadingClause != nil {
				firstReadingClause.Match.Where = typedExpression
			} else {
				singlePartQuery.AddReadingClause(&cypher.ReadingClause{
					Match: &cypher.Match{
						Where: typedExpression,
					},
					Unwind: nil,
				})
			}

		case *cypher.Return:
			singlePartQuery.Return = typedExpression

		case *cypher.Limit:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Limit = typedExpression
			}

		case *cypher.Skip:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Skip = typedExpression
			}

		case *cypher.Order:
			if singlePartQuery.Return != nil {
				singlePartQuery.Return.Projection.Order = typedExpression
			}

		case *cypher.UpdatingClause:
			singlePartQuery.AddUpdatingClause(typedExpression)

		case []*cypher.UpdatingClause:
			for _, updatingClause := range typedExpression {
				singlePartQuery.AddUpdatingClause(updatingClause)
			}

		default:
			singlePartQuery.AddError(fmt.Errorf("invalid type for dawgs query: %T %+v", expression, expression))
		}
	}

	return query
}

func EmptySinglePartQuery() *cypher.RegularQuery {
	return &cypher.RegularQuery{
		SingleQuery: &cypher.SingleQuery{
			SinglePartQuery: &cypher.SinglePartQuery{},
		},
	}
}
