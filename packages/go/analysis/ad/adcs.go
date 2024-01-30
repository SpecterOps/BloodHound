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
	"errors"
	"fmt"
	"slices"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

var (
	ErrNoCertParent     = errors.New("cert has no parent")
	EkuAnyPurpose       = "2.5.29.37.0"
	EkuCertRequestAgent = "1.3.6.1.4.1.311.20.2.1"
)

func fetchFirstDegreeNodes(tx graph.Transaction, targetNode *graph.Node, relKinds ...graph.Kind) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filter(
		query.And(
			query.Kind(query.Start(), ad.Entity),
			query.KindIn(query.Relationship(), relKinds...),
			query.Equals(query.EndID(), targetNode.ID),
		),
	))
}

func PostADCSESC1(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, expandedGroups impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap32()

	if publishedCertTemplates, err := FetchCertTemplatesPublishedToCA(tx, enterpriseCA); err != nil {
		return fmt.Errorf("error fetching cert templates for ECA %d: %w", enterpriseCA.ID, err)
	} else {
		for _, certTemplate := range publishedCertTemplates {
			if reqManagerApproval, err := certTemplate.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.RequiresManagerApproval.String(), certTemplate.ID, err)
				continue
			} else if authenticationEnabled, err := certTemplate.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.AuthenticationEnabled.String(), certTemplate.ID, err)
				continue
			} else if enrolleeSuppliesSubject, err := certTemplate.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.NoSecurityExtension.String(), certTemplate.ID, err)
				continue
			} else if schemaVersion, err := certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.SchemaVersion.String(), certTemplate.ID, err)
				continue
			} else if authorizedSignatures, err := certTemplate.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.AuthorizedSignatures.String(), certTemplate.ID, err)
				continue
			} else if !isCertTemplateValidForEsc1(reqManagerApproval, authenticationEnabled, enrolleeSuppliesSubject, schemaVersion, authorizedSignatures) {
				continue
			} else {
				results.Or(CalculateCrossProductNodeSets(expandedGroups, cache.CertTemplateControllers[certTemplate.ID], cache.EnterpriseCAEnrollers[enterpriseCA.ID]))
			}
		}

		results.Each(func(value uint32) (bool, error) {
			if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC1,
			}) {
				return false, nil
			} else {
				return true, nil
			}
		})

		return nil
	}
}

