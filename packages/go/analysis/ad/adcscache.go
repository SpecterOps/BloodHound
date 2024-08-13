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

	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

type ADCSCache struct {
	AuthStoreForChainValid          map[graph.ID]cardinality.Duplex[uint32]
	RootCAForChainValid             map[graph.ID]cardinality.Duplex[uint32]
	ExpandedCertTemplateControllers map[graph.ID][]uint32
	CertTemplateEnrollers           map[graph.ID][]*graph.Node // principals that have enrollment on a cert template via `enroll`, `generic all`, `all extended rights` edges
	CertTemplateControllers         map[graph.ID][]*graph.Node // principals that have privileges on a cert template via `owner`, `generic all`, `write dacl`, `write owner` edges
	EnterpriseCAEnrollers           map[graph.ID][]*graph.Node // principals that have enrollment rights on an enterprise ca via `enroll` edge
	PublishedTemplateCache          map[graph.ID][]*graph.Node // cert templates that are published to an enterprise ca
	HasUPNCertMappingInForest       map[graph.ID]struct{}      // domains where at least one DC in the forest has Schannel UPN cert mapping enabled
	HasWeakCertBindingInForest      map[graph.ID]struct{}      // domains where at least one DC in the forest has Kerberos weak cert binding enabled
}

func NewADCSCache() ADCSCache {
	return ADCSCache{
		AuthStoreForChainValid:          make(map[graph.ID]cardinality.Duplex[uint32]),
		RootCAForChainValid:             make(map[graph.ID]cardinality.Duplex[uint32]),
		ExpandedCertTemplateControllers: make(map[graph.ID][]uint32),
		CertTemplateEnrollers:           make(map[graph.ID][]*graph.Node),
		CertTemplateControllers:         make(map[graph.ID][]*graph.Node),
		EnterpriseCAEnrollers:           make(map[graph.ID][]*graph.Node),
		PublishedTemplateCache:          make(map[graph.ID][]*graph.Node),
		HasUPNCertMappingInForest:       make(map[graph.ID]struct{}),
		HasWeakCertBindingInForest:      make(map[graph.ID]struct{}),
	}
}

func (s ADCSCache) BuildCache(ctx context.Context, db graph.Database, enterpriseCAs, certTemplates, domains []*graph.Node) {
	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, ct := range certTemplates {
			// cert template enrollers
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Enroll, ad.GenericAll, ad.AllExtendedRights); err != nil {
				log.Errorf("Error fetching enrollers for cert template %d: %v", ct.ID, err)
			} else {
				s.CertTemplateEnrollers[ct.ID] = firstDegreePrincipals.Slice()
			}

			// cert template controllers
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Owns, ad.GenericAll, ad.WriteDACL, ad.WriteOwner); err != nil {
				log.Errorf("Error fetching controllers for cert template %d: %v", ct.ID, err)
			} else {
				s.CertTemplateControllers[ct.ID] = firstDegreePrincipals.Slice()
			}

		}

		for _, eca := range enterpriseCAs {
			if firstDegreeEnrollers, err := fetchFirstDegreeNodes(tx, eca, ad.Enroll); err != nil {
				log.Errorf("Error fetching enrollers for enterprise ca %d: %v", eca.ID, err)
			} else {
				s.EnterpriseCAEnrollers[eca.ID] = firstDegreeEnrollers.Slice()
			}

			if publishedTemplates, err := FetchCertTemplatesPublishedToCA(tx, eca); err != nil {
				log.Errorf("Error fetching published cert templates for enterprise ca %d: %v", eca.ID, err)
			} else {
				s.PublishedTemplateCache[eca.ID] = publishedTemplates.Slice()
			}
		}

		for _, domain := range domains {
			if rootCaForNodes, err := FetchEnterpriseCAsRootCAForPathToDomain(tx, domain); err != nil {
				log.Errorf("Error getting cas via rootcafor for domain %d: %v", domain.ID, err)
			} else if authStoreForNodes, err := FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, domain); err != nil {
				log.Errorf("Error getting cas via authstorefor for domain %d: %v", domain.ID, err)
			} else {
				s.AuthStoreForChainValid[domain.ID] = cardinality.NodeSetToDuplex(authStoreForNodes)
				s.RootCAForChainValid[domain.ID] = cardinality.NodeSetToDuplex(rootCaForNodes)
			}

			// Check for weak cert config on DCs
			if upnMapping, err := HasUPNCertMappingInForest(tx, domain); err != nil {
				log.Warnf("Error checking HasUPNCertMappingInForest for domain %d: %v", domain.ID, err)
				return nil
			} else if upnMapping {
				s.HasUPNCertMappingInForest[domain.ID] = struct{}{}
			}
			if weakCertBinding, err := HasWeakCertBindingInForest(tx, domain); err != nil {
				log.Warnf("Error checking HasWeakCertBindingInForest for domain %d: %v", domain.ID, err)
				return nil
			} else if weakCertBinding {
				s.HasWeakCertBindingInForest[domain.ID] = struct{}{}
			}
		}

		return nil
	})

	log.Infof("Finished building adcs cache")
}

