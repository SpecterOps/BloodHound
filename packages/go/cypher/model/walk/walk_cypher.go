package walk

import (
	"fmt"

	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func newCypherWalkCursor(expression cypher.SyntaxNode) (*Cursor[cypher.SyntaxNode], error) {
	cursor := &Cursor[cypher.SyntaxNode]{
		Expression: expression,
	}

	switch typedExpression := expression.(type) {
	// Types with no AST branches
	case *cypher.RangeQuantifier, *cypher.PropertyLookup, *cypher.Literal, cypher.Operator, *cypher.Properties, *cypher.KindMatcher, *cypher.Limit, *cypher.Skip, graph.Kinds:
		return cursor, nil

	case *cypher.Create:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Pattern),
		}, nil

	case *cypher.Unwind:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Expression, typedExpression.Binding},
		}, nil

	case *cypher.RemoveItem:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if typedExpression.KindMatcher != nil {
			nextCursor.AddBranches(typedExpression.KindMatcher)
		}

		nextCursor.AddBranches(typedExpression.Property)
		return nextCursor, nil

	case *cypher.Remove:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Items),
		}, nil

	case *cypher.Delete:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Expressions),
		}, nil

	case *cypher.SetItem:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Left, typedExpression.Right},
		}, nil

	case *cypher.Set:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Items),
		}, nil

	case *cypher.UpdatingClause:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Clause},
		}, nil

	case *cypher.PatternPredicate:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.PatternElements),
		}, nil

	case *cypher.Order:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Items),
		}, nil

	case *cypher.SortItem:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Expression},
		}, nil

	case *cypher.MultiPartQuery:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   append(pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Parts), typedExpression.SinglePartQuery),
		}, nil

	case *cypher.MultiPartQueryPart:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if len(typedExpression.ReadingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.ReadingClauses)...)
		}

		if len(typedExpression.UpdatingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.UpdatingClauses)...)
		}

		if typedExpression.With != nil {
			nextCursor.AddBranches(typedExpression.With)
		}

		return nextCursor, nil

	case *cypher.With:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if typedExpression.Where != nil {
			nextCursor.AddBranches(typedExpression.Where)
		}

		if typedExpression.Projection != nil {
			nextCursor.AddBranches(typedExpression.Projection)
		}

		return nextCursor, nil

	case *cypher.Quantifier:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Filter},
		}, nil

	case *cypher.FilterExpression:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Specifier},
		}

		if typedExpression.Where != nil {
			nextCursor.AddBranches(typedExpression.Where)
		}

		return nextCursor, nil

	case *cypher.IDInCollection:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Expression},
		}, nil

	case *cypher.FunctionInvocation:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Arguments),
		}, nil

	case *cypher.Parenthetical:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Expression},
		}, nil

	case *cypher.RegularQuery:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.SingleQuery},
		}, nil

	case *cypher.SingleQuery:
		if typedExpression.SinglePartQuery != nil {
			return &Cursor[cypher.SyntaxNode]{
				Expression: expression,
				Branches:   []cypher.SyntaxNode{typedExpression.SinglePartQuery},
			}, nil
		}

		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.MultiPartQuery},
		}, nil

	case *cypher.SinglePartQuery:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if len(typedExpression.ReadingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.ReadingClauses)...)
		}

		if len(typedExpression.UpdatingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.UpdatingClauses)...)
		}

		if typedExpression.Return != nil {
			nextCursor.AddBranches(typedExpression.Return)
		}

		return nextCursor, nil

	case *cypher.Return:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Projection},
		}, nil

	case *cypher.Projection:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Items),
		}

		if typedExpression.Order != nil {
			nextCursor.AddBranches(typedExpression.Order)
		}

		if typedExpression.Skip != nil {
			nextCursor.AddBranches(typedExpression.Skip)
		}

		if typedExpression.Limit != nil {
			nextCursor.AddBranches(typedExpression.Limit)
		}

		return nextCursor, nil

	case *cypher.ProjectionItem:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Expression},
		}

		if typedExpression.Binding != nil {
			nextCursor.AddBranches(typedExpression.Binding)
		}

		return nextCursor, nil

	case *cypher.ReadingClause:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if typedExpression.Match != nil {
			nextCursor.AddBranches(typedExpression.Match)
		}

		if typedExpression.Unwind != nil {
			nextCursor.AddBranches(typedExpression.Unwind)
		}

		return nextCursor, nil

	case *cypher.Match:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Pattern),
		}

		if typedExpression.Where != nil {
			nextCursor.AddBranches(typedExpression.Where)
		}

		return nextCursor, nil

		// pattern parts are delinated by commas.  ie: `match (s), (e) return s` has 2 "PattenParts"
	case *cypher.PatternPart:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if typedExpression.Binding != nil {
			nextCursor.AddBranches(typedExpression.Binding)
		}

		nextCursor.AddBranches(pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.PatternElements)...)
		return nextCursor, nil

	case *cypher.PatternElement:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   []cypher.SyntaxNode{typedExpression.Element},
		}, nil

	case *cypher.RelationshipPattern:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if typedExpression.Binding != nil {
			nextCursor.AddBranches(typedExpression.Binding)
		}

		if typedExpression.Properties != nil {
			nextCursor.AddBranches(typedExpression.Properties)
		}

		return nextCursor, nil

	case *cypher.NodePattern:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}

		if typedExpression.Binding != nil {
			nextCursor.AddBranches(typedExpression.Binding)
		}

		if typedExpression.Properties != nil {
			nextCursor.AddBranches(typedExpression.Properties)
		}

		return nextCursor, nil

	case *cypher.Where:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Expressions),
		}, nil

	case *cypher.Variable:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
		}, nil

	case *cypher.ArithmeticExpression:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   append([]cypher.SyntaxNode{typedExpression.Left}, pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Partials)...),
		}, nil

	case *cypher.PartialArithmeticExpression:
		return &Cursor[cypher.SyntaxNode]{
			Expression:  expression,
			Branches:    []cypher.SyntaxNode{typedExpression.Operator, typedExpression.Right},
			BranchIndex: 0,
		}, nil

	case *cypher.PartialComparison:
		return &Cursor[cypher.SyntaxNode]{
			Expression:  expression,
			Branches:    []cypher.SyntaxNode{typedExpression.Operator, typedExpression.Right},
			BranchIndex: 0,
		}, nil

	case *cypher.Negation:
		return cursor, SetBranches(cursor, typedExpression.Expression)

	case *cypher.Conjunction:
		return cursor, SetBranches(cursor, typedExpression.Expressions...)

	case *cypher.Disjunction:
		return cursor, SetBranches(cursor, typedExpression.Expressions...)

	case *cypher.Comparison:
		return &Cursor[cypher.SyntaxNode]{
			Expression: expression,
			Branches:   append([]cypher.SyntaxNode{typedExpression.Left}, pgsql.MustSliceAs[cypher.SyntaxNode](typedExpression.Partials)...),
		}, nil

	default:
		return nil, fmt.Errorf("unable to negotiate cypher model type %T into a translation cursor", expression)
	}
}
