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
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

type ADCSCache struct {
	AuthStoreForChainValid          map[graph.ID]cardinality.Duplex[uint32]
	RootCAForChainValid             map[graph.ID]cardinality.Duplex[uint32]
	ExpandedCertTemplateControllers map[graph.ID][]uint32
	CertTemplateControllers         map[graph.ID][]*graph.Node
	EnterpriseCAEnrollers           map[graph.ID][]*graph.Node
	PublishedTemplateCache          map[graph.ID][]*graph.Node
}

func NewADCSCache() ADCSCache {
	return ADCSCache{
		AuthStoreForChainValid:          make(map[graph.ID]cardinality.Duplex[uint32]),
		RootCAForChainValid:             make(map[graph.ID]cardinality.Duplex[uint32]),
		ExpandedCertTemplateControllers: make(map[graph.ID][]uint32),
		CertTemplateControllers:         make(map[graph.ID][]*graph.Node),
		EnterpriseCAEnrollers:           make(map[graph.ID][]*graph.Node),
		PublishedTemplateCache:          make(map[graph.ID][]*graph.Node),
	}
}

func (s ADCSCache) BuildCache(ctx context.Context, db graph.Database, enterpriseCAs, certTemplates []*graph.Node) {
	db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, ct := range certTemplates {
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Enroll, ad.GenericAll, ad.AllExtendedRights); err != nil {
				log.Errorf("error fetching enrollers for cert template %d: %v", ct.ID, err)
			} else {
				s.CertTemplateControllers[ct.ID] = firstDegreePrincipals.Slice()
			}
		}

		for _, eca := range enterpriseCAs {
			if firstDegreeEnrollers, err := fetchFirstDegreeNodes(tx, eca, ad.Enroll); err != nil {
				log.Errorf("error fetching enrollers for enterprise ca %d: %v", eca.ID, err)
			} else {
				s.EnterpriseCAEnrollers[eca.ID] = firstDegreeEnrollers.Slice()
			}

			if publishedTemplates, err := FetchCertTemplatesPublishedToCA(tx, eca); err != nil {
				log.Errorf("error fetching published cert templates for enterprise ca %d: %v", eca.ID, err)
			} else {
				s.PublishedTemplateCache[eca.ID] = publishedTemplates.Slice()
			}
		}

		if domains, err := FetchCollectedDomains(tx); err != nil {
			log.Errorf("error fetching collected domains for esc cache: %v", err)
		} else {
			for _, domain := range domains {
				if rootCaForNodes, err := FetchEnterpriseCAsRootCAForPathToDomain(tx, domain); err != nil {
					log.Errorf("error getting cas via rootcafor for domain %d: %v", domain.ID, err)
				} else if authStoreForNodes, err := FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, domain); err != nil {
					log.Errorf("error getting cas via authstorefor for domain %d: %v", domain.ID, err)
				} else {
					s.AuthStoreForChainValid[domain.ID] = cardinality.NodeSetToDuplex(authStoreForNodes)
					s.RootCAForChainValid[domain.ID] = cardinality.NodeSetToDuplex(rootCaForNodes)
				}
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
