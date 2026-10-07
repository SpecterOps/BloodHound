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
	"iter"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/errorlist"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util"
)

type relationshipIngester struct {
	ingestContext     *IngestContext
	stageMeasurements graphifyChunkStageMeasurements
}

// IngestRelationships resolves and writes a batch of ingestible relationships to the graph.
//
// This function first calls resolveRelationships to resolve node identifiers based on name and kind.
//
// Each resolved relationship update is applied to the graph via batch.UpdateRelationshipBy.
// Errors encountered during resolution or update are collected and returned as a single combined error.
func IngestRelationships(ingestCtx *IngestContext, sourceKind graph.Kind, relationships []ein.IngestibleRelationship) error {
	var (
		ingester = relationshipIngester{ingestContext: ingestCtx}
	)
	defer ingester.stageMeasurements.publish()

	return ingester.ingestRelationships(sourceKind, relationships)
}

func (s *relationshipIngester) ingestRelationships(sourceKind graph.Kind, relationships []ein.IngestibleRelationship) error {
	var (
		errs                  = errorlist.NewBuilder()
		resolvedRelationships []ein.IngestibleRelationship
		resolveErrors         error
		stageStartedAt        time.Time
	)

	stageStartedAt = time.Now()
	resolvedRelationships, resolveErrors = endpoint.ResolveAll(s.ingestContext.Ctx, s.ingestContext.EndpointResolver, relationships, s.ingestContext.UseRawObjectIDs)
	s.stageMeasurements.relationshipResolution.add(time.Since(stageStartedAt), len(relationships), len(relationships)-len(resolvedRelationships))

	if resolveErrors != nil {
		errs.Add(resolveErrors)
	}

	for update := range s.ingestibleRelationshipsToUpdates(resolvedRelationships, sourceKind) {
		if err := s.maybeSubmitRelationshipUpdate(update); err != nil {
			errs.Add(err)
		}
	}

	return errs.Build()
}

// maybeSubmitRelationshipUpdate decides whether to upsert a node directly, or route it
// through the changelog for deduplication and caching.
func (s *relationshipIngester) maybeSubmitRelationshipUpdate(update graph.RelationshipUpdate) error {
	var (
		stageStartedAt time.Time
	)

	// Track that we processed this relationship (regardless of whether it's written)
	s.ingestContext.Stats.RelationshipsProcessed.Add(1)

	if !s.ingestContext.HasChangelog() {
		// No changelog: always update via dawgs batch
		stageStartedAt = time.Now()
		err := s.ingestContext.Batch.UpdateRelationshipBy(update)
		if err != nil {
			s.stageMeasurements.relationshipBatchUpdate.add(time.Since(stageStartedAt), 1, 1)
		} else {
			s.stageMeasurements.relationshipBatchUpdate.add(time.Since(stageStartedAt), 1, 0)
		}
		return err
	}

	stageStartedAt = time.Now()
	sourceObjectID, err := update.Start.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		s.stageMeasurements.relationshipDeduplicate.add(time.Since(stageStartedAt), 1, 1)
		return fmt.Errorf("get source objectid: %w", err)
	}
	targetObjectID, err := update.End.Properties.Get(common.ObjectID.String()).String()
	if err != nil {
		s.stageMeasurements.relationshipDeduplicate.add(time.Since(stageStartedAt), 1, 1)
		return fmt.Errorf("get target objectid: %w", err)
	}

	change := changelog.NewEdgeChange(
		sourceObjectID,
		targetObjectID,
		update.Relationship.Kind,
		update.Relationship.Properties,
	)

	shouldSubmit, err := s.ingestContext.Manager.ResolveChange(change)
	if err != nil {
		s.stageMeasurements.relationshipDeduplicate.add(time.Since(stageStartedAt), 1, 1)
		return fmt.Errorf("resolve edge change: %w", err)
	}
	s.stageMeasurements.relationshipDeduplicate.add(time.Since(stageStartedAt), 1, 0)

	if shouldSubmit {
		// New/modified: update via dawgs batch
		stageStartedAt = time.Now()
		err = s.ingestContext.Batch.UpdateRelationshipBy(update)
		if err != nil {
			s.stageMeasurements.relationshipBatchUpdate.add(time.Since(stageStartedAt), 1, 1)
			return err
		} else {
			s.stageMeasurements.relationshipBatchUpdate.add(time.Since(stageStartedAt), 1, 0)
			return nil
		}
	}

	// Unchanged: enqueue change-- this is needed to maintain reconciliation
	if ok := s.ingestContext.Manager.Submit(s.ingestContext.Ctx, change); !ok {
		slog.WarnContext(s.ingestContext.Ctx, "Changelog submit dropped",
			slog.String("source_object_id", sourceObjectID),
			slog.String("target_object_id", targetObjectID),
			slog.String("kind", update.Relationship.Kind.String()))
	}

	return nil
}

