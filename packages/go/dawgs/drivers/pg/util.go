package pg

import "github.com/specterops/bloodhound/dawgs/graph"

func IsPostgreSQLGraph(db graph.Database) bool {
	return graph.IsDriver[*Driver](db)
}
