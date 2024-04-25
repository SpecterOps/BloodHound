package walk

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
)

func newSQLWalkCursor(node pgsql.SyntaxNode) (*Cursor[pgsql.SyntaxNode], error) {
	switch typedExpression := node.(type) {
	case pgsql.Query:
		nextCursor := &Cursor[pgsql.SyntaxNode]{
			Node: node,
		}

		if typedExpression.CommonTableExpressions != nil {
			nextCursor.Branches = append(nextCursor.Branches, *typedExpression.CommonTableExpressions)
		}

		nextCursor.Branches = append(nextCursor.Branches, typedExpression.Body)
		return nextCursor, nil

	case pgsql.With:
		return &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.Expressions),
		}, nil

	case pgsql.CommonTableExpression:
		return &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: []pgsql.SyntaxNode{typedExpression.Query, typedExpression.Alias},
		}, nil

	case pgsql.Select:
		nextCursor := &Cursor[pgsql.SyntaxNode]{
			Node: node,
		}

		nextCursor.AddBranches(pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.Projection)...)
		nextCursor.AddBranches(pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.From)...)

		if typedExpression.Where != nil {
			nextCursor.AddBranches(typedExpression.Where)
		}

		nextCursor.AddBranches(pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.From)...)

		if typedExpression.Having != nil {
			nextCursor.AddBranches(typedExpression.Having)
		}

		return nextCursor, nil

	case pgsql.FromClause:
		nextCursor := &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: []pgsql.SyntaxNode{typedExpression.Relation},
		}

		nextCursor.AddBranches(pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.Joins)...)
		return nextCursor, nil

	case pgsql.AliasedExpression:
		return &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: []pgsql.SyntaxNode{typedExpression.Expression, typedExpression.Alias},
		}, nil

	case pgsql.CompositeValue:
		return &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.Values),
		}, nil

	case pgsql.TableReference:
		nextCursor := &Cursor[pgsql.SyntaxNode]{
			Node:        node,
			Branches:    []pgsql.SyntaxNode{typedExpression.Name},
		}

		if typedExpression.Binding.Set {
			nextCursor.AddBranches(typedExpression.Binding.Value)
		}

		return nextCursor, nil

	case pgsql.TableAlias:
		nextCursor := &Cursor[pgsql.SyntaxNode]{
			Node: node,
			Branches: []pgsql.SyntaxNode{typedExpression.Name},
		}

		if typedExpression.Shape.Set {
			nextCursor.AddBranches(typedExpression.Shape.Value)
		}

		return nextCursor, nil

	case pgsql.RowShape:
		return &Cursor[pgsql.SyntaxNode]{
			Node: node,
			Branches: pgsql.MustSliceTypeConvert[pgsql.SyntaxNode](typedExpression.Columns),
		}, nil

	case pgsql.CompoundIdentifier, pgsql.Operator, pgsql.Literal, pgsql.Identifier:
		return &Cursor[pgsql.SyntaxNode]{
			Node: node,
		}, nil

	case *pgsql.UnaryExpression:
		return &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: []pgsql.SyntaxNode{typedExpression.Operator, typedExpression.Operand},
		}, nil

	case pgsql.BinaryExpression:
		return &Cursor[pgsql.SyntaxNode]{
			Node:     node,
			Branches: []pgsql.SyntaxNode{typedExpression.LOperand, typedExpression.ROperand},
		}, nil

	case *pgsql.BinaryExpression:
		// Choose to unwrap the pointer to force logic for the binary expression through one route
		return newSQLWalkCursor(*typedExpression)

	default:
		return nil, fmt.Errorf("unable to negotiate sql type %T into a translation cursor", node)
	}
}
