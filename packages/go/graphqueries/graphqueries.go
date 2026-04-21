//go:generate go run go.uber.org/mock/mockgen -destination=./mocks/graph.go -package=mocks . Graph

package graphqueries

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

type Graph interface {
	queries.Graph
}

type GraphQuery struct {
	*queries.GraphQuery
}

func NewGraphQuery(graphDB graph.Database, cache cache.Cache, cfg config.Configuration) *GraphQuery {
	return &GraphQuery{
		GraphQuery: queries.NewGraphQuery(graphDB, cache, cfg),
	}
}

func (s *GraphQuery) CountNodesByKind(ctx context.Context, kinds ...graph.Kind) (int64, error) {
	return s.CountFilteredNodes(ctx, query.KindIn(query.Node(), kinds...))
}

func (s *GraphQuery) CountFilteredNodes(ctx context.Context, filterCriteria graph.Criteria) (int64, error) {
	return s.GraphQuery.CountFilteredNodes(ctx, query.And(graphschema.IgnoreMetaFilter, filterCriteria))
}

func (s *GraphQuery) GetPrimaryNodeKindCounts(ctx context.Context, primaryDisplayKinds graphschema.PrimaryDisplayKinds, kind graph.Kind, additionalFilters ...graph.Criteria) (map[string]int, error) {
	return s.GraphQuery.GetPrimaryNodeKindCounts(ctx, primaryDisplayKinds, kind, append([]graph.Criteria{graphschema.IgnoreMetaFilter}, additionalFilters...)...)
}

func (s *GraphQuery) GetFilteredAndSortedNodesPaginated(sortItems query.SortItems, filterCriteria graph.Criteria, offset, limit int) ([]*graph.Node, error) {
	return s.GraphQuery.GetFilteredAndSortedNodesPaginated(sortItems, query.And(graphschema.IgnoreMetaFilter, filterCriteria), offset, limit)
}
