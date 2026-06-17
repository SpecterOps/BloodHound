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
	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

// PostCanUseBadSuccessor just iterates the result and emits EnsureRelationshipJob{FromID: principal, ToID: target, Kind: ad.CanUseBadSuccessor}.
func PostCanUseBadSuccessor(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing CanUseBadSuccessor",
		attr.Namespace("analysis"),
		attr.Function("PostCanUseBadSuccessor"),
		attr.Scope("process"),
	)()

	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "CanUseBadSuccessor Post Processing")

		for _, domain := range domainNodes {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				edgeMap, err := getCanUseBadSuccessorViaCreateChild(tx, innerDomain, localGroupData)
				if err != nil {
					return err
				}
				for targetID, duplex := range edgeMap {
					duplex.Each(func(principal uint64) bool {
						channels.Submit(ctx, outC, post.EnsureRelationshipJob{
							FromID: graph.ID(principal),
							ToID:   targetID,
							Kind:   ad.CanUseBadSuccessor,
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

// getCanUseBadSuccessorPostPatch
// calls both functions
// ANDs P1 into each entry of P2
// drops empty results
// returns the final edge map keyed by targetID.
func getCanUseBadSuccessorViaCreateChild(tx graph.Transaction, domain *graph.Node, localGroupData *LocalGroupData) (map[graph.ID]cardinality.Duplex[uint64], error) {
	// get the principals that can create child objects in this domain, regardless of where they can create them
	createChildPrincipals, err := getCreateChildPrincipalsFlat(tx, domain)
	if err != nil {
		return nil, err
	}
	if createChildPrincipals.Cardinality() == 0 {
		return make(map[graph.ID]cardinality.Duplex[uint64]), nil
	}

	tierZeroPrincipals, err := getTierZeroPrincipals(tx, domain)
	if err != nil {
		return nil, err
	}
	createChildPrincipals.AndNot(tierZeroPrincipals)
	if createChildPrincipals.Cardinality() == 0 {
		return make(map[graph.ID]cardinality.Duplex[uint64]), nil
	}
	// get the principals that can write the attributes required to set up a bad successor attack, per target
	domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String()
	if err != nil {
		return nil, err
	}
	entityIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.Entity),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
		)
	}))
	if err != nil {
		return nil, err
	}
	if len(entityIDs) == 0 {
		return make(map[graph.ID]cardinality.Duplex[uint64]), nil
	}
	supersededWriteManagedAccountPrincipals, err := getSupersededWriteManagedAccountPrincipals(tx, entityIDs)
	if err != nil {
		return nil, err
	}
	supersededWriteServiceAccountStatePrincipals, err := getSupersededWriteServiceAccountStatePrincipals(tx, entityIDs)
	if err != nil {
		return nil, err
	}
	genericWritePrincipals, err := getGenericWritePrincipals(tx, entityIDs)
	if err != nil {
		return nil, err
	}
	// must be AND on the first two, OR with the third, then we can AND with the createChildPrincipals
	principalsWithRequiredWritePermissions := make(map[graph.ID]cardinality.Duplex[uint64])
	for targetID, managedAccountDuplex := range supersededWriteManagedAccountPrincipals {
		serviceAccountStateDuplex, ok := supersededWriteServiceAccountStatePrincipals[targetID]
		if ok {
			managedAccountDuplex.And(serviceAccountStateDuplex)
			if managedAccountDuplex.Cardinality() > 0 {
				principalsWithRequiredWritePermissions[targetID] = managedAccountDuplex
			}
		}
	}
	for targetID, genericWriteDuplex := range genericWritePrincipals {
		existing, ok := principalsWithRequiredWritePermissions[targetID]
		if ok {
			existing.Or(genericWriteDuplex)
			principalsWithRequiredWritePermissions[targetID] = existing
		} else {
			principalsWithRequiredWritePermissions[targetID] = genericWriteDuplex
		}
	}
	// AND the results of who can CreateChild into who can Write the correct attributes, drop empty results, return the final edge map keyed by targetID.
	finalEdgeMap := make(map[graph.ID]cardinality.Duplex[uint64])
	for targetID, duplex := range principalsWithRequiredWritePermissions {
		duplex.And(createChildPrincipals)
		if duplex.Cardinality() > 0 {
			finalEdgeMap[targetID] = duplex
		}
	}
	return finalEdgeMap, nil
}

func getTierZeroPrincipals(tx graph.Transaction, domain *graph.Node) (cardinality.Duplex[uint64], error) {
	domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String()
	if err != nil {
		return nil, err
	}
	// find TierZero principals (by tag or SystemTags)
	// query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero)
	// ORed with query.Kind(query.Node(), tiering.KindTagTierZero)
	// this must be on the Start node

	adminCountPrincipals, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.User, ad.Group, ad.Computer, ad.DelegatedMSA),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
			query.Equals(query.NodeProperty(ad.AdminCount.String()), true),
		)
	}))
	if err != nil {
		return nil, err
	}

	tierZeroPrincipalIDs, err := ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.User, ad.Group, ad.Computer, ad.DelegatedMSA),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
			query.Or(
				query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero),
				query.Kind(query.Node(), tiering.KindTagTierZero),
			),
		)
	}))
	if err != nil {
		return nil, err
	}

	result := cardinality.NewBitmap64()
	for _, id := range adminCountPrincipals {
		result.Add(id.Uint64())
	}
	for _, id := range tierZeroPrincipalIDs {
		if !result.Contains(id.Uint64()) {
			result.Add(id.Uint64())
		}
	}
	return result, nil
}

