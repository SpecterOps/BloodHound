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

	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

// EnterpriseCAChainedDomains pairs an Enterprise CA node with the set of domain
// node IDs that are reachable from it through a valid certificate chain. A domain
// is considered reachable when the Enterprise CA has both a RootCAFor path and a
// TrustedForNTAuth path leading to that domain, meaning the CA is trusted to issue
// certificates that can authenticate against it.
type EnterpriseCAChainedDomains struct {
	// EnterpriseCA is the Enterprise CA node at the root of the certificate chain.
	EnterpriseCA *graph.Node
	// Domains is the set of domain node IDs reachable from EnterpriseCA through a
	// valid certificate chain (RootCAFor ∩ TrustedForNTAuth).
	Domains cardinality.Duplex[uint64]
}

// NewEnterpriseCAChainedDomains creates an EnterpriseCAChainedDomains for the
// given Enterprise CA node, seeded with a single domain node ID.
func NewEnterpriseCAChainedDomains(enterpriseCA *graph.Node) *EnterpriseCAChainedDomains {
	return &EnterpriseCAChainedDomains{EnterpriseCA: enterpriseCA, Domains: cardinality.NewBitmap64()}
}

// AddDomain adds a domain node ID to the set of domains reachable from this
// Enterprise CA. It is a no-op if the domain is already present.
func (s *EnterpriseCAChainedDomains) AddDomain(domainID uint64) {
	s.Domains.CheckedAdd(domainID)
}

type ADCSCache struct {
	mutex *sync.RWMutex

	enterpriseCertAuthorities []*graph.Node
	certTemplates             []*graph.Node
	domains                   []*graph.Node

	// To discourage direct access without getting a read lock, these are private
	chainedDomainsByEnterpriseCA    map[uint64]*EnterpriseCAChainedDomains  // chainedDomainsByEnterpriseCA maps each Enterprise CA node ID to the domains reachable from it through a valid certificate chain (RootCAFor ∩ TrustedForNTAuth).
	certTemplateHasSpecialEnrollers map[graph.ID]bool                       // whether Auth. Users or Everyone has enrollment rights on templates
	enterpriseCAHasSpecialEnrollers map[graph.ID]bool                       // whether Auth. Users or Everyone has enrollment rights on enterprise CAs
	certTemplateEnrollers           map[graph.ID][]*graph.Node              // principals that have enrollment on a cert template via `enroll`, `generic all`, `all extended rights` edges
	certTemplateControllers         map[graph.ID][]*graph.Node              // principals that have privileges on a cert template via `owner`, `generic all`, `write dacl`, `write owner` edges
	enterpriseCAEnrollers           map[graph.ID][]*graph.Node              // principals that have enrollment rights on an enterprise ca via `enroll` edge
	publishedTemplateCache          map[graph.ID][]*graph.Node              // cert templates that are published to an enterprise ca
	authUsersByDomain               map[graph.ID]graph.ID                   // domain node ID → Authenticated Users group node ID
	ecasWithHostingComputers        cardinality.Duplex[uint64]              // enterprise CAs with at least one hosting computer where the computer is enabled
	authStoreForChainValid          map[graph.ID]cardinality.Duplex[uint64] //Auth stores with a valid chain to the domain, key is domain ID
	rootCAForChainValid             map[graph.ID]cardinality.Duplex[uint64] //Root CA with a valid chain to the domain, key is domain ID
	hasHostingComputer              map[graph.ID]bool

	// ESC4-specific caches: principals with specific rights on cert templates, pre-computed to avoid per-ECA DB queries
	certTemplateGenericWriters              map[graph.ID][]*graph.Node // principals with GenericWrite on a cert template
	certTemplateEnrollOrAllExtendedRighters map[graph.ID][]*graph.Node // principals with Enroll or AllExtendedRights on a cert template
	certTemplateWritePKINameFlaggers        map[graph.ID][]*graph.Node // principals with WritePKINameFlag on a cert template
	certTemplateWritePKIEnrollmentFlaggers  map[graph.ID][]*graph.Node // principals with WritePKIEnrollmentFlag on a cert template
	hasUPNCertMappingInForest               cardinality.Duplex[uint64] // domains where at least one DC in the forest has Schannel UPN cert mapping enabled
	hasWeakCertBindingInForest              cardinality.Duplex[uint64] // domains where at least one DC in the forest has Kerberos weak cert binding enabled
}

