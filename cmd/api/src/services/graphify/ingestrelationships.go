// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package graphify

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/errorlist"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util"
)

// IngestRelationships resolves and writes a batch of ingestible relationships to the graph.
//
// This function first calls resolveRelationships to resolve node identifiers based on name and kind.
//
// Each resolved relationship update is applied to the graph via batch.UpdateRelationshipBy.
// Errors encountered during resolution or update are collected and returned as a single combined error.
func IngestRelationships(ingestCtx *IngestContext, sourceKind graph.Kind, relationships []ein.IngestibleRelationship) error {
	var (
		errs = util.NewErrorCollector()
	)

	updates, err := resolveRelationships(ingestCtx, relationships, sourceKind)
	if err != nil {
		errs.Add(err)
	}

	for _, update := range updates {
		if err := maybeSubmitRelationshipUpdate(ingestCtx, update); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}

// maybeSubmitRelationshipUpdate decides whether to upsert a node directly, or route it
// through the changelog for deduplication and caching.
func maybeSubmitRelationshipUpdate(ingestCtx *IngestContext, update graph.RelationshipUpdate) error {
	// Track that we processed this relationship (regardless of whether it's written)
	ingestCtx.Stats.RelationshipsProcessed.Add(1)

	if !ingestCtx.HasChangelog() {
		// No changelog: always update via dawgs batch
		return ingestCtx.Batch.UpdateRelationshipBy(update)
	}

	sourceObjectID, err := update.Start.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		return fmt.Errorf("get source objectid: %w", err)
	}
	targetObjectID, err := update.End.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		return fmt.Errorf("get target objectid: %w", err)
	}

	change := changelog.NewEdgeChange(
		sourceObjectID,
		targetObjectID,
		update.Relationship.Kind,
		update.Relationship.Properties,
	)

	shouldSubmit, err := ingestCtx.Manager.ResolveChange(change)
	if err != nil {
		return fmt.Errorf("resolve edge change: %w", err)
	}

	if shouldSubmit {
		// New/modified: update via dawgs batch (will increment RelationshipsWritten)
		return ingestCtx.Batch.UpdateRelationshipBy(update)
	}

	// Unchanged: enqueue change-- this is needed to maintain reconciliation
	if ok := ingestCtx.Manager.Submit(ingestCtx.Ctx, change); !ok {
		slog.WarnContext(ingestCtx.Ctx, "Changelog submit dropped",
			slog.String("source_object_id", sourceObjectID),
			slog.String("target_object_id", targetObjectID),
			slog.String("kind", update.Relationship.Kind.String()))
	}

	return nil
}