func PostADCSESC3(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca2, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap32()
	templates, ok := cache.PublishedTemplateCache[eca2.ID]
	if !ok {
		return nil
	} else if collected, err := eca2.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil {
		return fmt.Errorf("error getting enrollmentagentcollected for eca2 %d: %w", eca2.ID, err)
	} else if hasRestrictions, err := eca2.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
		return fmt.Errorf("error getting hasenrollmentagentrestrictions for ca %d: %w", eca2.ID, err)
	} else {
		for _, certTemplateTwo := range templates {
			if !isEndCertTemplateValidESC3(certTemplateTwo) {
				continue
			}

			if inboundTemplates, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Equals(query.EndID(), certTemplateTwo.ID),
					query.Kind(query.Relationship(), ad.EnrollOnBehalfOf),
					query.Kind(query.Start(), ad.CertTemplate),
				)
			})); err != nil {
				if !graph.IsErrNotFound(err) {
					log.Errorf("Error getting target nodes for esc3 for node %d: %v", certTemplateTwo.ID, err)
				}
			} else {
				for _, certTemplateOne := range inboundTemplates {
					if !isStartCertTemplateValidESC3(certTemplateOne) {
						continue
					}

					if publishedECAs, err := FetchCertTemplateCAs(tx, certTemplateOne); err != nil {
						log.Errorf("error getting cas for cert template %d: %v", certTemplateOne.ID, err)
					} else if publishedECAs.Len() == 0 {
						continue
					} else if collected && hasRestrictions {
						if delegatedAgents, err := fetchFirstDegreeNodes(tx, certTemplateTwo, ad.DelegatedEnrollmentAgent); err != nil {
							log.Errorf("error getting delegated agents for cert template %d: %v", certTemplateTwo.ID, err)
						} else {
							for _, eca1 := range publishedECAs {
								tempResults := CalculateCrossProductNodeSets(groupExpansions,
									cache.CertTemplateControllers[certTemplateOne.ID],
									cache.CertTemplateControllers[certTemplateTwo.ID],
									cache.EnterpriseCAEnrollers[eca1.ID],
									cache.EnterpriseCAEnrollers[eca2.ID],
									delegatedAgents.Slice())

								// Add principals to result set unless it's a user and DNS is required
								if filteredResults, err := filterUserDNSResults(tx, tempResults, certTemplateOne); err != nil {
									log.Errorf("Error filtering user dns results: %v", err)
								} else {
									results.Or(filteredResults)
								}
							}
						}
					} else {
						for _, eca1 := range publishedECAs {
							tempResults := CalculateCrossProductNodeSets(groupExpansions,
								cache.CertTemplateControllers[certTemplateOne.ID],
								cache.CertTemplateControllers[certTemplateTwo.ID],
								cache.EnterpriseCAEnrollers[eca1.ID],
								cache.EnterpriseCAEnrollers[eca2.ID])

							if filteredResults, err := filterUserDNSResults(tx, tempResults, certTemplateOne); err != nil {
								log.Errorf("Error filtering user dns results: %v", err)
							} else {
								results.Or(filteredResults)
							}
						}
					}
				}
			}
		}
	}

	results.Each(func(value uint32) (bool, error) {
		if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: graph.ID(value),
			ToID:   domain.ID,
			Kind:   ad.ADCSESC3,
		}) {
			return false, nil
		} else {
			return true, nil
		}
	})

	return nil
}

func filterUserDNSResults(tx graph.Transaction, tempResults cardinality.Duplex[uint32], certTemplate *graph.Node) (cardinality.Duplex[uint32], error) {
	if userNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.User),
			query.InIDs(query.NodeID(), cardinality.DuplexToGraphIDs(tempResults)...),
		)
	})); err != nil {
		if !graph.IsErrNotFound(err) {
			return nil, err
		}
	} else if len(userNodes) > 0 {
		if subjRequireDns, err := certTemplate.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
			log.Debugf("Failed to retrieve subjectAltRequireDNS for template %d: %v", certTemplate.ID, err)
			tempResults.Xor(cardinality.NodeSetToDuplex(userNodes))
		} else if subjRequireDomainDns, err := certTemplate.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool(); err != nil {
			log.Debugf("Failed to retrieve subjectAltRequireDomainDNS for template %d: %v", certTemplate.ID, err)
			tempResults.Xor(cardinality.NodeSetToDuplex(userNodes))
		} else if subjRequireDns || subjRequireDomainDns {
			//If either of these properties is true, we need to remove all these users from our victims list
			tempResults.Xor(cardinality.NodeSetToDuplex(userNodes))
		}
	}

	return tempResults, nil
}

func principalControlsCertTemplate(principal, certTemplate *graph.Node, groupExpansions impact.PathAggregator, cache ADCSCache) bool {
	var (
		expandedTemplateControllers = cache.ExpandedCertTemplateControllers[certTemplate.ID]
		principalID                 = principal.ID.Uint32()
	)

	if slices.Contains(expandedTemplateControllers, principalID) {
		return true
	}

	if CalculateCrossProductNodeSets(groupExpansions, graph.NewNodeSet(principal).Slice(), cache.CertTemplateControllers[certTemplate.ID]).Contains(principalID) {
		cache.ExpandedCertTemplateControllers[certTemplate.ID] = append(expandedTemplateControllers, principalID)
		return true
	}

	return false
}