func (s ADCSCache) DoesCAChainProperlyToDomain(enterpriseCA, domain *graph.Node) bool {
	var domainID = domain.ID
	var caID = enterpriseCA.ID.Uint32()

	if _, ok := s.RootCAForChainValid[domainID]; !ok {
		return false
	} else if _, ok := s.AuthStoreForChainValid[domainID]; !ok {
		return false
	} else {
		return s.RootCAForChainValid[domainID].Contains(caID) && s.AuthStoreForChainValid[domainID].Contains(caID)
	}
}

func HasUPNCertMappingInForest(tx graph.Transaction, domain *graph.Node) (bool, error) {
	if trustedByNodes, err := FetchNodesWithTrustedByParentChildRelationship(tx, domain); err != nil {
		log.Errorf("error in HasUPNCertMappingInForest: unable to fetch TrustedBy nodes: %v", err)
		return false, err
	} else {
		for _, trustedByDomain := range trustedByNodes {
			if dcForNodes, err := FetchNodesWithDCForEdge(tx, trustedByDomain); err != nil {
				log.Errorf("error in HasUPNCertMappingInForest: unable to fetch DCFor nodes: %v", err)
				continue
			} else {
				for _, dcForNode := range dcForNodes {
					if cmmrProperty, err := dcForNode.Properties.Get(ad.CertificateMappingMethodsRaw.String()).Int(); err != nil {
						// We do not want to throw an error here as this property only exists if privileged collection has been performed
						continue
					} else if cmmrProperty == ein.RegistryValueDoesNotExist {
						continue
					} else if cmmrProperty&int(ein.CertificateMappingUserPrincipalName) == int(ein.CertificateMappingUserPrincipalName) {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func HasWeakCertBindingInForest(tx graph.Transaction, domain *graph.Node) (bool, error) {
	if trustedByNodes, err := FetchNodesWithTrustedByParentChildRelationship(tx, domain); err != nil {
		log.Errorf("error in HasWeakCertBindingInForest: unable to fetch TrustedBy nodes: %v", err)
		return false, err
	} else {
		for _, trustedByDomain := range trustedByNodes {
			if dcForNodes, err := FetchNodesWithDCForEdge(tx, trustedByDomain); err != nil {
				log.Errorf("error in HasWeakCertBindingInForest: unable to fetch DCFor nodes: %v", err)
				continue
			} else {
				for _, dcForNode := range dcForNodes {
					if strongCertBindingEnforcement, err := dcForNode.Properties.Get(ad.StrongCertificateBindingEnforcementRaw.String()).Int(); err != nil {
						// We do not want to throw an error here as this property only exists if privileged collection has been performed
						continue
					} else if strongCertBindingEnforcement == 0 || strongCertBindingEnforcement == 1 {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}
