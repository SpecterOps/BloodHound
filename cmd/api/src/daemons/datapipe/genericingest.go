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
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/common"
)

type endpointKey struct {
	Name string
	Kind string
}

func addKey(endpoint ein.IngestibleEndpoint, cache map[endpointKey]struct{}) {
	if endpoint.MatchBy != ein.MatchByName {
		return
	}
	key := endpointKey{
		Name: strings.ToUpper(endpoint.Value),
	}
	if endpoint.Kind != nil {
		key.Kind = endpoint.Kind.String()
	}
	cache[key] = struct{}{}
}

func resolveAllEndpointsByName(batch graph.Batch, rels []ein.IngestibleRelationship) (map[endpointKey]string, error) {
	// seen deduplicates Name:Kind pairs from the input batch to ensure that each Name:Kind pairs is resolved once.
	seen := map[endpointKey]struct{}{}

	if len(rels) == 0 {
		return map[endpointKey]string{}, nil
	}

	for _, rel := range rels {
		addKey(rel.Source, seen)
		addKey(rel.Target, seen)
	}
	// if nothing to filter, return early
	if len(seen) == 0 {
		return map[endpointKey]string{}, nil
	}

	var (
		filters     = make([]graph.Criteria, 0, len(seen))
		buildFilter = func(key endpointKey) graph.Criteria {
			var criteria []graph.Criteria

			criteria = append(criteria, query.Equals(query.NodeProperty(common.Name.String()), key.Name))
			if key.Kind != "" {
				criteria = append(criteria, query.Kind(query.Node(), graph.StringKind(key.Kind)))
			}
			return query.And(criteria...)
		}
	)

	// aggregate all Name:Kind pairs in 1 DAWGs query for 1 round trip
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

				// edge case: resolve an empty key to match endpoints that provide no Kind filter
				node.Kinds = append(node.Kinds, graph.EmptyKind)

				// resolve all names found to objectids,
				// record ambiguous matches (when more than one match is found, we cannot disambiguate the requested node and must skip the update)
				for _, kind := range node.Kinds {
					key := endpointKey{Name: strings.ToUpper(nameVal), Kind: kind.String()}
					if existingID, exists := resolved[key]; exists && existingID != objectID {
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

func resolveRelationships(batch graph.Batch, rels []ein.IngestibleRelationship, identityKind graph.Kind) ([]*graph.RelationshipUpdate, error) {
	cache, err := resolveAllEndpointsByName(batch, rels)
	if err != nil {
		return nil, err
	}

	nowUTC := time.Now().UTC()
	var updates []*graph.RelationshipUpdate

	for _, rel := range rels {
		srcID, srcOK := resolveEndpointID(rel.Source, cache)
		targetID, targetOK := resolveEndpointID(rel.Target, cache)

		if !srcOK || !targetOK {
			slog.Warn("skipping unresolved relationship",
				slog.String("source", rel.Source.Value),
				slog.String("target", rel.Target.Value),
				slog.Bool("resolved_source", srcOK),
				slog.Bool("resolved_target", targetOK))
			continue
		}

		rel.RelProps[common.LastSeen.String()] = nowUTC

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

		if identityKind != graph.EmptyKind {
			update.StartIdentityKind = identityKind
			update.EndIdentityKind = identityKind
		}

		updates = append(updates, update)
	}

	return updates, nil
}

func resolveEndpointID(endpoint ein.IngestibleEndpoint, cache map[endpointKey]string) (string, bool) {
	if endpoint.MatchBy == ein.MatchByName {
		key := endpointKey{
			Name: strings.ToUpper(endpoint.Value),
			Kind: "",
		}
		if endpoint.Kind != nil {
			key.Kind = endpoint.Kind.String()
		}
		id, ok := cache[key]
		return id, ok
	}

	// Fallback to raw value if matching by ID
	return endpoint.Value, endpoint.Value != ""
}
