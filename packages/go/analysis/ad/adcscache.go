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
	"sync"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
)

type ADCSCache struct {
	mu *sync.RWMutex

	enterpriseCertAuthorities []*graph.Node
	certTemplates             []*graph.Node
	domains                   []*graph.Node

	// To discourage direct access without getting a read lock, these are private
	authStoreForChainValid map[graph.ID]cardinality.Duplex[uint64] //Auth stores with a valid chain to the domain, key is domain ID
	rootCAForChainValid    map[graph.ID]cardinality.Duplex[uint64] //Root CA with a valid chain to the domain, key is domain ID
	hasHostingComputer     map[graph.ID]bool
	//authStorePathToDomain           map[graph.ID]map[graph.ID]graph.Path
	//rootCAPathToDomain              map[graph.ID]map[graph.ID]graph.Path
	expandedCertTemplateControllers map[graph.ID]cardinality.Duplex[uint64]
	certTemplateHasSpecialEnrollers map[graph.ID]bool          // whether Auth. Users or Everyone has enrollment rights on templates
	enterpriseCAHasSpecialEnrollers map[graph.ID]bool          // whether Auth. Users or Everyone has enrollment rights on enterprise CAs
	certTemplateEnrollers           map[graph.ID][]*graph.Node // principals that have enrollment on a cert template via `enroll`, `generic all`, `all extended rights` edges
	certTemplateControllers         map[graph.ID][]*graph.Node // principals that have privileges on a cert template via `owner`, `generic all`, `write dacl`, `write owner` edges
	enterpriseCAEnrollers           map[graph.ID][]*graph.Node // principals that have enrollment rights on an enterprise ca via `enroll` edge
	publishedTemplateCache          map[graph.ID][]*graph.Node // cert templates that are published to an enterprise ca
	hasUPNCertMappingInForest       cardinality.Duplex[uint64] // domains where at least one DC in the forest has Schannel UPN cert mapping enabled
	hasWeakCertBindingInForest      cardinality.Duplex[uint64] // domains where at least one DC in the forest has Kerberos weak cert binding enabled
}

func NewADCSCache() ADCSCache {
	return ADCSCache{
		mu:                              &sync.RWMutex{},
		authStoreForChainValid:          make(map[graph.ID]cardinality.Duplex[uint64]),
		rootCAForChainValid:             make(map[graph.ID]cardinality.Duplex[uint64]),
		hasHostingComputer:              make(map[graph.ID]bool),
		expandedCertTemplateControllers: make(map[graph.ID]cardinality.Duplex[uint64]),
		certTemplateHasSpecialEnrollers: make(map[graph.ID]bool),
		enterpriseCAHasSpecialEnrollers: make(map[graph.ID]bool),
		certTemplateEnrollers:           make(map[graph.ID][]*graph.Node),
		certTemplateControllers:         make(map[graph.ID][]*graph.Node),
		enterpriseCAEnrollers:           make(map[graph.ID][]*graph.Node),
		publishedTemplateCache:          make(map[graph.ID][]*graph.Node),
		hasUPNCertMappingInForest:       cardinality.NewBitmap64(),
		hasWeakCertBindingInForest:      cardinality.NewBitmap64(),
	}
}