func NewADCSCache() *ADCSCache {
	return &ADCSCache{
		mutex:                                   &sync.RWMutex{},
		chainedDomainsByEnterpriseCA:            make(map[uint64]*EnterpriseCAChainedDomains),
		certTemplateHasSpecialEnrollers:         make(map[graph.ID]bool),
		enterpriseCAHasSpecialEnrollers:         make(map[graph.ID]bool),
		authStoreForChainValid:                  make(map[graph.ID]cardinality.Duplex[uint64]),
		rootCAForChainValid:                     make(map[graph.ID]cardinality.Duplex[uint64]),
		hasHostingComputer:                      make(map[graph.ID]bool),
		certTemplateEnrollers:                   make(map[graph.ID][]*graph.Node),
		certTemplateControllers:                 make(map[graph.ID][]*graph.Node),
		enterpriseCAEnrollers:                   make(map[graph.ID][]*graph.Node),
		publishedTemplateCache:                  make(map[graph.ID][]*graph.Node),
		authUsersByDomain:                       make(map[graph.ID]graph.ID),
		ecasWithHostingComputers:                cardinality.NewBitmap64(),
		hasUPNCertMappingInForest:               cardinality.NewBitmap64(),
		certTemplateGenericWriters:              make(map[graph.ID][]*graph.Node),
		certTemplateEnrollOrAllExtendedRighters: make(map[graph.ID][]*graph.Node),
		certTemplateWritePKINameFlaggers:        make(map[graph.ID][]*graph.Node),
		certTemplateWritePKIEnrollmentFlaggers:  make(map[graph.ID][]*graph.Node),
		hasWeakCertBindingInForest:              cardinality.NewBitmap64(),
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {

		if domains, err := FetchNodesByKind(ctx, db, ad.Domain); err != nil {
			return fmt.Errorf("failed fetching domain nodes: %w", err)
		} else {
			s.certTemplates = certTemplates
			s.enterpriseCertAuthorities = enterpriseCertAuthorities
			s.domains = domains
		}

		// Fetch Auth. Users and Everyone groups once for the entire BuildCache transaction
		// instead of re-fetching per cert template / enterprise CA.
		specialGroups, err := FetchAuthUsersAndEveryoneGroups(tx)
		if err != nil {
			return fmt.Errorf("failed fetching auth users and everyone groups: %w", err)
		}

		certTemplateMeasure := measure.ContextMeasure(
			ctx,
			slog.LevelInfo,
			"BuildCache cert template loop",
			attr.Namespace("analysis"),
			attr.Function("BuildCache.CertTemplates"),
			attr.Scope("routine"),
		)

		for _, ct := range s.certTemplates {
			if certTemplateEnrollers, err := fetchFirstDegreeNodes(tx, ct, ad.Enroll, ad.GenericAll, ad.AllExtendedRights); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching enrollers for cert template",
					slog.Uint64("cert_template", uint64(ct.ID)),
					attr.Error(err),
				)
			} else {
				s.certTemplateEnrollers[ct.ID] = certTemplateEnrollers.Slice()

				// Check if Auth. Users or Everyone has enroll
				if authUsersOrEveryoneHasEnroll, err := containsAuthUsersOrEveryone(tx, specialGroups, certTemplateEnrollers.Slice()); err != nil {
					slog.ErrorContext(
						ctx,
						"Error fetching if auth. users or everyone has enroll on certtemplate",
						slog.Uint64("cert_template", uint64(ct.ID)),
						attr.Error(err),
					)
				} else {
					s.certTemplateHasSpecialEnrollers[ct.ID] = authUsersOrEveryoneHasEnroll
				}
			}

			if certTemplateControllers, err := fetchFirstDegreeNodes(tx, ct, ad.Owns, ad.GenericAll, ad.WriteDACL, ad.WriteOwner); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching controllers for cert template",
					slog.Uint64("cert_template", uint64(ct.ID)),
					attr.Error(err),
				)
			} else {
				s.certTemplateControllers[ct.ID] = certTemplateControllers.Slice()
			}

			// ESC4-specific principal caches
			if principals, err := FetchPrincipalsWithGenericWriteOnCertTemplate(tx, ct); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching principals with GenericWrite on cert template",
					slog.Uint64("cert_template", uint64(ct.ID)),
					attr.Error(err),
				)
			} else {
				s.certTemplateGenericWriters[ct.ID] = principals.Slice()
			}

			if principals, err := FetchPrincipalsWithEnrollOrAllExtendedRightsOnCertTemplate(tx, ct); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching principals with Enroll/AllExtendedRights on cert template",
					slog.Uint64("cert_template", uint64(ct.ID)),
					attr.Error(err),
				)
			} else {
				s.certTemplateEnrollOrAllExtendedRighters[ct.ID] = principals.Slice()
			}

			if principals, err := FetchPrincipalsWithWritePKINameFlagOnCertTemplate(tx, ct); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching principals with WritePKINameFlag on cert template",
					slog.Uint64("cert_template", uint64(ct.ID)),
					attr.Error(err),
				)
			} else {
				s.certTemplateWritePKINameFlaggers[ct.ID] = principals.Slice()
			}

			if principals, err := FetchPrincipalsWithWritePKIEnrollmentFlagOnCertTemplate(tx, ct); err != nil {
				slog.ErrorContext(ctx,
					"Error fetching principals with WritePKIEnrollmentFlag on cert template",
					slog.Uint64("cert_template", uint64(ct.ID)),
					attr.Error(err),
				)
			} else {
				s.certTemplateWritePKIEnrollmentFlaggers[ct.ID] = principals.Slice()
			}
		}

		certTemplateMeasure()

		ecaMeasure := measure.ContextMeasure(
			ctx,
			slog.LevelInfo,
			"BuildCache enterprise CA loop",
			attr.Namespace("analysis"),
			attr.Function("BuildCache.EnterpriseCAs"),
			attr.Scope("routine"),
		)

		for _, eca := range s.enterpriseCertAuthorities {
			if enterpriseCAEnrollers, err := fetchFirstDegreeNodes(tx, eca, ad.Enroll); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching enrollers for enterprise ca",
					slog.Uint64("enterprise_ca", uint64(eca.ID)),
					attr.Error(err),
				)
			} else {
				s.enterpriseCAEnrollers[eca.ID] = enterpriseCAEnrollers.Slice()

				// Check if Auth. Users or Everyone has enroll
				if authUsersOrEveryoneHasEnroll, err := containsAuthUsersOrEveryone(tx, specialGroups, enterpriseCAEnrollers.Slice()); err != nil {
					slog.ErrorContext(
						ctx,
						"Error fetching if auth. users or everyone has enroll on enterprise ca",
						slog.Uint64("enterprise_ca", uint64(eca.ID)),
						attr.Error(err),
					)
				} else {
					s.enterpriseCAHasSpecialEnrollers[eca.ID] = authUsersOrEveryoneHasEnroll
				}
			}

			if publishedTemplates, err := FetchCertTemplatesPublishedToCA(tx, eca); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching published cert templates for enterprise ca",
					slog.Uint64("enterprise_ca", uint64(eca.ID)),
					attr.Error(err),
				)
			} else {
				s.publishedTemplateCache[eca.ID] = publishedTemplates.Slice()
			}

			if hostingComputers, err := fetchFirstDegreeNodes(tx, eca, ad.HostsCAService); err != nil {
				slog.ErrorContext(
					ctx,
					"Error fetching hosting computers for enterprise ca",
					slog.Uint64("enterprise_ca", uint64(eca.ID)),
					attr.Error(err),
				)
			} else {
				hasHostingComputer := false

				for _, computer := range hostingComputers.Slice() {
					if enabled, err := computer.Properties.Get(common.Enabled.String()).Bool(); err != nil {
						continue
					} else if enabled {
						hasHostingComputer = true
						break
					}
				}
				s.hasHostingComputer[eca.ID] = hasHostingComputer

			}
		}

		ecaMeasure()

		domainMeasure := measure.ContextMeasure(
			ctx,
			slog.LevelInfo,
			"BuildCache domain loop",
			attr.Namespace("analysis"),
			attr.Function("BuildCache.Domains"),
			attr.Scope("routine"),
		)

		for _, domain := range s.domains {
			if rootCaForNodes, err := FetchEnterpriseCAsRootCAForPathToDomain(tx, domain); err != nil {
				slog.ErrorContext(
					ctx,
					"Error getting cas via rootcafor for domain",
					slog.Uint64("domain_id", uint64(domain.ID)),
					attr.Error(err),
				)
			} else if authStoreForNodes, err := FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, domain); err != nil {
				slog.ErrorContext(
					ctx,
					"Error getting cas via authstorefor for domain",
					slog.Uint64("domain_id", uint64(domain.ID)),
					attr.Error(err),
				)
			} else {
				s.authStoreForChainValid[domain.ID] = graph.NodeSetToDuplex(authStoreForNodes)
				s.rootCAForChainValid[domain.ID] = graph.NodeSetToDuplex(rootCaForNodes)
			}

			// Check for weak cert config on DCs
			if upnMapping, err := hasUPNCertMappingInForest(tx, domain); err != nil {
				slog.WarnContext(
					ctx,
					"Error checking hasUPNCertMappingInForest for domain",
					slog.Uint64("domain_id", uint64(domain.ID)),
					attr.Error(err),
				)
			} else if upnMapping {
				s.hasUPNCertMappingInForest.Add(domain.ID.Uint64())
			}

			if weakCertBinding, err := hasWeakCertBindingInForest(tx, domain); err != nil {
				slog.WarnContext(
					ctx,
					"Error checking hasWeakCertBindingInForest for domain",
					slog.Uint64("domain_id", uint64(domain.ID)),
					attr.Error(err),
				)
			} else if weakCertBinding {
				s.hasWeakCertBindingInForest.Add(domain.ID.Uint64())
			}

			if authUserID, ok, err := fetchAuthUserForDomain(tx, domain); err != nil {
				slog.WarnContext(
					ctx,
					"Error fetching authenticated users group for domain",
					slog.Uint64("domain_id", uint64(domain.ID)),
					attr.Error(err),
				)
			} else if ok {
				s.authUsersByDomain[domain.ID] = authUserID
			}
		}

		domainMeasure()

		return nil
	})

	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error building adcs cache",
			attr.Error(err),
		)
	}

	return err
}

