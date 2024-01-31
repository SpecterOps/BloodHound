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
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
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

func PostADCSESC6b(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	//The enterpriseCA that is passed here has a valid certificate chain up to the domain through an NTAuthStore and a RootCA
	if isUserSpecifiesSanEnabled, err := enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabled.String()).Bool(); err != nil {
		return err
	} else if !isUserSpecifiesSanEnabled {
		//Invalid enterpriseCA because isUserSpecifiesSanEnabled is false
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		//Return early since there are no certTemplates with an outbound PublishedTo relationship to this enterpriseCA
		return nil
	} else if canAbuseUPNCertMappingRels, err := FetchCanAbuseUPNCertMappingRels(tx, enterpriseCA); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseUPNCertMappingRels) == 0 {
		//No outbound canAbuseUPNCertMappingRels relationships from this enterpriseCA means there will not be a valid ESC6b path here
		return nil
	} else {
		var (
			tempResults        = cardinality.NewBitmap32()
			validCertTemplates []*graph.Node
		)

		for _, publishedCertTemplate := range publishedCertTemplates {
			if reqManagerApproval, err := publishedCertTemplate.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.RequiresManagerApproval.String(), publishedCertTemplate.ID, err)
				continue
			} else if authenticationEnabled, err := publishedCertTemplate.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.AuthenticationEnabled.String(), publishedCertTemplate.ID, err)
				continue
			} else if schemaVersion, err := publishedCertTemplate.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.SchemaVersion.String(), publishedCertTemplate.ID, err)
				continue
			} else if authorizedSignatures, err := publishedCertTemplate.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.AuthorizedSignatures.String(), publishedCertTemplate.ID, err)
				continue
			} else if !isCertTemplateValidForEsc6b(reqManagerApproval, authenticationEnabled, schemaVersion, authorizedSignatures) {
				//Continue to the next certificateTemplate published to this enterpriseCA since this certificateTemplate's properties do not allow for ESC6b
				continue
			} else {
				validCertTemplates = append(validCertTemplates, publishedCertTemplate)

				for _, controller := range cache.CertTemplateControllers[publishedCertTemplate.ID] {
					tempResults.Or(
						CalculateCrossProductNodeSets(
							groupExpansions,
							graph.NewNodeSet(controller).Slice(),
							cache.EnterpriseCAEnrollers[enterpriseCA.ID],
						),
					)
				}

			}
		}

		if err := filterTempResultsForESC6(tx, tempResults, groupExpansions, validCertTemplates, cache).Each(
			func(value uint32) (bool, error) {
				if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
					FromID: graph.ID(value),
					ToID:   domain.ID,
					Kind:   ad.ADCSESC6b,
				}) {
					return false, nil
				} else {
					return true, nil
				}
			}); err != nil {
			return err
		}
	}
	return nil
}

func isCertTemplateValidForEsc6b(reqManagerApproval, authenticationEnabled bool, schemaVersion, authorizedSignatures float64) bool {
	if reqManagerApproval {
		return false
	} else if !authenticationEnabled {
		return false
	} else if schemaVersion == 1 {
		return true
	} else if schemaVersion > 1 && authorizedSignatures == 0 {
		return true
	} else {
		return false
	}
}

func filterTempResultsForESC6(tx graph.Transaction, tempResults cardinality.Duplex[uint32], groupExpansions impact.PathAggregator, validCertTemplates []*graph.Node, cache ADCSCache) cardinality.Duplex[uint32] {
	principalsEnabledForESC6 := cardinality.NewBitmap32()

	tempResults.Each(func(value uint32) (bool, error) {
		sourceID := graph.ID(value)

		if resultNode, err := tx.Nodes().Filter(query.Equals(query.NodeID(), sourceID)).First(); err != nil {
			return true, nil
		} else {
			if resultNode.Kinds.ContainsOneOf(ad.Group) {
				//A Group will be added to the list since it requires no further conditions
				principalsEnabledForESC6.Add(value)
			} else if resultNode.Kinds.ContainsOneOf(ad.User) {
				if checkDNSValidity(resultNode, validCertTemplates, groupExpansions, cache) {
					if checkEmailValidity(resultNode, validCertTemplates, groupExpansions, cache) {
						principalsEnabledForESC6.Add(value)
					}
				}
			} else if resultNode.Kinds.ContainsOneOf(ad.Computer) {
				if checkEmailValidity(resultNode, validCertTemplates, groupExpansions, cache) {
					principalsEnabledForESC6.Add(value)
				}
			}
		}
		return true, nil
	})
	return principalsEnabledForESC6
}
