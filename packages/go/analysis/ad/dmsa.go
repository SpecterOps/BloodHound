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

package ad

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

var badSuccessorPostProcessedEdges = graph.Kinds{
	ad.BadSuccessor,
}

var ouroborosSelfEnrollPostProcessedEdges = graph.Kinds{
	ad.OuroborosSelfEnroll,
}

// PostBadSuccessor creates BadSuccessor edges from any principal that can create a delegated Managed Service Account
// (a CreateDMSA edge, i.e. CreateChild for the dMSA object class on a container in the domain) to that Domain node.
//
// A principal who can create a dMSA controls it as its creator/owner, can set the new dMSA's
// msDS-ManagedAccountPrecededByLink to any victim account together with msDS-DelegatedMSAState=2, and the KDC will then
// release the victim's keys to the dMSA via S4U2self (KERB-DMSA-KEY-PACKAGE). Because the victim can be any account in
// the domain (including Tier-0 principals) and requires no ACE on the victim itself, the capability is modeled as a
// domain-compromise edge to the Domain node rather than fanning out an edge to every account. (CVE-2025-53779.)
func PostBadSuccessor(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing BadSuccessor",
		attr.Namespace("analysis"),
		attr.Function("PostBadSuccessor"),
		attr.Scope("process"),
	)()

	// Clear old post-processed edges that will not have a `firstseen` property
	if err := post.MigrationForDCAPostProcessedEdges(ctx, db, badSuccessorPostProcessedEdges); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	domainNodes, err := fetchCollectedDomainNodes(ctx, db)
	if err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	// Pull a subgraph to compare against for tracking changes
	badSuccessorTracker, err := post.FetchTracker(ctx, db, badSuccessorPostProcessedEdges)
	if err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	sink := post.NewFilteredRelationshipSink(ctx, "PostBadSuccessor", db, badSuccessorTracker)
	defer sink.Done()

	for _, domain := range domainNodes {
		if domainSID, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			// DomainSID is only created after we have performed collection of the domain
			slog.DebugContext(
				ctx,
				"Skipping domain for BadSuccessor. Missing DomainSID property",
				slog.Uint64("domain_id", uint64(domain.ID)),
			)
			continue
		} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if creators, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Kind(query.Relationship(), ad.CreateDMSA),
					query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
				)
			})); err != nil {
				return err
			} else {
				for _, creator := range creators {
					if !sink.Submit(ctx, post.EnsureRelationshipJob{
						FromID: creator.ID,
						ToID:   domain.ID,
						Kind:   ad.BadSuccessor,
					}) {
						return fmt.Errorf("unable to submit to channel in PostBadSuccessor")
					}
				}

				return nil
			}
		}); err != nil {
			return sink.Stats(), fmt.Errorf("error creating BadSuccessor edges: %w", err)
		}
	}

	return sink.Stats(), nil
}

// PostOuroborosSelfEnroll creates OuroborosSelfEnroll edges from any principal that owns or holds GenericAll over a
// delegated Managed Service Account (DMSA node) to that dMSA.
//
// With Owner (which yields WriteDACL -> GenericAll) or GenericAll over a dMSA, a principal can write the dMSA's own SID
// into msDS-GroupMSAMembership, authorizing itself to retrieve the dMSA's managed password -- the self-sustaining
// "Ouroboros" loop in which the dMSA authenticates and authorizes itself.
func PostOuroborosSelfEnroll(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing OuroborosSelfEnroll",
		attr.Namespace("analysis"),
		attr.Function("PostOuroborosSelfEnroll"),
		attr.Scope("process"),
	)()

	// Clear old post-processed edges that will not have a `firstseen` property
	if err := post.MigrationForDCAPostProcessedEdges(ctx, db, ouroborosSelfEnrollPostProcessedEdges); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	// Pull a subgraph to compare against for tracking changes
	ouroborosTracker, err := post.FetchTracker(ctx, db, ouroborosSelfEnrollPostProcessedEdges)
	if err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	sink := post.NewFilteredRelationshipSink(ctx, "PostOuroborosSelfEnroll", db, ouroborosTracker)
	defer sink.Done()

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if rels, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Relationship(), ad.Owns, ad.GenericAll),
				query.Kind(query.End(), ad.DMSA),
			)
		})); err != nil {
			return err
		} else {
			for _, rel := range rels {
				if !sink.Submit(ctx, post.EnsureRelationshipJob{
					FromID: rel.StartID,
					ToID:   rel.EndID,
					Kind:   ad.OuroborosSelfEnroll,
				}) {
					return fmt.Errorf("unable to submit to channel in PostOuroborosSelfEnroll")
				}
			}

			return nil
		}
	}); err != nil {
		return sink.Stats(), fmt.Errorf("error creating OuroborosSelfEnroll edges: %w", err)
	}

	return sink.Stats(), nil
}
