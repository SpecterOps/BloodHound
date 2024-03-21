package visualization

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"io"
	"os"
	"strings"
)

func WriteStrings(writer io.Writer, strings ...string) error {
	for _, str := range strings {
		if _, err := io.WriteString(writer, str); err != nil {
			return err
		}
	}

	return nil
}

func GraphToPUMLDigraph(graph Graph, writer io.Writer) error {
	if err := WriteStrings(writer, "@startuml\ndigraph syntaxTree {\nrankdir=BT\n\n"); err != nil {
		return err
	}

	if graph.Title != "" {
		if err := WriteStrings(writer, "label=\"", graph.Title, "\"\n\n"); err != nil {
			return err
		}
	}

	for _, node := range graph.Nodes {
		nodeLabel := strings.Join(node.Labels, ":")

		if value, hasValue := node.Properties["value"]; hasValue {
			nodeLabel = fmt.Sprintf("%v", value)
		}

		if err := WriteStrings(writer, node.ID, "[label=\"", nodeLabel, "\"]", "\n"); err != nil {
			return err
		}
	}

	if err := WriteStrings(writer, "\n"); err != nil {
		return err
	}

	for _, relationship := range graph.Relationships {
		if err := WriteStrings(writer, relationship.FromID, " -> ", relationship.ToID, "\n"); err != nil {
			return err
		}
	}

	return WriteStrings(writer, "}\n@enduml\n")
}

func MustWritePUML(expression pgsql.Expression, path string) {
	if graph, err := SQLToDigraph(expression); err != nil {
		panic(fmt.Sprintf("error translating SQL AST to digraph: %v", err))
	} else if fout, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		panic(fmt.Sprintf("error opening file at path %s: %v", path, err))
	} else {
		defer fout.Close()

		if err := GraphToPUMLDigraph(graph, fout); err != nil {
			panic(fmt.Sprintf("error writing graph to PUML wrapped digraph: %v", err))
		}
	}
}
