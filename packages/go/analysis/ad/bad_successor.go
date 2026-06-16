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
	// "github.com/specterops/dawgs/util/channels"
)

// PostCanUseBadSuccessor
func PostCanUseBadSuccessor(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing CanUseBadSuccessor",
		attr.Namespace("analysis"),
		attr.Function("PostCanUseBadSuccessor"),
		attr.Scope("process"),
	)()

	if _, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "CanUseBadSuccessor Post Processing")

		// for _, domain := range domainNodes {
		// 	innerDomain := domain
		// 	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		// 		badLineOfSuccession, err := getCanUseBadSuccessorViaCreateChild(tx, innerDomain, localGroupData)
		// 		if err != nil {
		// 			return err
		// 		} else if len(badLineOfSuccession) == 0 {
		// 			return nil
		// 		}

		// 		for _, duplex := range badLineOfSuccession {
		// 			duplex.Each(func(value uint64) bool {
		// 				channels.Submit(ctx, outC, post.EnsureRelationshipJob{
		// 					FromID: graph.ID(value),
		// 					ToID:   graph.ID(),
		// 					Kind:   ad.CanUseBadSuccessor,
		// 				})
		// 				return true
		// 			})
		// 		}
		// 	})

		// 	return nil

		// }
		return &operation.Stats, operation.Done()
	}
}

// // getCanUseBadSuccessorViaCreateChild(tx, domain)     → map[ouID]principals
// func getCanUseBadSuccessorViaCreateChild(tx graph.Transaction, domain *graph.Node, localGroupData *LocalGroupData) (map[graph.ID]cardinality.Duplex[uint64], error) {
// 	createChildNodes, err := getCreateChildDMSAPrincipalsForDomain(tx, domain)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return createChildNodes, nil
// }

// // getCanUseBadSuccessorViaUnmigratedDMSA(tx, domain)  → map[dmsaID]principals
// func getCanUseBadSuccessorViaUnmigratedDMSA(tx graph.Transaction, domain *graph.Node) (map[graph.ID]cardinality.Duplex[uint64], error) {
// 	// this is a separate function because it has a different set of criteria, and it may be worth including the results even if the CreateChildDMSA portion doesn't return any results. also, it was easier to write and test separately.
// 	domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String()
// 	if err != nil {
// 		return nil, err
// 	}

// 	nonMigratedDMSANodes, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
// 		return query.And(
// 			query.Kind(query.Node(), ad.DelegatedMSA),
// 			query.Not(query.Equals(query.NodeProperty(string(ad.DelegatedMSAState)), "2")),
// 			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
// 		)
// 	}))
// 	if err != nil {
// 		return nil, err
// 	}

// 	delegatedStateNodes, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
// 		return query.And(
// 			query.Kind(query.Start(), ad.Entity),
// 			query.Or(query.Kind(query.Relationship(), ad.WriteMsDSDelegatedMSAState), query.Kind(query.Relationship(), ad.GenericAll)),
// 			query.InIDs(query.EndID(), nonMigratedDMSANodes...),
// 		)
// 	}))
// 	if err != nil {
// 		return nil, err
// 	}

// 	precededByNodes, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
// 		return query.And(
// 			query.Kind(query.Start(), ad.Entity),
// 			query.Kind(query.Relationship(), ad.WriteMsDSManagedAccountPrecededByLink),
// 			query.InIDs(query.EndID(), nonMigratedDMSANodes...),
// 		)
// 	}))
// 	if err != nil {
// 		return nil, err
// 	}

// 	results := make(map[graph.ID]cardinality.Duplex[uint64])
// 	for _, rel := range delegatedStateNodes {
// 		duplex, ok := results[rel.EndID]
// 		if !ok {
// 			duplex = cardinality.NewBitmap64()
// 		}
// 		duplex.Add(rel.StartID.Uint64())
// 		results[rel.EndID] = duplex
// 	}
// 	for _, rel := range precededByNodes {
// 		duplex, ok := results[rel.EndID]
// 		if !ok {
// 			duplex = cardinality.NewBitmap64()
// 		}
// 		duplex.Add(rel.StartID.Uint64())
// 		results[rel.EndID] = duplex
// 	}
// 	return results, nil

// }