func (s *relationshipIngester) ingestDNRelationship(nextRel ein.IngestibleRelationship) error {
	var (
		stageStartedAt = time.Now()
	)

	nextRel.RelProps[common.LastSeen.String()] = s.ingestContext.IngestTime
	// The source is a distinguished name, which is always normalized to
	// upper case for resolution regardless of use_raw_object_id.
	nextRel.Source.Value = strings.ToUpper(nextRel.Source.Value)
	// The target is an object id; only normalize it when not preserving raw casing.
	if !s.ingestContext.UseRawObjectIDs {
		nextRel.Target.Value = strings.ToUpper(nextRel.Target.Value)
	}

	update := graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			ad.DistinguishedName: nextRel.Source,
			common.LastSeen:      s.ingestContext.IngestTime,
		}), nextRel.Source.Kind),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			ad.DistinguishedName.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: s.ingestContext.IngestTime,
		}), nextRel.Target.Kind),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	}

	s.stageMeasurements.dnRelationshipPrepare.add(time.Since(stageStartedAt), 1, 0)
	return s.maybeSubmitRelationshipUpdate(update)
}

func IngestDNRelationships(batch *IngestContext, relationships []ein.IngestibleRelationship) error {
	var (
		ingester = relationshipIngester{ingestContext: batch}
	)
	defer ingester.stageMeasurements.publish()

	return ingester.ingestDNRelationships(relationships)
}

func (s *relationshipIngester) ingestDNRelationships(relationships []ein.IngestibleRelationship) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range relationships {
		if err := s.ingestDNRelationship(next); err != nil {
			slog.Error("Error ingesting relationship", attr.Error(err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}

func (s *relationshipIngester) ingestSession(nextSession ein.IngestibleSession) error {
	var (
		stageStartedAt = time.Now()
	)

	nextSession.Target = strings.ToUpper(nextSession.Target)
	nextSession.Source = strings.ToUpper(nextSession.Source)

	update := graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(graph.PropertyMap{
			common.LastSeen: s.ingestContext.IngestTime,
			ad.LogonType:    nextSession.LogonType,
		}), ad.HasSession),
		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Source,
			common.LastSeen: s.ingestContext.IngestTime,
		}), ad.Computer),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},
		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Target,
			common.LastSeen: s.ingestContext.IngestTime,
		}), ad.User),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	}

	s.stageMeasurements.sessionPrepare.add(time.Since(stageStartedAt), 1, 0)
	return s.maybeSubmitRelationshipUpdate(update)
}

func IngestSessions(batch *IngestContext, sessions []ein.IngestibleSession) error {
	var (
		ingester = relationshipIngester{ingestContext: batch}
	)
	defer ingester.stageMeasurements.publish()

	return ingester.ingestSessions(sessions)
}

func (s *relationshipIngester) ingestSessions(sessions []ein.IngestibleSession) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range sessions {
		if err := s.ingestSession(next); err != nil {
			slog.Error("Error ingesting sessions", attr.Error(err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}

// ingestibleRelationshipsToUpdates transforms a list of ingestible relationships into
// graph.RelationshipUpdate objects as an iterator.
func (s *relationshipIngester) ingestibleRelationshipsToUpdates(rels []ein.IngestibleRelationship, sourceKind graph.Kind) iter.Seq[graph.RelationshipUpdate] {
	return func(yield func(graph.RelationshipUpdate) bool) {
		for _, rel := range rels {
			stageStartedAt := time.Now()
			rel.RelProps[common.LastSeen.String()] = s.ingestContext.IngestTime

			var (
				startKinds = MergeNodeKinds(sourceKind, rel.Source.Kind)
				endKinds   = MergeNodeKinds(sourceKind, rel.Target.Kind)

				startObjID = rel.Source.Value
				endObjID   = rel.Target.Value
			)

			if !s.ingestContext.UseRawObjectIDs {
				startObjID = strings.ToUpper(startObjID)
				endObjID = strings.ToUpper(endObjID)
			}

			update := graph.RelationshipUpdate{
				Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: startObjID,
					common.LastSeen: s.ingestContext.IngestTime,
				}), startKinds...),
				StartIdentityProperties: []string{common.ObjectID.String()},
				StartIdentityKind:       sourceKind,
				End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
					common.ObjectID: endObjID,
					common.LastSeen: s.ingestContext.IngestTime,
				}), endKinds...),
				EndIdentityKind:       sourceKind,
				EndIdentityProperties: []string{common.ObjectID.String()},
				Relationship:          graph.PrepareRelationship(graph.AsProperties(rel.RelProps), rel.RelType),
			}

			s.stageMeasurements.relationshipPrepare.add(time.Since(stageStartedAt), 1, 0)
			if !yield(update) {
				break
			}
		}
	}
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