// getCreateChildPrincipalsFlat(tx, domain) (Duplex[uint64], error) this gets the first part, who can create what and where
func getCreateChildPrincipalsFlat(tx graph.Transaction, domain *graph.Node) (cardinality.Duplex[uint64], error) {
	// calls the other fancy function
	createChildNodes, err := getCreateChildForBadSuccessorPrincipalsForDomain(tx, domain)
	if err != nil {
		return nil, err
	}
	// squash it flat! we need a bitmap of all the principals that can create child objects in this domain, regardless of where they can create them
	result := cardinality.NewBitmap64()
	for _, duplex := range createChildNodes {
		duplex.Each(func(value uint64) bool {
			result.Add(value)
			return true
		})
	}
	return result, nil

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

// getSupersededWritePrincipals(tx, domain) (map[graph.ID]Duplex[uint64], error):
// we finding a combination of either:
// WriteMsDSSupersededManagedAccountLink and WriteMsDSSupersededServiceAccountState
// GenericAll, GenericWrite, WriteOwner, WriteDACL, Owns
// builds a map per relationship, then intersects them per targetID. Returns Part2.
func getSupersededWriteManagedAccountPrincipals(tx graph.Transaction, entityIDs []graph.ID) (map[graph.ID]cardinality.Duplex[uint64], error) {

	relMap, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.Entity),
			query.Kind(query.Relationship(), ad.WriteMsDSSupersededManagedAccountLink),
			query.InIDs(query.EndID(), entityIDs...),
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

func getSupersededWriteServiceAccountStatePrincipals(tx graph.Transaction, entityIDs []graph.ID) (map[graph.ID]cardinality.Duplex[uint64], error) {

	relMap, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.Entity),
			query.Kind(query.Relationship(), ad.WriteMsDSSupersededServiceAccountState),
			query.InIDs(query.EndID(), entityIDs...),
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

func getGenericWritePrincipals(tx graph.Transaction, entityIDs []graph.ID) (map[graph.ID]cardinality.Duplex[uint64], error) {

	relMap, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), ad.Entity),
			query.Or(query.Kind(query.Relationship(), ad.GenericAll), query.Kind(query.Relationship(), ad.GenericWrite), query.Kind(query.Relationship(), ad.WriteOwner), query.Kind(query.Relationship(), ad.WriteDACL), query.Kind(query.Relationship(), ad.Owns)),
			query.InIDs(query.EndID(), entityIDs...),
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