func checkEmailValidity(node *graph.Node, validCertTemplates []*graph.Node, groupExpansions impact.PathAggregator, cache ADCSCache) bool {
	email, err := node.Properties.Get(common.Email.String()).String()
	if err != nil {
		log.Errorf("%s property access error %d: %v", common.Email.String(), node.ID, err)
	}

	if email == "" {
		for _, certTemplate := range validCertTemplates {
			if principalControlsCertTemplate(node, certTemplate, groupExpansions, cache) {
				if schemaVersion, err := certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
					log.Errorf("%s property access error %d: %v", ad.SchemaVersion.String(), certTemplate.ID, err)
					continue
				} else if subjectAltRequireEmail, err := certTemplate.Properties.Get(ad.SubjectAltRequireEmail.String()).Bool(); err != nil {
					log.Errorf("%s property access error %d: %v", ad.SubjectAltRequireEmail.String(), certTemplate.ID, err)
					continue
				} else if subjectRequireEmail, err := certTemplate.Properties.Get(ad.SubjectRequireEmail.String()).Bool(); err != nil {
					log.Errorf("%s property access error %d: %v", ad.SubjectRequireEmail.String(), certTemplate.ID, err)
					continue
				} else if (!subjectAltRequireEmail && !subjectRequireEmail) || schemaVersion == 1 {
					//Principal does not have an email set but at least one of the certTemplates it controls does not require it so add it to the list of principals enabled for ESC6a
					return true
				}
			}
		}
	} else {
		//Principal has an email set so add it to the list of principals enabled for ESC6a
		return true
	}
	return false
}

func checkDNSValidity(node *graph.Node, validCertTemplates []*graph.Node, groupExpansions impact.PathAggregator, cache ADCSCache) bool {
	for _, certTemplate := range validCertTemplates {
		if principalControlsCertTemplate(node, certTemplate, groupExpansions, cache) {
			if subjectAltRequireDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.SubjectAltRequireDNS.String(), certTemplate.ID, err)
				continue
			} else if subjectAltRequireDomainDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.SubjectAltRequireDomainDNS.String(), certTemplate.ID, err)
				continue
			} else if !subjectAltRequireDNS && !subjectAltRequireDomainDNS {
				return true
			}
		}
	}
	return false
}

func filterTempResultsForESC6a(tx graph.Transaction, tempResults cardinality.Duplex[uint32], groupExpansions impact.PathAggregator, validCertTemplates []*graph.Node, cache ADCSCache) cardinality.Duplex[uint32] {
	principalsEnabledForESC6a := cardinality.NewBitmap32()

	tempResults.Each(func(value uint32) (bool, error) {
		sourceID := graph.ID(value)

		if resultNode, err := tx.Nodes().Filter(query.Equals(query.NodeID(), sourceID)).First(); err != nil {
			return true, nil
		} else {
			if resultNode.Kinds.ContainsOneOf(ad.Group) {
				//A Group will be added to the list since it requires no further conditions
				principalsEnabledForESC6a.Add(value)
			} else if resultNode.Kinds.ContainsOneOf(ad.User) {
				if checkDNSValidity(resultNode, validCertTemplates, groupExpansions, cache) {
					if checkEmailValidity(resultNode, validCertTemplates, groupExpansions, cache) {
						principalsEnabledForESC6a.Add(value)
					}
				}
			} else if resultNode.Kinds.ContainsOneOf(ad.Computer) {
				if checkEmailValidity(resultNode, validCertTemplates, groupExpansions, cache) {
					principalsEnabledForESC6a.Add(value)
				}
			}
		}
		return true, nil
	})
	return principalsEnabledForESC6a
}