func (s *ADCSCache) GetECAHostedChainedDomains() map[uint64]*EnterpriseCAChainedDomains {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filtered := make(map[uint64]*EnterpriseCAChainedDomains)

	for _, enterpriseCA := range s.enterpriseCertAuthorities {
		innerEnterpriseCA := enterpriseCA

		targetDomains := NewEnterpriseCAChainedDomains(enterpriseCA)
		for _, domain := range s.domains {
			innerDomain := domain

			if hasHost, ok := s.hasHostingComputer[innerEnterpriseCA.ID]; !ok {
				continue
			} else if !hasHost {
				continue
			} else if _, ok := s.rootCAForChainValid[innerDomain.ID]; !ok {
				continue
			} else if _, ok := s.authStoreForChainValid[innerDomain.ID]; !ok {
				continue
			} else if s.rootCAForChainValid[innerDomain.ID].Contains(innerEnterpriseCA.ID.Uint64()) && s.authStoreForChainValid[innerDomain.ID].Contains(innerEnterpriseCA.ID.Uint64()) {
				targetDomains.AddDomain(innerDomain.ID.Uint64())
			}

		}
		if targetDomains.Domains.Cardinality() > 0 {
			filtered[enterpriseCA.ID.Uint64()] = targetDomains
		}
	}

	return filtered
}

