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

const (
	windowsServer2025Build           = 26100
	badSuccessorPatchedBuildRevision = 4946
)

type windowsBuildVersion struct {
	Build    int
	Revision int
}

type badSuccessorTarget struct {
	CreatorID graph.ID
	VictimID  graph.ID
}

// PostBadSuccessor creates a domain-level BadSuccessor edge for each CreateDMSA principal when collected registry data
// confirms at least one Windows Server 2025 domain controller predates build 26100.4946. Otherwise it emits a
// target-specific edge only when the same principal has GenericAll, GenericWrite, or both attribute-specific write
// permissions required by the patched KDC over a user or computer in that domain.
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
			return postBadSuccessorForDomain(ctx, tx, sink, domain, domainSID)
		}); err != nil {
			return sink.Stats(), fmt.Errorf("error creating BadSuccessor edges: %w", err)
		}
	}

	return sink.Stats(), nil
}

// postBadSuccessorForDomain applies the patch-aware edge model to one domain.
func postBadSuccessorForDomain(
	ctx context.Context,
	tx graph.Transaction,
	sink *post.FilteredRelationshipSink,
	domain *graph.Node,
	domainSID string,
) error {
	creators, err := fetchDMSACreators(tx, domainSID)
	if err != nil {
		return err
	}
	if len(creators) == 0 {
		return nil
	}

	builds, err := fetchDomainControllerBuilds(tx, domain.ID)
	if err != nil {
		return err
	}

	if domainHasUnpatchedDMSAKDC(builds) {
		for _, creator := range creators {
			if err := submitBadSuccessor(ctx, sink, creator.ID, domain.ID); err != nil {
				return err
			}
		}

		return nil
	}

	targets, err := fetchBadSuccessorTargets(tx, creators, domainSID)
	if err != nil {
		return err
	}

	for target, rights := range targets {
		if qualifiesBadSuccessorVictim(rights) {
			if err := submitBadSuccessor(ctx, sink, target.CreatorID, target.VictimID); err != nil {
				return err
			}
		}
	}

	return nil
}

// fetchDMSACreators returns principals with CreateDMSA rights scoped to a domain.
func fetchDMSACreators(tx graph.Transaction, domainSID string) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), ad.CreateDMSA),
			query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
		)
	}))
}

// fetchDomainControllerBuilds returns complete registry-backed build versions for domain controllers in a domain.
func fetchDomainControllerBuilds(tx graph.Transaction, domainID graph.ID) ([]windowsBuildVersion, error) {
	domainControllers, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), ad.DCFor),
			query.Equals(query.EndID(), domainID),
		)
	}))
	if err != nil {
		return nil, err
	}

	builds := make([]windowsBuildVersion, 0, len(domainControllers))
	for _, domainController := range domainControllers {
		collected, collectedErr := domainController.Properties.Get(ad.WindowsBuildCollected.String()).Bool()
		build, buildErr := domainController.Properties.Get(ad.WindowsBuildNumber.String()).Int()
		revision, revisionErr := domainController.Properties.Get(ad.WindowsBuildRevision.String()).Int()
		if collectedErr == nil && collected && buildErr == nil && revisionErr == nil {
			builds = append(builds, windowsBuildVersion{Build: build, Revision: revision})
		}
	}

	return builds, nil
}

// fetchBadSuccessorTargets groups qualifying victim rights by creator and target.
func fetchBadSuccessorTargets(tx graph.Transaction, creators graph.NodeSet, domainSID string) (map[badSuccessorTarget]graph.Kinds, error) {
	if len(creators) == 0 {
		return map[badSuccessorTarget]graph.Kinds{}, nil
	}

	creatorIDs := make([]graph.ID, 0, len(creators))
	for _, creator := range creators {
		creatorIDs = append(creatorIDs, creator.ID)
	}

	relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.StartID(), creatorIDs...),
			query.KindIn(
				query.Relationship(),
				ad.GenericAll,
				ad.GenericWrite,
				ad.WriteSupersededManagedAccountLink,
				ad.WriteSupersededServiceAccountState,
			),
			query.KindIn(query.End(), ad.User, ad.Computer),
			query.Equals(query.EndProperty(ad.DomainSID.String()), domainSID),
		)
	}))
	if err != nil {
		return nil, err
	}

	targets := make(map[badSuccessorTarget]graph.Kinds)
	for _, relationship := range relationships {
		target := badSuccessorTarget{CreatorID: relationship.StartID, VictimID: relationship.EndID}
		targets[target] = append(targets[target], relationship.Kind)
	}

	return targets, nil
}

// domainHasUnpatchedDMSAKDC reports whether any collected Windows Server 2025 build predates the KDC fix.
func domainHasUnpatchedDMSAKDC(builds []windowsBuildVersion) bool {
	for _, build := range builds {
		if build.Build == windowsServer2025Build && build.Revision < badSuccessorPatchedBuildRevision {
			return true
		}
	}

	return false
}

// qualifiesBadSuccessorVictim reports whether a right set can update both victim attributes required by the patched KDC.
func qualifiesBadSuccessorVictim(rights graph.Kinds) bool {
	var (
		canWriteManagedAccountLink  bool
		canWriteServiceAccountState bool
	)

	for _, right := range rights {
		switch {
		case right.Is(ad.GenericAll), right.Is(ad.GenericWrite):
			return true
		case right.Is(ad.WriteSupersededManagedAccountLink):
			canWriteManagedAccountLink = true
		case right.Is(ad.WriteSupersededServiceAccountState):
			canWriteServiceAccountState = true
		}
	}

	return canWriteManagedAccountLink && canWriteServiceAccountState
}

// submitBadSuccessor queues a derived edge for the post-processing sink.
func submitBadSuccessor(ctx context.Context, sink *post.FilteredRelationshipSink, creatorID, targetID graph.ID) error {
	if !sink.Submit(ctx, post.EnsureRelationshipJob{
		FromID: creatorID,
		ToID:   targetID,
		Kind:   ad.BadSuccessor,
	}) {
		return fmt.Errorf("unable to submit to channel in PostBadSuccessor")
	}

	return nil
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
