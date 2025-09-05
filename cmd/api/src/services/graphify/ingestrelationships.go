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
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
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
		// New/modified: update via dawgs batch
		return ingestCtx.Batch.UpdateRelationshipBy(update)
	}

	// Unchanged: enqueue change-- this is needed to maintain reconciliation
	ingestCtx.Manager.Submit(ingestCtx.Ctx, change)

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
			slog.Error(fmt.Sprintf("Error ingesting relationship: %v", err))
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
			slog.Error(fmt.Sprintf("Error ingesting sessions: %v", err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}