func (s *ADCSCache) GetChainedDomains() map[uint64]*EnterpriseCAChainedDomains {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filtered := make(map[uint64]*EnterpriseCAChainedDomains)

	for _, enterpriseCA := range s.enterpriseCertAuthorities {
		innerEnterpriseCA := enterpriseCA

		targetDomains := NewEnterpriseCAChainedDomains(enterpriseCA)
		for _, domain := range s.domains {
			innerDomain := domain

			if _, ok := s.rootCAForChainValid[innerDomain.ID]; !ok {
				continue
			} else if _, ok := s.authStoreForChainValid[innerDomain.ID]; !ok {
				continue
			} else if s.rootCAForChainValid[innerDomain.ID].Contains(innerEnterpriseCA.ID.Uint64()) && s.authStoreForChainValid[innerDomain.ID].Contains(innerEnterpriseCA.ID.Uint64()) {
				targetDomains.AddDomain(innerDomain.ID.Uint64())
			}

		}
		if targetDomains.Domains.Cardinality() > 0 {
			filtered[enterpriseCA.ID.Uint64()] = targetDomains
		}
	}

	return filtered
}

func (s *ADCSCache) GetCertTemplateHasSpecialEnrollers(id graph.ID) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateHasSpecialEnrollers[id]
}

