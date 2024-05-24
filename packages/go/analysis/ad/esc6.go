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
	"slices"
	"sync"

	"github.com/specterops/bloodhound/ein"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC6a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	if isUserSpecifiesSanEnabled, err := enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabled.String()).Bool(); err != nil {
		return err
	} else if !isUserSpecifiesSanEnabled {
		return nil
	} else if canAbuseWeakCertBindingRels, err := FetchCanAbuseWeakCertBindingRels(tx, enterpriseCA); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseWeakCertBindingRels) == 0 {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		return nil
	} else {
		var (
			tempResults        = cardinality.NewBitmap32()
			validCertTemplates []*graph.Node
		)
		for _, publishedCertTemplate := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC6(publishedCertTemplate, false); err != nil {
				log.Warnf("Error validating cert template %d: %v", publishedCertTemplate.ID, err)
				continue
			} else if !valid {
				continue
			} else {
				validCertTemplates = append(validCertTemplates, publishedCertTemplate)

				for _, enroller := range cache.CertTemplateEnrollers[publishedCertTemplate.ID] {
					tempResults.Or(CalculateCrossProductNodeSets(groupExpansions, graph.NewNodeSet(enroller).Slice(), cache.EnterpriseCAEnrollers[enterpriseCA.ID]))
				}

			}
		}

		filterTempResultsForESC6(tx, tempResults, groupExpansions, validCertTemplates, cache).Each(
			func(value uint32) bool {
				return channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
					FromID: graph.ID(value),
					ToID:   domain.ID,
					Kind:   ad.ADCSESC6a,
				})
			})
	}
	return nil
}

func PostADCSESC6b(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	if isUserSpecifiesSanEnabled, err := enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabled.String()).Bool(); err != nil {
		return err
	} else if !isUserSpecifiesSanEnabled {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		return nil
	} else if canAbuseUPNCertMappingRels, err := FetchCanAbuseUPNCertMappingRels(tx, enterpriseCA); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseUPNCertMappingRels) == 0 {
		return nil
	} else {
		var (
			tempResults        = cardinality.NewBitmap32()
			validCertTemplates []*graph.Node
		)
		for _, publishedCertTemplate := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC6(publishedCertTemplate, true); err != nil {
				log.Warnf("Error validating cert template %d: %v", publishedCertTemplate.ID, err)
				continue
			} else if !valid {
				continue
			} else {
				validCertTemplates = append(validCertTemplates, publishedCertTemplate)

				for _, enroller := range cache.CertTemplateEnrollers[publishedCertTemplate.ID] {
					tempResults.Or(
						CalculateCrossProductNodeSets(
							groupExpansions,
							graph.NewNodeSet(enroller).Slice(),
							cache.EnterpriseCAEnrollers[enterpriseCA.ID],
						),
					)
				}

			}
		}

		filterTempResultsForESC6(tx, tempResults, groupExpansions, validCertTemplates, cache).Each(
			func(value uint32) bool {
				return channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
					FromID: graph.ID(value),
					ToID:   domain.ID,
					Kind:   ad.ADCSESC6b,
				})
			})
	}
	return nil
}

