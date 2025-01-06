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

package visualization

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cypher/models/walk"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
)

type SQLVisualizer struct {
	walk.HierarchicalVisitor[pgsql.SyntaxNode]

	Graph  Graph
	stack  []Node
	nextID int
}

func (s *SQLVisualizer) getNextID(prefix string) string {
	nextID := s.nextID
	s.nextID += 1

	return prefix + strconv.Itoa(nextID)
}

func (s *SQLVisualizer) Enter(node pgsql.SyntaxNode) {
	nextNode := Node{
		ID:         s.getNextID("n"),
		Labels:     []string{node.NodeType()},
		Properties: map[string]any{},
	}

	switch typedExpression := node.(type) {
	case pgsql.BinaryExpression:
		nextNode.Properties["value"] = typedExpression.Operator

	case pgsql.Identifier:
		nextNode.Properties["value"] = typedExpression

	case pgsql.CompoundIdentifier:
		nextNode.Properties["value"] = strings.Join(typedExpression.Strings(), ".")

	case pgsql.Literal:
		nextNode.Properties["value"] = fmt.Sprintf("%v::%T", typedExpression.Value, typedExpression.Value)
	}

	s.Graph.Nodes = append(s.Graph.Nodes, nextNode)

	if len(s.stack) > 0 {
		s.Graph.Relationships = append(s.Graph.Relationships, Relationship{
			ID:     s.getNextID("r"),
			FromID: nextNode.ID,
			ToID:   s.stack[len(s.stack)-1].ID,
		})
	}

	s.stack = append(s.stack, nextNode)
}

func (s *SQLVisualizer) Exit(node pgsql.SyntaxNode) {
	s.stack = s.stack[0 : len(s.stack)-1]
}

func SQLToDigraph(node pgsql.SyntaxNode) (Graph, error) {
	visualizer := &SQLVisualizer{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[pgsql.SyntaxNode](),
	}

	if title, err := format.SyntaxNode(node, format.NewOutputBuilder()); err != nil {
		return Graph{}, err
	} else {
		visualizer.Graph.Title = title
	}

	return visualizer.Graph, walk.WalkPgSQL(node, visualizer)
}
