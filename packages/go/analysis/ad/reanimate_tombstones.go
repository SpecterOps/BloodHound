// Copyright 2026 Specter Ops, Inc.
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
package ad

import (
	"context"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

// PostCanReanimateTombstone
func PostCanReanimateTombstone(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing CanReanimateTombstone",
		attr.Namespace("analysis"),
		attr.Function("PostCanReanimateTombstone"),
		attr.Scope("process"),
	)()

	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "CanReanimateTombstone Post Processing")

		for _, domain := range domainNodes {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				reanimatorsMap, err := getCanReanimateTombstone(tx, innerDomain, localGroupData)
				if err != nil {
					return err
				}
				if len(reanimatorsMap) == 0 {
					return nil
				}

				for targetID, reanimators := range reanimatorsMap {
					reanimators.Each(func(value uint64) bool {
						channels.Submit(ctx, outC, post.EnsureRelationshipJob{
							FromID: graph.ID(value),
							ToID:   graph.ID(targetID),
							Kind:   ad.CanReanimateTombstone,
						})
						return true
					})
				}
				return nil
			})
		}
		return &operation.Stats, operation.Done()
	}
}

func getCanReanimateTombstone(tx graph.Transaction, domain *graph.Node, localGroupData *LocalGroupData) (map[graph.ID]cardinality.Duplex[uint64], error) {
	reanimateNodes, err := ops.FetchStartNodes(fromEntityToEntityWithRelationshipKind(tx, domain, ad.ReanimateTombstones))
	if err != nil {
		return nil, err
	}
	// Short circuit if there are no nodes with ReanimateTombstones relationship to the domain, since that relationship is a prerequisite for CanReanimateTombstone
	if reanimateNodes.Len() == 0 {
		return map[graph.ID]cardinality.Duplex[uint64]{}, nil
	}
	createChildNodes, err := getCreateChildPrincipalsForDomain(tx, domain)
	if err != nil {
		return nil, err
	}
	// short circuit again if createChildNodes is nil or 0 is required for CanReanimateTombstones
	if createChildNodes.Len() == 0 {
		return map[graph.ID]cardinality.Duplex[uint64]{}, nil
	}
	writeNameNodes, err := getWriteEdgesByTargetForDomain(tx, domain)
	if err != nil {
		return nil, err
	}
	results := make(map[graph.ID]cardinality.Duplex[uint64])
	for targetID, writePrincipals := range writeNameNodes {
		eligible := CalculateCrossProductNodeSets(localGroupData,
			NewCachedPrincipalSet(writePrincipals),
			NewCachedPrincipalSet(reanimateNodes.Slice()),
			NewCachedPrincipalSet(createChildNodes.Slice()))
		if eligible.Cardinality() > 0 {
			results[targetID] = eligible
		}
	}
	return results, nil
}

func getCreateChildPrincipalsForDomain(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String()
	if err != nil {
		return nil, err
	}

	// initially this wasn't working as a single query, so I split it in two.
	// get the containerIDs first, then use them to walkthrough the tree
	containerIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.Container, ad.OU),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
		)
	}))
	if err != nil {
		return nil, err
	}
	if len(containerIDs) == 0 {
		return graph.NodeSet{}, nil
	}
	return ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		// there are more relationships than this, but they are transitive. these are the relationships that include or can directly grant CreateChild
		return query.And(
			query.Kind(query.Start(), ad.Entity),
			query.Or(query.Kind(query.Relationship(), ad.CreateChild), query.Kind(query.Relationship(), ad.CreateChildAll), query.Kind(query.Relationship(), ad.GenericAll), query.Kind(query.Relationship(), ad.GenericWrite), query.Kind(query.Relationship(), ad.WriteOwner), query.Kind(query.Relationship(), ad.WriteDACL), query.Kind(query.Relationship(), ad.Owns)),
			query.InIDs(query.EndID(), containerIDs...),
		)
	}))
}

func getWriteEdgesByTargetForDomain(tx graph.Transaction, domain *graph.Node) (map[graph.ID][]*graph.Node, error) {
	domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String()
	if err != nil {
		return nil, err
	}
	containerIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.User, ad.Group, ad.Computer),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
		)
	}))
	if err != nil {
		return nil, err
	}

	relationships, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.Entity),
			query.InIDs(query.EndID(), containerIDs...),
			query.Or(query.Kind(query.Relationship(), ad.WriteRDN, ad.WriteCommonName), query.Kind(query.Relationship(), ad.GenericAll), query.Kind(query.Relationship(), ad.GenericWrite)),
		)
	}))
	if err != nil {
		return nil, err
	}
	targetToPrincipalIDs := map[graph.ID][]graph.ID{}
	allStartIDs := []graph.ID{}
	for _, rel := range relationships {
		targetToPrincipalIDs[rel.EndID] = append(targetToPrincipalIDs[rel.EndID], rel.StartID)
		allStartIDs = append(allStartIDs, rel.StartID)
	}

	nodesByID := map[graph.ID]*graph.Node{}
	nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.InIDs(query.NodeID(), allStartIDs...)
	}))
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		nodesByID[n.ID] = n
	}
	result := map[graph.ID][]*graph.Node{}
	for targetID, ids := range targetToPrincipalIDs {
		for _, id := range ids {
			result[targetID] = append(result[targetID], nodesByID[id])
		}
	}
	return result, nil
}
