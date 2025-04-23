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

func processRelationshipWithoutOIDs(batch graph.Batch, rel ein.IngestibleRelationship) error {
	if update, err := resolveRelationshipByName(batch, rel); err != nil {
		return err
	} else if update != nil { // nil updates are skipped. a nil update can represent a rel with endpoints that couldn't be fully resolved
		return batch.UpdateRelationshipBy(*update)
	}
	return nil
}

// this func attempts to resolve objectids for a rel's source and target nodes.
// name (and optional kind filter) --> objectid --> submit to batch processor for standard processing flow
// todo: change rel -> rels and refactor this to do resolution for many rels in one dawg query
func resolveRelationshipByName(batch graph.Batch, rel ein.IngestibleRelationship) (*graph.RelationshipUpdate, error) {
	result := &graph.RelationshipUpdate{}
	if srcID, srcOK, err := resolveNodeByName(batch, rel.Source); err != nil {
		return result, err
	} else if targetID, targetOK, err := resolveNodeByName(batch, rel.Target); err != nil {
		return result, err
	} else if !srcOK || !targetOK {
		slog.Warn("failed to resolve relationship endpoints by name",
			slog.String("source", rel.Source.Value),
			slog.String("target", rel.Target.Value),
			slog.Bool("resolved_source", srcOK),
			slog.Bool("resolved_target", targetOK))
		return result, nil
	} else {
		nowUTC := time.Now().UTC()

		start := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: srcID,
			common.LastSeen: nowUTC,
		}), rel.Source.Kind)

		end := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: targetID,
			common.LastSeen: nowUTC,
		}), rel.Target.Kind)

		result = &graph.RelationshipUpdate{
			Start: start,
			StartIdentityProperties: []string{
				common.ObjectID.String(),
			},
			End: end,
			EndIdentityProperties: []string{
				common.ObjectID.String(),
			},
			Relationship: graph.PrepareRelationship(graph.AsProperties(rel.RelProps), rel.RelType),
			// note: no need to set start/end identitykind because this code path is for generic-ingest only which has no base kind.
		}

		return result, nil
	}
}

func resolveNodeByName(batch graph.Batch, endpoint ein.IngestibleEndpoint) (string, bool, error) {
	if endpoint.MatchBy != ein.MatchByName {
		// Fallback: assume Value is already an objectid
		return endpoint.Value, true, nil
	}

	var (
		match               string
		ambiguousResolution bool
		filter              = func() graph.Criteria {
			var criteria []graph.Criteria
			if endpoint.MatchBy == ein.MatchByName {
				criteria = append(criteria, query.Equals(query.NodeProperty(common.Name.String()), strings.ToUpper(endpoint.Value)))
			}
			if endpoint.Kind != nil && !endpoint.Kind.Is(graph.EmptyKind) {
				criteria = append(criteria, query.Kind(query.Node(), endpoint.Kind))
			}
			return query.And(criteria...)
		}
	)

	if err := batch.Nodes().Filterf(filter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			props := node.Properties

			if oid, err := props.Get(string(common.ObjectID)).String(); err != nil {
				slog.Warn(fmt.Sprintf("matched node missing objectid for %s", endpoint.Value))
			} else {
				if match != "" {
					slog.Warn("ambiguous name match on node. multiple results match.",
						slog.String("value", endpoint.Value))
					ambiguousResolution = true
					return nil
				} else {
					match = oid
				}
			}

		}
		return nil
	}); err != nil {
		return "", false, err
	}

	// couldn't find exactly 1 match. found 0 or > 1 matches!
	if ambiguousResolution || match == "" {
		return "", false, nil
	}

	return match, true, nil
}

type endpointKey struct {
	Name string
	Kind string
}