func PostCanAbuseUPNCertMapping(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		collector := errors.ErrorCollector{}
		for _, eca := range enterpriseCertAuthorities {
			if ecaDomainSID, err := eca.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				collector.Collect(fmt.Errorf("error in PostCanAbuseUPNCertMapping: unable to find domainsid for node ID %v: %v", eca.ID, err))
				continue
			} else if ecaDomain, err := analysis.FetchNodeByObjectID(tx, ecaDomainSID); err != nil {
				collector.Collect(fmt.Errorf("error in PostCanAbuseUPNCertMapping: unable to find node corresponding to domainsid %v: %v", ecaDomainSID, err))
				continue
			} else if trustedByNodes, err := fetchNodesWithTrustedByParentChildRelationship(tx, ecaDomain); err != nil {
				collector.Collect(fmt.Errorf("error in PostCanAbuseUPNCertMapping: unable to fetch TrustedBy nodes: %v", err))
				continue
			} else {
				for _, trustedByDomain := range trustedByNodes {
					if dcForNodes, err := fetchNodesWithDCForEdge(tx, trustedByDomain); err != nil {
						collector.Collect(fmt.Errorf("error in PostCanAbuseUPNCertMapping: unable to fetch DCFor nodes: %v", err))
						continue
					} else {
						for _, dcForNode := range dcForNodes {
							if cmmrProperty, err := dcForNode.Properties.Get(ad.CertificateMappingMethodsRaw.String()).Int(); err != nil {
								log.Warnf("Error in PostCanAbuseUPNCertMapping: unable to fetch %v property for node ID %v: %v", ad.CertificateMappingMethodsRaw.String(), dcForNode.ID, err)
								continue
							} else if cmmrProperty == ein.RegistryValueDoesNotExist {
								continue
							} else if cmmrProperty&int(ein.CertificateMappingUserPrincipalName) == int(ein.CertificateMappingUserPrincipalName) {
								if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
									FromID: eca.ID,
									ToID:   dcForNode.ID,
									Kind:   ad.CanAbuseUPNCertMapping,
								}) {
									return fmt.Errorf("context timed out while creating CanAbuseUPNCertMapping edge")
								}
							}
						}
					}
				}
			}
		}
		if err := collector.Return(); err != nil {
			log.Debugf("Errors in %s processing: %v", ad.CanAbuseUPNCertMapping.String(), err)
		}
		return nil
	})
}

func PostCanAbuseWeakCertBinding(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	return operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		collector := errors.ErrorCollector{}
		for _, eca := range enterpriseCertAuthorities {
			if ecaDomainSID, err := eca.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				collector.Collect(fmt.Errorf("error in PostCanAbuseWeakCertBinding: unable to find domainsid for node ID %v: %v", eca.ID, err))
				continue
			} else if ecaDomain, err := analysis.FetchNodeByObjectID(tx, ecaDomainSID); err != nil {
				collector.Collect(fmt.Errorf("error in PostCanAbuseWeakCertBinding: unable to find node corresponding to domainsid %v: %v", ecaDomainSID, err))
				continue
			} else if trustedByNodes, err := fetchNodesWithTrustedByParentChildRelationship(tx, ecaDomain); err != nil {
				collector.Collect(fmt.Errorf("error in PostCanAbuseWeakCertBinding: unable to fetch TrustedBy nodes: %v", err))
				continue
			} else {
				for _, trustedByDomain := range trustedByNodes {
					if dcForNodes, err := fetchNodesWithDCForEdge(tx, trustedByDomain); err != nil {
						collector.Collect(fmt.Errorf("error in PostCanAbuseWeakCertBinding: unable to fetch DCFor nodes: %v", err))
						continue
					} else {
						for _, dcForNode := range dcForNodes {
							if strongCertBindingEnforcement, err := dcForNode.Properties.Get(ad.StrongCertificateBindingEnforcementRaw.String()).Int(); err != nil {
								log.Warnf("Error in PostCanAbuseWeakCertBinding: unable to fetch %v property for node ID %v: %v", ad.StrongCertificateBindingEnforcementRaw.String(), dcForNode.ID, err)
								continue
							} else if strongCertBindingEnforcement == 0 || strongCertBindingEnforcement == 1 {
								if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
									FromID: eca.ID,
									ToID:   dcForNode.ID,
									Kind:   ad.CanAbuseWeakCertBinding,
								}) {
									return fmt.Errorf("context timed out while creating CanAbuseWeakCertBinding edge")
								}
							}
						}
					}
				}
			}
		}
		if err := collector.Return(); err != nil {
			log.Debugf("Errors in %s processing: %v", ad.CanAbuseWeakCertBinding.String(), err)
		}
		return nil
	})
}

