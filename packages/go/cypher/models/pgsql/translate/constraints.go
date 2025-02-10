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

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func leftNodeConstraint(edgeIdentifier, nodeIdentifier pgsql.Identifier, direction graph.Direction) (pgsql.Expression, error) {
	switch direction {
	case graph.DirectionOutbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionInbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionBoth:
		return pgsql.NewBinaryExpression(
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			),
			pgsql.OperatorOr,
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			),
		), nil

	default:
		return nil, fmt.Errorf("unsupported direction: %d", direction)
	}
}

func leftNodeTraversalStepConstraint(traversalStep *PatternSegment) (pgsql.Expression, error) {
	return leftNodeConstraint(
		traversalStep.Edge.Identifier,
		traversalStep.LeftNode.Identifier,
		traversalStep.Direction)
}

func rightEdgeConstraint(segment *PatternSegment, terminalEdge pgsql.Identifier, direction graph.Direction) (pgsql.Expression, error) {
	switch segment.Edge.DataType {
	case pgsql.EdgeComposite:
		switch direction {
		case graph.DirectionOutbound:
			return &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{segment.RightNode.Identifier, pgsql.ColumnID},
				ROperand: pgsql.CompoundIdentifier{terminalEdge, pgsql.ColumnStartID},
			}, nil

		case graph.DirectionInbound:
			return &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{segment.RightNode.Identifier, pgsql.ColumnID},
				ROperand: pgsql.CompoundIdentifier{terminalEdge, pgsql.ColumnEndID},
			}, nil

		default:
			return nil, fmt.Errorf("unsupported direction: %d", direction)
		}

	case pgsql.ExpansionEdge:
		switch direction {
		case graph.DirectionOutbound:
			return pgsql.NewBinaryExpression(
				pgsql.RowColumnReference{
					Identifier: &pgsql.ArrayIndex{
						Expression: segment.Edge.Identifier,
						Indexes: []pgsql.Expression{
							pgsql.FunctionCall{
								Function: pgsql.FunctionArrayLength,
								Parameters: []pgsql.Expression{
									segment.Edge.Identifier,
									pgsql.NewLiteral(1, pgsql.Int),
								},
								CastType: pgsql.Int,
							},
						},
					},
					Column: pgsql.ColumnEndID,
				},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{terminalEdge, pgsql.ColumnStartID},
			), nil

		case graph.DirectionInbound:
			return &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{segment.Edge.Identifier, pgsql.ColumnStartID},
				ROperand: pgsql.CompoundIdentifier{terminalEdge, pgsql.ColumnEndID},
			}, nil

		default:
			return nil, fmt.Errorf("unsupported direction: %d", direction)
		}

	default:
		return nil, fmt.Errorf("invalid root edge type: %s", segment.Edge.DataType)
	}
}

func terminalNodeConstraint(edgeIdentifier, nodeIdentifier pgsql.Identifier, direction graph.Direction) (pgsql.Expression, error) {
	switch direction {
	case graph.DirectionOutbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionInbound:
		return &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
		}, nil

	case graph.DirectionBoth:
		return pgsql.NewBinaryExpression(
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnEndID},
			),
			pgsql.OperatorOr,
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{nodeIdentifier, pgsql.ColumnID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{edgeIdentifier, pgsql.ColumnStartID},
			),
		), nil

	default:
		return nil, fmt.Errorf("unsupported direction: %d", direction)
	}
}

func rightNodeTraversalStepConstraint(traversalStep *PatternSegment) (pgsql.Expression, error) {
	return terminalNodeConstraint(
		traversalStep.Edge.Identifier,
		traversalStep.RightNode.Identifier,
		traversalStep.Direction)
}
