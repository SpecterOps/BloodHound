package walk

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func newCypherWalkCursor(node cypher.SyntaxNode) (*Cursor[cypher.SyntaxNode], error) {
	cursor := &Cursor[cypher.SyntaxNode]{
		Node: node,
	}

	switch typedNode := node.(type) {
	// Types with no AST branches
	case *cypher.RangeQuantifier, *cypher.PropertyLookup, *cypher.Literal, cypher.Operator, *cypher.Properties, *cypher.KindMatcher, *cypher.Limit, *cypher.Skip, graph.Kinds:
		return cursor, nil

	case *cypher.Create:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Pattern),
		}, nil

	case *cypher.Unwind:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Expression, typedNode.Binding},
		}, nil

	case *cypher.RemoveItem:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if typedNode.KindMatcher != nil {
			nextCursor.AddBranches(typedNode.KindMatcher)
		}

		nextCursor.AddBranches(typedNode.Property)
		return nextCursor, nil

	case *cypher.Remove:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Items),
		}, nil

	case *cypher.Delete:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Expressions),
		}, nil

	case *cypher.SetItem:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Left, typedNode.Right},
		}, nil

	case *cypher.Set:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Items),
		}, nil

	case *cypher.UpdatingClause:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Clause},
		}, nil

	case *cypher.PatternPredicate:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.PatternElements),
		}, nil

	case *cypher.Order:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Items),
		}, nil

	case *cypher.SortItem:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Expression},
		}, nil

	case *cypher.MultiPartQuery:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: append(pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Parts), typedNode.SinglePartQuery),
		}, nil

	case *cypher.MultiPartQueryPart:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if len(typedNode.ReadingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.ReadingClauses)...)
		}

		if len(typedNode.UpdatingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.UpdatingClauses)...)
		}

		if typedNode.With != nil {
			nextCursor.AddBranches(typedNode.With)
		}

		return nextCursor, nil

	case *cypher.With:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if typedNode.Where != nil {
			nextCursor.AddBranches(typedNode.Where)
		}

		if typedNode.Projection != nil {
			nextCursor.AddBranches(typedNode.Projection)
		}

		return nextCursor, nil

	case *cypher.Quantifier:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Filter},
		}, nil

	case *cypher.FilterExpression:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Specifier},
		}

		if typedNode.Where != nil {
			nextCursor.AddBranches(typedNode.Where)
		}

		return nextCursor, nil

	case *cypher.IDInCollection:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Expression},
		}, nil

	case *cypher.FunctionInvocation:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Arguments),
		}, nil

	case *cypher.Parenthetical:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Expression},
		}, nil

	case *cypher.RegularQuery:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.SingleQuery},
		}, nil

	case *cypher.SingleQuery:
		if typedNode.SinglePartQuery != nil {
			return &Cursor[cypher.SyntaxNode]{
				Node:     node,
				Branches: []cypher.SyntaxNode{typedNode.SinglePartQuery},
			}, nil
		}

		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.MultiPartQuery},
		}, nil

	case *cypher.SinglePartQuery:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if len(typedNode.ReadingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.ReadingClauses)...)
		}

		if len(typedNode.UpdatingClauses) > 0 {
			nextCursor.AddBranches(pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.UpdatingClauses)...)
		}

		if typedNode.Return != nil {
			nextCursor.AddBranches(typedNode.Return)
		}

		return nextCursor, nil

	case *cypher.Return:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Projection},
		}, nil

	case *cypher.Projection:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Items),
		}

		if typedNode.Order != nil {
			nextCursor.AddBranches(typedNode.Order)
		}

		if typedNode.Skip != nil {
			nextCursor.AddBranches(typedNode.Skip)
		}

		if typedNode.Limit != nil {
			nextCursor.AddBranches(typedNode.Limit)
		}

		return nextCursor, nil

	case *cypher.ProjectionItem:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Expression},
		}

		if typedNode.Binding != nil {
			nextCursor.AddBranches(typedNode.Binding)
		}

		return nextCursor, nil

	case *cypher.ReadingClause:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if typedNode.Match != nil {
			nextCursor.AddBranches(typedNode.Match)
		}

		if typedNode.Unwind != nil {
			nextCursor.AddBranches(typedNode.Unwind)
		}

		return nextCursor, nil

	case *cypher.Match:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Pattern),
		}

		if typedNode.Where != nil {
			nextCursor.AddBranches(typedNode.Where)
		}

		return nextCursor, nil

	case *cypher.PatternPart:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if typedNode.Binding != nil {
			nextCursor.AddBranches(typedNode.Binding)
		}

		nextCursor.AddBranches(pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.PatternElements)...)
		return nextCursor, nil

	case *cypher.PatternElement:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: []cypher.SyntaxNode{typedNode.Element},
		}, nil

	case *cypher.RelationshipPattern:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if typedNode.Binding != nil {
			nextCursor.AddBranches(typedNode.Binding)
		}

		if typedNode.Properties != nil {
			nextCursor.AddBranches(typedNode.Properties)
		}

		return nextCursor, nil

	case *cypher.NodePattern:
		nextCursor := &Cursor[cypher.SyntaxNode]{
			Node: node,
		}

		if typedNode.Binding != nil {
			nextCursor.AddBranches(typedNode.Binding)
		}

		if typedNode.Properties != nil {
			nextCursor.AddBranches(typedNode.Properties)
		}

		return nextCursor, nil

	case *cypher.Where:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Expressions),
		}, nil

	case *cypher.Variable:
		return &Cursor[cypher.SyntaxNode]{
			Node: node,
		}, nil

	case *cypher.ArithmeticExpression:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: append([]cypher.SyntaxNode{typedNode.Left}, pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Partials)...),
		}, nil

	case *cypher.PartialArithmeticExpression:
		return &Cursor[cypher.SyntaxNode]{
			Node:        node,
			Branches:    []cypher.SyntaxNode{typedNode.Operator, typedNode.Right},
			BranchIndex: 0,
		}, nil

	case *cypher.PartialComparison:
		return &Cursor[cypher.SyntaxNode]{
			Node:        node,
			Branches:    []cypher.SyntaxNode{typedNode.Operator, typedNode.Right},
			BranchIndex: 0,
		}, nil

	case *cypher.Negation:
		return cursor, SetBranches(cursor, typedNode.Expression)

	case *cypher.Conjunction:
		return cursor, SetBranches(cursor, typedNode.Expressions...)

	case *cypher.Disjunction:
		return cursor, SetBranches(cursor, typedNode.Expressions...)

	case *cypher.Comparison:
		return &Cursor[cypher.SyntaxNode]{
			Node:     node,
			Branches: append([]cypher.SyntaxNode{typedNode.Left}, pgsql.MustSliceTypeConvert[cypher.SyntaxNode](typedNode.Partials)...),
		}, nil

	default:
		return nil, fmt.Errorf("unable to negotiate cypher model type %T into a translation cursor", node)
	}
}
