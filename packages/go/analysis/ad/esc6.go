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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostCanAbuseUPNCertMapping(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
								log.Warnf("error in PostCanAbuseUPNCertMapping: unable to fetch %v property for node ID %v: %v", ad.StrongCertificateBindingEnforcementRaw.String(), dcForNode.ID, err)
								continue
							} else if cmmrProperty&0x04 == 0x04 {
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
		return collector.Return()
	})
	return nil
}

func PostCanAbuseWeakCertBinding(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
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
								log.Warnf("error in PostCanAbuseWeakCertBinding: unable to fetch %v property for node ID %v: %v", ad.StrongCertificateBindingEnforcementRaw.String(), dcForNode.ID, err)
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
		return collector.Return()
	})
	return nil
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