func (s *ADCSCache) GetEnterpriseCAHasSpecialEnrollers(id graph.ID) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.enterpriseCAHasSpecialEnrollers[id]
}

func (s *ADCSCache) GetCertTemplateEnrollers(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateEnrollers[id]
}

func (s *ADCSCache) GetCertTemplateControllers(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateControllers[id]
}

func (s *ADCSCache) GetCertTemplateGenericWriters(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateGenericWriters[id]
}

func (s *ADCSCache) GetCertTemplateEnrollOrAllExtendedRighters(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateEnrollOrAllExtendedRighters[id]
}

func (s *ADCSCache) GetCertTemplateWritePKINameFlaggers(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateWritePKINameFlaggers[id]
}

func (s *ADCSCache) GetCertTemplateWritePKIEnrollmentFlaggers(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplateWritePKIEnrollmentFlaggers[id]
}

func (s *ADCSCache) GetEnterpriseCAEnrollers(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.enterpriseCAEnrollers[id]
}

func (s *ADCSCache) GetPublishedTemplateCache(id graph.ID) []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.publishedTemplateCache[id]
}

func (s *ADCSCache) HasUPNCertMappingInForest(id uint64) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.hasUPNCertMappingInForest.Contains(id)
}

func (s *ADCSCache) HasWeakCertBindingInForest(id uint64) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.hasWeakCertBindingInForest.Contains(id)
}

func (s *ADCSCache) GetCertTemplates() []*graph.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.certTemplates
}

// fetchAuthUserForDomain looks up the Authenticated Users well-known group that belongs
// to the given domain. It returns the group's graph ID, a boolean indicating whether the
// group was found, and any error encountered. At most one Authenticated Users group
// is expected per domain, so the first match is returned.
func fetchAuthUserForDomain(tx graph.Transaction, domain *graph.Node) (graph.ID, bool, error) {
	var (
		domainSID string
		err       error
	)

	if domainSID, err = domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return 0, false, err
	}

	nodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Group),
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.AuthenticatedUsersSIDSuffix.String()),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
		)
	}))
	if err != nil {
		return 0, false, err
	}

	for _, node := range nodes {
		return node.ID, true, nil
	}

	return 0, false, nil
}

func (s *ADCSCache) GetAuthUserForDomain(domainID graph.ID) (graph.ID, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	authUserID, ok := s.authUsersByDomain[domainID]
	return authUserID, ok
}

func hasUPNCertMappingInForest(tx graph.Transaction, domain *graph.Node) (bool, error) {
	if sameForestTrustNodes, err := FetchNodesWithSameForestTrustRelationship(tx, domain); err != nil {
		return false, err
	} else {
		for _, sameForestTrustDomain := range sameForestTrustNodes {
			if dcForNodes, err := FetchNodesWithDCForEdge(tx, sameForestTrustDomain); err != nil {
				slog.Warn(
					"Unable to fetch DCFor nodes in hasUPNCertMappingInForest",
					attr.Error(err),
				)
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
				slog.Warn(
					"Unable to fetch DCFor nodes in hasWeakCertBindingInForest",
					attr.Error(err),
				)
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
