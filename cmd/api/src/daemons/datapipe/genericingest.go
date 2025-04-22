// Copyright 2025 Specter Ops, Inc.
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

package datapipe

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/common"
)

// this func attempts to resolve objectids for a rel's source and target nodes.
// name (and optional kind filter) --> objectid --> submit to batch processor for standard processing flow
// todo: change rel -> rels and refactor this to do resolution for many rels in one dawg query
func resolveRelationshipByName(batch graph.Batch, rel ein.IngestibleRelationship) (graph.RelationshipUpdate, error) {
	var (
		nowUTC              = time.Now().UTC()
		matches             = map[string]string{}
		ambiguousResolution = false // if multiple nodes matched source/target
		filter              = func() graph.Criteria {
			buildCriteria := func(endpoint ein.IngestibleEndpoint) []graph.Criteria {
				var criteria []graph.Criteria

				if endpoint.MatchBy == ein.MatchByName {
					criteria = append(criteria, query.Equals(query.NodeProperty(common.Name.String()), strings.ToUpper(endpoint.Value)))
				}

				if !endpoint.Kind.Is(graph.EmptyKind) { // TODO: its empty string, not nil
					criteria = append(criteria, query.Kind(query.Node(), endpoint.Kind))
				}

				return criteria
			}

			return query.Or(
				query.And(buildCriteria(rel.Source)...),
				query.And(buildCriteria(rel.Target)...),
			)
		}

		result graph.RelationshipUpdate
	)

	if rel.Source.Value == "" || rel.Target.Value == "" {
		return result, nil
	}

	if err := batch.Nodes().Filterf(filter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			props := node.Properties

			if val, _ := props.Get(common.Name.String()).String(); strings.EqualFold(val, rel.Source.Value) {
				if oid, err := props.Get(string(common.ObjectID)).String(); err != nil {
					slog.Warn(fmt.Sprintf("matched source node missing objectid for %s", rel.Source.Value))
				} else {
					if _, hasExisting := matches["source"]; hasExisting {
						slog.Warn("ambiguous name match on source node. multiple results match.",
							slog.String("value", rel.Source.Value))
						ambiguousResolution = true
						return nil
					} else {
						matches["source"] = oid
					}
				}
			}

			if val, _ := props.Get(common.Name.String()).String(); strings.EqualFold(val, rel.Target.Value) {
				if oid, err := props.Get(string(common.ObjectID)).String(); err != nil {
					slog.Warn(fmt.Sprintf("matched target node missing objectid for %s", rel.Target.Value))
				} else {
					if _, hasExisting := matches["target"]; hasExisting {
						slog.Warn("ambiguous property match on target node. multiple results match.",
							slog.String("value", rel.Target.Value))
						ambiguousResolution = true
						return nil
					} else {
						matches["target"] = oid
					}
				}
			}
		}
		return nil
	}); err != nil {
		return result, err
	}

	if ambiguousResolution {
		return result, nil
	}

	srcID, srcOk := matches["source"]
	targetID, targetOk := matches["target"]

	if !srcOk || !targetOk {
		slog.Warn("failed to resolve both nodes by name",
			slog.String("source", rel.Source.Value),
			slog.String("target", rel.Target.Value),
			slog.Bool("resolved_source", srcOk),
			slog.Bool("resolved_target", targetOk))
		return result, nil
	}

	start := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
		common.ObjectID: srcID,
		common.LastSeen: nowUTC,
	}), rel.Source.Kind)

	end := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
		common.ObjectID: targetID,
		common.LastSeen: nowUTC,
	}), rel.Target.Kind)

	result = graph.RelationshipUpdate{
		Start: start,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},
		End: end,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
		Relationship: graph.PrepareRelationship(graph.AsProperties(rel.RelProps), rel.RelType),
		// note: no need to set start/end identitykind because this code path is generic-ingest only which has no base kind.
	}

	return result, nil

}
