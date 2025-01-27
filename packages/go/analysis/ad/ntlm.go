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
		authenticatedUsersCache map[string]graph.ID
	)

	// TODO: after adding all of our new NTLM edges, benchmark performance between submitting multiple readers per computer or single reader per computer
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {

		// Fetch all nodes where the node is a Group and is an Authenticated User
		if innerAuthenticatedUsersCache, err := FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		} else {
			authenticatedUsersCache = innerAuthenticatedUsersCache
			// Fetch all nodes where the type is Computer
			return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for computer := range cursor.Chan() {
					innerComputer := computer

					domain, err := innerComputer.Properties.Get(ad.Domain.String()).String()

					if err != nil {
						continue
					} else if authenticatedUserID, ok := authenticatedUsersCache[domain]; !ok {
						continue
					} else if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						return PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserID)
					}); err != nil {
						slog.WarnContext(ctx, fmt.Sprintf("Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToSMB, err))
						// Additional analysis may occur if one of our analysis errors
						continue
					}

					if webclientRunning, err := innerComputer.Properties.Get(ad.WebClientRunning.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						log.Warnf("Error getting webclientrunningproperty from computer %d", innerComputer.ID)
					} else if restrictOutboundNtlm, err := innerComputer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						log.Warnf("Error getting restrictoutboundntlm from computer %d", innerComputer.ID)
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
		for _, outerEnterpriseCA := range adcsCache.enterpriseCertAuthorities {
			domain := outerDomain
			enterpriseCA := outerEnterpriseCA
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if publishedCertTemplates := adcsCache.GetPublishedTemplateCache(enterpriseCA.ID); len(publishedCertTemplates) == 0 {
					//If this enterprise CA has no published templates, then there's no reason to check further
					return nil
				} else if !adcsCache.DoesCAChainProperlyToDomain(enterpriseCA, domain) {
					//If the CA doesn't chain up to the domain properly than its invalid
					return nil
				} else if ecaValid, err := isEnterpriseCAValidForADCS(enterpriseCA); err != nil {
					log.Errorf("Error validating EnterpriseCA %d for ADCS relay: %v", enterpriseCA.ID, err)
					return nil
				} else if !ecaValid {
					//Check some prereqs on the enterprise CA. If the enterprise CA is invalid, we can fast skip it
					return nil
				} else if domainsid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					log.Warnf("Error getting domainsid for domain %d: %v", domain.ID, err)
					return nil
				} else if authUsersGroup, ok := authUsersCache[domainsid]; !ok {
					//If we cant find an auth users group for this domain then we're not going to be able to make an edge regardless
					log.Warnf("Unable to find auth users group for domain %s", domainsid)
					return nil
				} else {
					//If auth users doesn't have enroll rights here than it's not valid either. Unroll enrollers into a slice and check if auth users is in it
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
							log.Errorf("Error validating cert template %d for NTLM ADCS relay: %v", certTemplate.ID, err)
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
	} else if httpsEnrollmentEpa, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTPSEPA.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else {
		return httpsEnrollment && !httpsEnrollmentEpa, nil
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
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// PostCoerceAndRelayNtlmToSmb creates edges that allow a computer with unrolled admin access to one or more computers where SMB signing is disabled.
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
				allAdminGroups.And(expandedGroups.Cardinality(group.Uint64()))
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

// FetchAuthUsersMappedToDomains Fetch all nodes where the node is a Group and is an Authenticated User
func FetchAuthUsersMappedToDomains(tx graph.Transaction) (map[string]graph.ID, error) {
	authenticatedUsers := make(map[string]graph.ID)

	err := tx.Nodes().Filter(
		query.And(
			query.Kind(query.Node(), ad.Group),
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), AuthenticatedUsersSuffix)),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for authenticatedUser := range cursor.Chan() {
			if domain, err := authenticatedUser.Properties.Get(ad.Domain.String()).String(); err != nil {
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