func (s *ADCSCache) BuildCache(ctx context.Context, db graph.Database, enterpriseCertAuthorities, certTemplates []*graph.Node) error {
	defer measure.ContextMeasure(
		ctx,
		slog.LevelInfo,
		"ADCSCache.BuildCache",
		attr.Namespace("analysis"),
		attr.Function("BuildCache"),
		attr.Scope("routine"),
	)()

	s.mu.Lock()
	defer s.mu.Unlock()

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if domains, err := FetchNodesByKind(ctx, db, ad.Domain); err != nil {
			return fmt.Errorf("failed fetching domain nodes: %w", err)
		} else {
			s.certTemplates = certTemplates
			s.enterpriseCertAuthorities = enterpriseCertAuthorities
			s.domains = domains
		}
		for _, ct := range s.certTemplates {
			// cert template enrollers
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Enroll, ad.GenericAll, ad.AllExtendedRights); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error fetching enrollers for cert template %d: %v", ct.ID, err))
			} else {
				s.certTemplateEnrollers[ct.ID] = firstDegreePrincipals.Slice()

				// Check if Auth. Users or Everyone has enroll
				if authUsersOrEveryoneHasEnroll, err := containsAuthUsersOrEveryone(tx, firstDegreePrincipals.Slice()); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error fetching if auth. users or everyone has enroll on certtemplate %d: %v", ct.ID, err))
				} else {
					s.certTemplateHasSpecialEnrollers[ct.ID] = authUsersOrEveryoneHasEnroll
				}
			}

			// cert template controllers
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Owns, ad.GenericAll, ad.WriteDACL, ad.WriteOwner); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error fetching controllers for cert template %d: %v", ct.ID, err))
			} else {
				s.certTemplateControllers[ct.ID] = firstDegreePrincipals.Slice()
			}
		}

		for _, eca := range s.enterpriseCertAuthorities {
			if firstDegreeEnrollers, err := fetchFirstDegreeNodes(tx, eca, ad.Enroll); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error fetching enrollers for enterprise ca %d: %v", eca.ID, err))
			} else {
				s.enterpriseCAEnrollers[eca.ID] = firstDegreeEnrollers.Slice()

				// Check if Auth. Users or Everyone has enroll
				if authUsersOrEveryoneHasEnroll, err := containsAuthUsersOrEveryone(tx, firstDegreeEnrollers.Slice()); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error fetching if auth. users or everyone has enroll on enterprise ca %d: %v", eca.ID, err))
				} else {
					s.enterpriseCAHasSpecialEnrollers[eca.ID] = authUsersOrEveryoneHasEnroll
				}
			}

			if publishedTemplates, err := FetchCertTemplatesPublishedToCA(tx, eca); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error fetching published cert templates for enterprise ca %d: %v", eca.ID, err))
			} else {
				s.publishedTemplateCache[eca.ID] = publishedTemplates.Slice()
			}

			if hostingComputers, err := fetchFirstDegreeNodes(tx, eca, ad.HostsCAService); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error fetching hosting computers for enterprise ca %d: %v", eca.ID, err))
			} else {
				hasHostingComputer := false
				for _, computer := range hostingComputers.Slice() {
					if enabled, err := computer.Properties.Get(common.Enabled.String()).Bool(); err != nil {
						continue
					} else if enabled {
						hasHostingComputer = true
					}
				}
				s.hasHostingComputer[eca.ID] = hasHostingComputer
			}
		}

		for _, domain := range s.domains {
			//TODO: This code is necessary for ADCS composition during post, but we have scrapped that as part of this initiative. Leaving this code for later use
			//if rootCaPaths, err := FetchEnterpriseCAsRootCAForPathToDomainFull(tx, domain); err != nil {
			//	slog.ErrorContext(ctx, fmt.Sprintf("Error getting cas via rootcafor for domain %d: %v", domain.ID, err))
			//} else {
			//	s.rootCAForChainValid[domain.ID] = graph.NodeSetToDuplex(rootCaPaths.Terminals())
			//	for _, path := range rootCaPaths {
			//		s.rootCAPathToDomain[domain.ID][path.Terminal().ID] = path
			//	}
			//}
			//
			//if authStoreForPaths, err := FetchEnterpriseCAsTrustedForNTAuthToDomainFull(tx, domain); err != nil {
			//	slog.ErrorContext(ctx, fmt.Sprintf("Error getting cas via authstorefor for domain %d: %v", domain.ID, err))
			//} else {
			//	s.authStoreForChainValid[domain.ID] = graph.NodeSetToDuplex(authStoreForPaths.Terminals())
			//	for _, path := range authStoreForPaths {
			//		s.authStorePathToDomain[domain.ID][path.Terminal().ID] = path
			//	}
			//}

			if rootCaForNodes, err := FetchEnterpriseCAsRootCAForPathToDomain(tx, domain); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error getting cas via rootcafor for domain %d: %v", domain.ID, err))
			} else if authStoreForNodes, err := FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, domain); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error getting cas via authstorefor for domain %d: %v", domain.ID, err))
			} else {
				s.authStoreForChainValid[domain.ID] = graph.NodeSetToDuplex(authStoreForNodes)
				s.rootCAForChainValid[domain.ID] = graph.NodeSetToDuplex(rootCaForNodes)
			}

			// Check for weak cert config on DCs
			if upnMapping, err := hasUPNCertMappingInForest(tx, domain); err != nil {
				slog.WarnContext(ctx, fmt.Sprintf("Error checking hasUPNCertMappingInForest for domain %d: %v", domain.ID, err))
				return nil
			} else if upnMapping {
				s.hasUPNCertMappingInForest.Add(domain.ID.Uint64())
			}
			if weakCertBinding, err := hasWeakCertBindingInForest(tx, domain); err != nil {
				slog.WarnContext(ctx, fmt.Sprintf("Error checking hasWeakCertBindingInForest for domain %d: %v", domain.ID, err))
				return nil
			} else if weakCertBinding {
				s.hasWeakCertBindingInForest.Add(domain.ID.Uint64())
			}
		}

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error building adcs cache %v", err))
	}

	return err
}

