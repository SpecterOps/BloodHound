// Copyright 2023 Specter Ops, Inc.
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
	"slices"
	"sync"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/impact"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/traversal"
	"github.com/specterops/dawgs/util/channels"
)

func PostADCSESC16(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA *graph.Node, targetDomains *graph.NodeSet, cache ADCSCache) error {
	if isUserSpecifiesSanEnabledCollected, err := enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabledCollected.String()).Bool(); err != nil {
		return err
	} else if !isUserSpecifiesSanEnabledCollected {
		return nil
	} else if isUserSpecifiesSanEnabled, err := enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabled.String()).Bool(); err != nil {
		return err
	} else if !isUserSpecifiesSanEnabled {
		return nil
	} else if disabledExtensionsCollected, err := enterpriseCA.Properties.Get(ad.DisabledExtensionsCollected.String()).Bool(); err != nil {
		return err
	} else if !disabledExtensionsCollected {
		return nil
	} else if disabledExtensions, err := enterpriseCA.Properties.Get(ad.DisabledExtensions.String()).StringSlice(); err != nil {
		return err
	} else if !slices.Contains(disabledExtensions, "1.3.6.1.4.1.311.25.2") { // szOID_NTDS_CA_SECURITY_EXT
		return nil
	} else if publishedCertTemplates := cache.GetPublishedTemplateCache(enterpriseCA.ID); len(publishedCertTemplates) == 0 {
		return nil
	} else {
		enterpriseCAEnrollers := cache.GetEnterpriseCAEnrollers(enterpriseCA.ID)

		for _, publishedCertTemplate := range publishedCertTemplates {

			if valid, err := isCertTemplateValidForESC16(publishedCertTemplate); err != nil {
				slog.WarnContext(ctx, fmt.Sprintf("Error validating cert template %d: %v", publishedCertTemplate.ID, err))
				continue
			} else if !valid {
				continue
			} else {
				enrollers := CalculateCrossProductNodeSets(tx, groupExpansions, cache.GetCertTemplateEnrollers(publishedCertTemplate.ID), enterpriseCAEnrollers)

				if filteredEnrollers, err := filterUserDNSResults(tx, enrollers, publishedCertTemplate); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error filtering users in ESC16: %v", err))
					continue
				} else {
					filteredEnrollers.Each(func(value uint64) bool {
						for _, domain := range targetDomains.Slice() {
							channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
								FromID: graph.ID(value),
								ToID:   domain.ID,
								Kind:   ad.ADCSESC16,
							})
						}
						return true
					})
				}
			}
		}
	}
	return nil
}

func isCertTemplateValidForESC16(ct *graph.Node) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return false, err
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else {
		return true, nil
	}
}

func GetADCSESC16EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH p1 = (n)-[:MemberOf*0..]->()-[:Enroll]->(ca:EnterpriseCA)-[:TrustedForNTAuth]->(nt:NTAuthStore)-[:NTAuthStoreFor]->(d:Domain)
		WHERE ca.isuserspecifiessanenabled = true
		AND "1.3.6.1.4.1.311.25.2" IN ca.disabledextensions

		MATCH p2 = (n)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		WHERE ct.authenticationenabled = true
			AND ct.requiresmanagerapproval = false
			AND (ct.schemaversion = 1 OR ct.authorizedsignatures = 0)
			AND (
				n:Group
				OR n:Computer
				OR (
					n:User
					AND ct.subjectaltrequiredns = false
					AND ct.subjectaltrequiredomaindns = false
				)
			)

		RETURN p1,p2
	*/

	var (
		startNode  *graph.Node
		startNodes = graph.NodeSet{}

		traversalInst      = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock               = &sync.Mutex{}
		paths              = graph.PathSet{}
		path1Segments      = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs = cardinality.NewBitmap64()
		finalEnterpriseCAs = cardinality.NewBitmap64()
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	// Add startnode, Auth. Users, and Everyone to start nodes
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if nodeSet, err := FetchAuthUsersAndEveryoneGroups(tx); err != nil {
			return err
		} else {
			startNodes.AddSet(nodeSet)
			return nil
		}
	}); err != nil {
		return nil, err
	}
	startNodes.Add(startNode)

	// P1
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx,
			traversal.Plan{
				Root: n,
				Driver: ADCSESC16Path1Pattern(edge.EndID).Do(
					func(terminal *graph.PathSegment) error {
						enterpriseCA := terminal.Search(
							func(nextSegment *graph.PathSegment) bool {
								return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
							})

						lock.Lock()
						path1EnterpriseCAs.Add(enterpriseCA.ID.Uint64())
						path1Segments[enterpriseCA.ID] = append(path1Segments[enterpriseCA.ID], terminal)
						lock.Unlock()

						return nil
					}),
			}); err != nil {
			return nil, err
		}
	}

	// P2
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
			Driver: ADCSESC16Path2Pattern(edge.EndID, path1EnterpriseCAs, edge.Kind).Do(
				func(terminal *graph.PathSegment) error {
					certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					})

					if !startNode.Kinds.ContainsOneOf(ad.User) || certTemplateValidForUserVictim(certTemplate) {
						lock.Lock()
						paths.AddPath(terminal.Path())
						lock.Unlock()

						// add the ECA where the template is published (first ECA in the path in case of multi-tier hierarchy) to final list of ECAs
						terminal.Path().Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
							if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
								lock.Lock()
								finalEnterpriseCAs.Add(end.ID.Uint64())
								lock.Unlock()
								return false
							}
							return true
						})
					}
					return nil
				})}); err != nil {
			return nil, err
		}
	}

	if paths.Len() > 0 {
		finalEnterpriseCAs.Each(func(value uint64) bool {
			for _, segment := range path1Segments[graph.ID(value)] {
				paths.AddPath(segment.Path())
			}
			return true
		})
	}

	return paths, nil
}

func ADCSESC16Path1Pattern(domainId graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll),
				query.KindIn(query.End(), ad.EnterpriseCA),
				query.Equals(query.EndProperty(ad.IsUserSpecifiesSanEnabled.String()), true),
				query.InInverted(query.EndProperty(ad.DisabledExtensions.String()), "1.3.6.1.4.1.311.25.2"),
			)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC16Path2Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint64], edgeKind graph.Kind) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
			query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
			query.Or(
				query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
				query.And(
					query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), graph.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		OutboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.KindIn(query.End(), ad.EnterpriseCA, ad.AIACA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainId),
		))
}