func PostADCSESC6a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	//The enterpriseCA that is passed here has a valid certificate chain up to the domain through an NTAuthStore and a RootCA
	if isUserSpecifiesSanEnabled, err := enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabled.String()).Bool(); err != nil {
		return err
	} else if !isUserSpecifiesSanEnabled {
		//Invalid enterpriseCA because isUserSpecifiesSanEnabled is false
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		//Return early since there are no certTemplates with an outbound PublishedTo relationship to this enterpriseCA
		return nil
	} else if canAbuseWeakCertBindingRels, err := FetchCanAbuseWeakCertBindingRels(tx, enterpriseCA); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseWeakCertBindingRels) == 0 {
		//No outbound canAbuseWeakCertBinding relationships from this enterpriseCA means there will not be a valid ESC6a path here
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
			} else if noSecurityExtension, err := publishedCertTemplate.Properties.Get(ad.NoSecurityExtension.String()).Bool(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.NoSecurityExtension.String(), publishedCertTemplate.ID, err)
				continue
			} else if schemaVersion, err := publishedCertTemplate.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.SchemaVersion.String(), publishedCertTemplate.ID, err)
				continue
			} else if authorizedSignatures, err := publishedCertTemplate.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
				log.Errorf("%s property access error %d: %v", ad.AuthorizedSignatures.String(), publishedCertTemplate.ID, err)
				continue
			} else if !isCertTemplateValidForEsc6a(reqManagerApproval, authenticationEnabled, noSecurityExtension, schemaVersion, authorizedSignatures) {
				//Continue to the next certificateTemplate published to this enterpriseCA since this certificateTemplate's properties do not allow for ESC6a
				continue
			} else {
				validCertTemplates = append(validCertTemplates, publishedCertTemplate)

				for _, controller := range cache.CertTemplateControllers[publishedCertTemplate.ID] {
					tempResults.Or(CalculateCrossProductNodeSets(groupExpansions, graph.NewNodeSet(controller).Slice(), cache.EnterpriseCAEnrollers[enterpriseCA.ID]))
				}

			}
		}

		if err := filterTempResultsForESC6a(tx, tempResults, groupExpansions, validCertTemplates, cache).Each(func(value uint32) (bool, error) {
			if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC6a,
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

func isStartCertTemplateValidESC3(template *graph.Node) bool {
	if reqManagerApproval, err := template.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		log.Errorf("error getting reqmanagerapproval for certtemplate %d: %v", template.ID, err)
	} else if reqManagerApproval {
		return false
	} else if schemaVersion, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		log.Errorf("error getting schemaversion for certtemplate %d: %v", template.ID, err)
	} else if schemaVersion == 1 {
		return true
	} else if schemaVersion > 1 {
		if authorizedSignatures, err := template.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
			log.Errorf("error getting authorizedsignatures for certtemplate %d: %v", template.ID, err)
		} else if authorizedSignatures > 0 {
			return false
		} else {
			return true
		}
	}

	return false
}

func isEndCertTemplateValidESC3(template *graph.Node) bool {
	if authEnabled, err := template.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		log.Errorf("error getting authenabled for cert template %d: %v", template.ID, err)
		return false
	} else if !authEnabled {
		return false
	} else if reqManagerApproval, err := template.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		log.Errorf("error getting reqManagerApproval for cert template %d: %v", template.ID, err)
		return false
	} else if reqManagerApproval {
		return false
	} else {
		return true
	}
}

func isCertTemplateValidForEsc6a(reqManagerApproval, authenticationEnabled, noSecurityExtension bool, schemaVersion, authorizedSignatures float64) bool {
	if reqManagerApproval {
		return false
	} else if !authenticationEnabled {
		return false
	} else if !noSecurityExtension {
		return false
	} else if schemaVersion == 1 {
		return true
	} else if schemaVersion > 1 && authorizedSignatures == 0 {
		return true
	} else {
		return false
	}
}

func isCertTemplateValidForEsc1(reqManagerApproval, authenticationEnabled, enrolleeSuppliesSubject bool, schemaVersion, authorizedSignatures float64) bool {
	if reqManagerApproval {
		return false
	} else if !authenticationEnabled {
		return false
	} else if !enrolleeSuppliesSubject {
		return false
	} else if schemaVersion == 1 {
		return true
	} else if schemaVersion > 1 && authorizedSignatures == 0 {
		return true
	} else {
		return false
	}
}