func resolveAllEndpointsByName(batch graph.Batch, rels []ein.IngestibleRelationship) (map[endpointKey]string, error) {
	// seen deduplicates Name:Kind pairs from the input batch to ensure that Name:Kind pairs are resolved once.
	seen := map[endpointKey]struct{}{}

	if len(rels) == 0 {
		return map[endpointKey]string{}, nil
	}

	for _, rel := range rels {
		if rel.Source.MatchBy == ein.MatchByName {
			kind := ""
			if rel.Source.Kind != nil {
				kind = rel.Source.Kind.String()
			}
			key := endpointKey{Name: strings.ToUpper(rel.Source.Value), Kind: kind}
			seen[key] = struct{}{}
		}
		if rel.Target.MatchBy == ein.MatchByName {
			kind := ""
			if rel.Source.Kind != nil {
				kind = rel.Target.Kind.String()
			}
			key := endpointKey{Name: strings.ToUpper(rel.Target.Value), Kind: kind}
			seen[key] = struct{}{}
		}
	}

	var (
		filters     []graph.Criteria
		buildFilter = func(key endpointKey) graph.Criteria {
			var criteria []graph.Criteria

			criteria = append(criteria, query.Equals(query.NodeProperty(common.Name.String()), key.Name))
			if key.Kind != "" {
				criteria = append(criteria, query.Kind(query.Node(), graph.StringKind(key.Kind)))
			}
			return query.And(criteria...)
		}
	)

	// aggregate all Name:Kind pairs in one DAWGs query for 1 round trip
	for key := range seen {
		filters = append(filters, buildFilter(key))
	}

	var (
		resolved  = map[endpointKey]string{}
		ambiguous = map[endpointKey]bool{}
	)

	if err := batch.Nodes().Filter(query.Or(filters...)).Fetch(
		func(cursor graph.Cursor[*graph.Node]) error {

			for node := range cursor.Chan() {
				nameVal, _ := node.Properties.Get(common.Name.String()).String()
				objectID, err := node.Properties.Get(string(common.ObjectID)).String()
				if err != nil || objectID == "" {
					slog.Warn("matched node missing objectid",
						slog.String("name", nameVal),
						slog.Any("kinds", node.Kinds))
					continue
				}

				// resolve all names found to objectids,
				// record ambiguous matches (when more than one match is found, we cannot disambiguate the requested node and must skip the update)
				for _, kind := range node.Kinds {
					key := endpointKey{Name: strings.ToUpper(nameVal), Kind: kind.String()}
					if _, exists := resolved[key]; exists {
						ambiguous[key] = true
					} else {
						resolved[key] = objectID
					}
				}
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	// remove ambiguous matches
	for key := range ambiguous {
		delete(resolved, key)
	}

	return resolved, nil
}

func resolveRelationshipsByName(
	batch graph.Batch,
	rels []ein.IngestibleRelationship,
) ([]*graph.RelationshipUpdate, error) {

	if cache, err := resolveAllEndpointsByName(batch, rels); err != nil {
		return nil, err
	} else {

		var (
			nowUTC  = time.Now().UTC()
			updates []*graph.RelationshipUpdate
		)

		for _, rel := range rels {
			var srcID, targetID string
			var srcOK, targetOK bool

			if rel.Source.MatchBy == ein.MatchByName {
				key := endpointKey{
					Name: strings.ToUpper(rel.Source.Value),
					Kind: rel.Source.Kind.String(),
				}
				srcID, srcOK = cache[key]
			} else { // assume value is already objectid when matchby == MatchByID or is unset (for existing paths)
				srcID, srcOK = rel.Source.Value, rel.Source.Value != ""
			}

			if rel.Target.MatchBy == ein.MatchByName {
				key := endpointKey{
					Name: strings.ToUpper(rel.Target.Value),
					Kind: rel.Target.Kind.String(),
				}
				targetID, targetOK = cache[key]
			} else { // assume value is already objectid when matchby == MatchByID or is unset (for existing paths)
				targetID, targetOK = rel.Target.Value, rel.Target.Value != ""
			}

			if !srcOK || !targetOK {
				slog.Warn("skipping unresolved relationship",
					slog.String("source", rel.Source.Value),
					slog.String("target", rel.Target.Value),
					slog.Bool("resolved_source", srcOK),
					slog.Bool("resolved_target", targetOK))
				continue
			}

			update := &graph.RelationshipUpdate{
				Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: srcID,
					common.LastSeen: nowUTC,
				}), rel.Source.Kind),
				StartIdentityProperties: []string{common.ObjectID.String()},
				End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: targetID,
					common.LastSeen: nowUTC,
				}), rel.Target.Kind),
				EndIdentityProperties: []string{common.ObjectID.String()},
				Relationship:          graph.PrepareRelationship(graph.AsProperties(rel.RelProps), rel.RelType),
			}

			updates = append(updates, update)
		}

		return updates, nil
	}

}