func ingestDNRelationship(batch *IngestContext, nextRel ein.IngestibleRelationship) error {
	nextRel.RelProps[common.LastSeen.String()] = batch.IngestTime
	nextRel.Source.Value = strings.ToUpper(nextRel.Source.Value)
	nextRel.Target.Value = strings.ToUpper(nextRel.Target.Value)

	return batch.Batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			ad.DistinguishedName: nextRel.Source,
			common.LastSeen:      batch.IngestTime,
		}), nextRel.Source.Kind),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			ad.DistinguishedName.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: batch.IngestTime,
		}), nextRel.Target.Kind),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestDNRelationships(batch *IngestContext, relationships []ein.IngestibleRelationship) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range relationships {
		if err := ingestDNRelationship(batch, next); err != nil {
			slog.Error("Error ingesting relationship", attr.Error(err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}

func ingestSession(batch *IngestContext, nextSession ein.IngestibleSession) error {
	nextSession.Target = strings.ToUpper(nextSession.Target)
	nextSession.Source = strings.ToUpper(nextSession.Source)

	return batch.Batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(graph.PropertyMap{
			common.LastSeen: batch.IngestTime,
			ad.LogonType:    nextSession.LogonType,
		}), ad.HasSession),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Source,
			common.LastSeen: batch.IngestTime,
		}), ad.Computer),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Target,
			common.LastSeen: batch.IngestTime,
		}), ad.User),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestSessions(batch *IngestContext, sessions []ein.IngestibleSession) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range sessions {
		if err := ingestSession(batch, next); err != nil {
			slog.Error("Error ingesting sessions", attr.Error(err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}

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

// resolveAllEndpointsByName attempts to resolve all unique source and target
// endpoints from a list of ingestible relationships into their corresponding object IDs.
//
// Each endpoint is identified by a Name, (optional) Kind pair. A single batch query is
// used to resolve all endpoints in one round trip.
//
// If multiple nodes match a given Name, Kind pair with conflicting object IDs,
// the match is considered ambiguous and excluded from the result. This can happen because there are no
// uniqueness guarantees on a node's `Name` property.
//
// Returns a map of resolved object IDs. If no matches are found or the input is empty, an empty map is returned.
func resolveAllEndpointsByName(batch BatchUpdater, rels []ein.IngestibleRelationship) (map[endpointKey]string, error) {
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
					slog.Warn("Matched node missing objectid",
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

// resolveRelationships transforms a list of ingestible relationships into a
// slice of graph.RelationshipUpdate objects, suitable for ingestion into the
// graph database.
//
// The function resolves all source and target endpoints to their corresponding
// object IDs if MatchByName is set on an endpoint. Relationships with unresolved
// or ambiguous endpoints are skipped and logged with a warning.
//
// The identityKind parameter determines the identity kind used for both start
// and end nodes if provided. eg. ad.Base and az.Base are used for *hound collections, and generic ingest has no base kind.
//
// Each resolved relationship is stamped with the current UTC timestamp as the "last seen" property.
//
// Returns a slice of valid relationship updates or an error if resolution fails.
func resolveRelationships(batch *IngestContext, rels []ein.IngestibleRelationship, sourceKind graph.Kind) ([]graph.RelationshipUpdate, error) {
	if cache, err := resolveAllEndpointsByName(batch.Batch, rels); err != nil {
		return nil, err
	} else {
		var (
			updates []graph.RelationshipUpdate
			errs    = errorlist.NewBuilder()
		)

		for _, rel := range rels {
			srcID, srcOK := resolveEndpointID(rel.Source, cache)
			targetID, targetOK := resolveEndpointID(rel.Target, cache)

			if !srcOK || !targetOK {
				slog.Warn("Skipping unresolved relationship",
					slog.String("source", rel.Source.Value),
					slog.String("target", rel.Target.Value),
					slog.Bool("resolved_source", srcOK),
					slog.Bool("resolved_target", targetOK))
				errs.Add(
					fmt.Errorf("skipping invalid relationship. unable to resolve endpoints. source: %s, target: %s", rel.Source.Value, rel.Target.Value),
				)
				continue
			}

			rel.RelProps[common.LastSeen.String()] = batch.IngestTime

			startKinds := MergeNodeKinds(sourceKind, rel.Source.Kind)
			endKinds := MergeNodeKinds(sourceKind, rel.Target.Kind)

			update := graph.RelationshipUpdate{
				Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: srcID,
					common.LastSeen: batch.IngestTime,
				}), startKinds...),
				StartIdentityProperties: []string{common.ObjectID.String()},
				StartIdentityKind:       sourceKind,
				End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: targetID,
					common.LastSeen: batch.IngestTime,
				}), endKinds...),
				EndIdentityKind:       sourceKind,
				EndIdentityProperties: []string{common.ObjectID.String()},
				Relationship:          graph.PrepareRelationship(graph.AsProperties(rel.RelProps), rel.RelType),
			}

			updates = append(updates, update)
		}

		return updates, errs.Build()
	}
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

// MergeNodeKinds combines a source kind with any additional kinds,
// then removes any occurrences of graph.EmptyKind from the result.
// Ensures a clean, usable kind list for downstream logic.
func MergeNodeKinds(sourceKind graph.Kind, additionalKinds ...graph.Kind) []graph.Kind {
	merged := mergeKinds(sourceKind, additionalKinds...)
	filtered := filterOutEmptyKind(merged)
	return deduplicateKinds(filtered)
}

// mergeKinds appends the sourceKind (if not EmptyKind)
// to the front of the additionalKinds slice, preserving order.
func mergeKinds(sourceKind graph.Kind, additionalKinds ...graph.Kind) []graph.Kind {
	if sourceKind == graph.EmptyKind {
		return append([]graph.Kind(nil), additionalKinds...)
	}
	return append([]graph.Kind{sourceKind}, additionalKinds...)
}

// filterOutEmptyKind removes any graph.EmptyKind values from the provided slice.
// Used to ensure the final kind list contains only meaningful entries.
func filterOutEmptyKind(kinds []graph.Kind) []graph.Kind {
	var result []graph.Kind
	for _, k := range kinds {
		if k != graph.EmptyKind {
			result = append(result, k)
		}
	}
	return result
}

// deduplicateKinds removes duplicate kinds from the list while preserving order.
// prevents duplicates like ["Base", "Base", "Person"]
func deduplicateKinds(kinds []graph.Kind) []graph.Kind {
	seen := make(map[graph.Kind]struct{})
	var result []graph.Kind
	for _, key := range kinds {
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, key)
		}
	}
	return result
}
