package translate

import (
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/walk"
)

const (
	// Below are a select set of constants to represent different weights to represent, roughly, the selectivity
	// of a given PGSQL expression. These weights are meant to be inexact and are only useful in comparison to other
	// summed weights.
	//
	// The goal of these weights are to enable reordering of queries such that the more selective side of a traversal
	// step is expanded first. Eventually, these weights may also enable reordering of multipart queries.

	// Entity ID references are a safe selectivity bet. A direct reference will typically take the form of:
	// `n0.id = 1` or some other direct comparison against the entity's ID. All entity IDs are covered by a unique
	// b-tree index, making them both highly selective and lucrative to weight higher.
	selectivityWeightEntityIDReference = 125

	// Unique node properties are both covered by a compatible index and unique, making them highly selective
	selectivityWeightUniqueNodeProperty = 100

	// Operators that narrow the search space are given a higher selectivity
	selectivityWeightNarrowSearch = 30

	// Operators that perform string searches are given a higher selectivity
	selectivityWeightStringSearch = 20

	// Operators that perform range comparisons are reasonably selective
	selectivityWeightRangeComparison = 10

	// Conjunctions can narrow search space, especially when compounded, but may be order dependent and unreliable as
	// a good selectivity heuristic
	selectivityWeightConjunction = 5

	// Exclusions can narrow the search space but often only slightly
	selectivityWeightNotEquals = 1

	// Disjunctions expand search space by adding a secondary, conditional operation
	selectivityWeightDisjunction = -10
)

// knownNodePropertySelectivity is a hack to enable the selectivity measurement to take advantage of known property indexes
// or uniqueness constraints.
//
// Eventually, this should be replaced by a tool that can introspect a graph schema and derive this map.
var knownNodePropertySelectivity = map[string]int{
	"objectid": selectivityWeightUniqueNodeProperty, // Object ID contains a unique constraint giving this a high degree of selectivity
	"name":     selectivityWeightUniqueNodeProperty, // Name contains a unique constraint giving this a high degree of selectivity
}

type measureSelectivityVisitor struct {
	scope       *Scope
	selectivity int
}

func (s *measureSelectivityVisitor) Measure(node pgsql.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
	switch typedNode := node.(type) {
	case *pgsql.BinaryExpression:
		switch typedLOperand := typedNode.LOperand.(type) {
		case pgsql.CompoundIdentifier:
			if typedLOperand.HasField() {
				switch typedLOperand.Field() {
				case pgsql.ColumnID:
					// Identifier references typically have high selectivity. This might be a nested reference, reducing the
					// effectiveness of the heuristic but the benefits outweigh this deficiency
					s.selectivity += selectivityWeightEntityIDReference
				}
			}
		}

		switch typedROperand := typedNode.ROperand.(type) {
		case pgsql.CompoundIdentifier:
			if typedROperand.HasField() {
				switch typedROperand.Field() {
				case pgsql.ColumnID:
					// Identifier references typically have high selectivity. This might be a nested reference, reducing the
					// effectiveness of the heuristic but the benefits outweigh this deficiency
					s.selectivity += selectivityWeightEntityIDReference
				}
			}
		}

		switch typedNode.Operator {
		case pgsql.OperatorOr:
			s.selectivity -= selectivityWeightDisjunction

		case pgsql.OperatorNotEquals:
			s.selectivity += selectivityWeightNotEquals

		case pgsql.OperatorAnd:
			s.selectivity += selectivityWeightConjunction

		case pgsql.OperatorLessThan, pgsql.OperatorGreaterThan, pgsql.OperatorLessThanOrEqualTo, pgsql.OperatorGreaterThanOrEqualTo:
			s.selectivity += selectivityWeightRangeComparison

		case pgsql.OperatorLike, pgsql.OperatorILike, pgsql.OperatorRegexMatch, pgsql.OperatorSimilarTo:
			s.selectivity += selectivityWeightStringSearch

		case pgsql.OperatorIn, pgsql.OperatorEquals, pgsql.OperatorIs, pgsql.OperatorPGArrayOverlap, pgsql.OperatorArrayOverlap:
			s.selectivity += selectivityWeightNarrowSearch

		case pgsql.OperatorJSONField, pgsql.OperatorJSONTextField, pgsql.OperatorPropertyLookup:
			if propertyLookup, err := binaryExpressionToPropertyLookup(typedNode); err != nil {
				errorHandler.SetError(err)
			} else {
				// Lookup the reference
				leftIdentifier := propertyLookup.Reference.Root()

				if binding, bound := s.scope.Lookup(leftIdentifier); !bound {
					errorHandler.SetErrorf("unable to lookup identifier %s", leftIdentifier)
				} else {
					switch binding.DataType {
					case pgsql.ExpansionRootNode, pgsql.ExpansionTerminalNode, pgsql.NodeComposite:
						// This is a node property, search through the available node property selectivity weights
						if selectivity, hasKnownSelectivity := knownNodePropertySelectivity[propertyLookup.Field]; hasKnownSelectivity {
							s.selectivity += selectivity
						}
					}
				}
			}
		}
	}
}

// MeasureSelectivity attempts to measure how selective (i.e. how narrow) the query expression passed in is. This is
// a simple heuristic that is best-effort for attempting to find which side of a traversal step ()-[]->() is more
// selective.
//
// The boolean parameter owningIdentifierBound is intended to represent if the identifier the expression constraints
// is part of a materialized set of nodes where the entity IDs of each are known at time of query. In this case the
// bound component is considered to be highly-selective.
//
// Many numbers are magic values selected based on implementor's perception of selectivity of certain operators.
func MeasureSelectivity(scope *Scope, owningIdentifierBound bool, expression pgsql.Expression) (int, error) {
	visitor := &measureSelectivityVisitor{
		scope:       scope,
		selectivity: 0,
	}

	// If the identifier is reified at this stage in the query then it's already selected
	if owningIdentifierBound {
		visitor.selectivity += 1000
	}

	return visitor.selectivity, walk.PgSQL(expression, walk.NewSimpleVisitor[pgsql.SyntaxNode](visitor.Measure))
}
