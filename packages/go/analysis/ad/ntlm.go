// Copyright 2024 Specter Ops, Inc.
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
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

// PostNTLM is the initial function used to execute our NTLM analysis
func PostNTLM(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (*analysis.AtomicPostProcessingStats, error) {
	var (
		adcsComputerCache       = make(map[string]cardinality.Duplex[uint64])
		operation               = analysis.NewPostRelationshipOperation(ctx, db, "PostNTLM")
		authenticatedUsersCache = make(map[string]graph.ID)
		protectedUsersCache     = make(map[string]cardinality.Duplex[uint64])
	)

	// TODO: after adding all of our new NTLM edges, benchmark performance between submitting multiple readers per computer or single reader per computer
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Fetch all nodes where the node is a Group and is an Authenticated User
		if innerAuthenticatedUsersCache, err := FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		} else if innerProtectedUsersCache, err := FetchProtectedUsersMappedToDomains(tx, groupExpansions); err != nil {
			return err
		} else if ldapSigningCache, err := FetchLDAPSigningCache(ctx, db); err != nil {
			return err
		} else {
			authenticatedUsersCache = innerAuthenticatedUsersCache
			protectedUsersCache = innerProtectedUsersCache

			// Fetch all nodes where the type is Computer
			return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for computer := range cursor.Chan() {
					innerComputer := computer

					domain, err := innerComputer.Properties.Get(ad.DomainSID.String()).String()

					if err != nil {
						continue
					} else if authenticatedUserGroupID, ok := authenticatedUsersCache[domain]; !ok {
						continue
					} else if protectedUsersForDomain, ok := protectedUsersCache[domain]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domain]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
						continue
					} else if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						return PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserGroupID)
					}); err != nil {
						slog.WarnContext(ctx, fmt.Sprintf("Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToSMB, err))
						// Additional analysis may occur if one of our analysis errors
						continue
					} else if err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						return PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserGroupID, ldapSigningCache)
					}); err != nil {
						slog.WarnContext(ctx, fmt.Sprintf("Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToLDAP, err))
						continue
					}

					if webclientRunning, err := innerComputer.Properties.Get(ad.WebClientRunning.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						slog.WarnContext(ctx, fmt.Sprintf("Error getting webclientrunningproperty from computer %d", innerComputer.ID))
					} else if restrictOutboundNtlm, err := innerComputer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						slog.WarnContext(ctx, fmt.Sprintf("Error getting restrictoutboundntlm from computer %d", innerComputer.ID))
					} else if webclientRunning && !restrictOutboundNtlm {
						adcsComputerCache[domain].Add(innerComputer.ID.Uint64())
					}
				}

				return cursor.Error()
			})
		}
	})
	if err != nil {
		operation.Done()
		return nil, err
	}

	if err := PostCoerceAndRelayNTLMToADCS(ctx, db, operation, authenticatedUsersCache, adcsComputerCache); err != nil {
		return nil, err
	}

	return &operation.Stats, operation.Done()
}

func PostCoerceAndRelayNTLMToADCS(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], authUsersCache map[string]graph.ID, adcsComputerCache map[string]cardinality.Duplex[uint64]) error {
	adcsCache := NewADCSCache()
	if err := adcsCache.BuildCache(ctx, db); err != nil {
		return err
	}
	for _, outerDomain := range adcsCache.domains {
		for _, outerEnterpriseCA := range adcsCache.GetEnterpriseCertAuthorities() {
			domain := outerDomain
			enterpriseCA := outerEnterpriseCA
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if publishedCertTemplates := adcsCache.GetPublishedTemplateCache(enterpriseCA.ID); len(publishedCertTemplates) == 0 {
					// If this enterprise CA has no published templates, then there's no reason to check further
					return nil
				} else if !adcsCache.DoesCAChainProperlyToDomain(enterpriseCA, domain) {
					// If the CA doesn't chain up to the domain properly then its invalid
					return nil
				} else if ecaValid, err := isEnterpriseCAValidForADCS(enterpriseCA); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error validating EnterpriseCA %d for ADCS relay: %v", enterpriseCA.ID, err))
					return nil
				} else if !ecaValid {
					// Check some prereqs on the enterprise CA. If the enterprise CA is invalid, we can fast skip it
					return nil
				} else if domainsid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error getting domainsid for domain %d: %v", domain.ID, err))
					return nil
				} else if authUsersGroup, ok := authUsersCache[domainsid]; !ok {
					// If we cant find an auth users group for this domain then we're not going to be able to make an edge regardless
					slog.WarnContext(ctx, fmt.Sprintf("Unable to find auth users group for domain %s", domainsid))
					return nil
				} else {
					// If auth users doesn't have enroll rights here than it's not valid either. Unroll enrollers into a slice and check if auth users is in it
					ecaEnrollers := adcsCache.GetEnterpriseCAEnrollers(enterpriseCA.ID)
					authUsersHasEnrollmentRights := false
					for _, l := range ecaEnrollers {
						if l.ID == authUsersGroup {
							authUsersHasEnrollmentRights = true
							break
						}
					}

					if !authUsersHasEnrollmentRights {
						return nil
					}

					for _, certTemplate := range publishedCertTemplates {
						if valid, err := isCertTemplateValidForADCSRelay(certTemplate); err != nil {
							slog.ErrorContext(ctx, fmt.Sprintf("Error validating cert template %d for NTLM ADCS relay: %v", certTemplate.ID, err))
							continue
						} else if !valid {
							continue
						} else if computers, ok := adcsComputerCache[domainsid]; !ok {
							continue
						} else {
							computers.Each(func(value uint64) bool {
								outC <- analysis.CreatePostRelationshipJob{
									FromID: authUsersGroup,
									ToID:   graph.ID(value),
									Kind:   ad.CoerceAndRelayNTLMToADCS,
								}
								return true
							})
						}
					}

					return nil
				}
			})
		}
	}

	return nil
}

