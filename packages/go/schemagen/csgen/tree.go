package csgen

type cursor struct {
	node     SyntaxNode
	branches []SyntaxNode
	idx      int
}

func WalkSyntaxTree(node SyntaxNode, builder *OutputBuilder) error {

	stack := []cursor{newCursor(node)}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// This is the first time we are visiting this node
		if current.idx == 0 {
			switch typedNode := current.node.(type) {
			case Namespace:
				builder.Write(typedNode.Enter())
			case Class:
				builder.Write(typedNode.Enter())
			}
		}

		if current.idx == len(current.branches) {
			switch typedNode := current.node.(type) {
			case Namespace:
				builder.Write(typedNode.Exit())
			case Class:
				builder.Write(typedNode.Exit())
			case BinaryExpression:
				stack = append(stack,
					newCursor(FormattingLiteralNewline),
					newCursor(FormattingLiteralSemicolon),
					newCursor(typedNode.RightOperand),
					newCursor(FormattingLiteralSpace),
					newCursor(typedNode.Operator),
					newCursor(FormattingLiteralSpace),
					newCursor(typedNode.LeftOperand))
			case ClassMemberAssignment:
				stack = append(stack,
					newCursor(typedNode.Symbol),
					newCursor(FormattingLiteralSpace),
					newCursor(typedNode.Type),
					newCursor(FormattingLiteralSpace),
					newCursor(typedNode.Modifier),
					newCursor(FormattingLiteralSpace),
					newCursor(typedNode.Visibility))
			case Type:
				builder.Write(typedNode.String())
			case Symbol:
				builder.Write(string(typedNode))
			case FormattingLiteral:
				builder.Write(typedNode.String())
			case Operator:
				builder.Write(typedNode.String())
			case Visibility:
				builder.Write(typedNode.String())
			case Modifier:
				builder.Write(typedNode.String())
			case Modifiers:
				builder.Write(typedNode.String())
			case Literal:
				if err := formatLiteral(builder, typedNode); err != nil {
					return err
				}
			}

		} else {
			// Next child
			nextChild := current.branches[current.idx]
			current.idx += 1

			stack = append(stack, current, cursor{
				node:     nextChild,
				branches: nextChild.Children(),
				idx:      0,
			})

			continue
		}
	}

	return nil
}

func newCursor(node SyntaxNode) cursor {
	return cursor{
		node:     node,
		branches: node.Children(),
		idx:      0,
	}
}