func PostADCS(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator, adcsEnabled bool) (*analysis.AtomicPostProcessingStats, error) {
	if !adcsEnabled {
		return &analysis.AtomicPostProcessingStats{}, nil
	}
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing")

	if enterpriseCertAuthorities, err := FetchNodesByKind(ctx, db, ad.EnterpriseCA); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching enterpriseCA nodes: %w", err)
	} else if rootCertAuthorities, err := FetchNodesByKind(ctx, db, ad.RootCA); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching rootCA nodes: %w", err)
	} else if certTemplates, err := FetchNodesByKind(ctx, db, ad.CertTemplate); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching cert template nodes: %w", err)
	} else if domains, err := FetchNodesByKind(ctx, db, ad.Domain); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching domain nodes: %w", err)
	} else if step1Stats, err := postADCSPreProcessStep1(ctx, db, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing step 1: %w", err)
	} else if step2Stats, err := postADCSPreProcessStep2(ctx, db, certTemplates); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing step 2: %w", err)
	} else {
		operation.Stats.Merge(step1Stats)
		operation.Stats.Merge(step2Stats)
		var cache = NewADCSCache()
		cache.BuildCache(ctx, db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			for _, enterpriseCA := range enterpriseCertAuthorities {
				if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
					innerEnterpriseCA := enterpriseCA

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := PostGoldenCert(ctx, tx, outC, innerDomain, innerEnterpriseCA); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.GoldenCert.String(), err)
						} else {
							return nil
						}
						return nil
					})

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := PostADCSESC1(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
						} else {
							return nil
						}
						return nil
					})

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := PostADCSESC3(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
						} else {
							return nil
						}

						return nil
					})

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := PostADCSESC6a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC6a.String(), err)
						}
						return nil
					})

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := PostADCSESC9a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC9a.String(), err)
						}

						return nil
					})

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if err := PostADCSESC10a(ctx, tx, outC, groupExpansions, innerEnterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC10a.String(), err)
						}

						return nil
					})
				}
			}

		}

		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep1 processes the edges that are not dependent on any other post-processed edges
func postADCSPreProcessStep1(ctx context.Context, db graph.Database, enterpriseCertAuthorities, rootCertAuthorities []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 1")
	// TODO clean up the operation.Done() calls below

	if err := PostTrustedForNTAuth(ctx, db, operation); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.TrustedForNTAuth.String(), err)
	} else if err := PostIssuedSignedBy(operation, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.IssuedSignedBy.String(), err)
	} else if err := PostEnterpriseCAFor(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnterpriseCAFor.String(), err)
	} else if err = PostCanAbuseUPNCertMapping(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.CanAbuseUPNCertMapping.String(), err)
	} else if err = PostCanAbuseWeakCertBinding(operation, enterpriseCertAuthorities); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.CanAbuseWeakCertBinding.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep2 Processes the edges that are dependent on those processed in postADCSPreProcessStep1
func postADCSPreProcessStep2(ctx context.Context, db graph.Database, certTemplates []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 2")

	if err := PostEnrollOnBehalfOf(certTemplates, operation); err != nil {
		operation.Done()
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnrollOnBehalfOf.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

func PostGoldenCert(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, domain, enterpriseCA *graph.Node) error {
	if hostCAServiceComputers, err := FetchHostsCAServiceComputers(tx, enterpriseCA); err != nil {
		log.Errorf("error fetching host ca computer for enterprise ca %d: %v", enterpriseCA.ID, err)
	} else {
		for _, computer := range hostCAServiceComputers {
			if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: computer.ID,
				ToID:   domain.ID,
				Kind:   ad.GoldenCert,
			}) {
				return nil
			}
		}
	}
	return nil
}