func isEnterpriseCAValidForADCS(eca *graph.Node) (bool, error) {
	if httpEnrollment, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTP.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else if httpEnrollment {
		return true, nil
	} else if httpsEnrollment, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTPS.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else if !httpsEnrollment {
		return false, nil
	} else if httpsEnrollmentEpa, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTPSEPA.String()).Bool(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			return false, nil
		}
		return false, err
	} else {
		return !httpsEnrollmentEpa, nil
	}
}

func isCertTemplateValidForADCSRelay(ct *graph.Node) (bool, error) {
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
	} else if schemaVersion <= 1 {
		return true, nil
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else {
		return authorizedSignatures == 0, nil
	}
}

// PostCoerceAndRelayNTLMToSMB creates edges that allow a computer with unrolled admin access to one or more computers where SMB signing is disabled.
// Comprised solely of adminTo and memberOf edges
func PostCoerceAndRelayNTLMToSMB(tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, expandedGroups impact.PathAggregator, computer *graph.Node, authenticatedUserID graph.ID) error {
	if smbSigningEnabled, err := computer.Properties.Get(ad.SMBSigning.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if restrictOutboundNtlm, err := computer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if !smbSigningEnabled && !restrictOutboundNtlm {
		// Fetch the admins with edges to the provided computer
		if firstDegreeAdmins, err := fetchFirstDegreeNodes(tx, computer, ad.AdminTo); err != nil {
			return err
		} else if firstDegreeAdmins.ContainingNodeKinds(ad.Computer).Len() > 0 {
			outC <- analysis.CreatePostRelationshipJob{
				FromID: authenticatedUserID,
				ToID:   computer.ID,
				Kind:   ad.CoerceAndRelayNTLMToSMB,
			}
		} else {
			allAdminGroups := cardinality.NewBitmap64()
			for group := range firstDegreeAdmins.ContainingNodeKinds(ad.Group) {
				allAdminGroups.Or(expandedGroups.Cardinality(group.Uint64()))
			}

			// Fetch nodes where the node id is in our allAdminGroups bitmap and are of type Computer
			if computerIds, err := ops.FetchNodeIDs(tx.Nodes().Filter(
				query.And(
					query.InIDs(query.Node(), graph.DuplexToGraphIDs(allAdminGroups)...),
					query.Kind(query.Node(), ad.Computer),
				),
			)); err != nil {
				return err
			} else if len(computerIds) > 0 {
				outC <- analysis.CreatePostRelationshipJob{
					FromID: authenticatedUserID,
					ToID:   computer.ID,
					Kind:   ad.CoerceAndRelayNTLMToSMB,
				}
			}
		}
	}

	return nil
}

// PostCoerceAndRelayNTLMToLDAP creates edges where an authenticated user group, for a given domain, is able to target the provided computer.
// This will create either a CoerceAndRelayNTLMToLDAP or CoerceAndRelayNTLMToLDAPS edges, depending on the ldapSigning property of the domain
func PostCoerceAndRelayNTLMToLDAP(outC chan<- analysis.CreatePostRelationshipJob, computer *graph.Node, authenticatedUserID graph.ID, ldapSigningCache map[string]LDAPSigningCache) error {
	if restrictOutboundNtlm, err := computer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return err
	} else if restrictOutboundNtlm {
		return nil
	} else if webClientRunning, err := computer.Properties.Get(ad.WebClientRunning.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return err
	} else if webClientRunning {
		if domainSid, err := computer.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			if errors.Is(err, graph.ErrPropertyNotFound) {
				return nil
			} else {
				return err
			}
		} else {
			if signingCache, ok := ldapSigningCache[domainSid]; !ok {
				return nil
			} else {
				if len(signingCache.relayableToDCLDAP) > 0 {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAP,
					}
				}

				if len(signingCache.relayableToDCLDAPS) > 0 {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAPS,
					}
				}
			}
		}
	}

	return nil
}

