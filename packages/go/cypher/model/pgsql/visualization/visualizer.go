package visualization

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/cypher/model/pgsql/format"
	"github.com/specterops/bloodhound/cypher/model/walk"
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

func (s *SQLVisualizer) Enter(expression pgsql.SyntaxNode) {
	nextNode := Node{
		ID:         s.getNextID("n"),
		Labels:     []string{expression.NodeType()},
		Properties: map[string]any{},
	}

	switch typedExpression := expression.(type) {
	case pgsql.Operator:
		nextNode.Properties["value"] = typedExpression

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

func (s *SQLVisualizer) Visit(expression pgsql.SyntaxNode) {}

func (s *SQLVisualizer) Exit(expression pgsql.SyntaxNode) {
	s.stack = s.stack[0 : len(s.stack)-1]
}

func SQLToDigraph(expression pgsql.SyntaxNode) (Graph, error) {
	visualizer := &SQLVisualizer{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[pgsql.SyntaxNode](),
	}

	if title, err := format.Expression(expression); err != nil {
		return Graph{}, err
	} else {
		visualizer.Graph.Title = title.Value
	}

	return visualizer.Graph, walk.PgSQL(expression, visualizer)
}