func fetchNodesWithTrustedByParentChildRelationship(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	if nodeSet, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      root,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Relationship(), ad.TrustedBy),
				query.Equals(query.RelationshipProperty(ad.TrustType.String()), "ParentChild"),
			)
		},
	}); err != nil {
		return graph.NodeSet{}, err
	} else {
		nodeSet.Add(root)
		return nodeSet, nil
	}
}

func fetchNodesWithDCForEdge(tx graph.Transaction, rootNode *graph.Node) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      rootNode,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Start(), ad.Computer),
				query.KindIn(query.Relationship(), ad.DCFor),
			)
		},
	})
}

func filterTempResultsForESC6(tx graph.Transaction, tempResults cardinality.Duplex[uint32], groupExpansions impact.PathAggregator, validCertTemplates []*graph.Node, cache ADCSCache) cardinality.Duplex[uint32] {
	principalsEnabledForESC6 := cardinality.NewBitmap32()

	tempResults.Each(func(value uint32) bool {
		sourceID := graph.ID(value)

		if resultNode, err := tx.Nodes().Filter(query.Equals(query.NodeID(), sourceID)).First(); err != nil {
			return true
		} else {
			if resultNode.Kinds.ContainsOneOf(ad.Group) {
				//A Group will be added to the list since it requires no further conditions
				principalsEnabledForESC6.Add(value)
			} else if resultNode.Kinds.ContainsOneOf(ad.User) {
				for _, certTemplate := range validCertTemplates {
					if principalControlsCertTemplate(resultNode, certTemplate, groupExpansions, cache) {
						if certTemplateValidForUserVictim(certTemplate) {
							principalsEnabledForESC6.Add(value)
						}
					}
				}
			} else if resultNode.Kinds.ContainsOneOf(ad.Computer) {
				for _, certTemplate := range validCertTemplates {
					if principalControlsCertTemplate(resultNode, certTemplate, groupExpansions, cache) {
						principalsEnabledForESC6.Add(value)
					}
				}
			}
		}
		return true
	})
	return principalsEnabledForESC6
}

func principalControlsCertTemplate(principal, certTemplate *graph.Node, groupExpansions impact.PathAggregator, cache ADCSCache) bool {
	var (
		expandedTemplateControllers = cache.ExpandedCertTemplateControllers[certTemplate.ID]
		principalID                 = principal.ID.Uint32()
	)

	if slices.Contains(expandedTemplateControllers, principalID) {
		return true
	}

	if CalculateCrossProductNodeSets(groupExpansions, graph.NewNodeSet(principal).Slice(), cache.CertTemplateEnrollers[certTemplate.ID]).Contains(principalID) {
		cache.ExpandedCertTemplateControllers[certTemplate.ID] = append(expandedTemplateControllers, principalID)
		return true
	}

	return false
}

func isCertTemplateValidForESC6(ct *graph.Node, scenarioB bool) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return false, err
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else if !scenarioB {
		if noSecurityExtension, err := ct.Properties.Get(ad.NoSecurityExtension.String()).Bool(); err != nil {
			return false, err
		} else if !noSecurityExtension {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		return true, nil
	}
}