// FetchAuthUsersMappedToDomains fetches all nodes where the node is a Group and is an Authenticated User
func FetchAuthUsersMappedToDomains(tx graph.Transaction) (map[string]graph.ID, error) {
	authenticatedUsers := make(map[string]graph.ID)

	err := tx.Nodes().Filter(
		query.And(
			query.Kind(query.Node(), ad.Group),
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), AuthenticatedUsersSuffix)),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for authenticatedUser := range cursor.Chan() {
			if domain, err := authenticatedUser.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				continue
			} else {
				authenticatedUsers[domain] = authenticatedUser.ID
			}
		}

		return cursor.Error()
	},
	)

	return authenticatedUsers, err
}

// FetchProtectedUsersMappedToDomains fetches all protected users groups mapped by their domain SID
func FetchProtectedUsersMappedToDomains(tx graph.Transaction, groupExpansions impact.PathAggregator) (map[string]cardinality.Duplex[uint64], error) {
	protectedUsers := make(map[string]cardinality.Duplex[uint64])

	err := tx.Nodes().Filter(
		query.And(
			query.Kind(query.Node(), ad.Group),
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), ProtectedUsersSuffix)),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for protectedUserGroup := range cursor.Chan() {
			if domain, err := protectedUserGroup.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				continue
			} else {
				set := cardinality.NewBitmap64()
				set.Or(groupExpansions.Cardinality(protectedUserGroup.ID.Uint64()))
				protectedUsers[domain] = set
			}
		}

		return cursor.Error()
	},
	)

	return protectedUsers, err
}

// LDAPSigningCache encapsulates whether a domain had a valid functionallevel property and slices of node ids that meet the criteria
// for a CoerceAndRelayNTLMToLDAP or CoerceAndRelayNTLMToLDAPS edge
type LDAPSigningCache struct {
	IsValidFunctionalLevel bool
	relayableToDCLDAP      []graph.ID
	relayableToDCLDAPS     []graph.ID
}

// FetchLDAPSigningCache will check all Domain Controllers (DCs) for LDAP signing. If the DC has the "ldap_signing" set to true along with "ldaps_available" to true and "ldaps_epa" to false,
// we add the DC to the relayableToDCLDAPS slice. If the DC has "ldap_signing" set to false then we simply set the DC to be a relayableToDCLDAP" slice
func FetchLDAPSigningCache(ctx context.Context, db graph.Database) (map[string]LDAPSigningCache, error) {
	if domains, err := FetchAllDomains(ctx, db); err != nil {
		return nil, err
	} else {
		cache := make(map[string]LDAPSigningCache)
		// Iterate all domains to obtain the DomainSID, which we can use to query for DCs that control the domain
		for _, domain := range domains {
			if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				if errors.Is(err, graph.ErrPropertyNotFound) {
					continue
				} else {
					return nil, err
				}
			} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
				if relayableToDcLdap, err := ops.FetchNodeIDs(tx.Nodes().Filter(
					query.And(
						query.Equals(
							query.NodeProperty(ad.DomainSID.String()), domainSid,
						),
						query.Equals(
							query.NodeProperty(ad.IsDC.String()), domainSid,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPSigning.String()), false,
						),
					),
				)); err != nil {
					return err
				} else if relayableToDcLdaps, err := ops.FetchNodeIDs(tx.Nodes().Filter(
					query.And(
						query.Equals(
							query.NodeProperty(ad.DomainSID.String()), domainSid,
						),
						query.Equals(
							query.NodeProperty(ad.IsDC.String()), domainSid,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPSAvailable.String()), true,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPSEPA.String()), false,
						),
					),
				)); err != nil {
					return err
				} else {
					isFunctionalLevelValid := false
					if functionalLevel, err := domain.Properties.Get(ad.FunctionalLevel.String()).String(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						return err
					} else if slices.Contains(vulnerableFunctionalLevels(), functionalLevel) {
						isFunctionalLevelValid = true
					}

					cache[domainSid] = LDAPSigningCache{
						IsValidFunctionalLevel: isFunctionalLevelValid,
						relayableToDCLDAP:      relayableToDcLdap,
						relayableToDCLDAPS:     relayableToDcLdaps,
					}

					return nil
				}
			}); err != nil {
				return nil, err
			}
		}

		return cache, nil
	}
}

// vulnerableFunctionalLevels is a simple constant slice of releases that are vulnerable to a CoerceAndRelayNTLMToLDAP(S) attack path
// They can be used by checking a node's functionalllevel property
func vulnerableFunctionalLevels() []string {
	return []string{
		"2000 Mixed/Native",
		"2003 Interim",
		"2003",
		"2008",
		"2008 R2",
		"2012",
	}
}