/*
   **Two** accounts are required for this.
   1. "attacker_dmsa" can be achieved with **EITHER:**
     - CreateChild(all) or CreateChild (msDS-DelegatedManagedServiceAccount) on an OU and maq != 0
     - Having Write on `msDS-DelegatedMSAState` and `msDS-ManagedAccountPrecededByLink` or GenericWrite or GenericAll on an existing dMSA that is in a non-migrated state (msDS-DelegatedMSAState != 2)

   2. Must have the ability to Write to/modify these attributes on ANOTHER account, doesn't have to be a dMSA though.
	  - `msDS-SupersededManagedAccountLink`
		- `msDS-SupersededServiceAccountState`
*/
// getCanUseBadSuccessorPostPatch(tx, domain) → map[dmsaID]principals
func getCanUseBadSuccessorPostPatch(tx graph.Transaction, domain *graph.Node, createChildNodes map[graph.ID]cardinality.Duplex[uint64]) (map[graph.ID]cardinality.Duplex[uint64], error) {
	var firstDMSACandidates []graph.ID
	var secondDMSACandidates []graph.ID
	results := make(map[graph.ID]cardinality.Duplex[uint64])

	writeSupersededManagedAccountLinkNodes, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.DelegatedMSA),
			query.Kind(query.Relationship(), ad.WriteMsDSSupersededManagedAccountLink),
		)
	}))
	if err != nil {
		return nil, err
	}

	writeSupersededServiceAccountStateNodes, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.DelegatedMSA),
			query.Kind(query.Relationship(), ad.WriteMsDSSupersededServiceAccountState),
		)
	}))
	if err != nil {
		return nil, err
	}

	for _, rel := range writeSupersededManagedAccountLinkNodes {
		secondDMSACandidates = append(secondDMSACandidates, rel.EndID)
	}
	for _, rel := range writeSupersededServiceAccountStateNodes {
		secondDMSACandidates = append(secondDMSACandidates, rel.EndID)
	}
	if len(secondDMSACandidates) == 0 {
		return results, nil
	}
	if len(createChildNodes) > 0 {
		for _, dmsaID := range secondDMSACandidates {
			for _, duplex := range createChildNodes {
				duplex.Each(func(value uint64) bool {
					duplex2, ok := results[graph.ID(dmsaID)]
					if !ok {
						duplex2 = cardinality.NewBitmap64()
					}
					duplex2.Add(value)
					results[graph.ID(dmsaID)] = duplex2
					return true
				})
			}
		}
		return results, nil
	}
	precededByNodes, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.DelegatedMSA),
			query.Kind(query.Relationship(), ad.WriteMsDSManagedAccountPrecededByLink),
		)
	}))
	if err != nil {
		return nil, err
	}

	writableGroupMembershipNodes, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.DelegatedMSA),
			query.Kind(query.Relationship(), ad.WriteMsDSGroupMSAMembership),
		)
	}))
	if err != nil {
		return nil, err
	}

	for _, rel := range precededByNodes {
		firstDMSACandidates = append(firstDMSACandidates, rel.EndID)
	}
	for _, rel := range writableGroupMembershipNodes {
		firstDMSACandidates = append(firstDMSACandidates, rel.EndID)
	}
	// then we need to find the intersection of the first and second dmsa candidates, and then find the principals that have access to both, keyed by the secondDMSAId
	for _, secondDMSAID := range secondDMSACandidates {
		for _, firstDMSAID := range firstDMSACandidates {
			if secondDMSAID == firstDMSAID {
				// this means that the same dmsa has both sets of permissions, which is not valid for this attack path, so we skip it.
				continue
			}
			var firstDMSAPrincipals []uint64
			var secondDMSAPrincipals []uint64
			for _, rel := range writeSupersededManagedAccountLinkNodes {
				if rel.EndID == secondDMSAID {
					secondDMSAPrincipals = append(secondDMSAPrincipals, rel.StartID.Uint64())
				}
			}
			for _, rel := range writeSupersededServiceAccountStateNodes {
				if rel.EndID == secondDMSAID {
					secondDMSAPrincipals = append(secondDMSAPrincipals, rel.StartID.Uint64())
				}
			}
			for _, rel := range precededByNodes {
				if rel.EndID == firstDMSAID {
					firstDMSAPrincipals = append(firstDMSAPrincipals, rel.StartID.Uint64())
				}
			}
			for _, rel := range writableGroupMembershipNodes {
				if rel.EndID == firstDMSAID {
					firstDMSAPrincipals = append(firstDMSAPrincipals, rel.StartID.Uint64())
				}
			}
			if len(firstDMSAPrincipals) == 0 || len(secondDMSAPrincipals) == 0 {
				continue
			}
		}
	}
	return results, nil
}

func getCreateChildForBadSuccessorPrincipalsForDomain(tx graph.Transaction, domain *graph.Node) (map[graph.ID]cardinality.Duplex[uint64], error) {
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
		return make(map[graph.ID]cardinality.Duplex[uint64]), nil
	}
	relMap, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		// there are more relationships than this, but they are transitive. these are the relationships that include or can directly grant CreateChild
		return query.And(
			query.Kind(query.Start(), ad.Entity),
			query.Or(query.Kind(query.Relationship(), ad.CreateChildDMSA), query.Kind(query.Relationship(), ad.CreateChildAll), query.Kind(query.Relationship(), ad.GenericAll), query.Kind(query.Relationship(), ad.GenericWrite), query.Kind(query.Relationship(), ad.WriteOwner), query.Kind(query.Relationship(), ad.WriteDACL), query.Kind(query.Relationship(), ad.Owns)),
			query.InIDs(query.EndID(), containerIDs...),
		)
	}))
	if err != nil {
		return nil, err
	}
	// then group by rel.EndID into a map[graph.ID]cardinality.Duplex[uint64]
	relMapByEndID := make(map[graph.ID]cardinality.Duplex[uint64])
	for _, rel := range relMap {
		duplex, ok := relMapByEndID[rel.EndID]
		if !ok {
			duplex = cardinality.NewBitmap64()
		}
		duplex.Add(rel.StartID.Uint64())
		relMapByEndID[rel.EndID] = duplex
	}
	return relMapByEndID, nil
}