func (s *ADCSCache) DoesCAChainProperlyToDomain(enterpriseCA, domain *graph.Node) bool {
	var domainID = domain.ID
	var caID = enterpriseCA.ID.Uint64()

	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.rootCAForChainValid[domainID]; !ok {
		return false
	} else if _, ok := s.authStoreForChainValid[domainID]; !ok {
		return false
	} else {
		return s.rootCAForChainValid[domainID].Contains(caID) && s.authStoreForChainValid[domainID].Contains(caID)
	}
}

func (s *ADCSCache) DoesCAHaveHostingComputer(enterpriseCA *graph.Node) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hasHost, ok := s.hasHostingComputer[enterpriseCA.ID]; !ok {
		return false
	} else {
		return hasHost
	}
}

func (s *ADCSCache) GetExpandedCertTemplateControllers(id graph.ID) cardinality.Duplex[uint64] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if expandedCertTemplateControllers, ok := s.expandedCertTemplateControllers[id]; !ok {
		return cardinality.NewBitmap64()
	} else {
		return expandedCertTemplateControllers
	}
}

func (s *ADCSCache) SetExpandedCertTemplateControllers(certId graph.ID, principalId uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.expandedCertTemplateControllers[certId]; !ok {
		s.expandedCertTemplateControllers[certId] = cardinality.NewBitmap64With(principalId)
	} else {
		s.expandedCertTemplateControllers[certId].Add(principalId)
	}
}

func (s *ADCSCache) GetCertTemplateHasSpecialEnrollers(id graph.ID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.certTemplateHasSpecialEnrollers[id]
}

func (s *ADCSCache) GetEnterpriseCAHasSpecialEnrollers(id graph.ID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.enterpriseCAHasSpecialEnrollers[id]
}

func (s *ADCSCache) GetCertTemplateEnrollers(id graph.ID) []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.certTemplateEnrollers[id]
}

func (s *ADCSCache) GetCertTemplateControllers(id graph.ID) []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.certTemplateControllers[id]
}

func (s *ADCSCache) GetEnterpriseCAEnrollers(id graph.ID) []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.enterpriseCAEnrollers[id]
}

func (s *ADCSCache) GetPublishedTemplateCache(id graph.ID) []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.publishedTemplateCache[id]
}

func (s *ADCSCache) HasUPNCertMappingInForest(id uint64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.hasUPNCertMappingInForest.Contains(id)
}

func (s *ADCSCache) HasWeakCertBindingInForest(id uint64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.hasWeakCertBindingInForest.Contains(id)
}

func (s *ADCSCache) GetEnterpriseCertAuthorities() []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.enterpriseCertAuthorities
}

func (s *ADCSCache) GetCertTemplates() []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.certTemplates
}

func (s *ADCSCache) GetDomains() []*graph.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.domains
}

func hasUPNCertMappingInForest(tx graph.Transaction, domain *graph.Node) (bool, error) {
	if sameForestTrustNodes, err := FetchNodesWithSameForestTrustRelationship(tx, domain); err != nil {
		return false, err
	} else {
		for _, sameForestTrustDomain := range sameForestTrustNodes {
			if dcForNodes, err := FetchNodesWithDCForEdge(tx, sameForestTrustDomain); err != nil {
				slog.Warn(fmt.Sprintf("unable to fetch DCFor nodes in hasUPNCertMappingInForest: %v", err))
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

func hasWeakCertBindingInForest(tx graph.Transaction, domain *graph.Node) (bool, error) {
	if sameForestTrustNodes, err := FetchNodesWithSameForestTrustRelationship(tx, domain); err != nil {
		return false, err
	} else {
		for _, sameForestTrustDomain := range sameForestTrustNodes {
			if dcForNodes, err := FetchNodesWithDCForEdge(tx, sameForestTrustDomain); err != nil {
				slog.Warn(fmt.Sprintf("unable to fetch DCFor nodes in hasWeakCertBindingInForest: %v", err))
				continue
			} else {
				for _, dcForNode := range dcForNodes {
					if strongCertBindingEnforcement, err := dcForNode.Properties.Get(ad.StrongCertificateBindingEnforcementRaw.String()).Int(); err != nil {
						// We do not want to throw an error here as this property only exists if privileged collection has been performed
						continue
					} else if strongCertBindingEnforcement == 0 || strongCertBindingEnforcement == 1 {
						return true, nil
					} else if strongCertBindingEnforcement == -1 { // We have confirmed the registry value does not exist. Compatibility mode is default.
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}