func GetADCSESC6EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
	MATCH p1 = (n {objectid:'S-1-5-21-2697957641-2271029196-387917394-2227'})-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d {objectid:'S-1-5-21-2697957641-2271029196-387917394'})
	WHERE ca.isuserspecifiessanenabled = true
	
	MATCH p2 = (ca)-[:CanAbuseWeakCertBinding|DCFor|TrustedBy*1..]->(d)                  <- ESC6a only
	MATCH p2 = (ca)-[:CanAbuseUPNCertMapping|DCFor|TrustedBy*1..]->(d)                   <- ESC6b only
	
	MATCH p3 = (n)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d:Domain)
	WHERE ct.nosecurityextension = true                                                  <- ESC6a only
		AND ct.authenticationenabled = true
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

	RETURN p1,p2,p3
	*/

	var (
		startNode            *graph.Node
		endNode              *graph.Node
		traversalInst        = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock                 = &sync.Mutex{}
		paths                = graph.PathSet{}
		path1Segments        = map[graph.ID][]*graph.PathSegment{}
		path2Segments        = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs   = cardinality.NewBitmap32()
		path2EnterpriseCAs   = cardinality.NewBitmap32()
		finalEnterpriseCAs   = cardinality.NewBitmap32()
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	// P1
	if err := traversalInst.BreadthFirst(ctx,
		traversal.Plan{
			Root: startNode,
			Driver: ADCSESC6Path1Pattern(edge.EndID).Do(
				func(terminal *graph.PathSegment) error {
					enterpriseCA := terminal.Search(
						func(nextSegment *graph.PathSegment) bool {
							return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
						})

					lock.Lock()
					path1EnterpriseCAs.Add(enterpriseCA.ID.Uint32())
					path1Segments[enterpriseCA.ID] = append(path1Segments[enterpriseCA.ID], terminal)
					lock.Unlock()

					return nil
				}),
		}); err != nil {
		return nil, err
	}

	// P2
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: endNode,
		Driver: ADCSESC6Path2Pattern(path1EnterpriseCAs, edge.Kind).Do(func(terminal *graph.PathSegment) error {
			enterpriseCA := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			path2Segments[enterpriseCA.ID] = append(path2Segments[enterpriseCA.ID], terminal)
			path2EnterpriseCAs.Add(enterpriseCA.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}


	// P3
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6Path3Pattern(edge.EndID, path2EnterpriseCAs, edge.Kind).Do(
			func(terminal *graph.PathSegment) error {
				certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				if !startNode.Kinds.ContainsOneOf(ad.User) || certTemplateValidForUserVictim(certTemplate) {
					paths.AddPath(terminal.Path())

					// add the ECA where the template is published (first ECA in the path in case of multi-tier hierarchy) to final list of ECAs
					terminal.Path().Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {						
						if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
							finalEnterpriseCAs.Add(end.ID.Uint32())
							return false
						}	
						return true
					})
				}
				return nil
			})}); err != nil {
		return nil, err
	}

	if paths.Len() > 0 {
		finalEnterpriseCAs.Each(func(value uint32) bool {
			for _, segment := range path1Segments[graph.ID(value)] {
				paths.AddPath(segment.Path())
			}
			for _, segment := range path2Segments[graph.ID(value)] {
				paths.AddPath(segment.Path())
			}
			return true
		})
	}

	return paths, nil
}

func getESC6CertTemplateCriteria(edgeKind graph.Kind) graph.Criteria {
	criteria := query.And(
		query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
		query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
		query.Or(
			query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
			query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
		),
	)

	if edgeKind == ad.ADCSESC6a {
		criteria = query.And(
			criteria,
			query.Equals(query.EndProperty(ad.NoSecurityExtension.String()), true),
		)
	}

	return criteria
}

func getESC6AbuseEdgeCriteria(edgeKind graph.Kind) graph.Criteria {
	if edgeKind == ad.ADCSESC6a {
		return query.KindIn(query.Relationship(), ad.CanAbuseWeakCertBinding)
	}
	return query.KindIn(query.Relationship(), ad.CanAbuseUPNCertMapping)
}

func ADCSESC6Path1Pattern(domainId graph.ID) traversal.PatternContinuation {
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

func ADCSESC6Path2Pattern(enterpriseCAs cardinality.Duplex[uint32], edgeKind graph.Kind) traversal.PatternContinuation {
	return traversal.NewPattern().
		InboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.TrustedBy),
			query.Kind(query.Start(), ad.Domain),
		)).
		Inbound(query.And(
			query.KindIn(query.Relationship(), ad.DCFor),
			query.Kind(query.Start(), ad.Computer),
		)).
		Inbound(query.And(
			getESC6AbuseEdgeCriteria(edgeKind),
			query.InIDs(query.Start(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
		))
}

func ADCSESC6Path3Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32], edgeKind graph.Kind) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			getESC6CertTemplateCriteria(edgeKind),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.IssuedSignedBy),
			query.Kind(query.End(), ad.EnterpriseCA),
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
