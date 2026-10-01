// Copyright 2026 Specter Ops, Inc.
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

package appdb

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

const graphExpansionBuiltInKinds = "ALL_ATTACK_PATHS"

var graphExpansionRelationshipKindNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (s *Store) ExpandGraph(ctx context.Context, nodeID int64, direction string, limit int) (model.UnifiedGraph, error) {
	if s.appDB == nil || s.cypherQuery == nil {
		return model.UnifiedGraph{}, fmt.Errorf("graph expansion dependencies are not configured")
	}

	relationshipKinds, err := s.graphExpansionRelationshipKinds(ctx)
	if err != nil {
		return model.UnifiedGraph{}, err
	}

	query, err := buildGraphExpansionQuery(nodeID, direction, relationshipKinds, limit)
	if err != nil {
		return model.UnifiedGraph{}, err
	}

	return s.rawCypherQuery(ctx, query, true)
}

func (s *Store) graphExpansionRelationshipKinds(ctx context.Context) ([]string, error) {
	openGraphExtensionManagementFeatureFlag, err := s.appDB.GetFlagByKey(ctx, appcfg.FeatureOpenGraphExtensionManagement)
	if err != nil {
		return nil, err
	}

	if !openGraphExtensionManagementFeatureFlag.Enabled {
		return nil, nil
	}

	relationshipKindFilters := model.Filters{
		"is_traversable": []model.Filter{{Operator: model.Equals, Value: "true"}},
	}
	openGraphRelationships, _, err := s.appDB.GetGraphSchemaRelationshipKinds(ctx, relationshipKindFilters, model.Sort{}, 0, 0)
	if err != nil {
		return nil, err
	}

	relationshipKinds := make([]string, 0, len(openGraphRelationships))
	for _, relationship := range openGraphRelationships {
		relationshipKinds = append(relationshipKinds, relationship.Name)
	}

	return relationshipKinds, nil
}

func (s *Store) rawCypherQuery(ctx context.Context, cypher string, includeProperties bool) (model.UnifiedGraph, error) {
	preparedQuery, err := s.cypherQuery.PrepareCypherQuery(cypher, queries.DefaultQueryFitnessLowerBoundExplore)
	if err != nil {
		return model.UnifiedGraph{}, err
	}

	primaryDisplayKinds, err := s.appDB.GetPrimaryDisplayKinds(ctx)
	if err != nil {
		return model.UnifiedGraph{}, err
	}

	return s.cypherQuery.RawCypherQuery(ctx, primaryDisplayKinds, preparedQuery, includeProperties)
}

func buildGraphExpansionQuery(nodeID int64, direction string, relationshipKinds []string, limit int) (string, error) {
	filteredRelationshipKinds := make([]string, 0, len(relationshipKinds)+1)
	seenRelationshipKinds := map[string]struct{}{}

	for _, relationshipKind := range append([]string{graphExpansionBuiltInKinds}, relationshipKinds...) {
		if !graphExpansionRelationshipKindNamePattern.MatchString(relationshipKind) {
			return "", services.NewGraphExpansionValidationError(
				services.ErrInvalidGraphExpansionRelationshipKind,
				fmt.Sprintf("invalid relationship kind: %s", relationshipKind),
			)
		}

		if _, seen := seenRelationshipKinds[relationshipKind]; seen {
			continue
		}

		seenRelationshipKinds[relationshipKind] = struct{}{}
		filteredRelationshipKinds = append(filteredRelationshipKinds, relationshipKind)
	}

	relationshipMatch := ""
	relationshipFilter := strings.Join(filteredRelationshipKinds, "|")

	switch direction {
	case services.GraphExpansionDirectionOutbound:
		relationshipMatch = fmt.Sprintf("MATCH (source)-[r:%s]->(target)", relationshipFilter)
	case services.GraphExpansionDirectionInbound:
		relationshipMatch = fmt.Sprintf("MATCH (target)-[r:%s]->(source)", relationshipFilter)
	default:
		return "", services.NewGraphExpansionValidationError(
			services.ErrInvalidGraphExpansionDirection,
			"direction must be either inbound or outbound",
		)
	}

	return fmt.Sprintf(`MATCH (source)
WHERE ID(source) = %d
%s
RETURN source, r, target
ORDER BY ID(r)
LIMIT %d`, nodeID, relationshipMatch, limit), nil
}
