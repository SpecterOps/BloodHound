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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
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
			if validationProperties, err := getValidatePublishedCertTemplateForEsc1PropertyValues(certTemplate); err != nil {
				log.Errorf("error getting published certtemplate validation properties, %v", err)
				continue
			} else if !validatePublishedCertTemplateForEsc1(validationProperties) {
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
								results.Or(tempResults)
							}
						}
					} else {
						for _, eca1 := range publishedECAs {
							tempResults := CalculateCrossProductNodeSets(groupExpansions,
								cache.CertTemplateControllers[certTemplateOne.ID],
								cache.CertTemplateControllers[certTemplateTwo.ID],
								cache.EnterpriseCAEnrollers[eca1.ID],
								cache.EnterpriseCAEnrollers[eca2.ID])
							results.Or(tempResults)
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

type PublishedCertTemplateValidationProperties struct {
	reqManagerApproval      bool
	authenticationEnabled   bool
	enrolleeSuppliesSubject bool
	schemaVersion           float64
	authorizedSignatures    float64
}

func getValidatePublishedCertTemplateForEsc1PropertyValues(certTemplate *graph.Node) (PublishedCertTemplateValidationProperties, error) {
	if reqManagerApproval, err := certTemplate.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return PublishedCertTemplateValidationProperties{}, fmt.Errorf("error getting reqmanagerapproval for certtemplate %d: %w", certTemplate.ID, err)
	} else if authenticationEnabled, err := certTemplate.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return PublishedCertTemplateValidationProperties{}, fmt.Errorf("error getting authenticationenabled for certtemplate %d: %w", certTemplate.ID, err)
	} else if enrolleeSuppliesSubject, err := certTemplate.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
		return PublishedCertTemplateValidationProperties{}, fmt.Errorf("error getting enrollesuppliessubject for certtemplate %d: %w", certTemplate.ID, err)
	} else if schemaVersion, err := certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return PublishedCertTemplateValidationProperties{}, fmt.Errorf("error getting schemaversion for certtemplate %d: %w", certTemplate.ID, err)
	} else if authorizedSignatures, err := certTemplate.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return PublishedCertTemplateValidationProperties{}, fmt.Errorf("error getting authorizedsignatures for certtemplate %d: %w", certTemplate.ID, err)
	} else {
		return PublishedCertTemplateValidationProperties{
			reqManagerApproval:      reqManagerApproval,
			authenticationEnabled:   authenticationEnabled,
			enrolleeSuppliesSubject: enrolleeSuppliesSubject,
			schemaVersion:           schemaVersion,
			authorizedSignatures:    authorizedSignatures,
		}, nil
	}
}

func validatePublishedCertTemplateForEsc1(properties PublishedCertTemplateValidationProperties) bool {
	if properties.reqManagerApproval {
		return false
	} else if !properties.authenticationEnabled {
		return false
	} else if !properties.enrolleeSuppliesSubject {
		return false
	} else if properties.schemaVersion == 1 {
		return true
	} else if properties.schemaVersion > 1 && properties.authorizedSignatures == 0 {
		return true
	} else {
		return false
	}
}

func PostEnrollOnBehalfOf(certTemplates []*graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {
	versionOneTemplates := make([]*graph.Node, 0)
	versionTwoTemplates := make([]*graph.Node, 0)

	for _, node := range certTemplates {
		if version, err := node.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
			log.Errorf("Error getting schema version for cert template %d: %v", node.ID, err)
		} else {
			if version == 1 {
				versionOneTemplates = append(versionOneTemplates, node)
			} else if version >= 2 {
				versionTwoTemplates = append(versionTwoTemplates, node)
			} else {
				log.Warnf("Got cert template %d with an invalid version %d", node.ID, version)
			}
		}
	}

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if results, err := EnrollOnBehalfOfVersionTwo(tx, versionTwoTemplates, certTemplates); err != nil {
			return err
		} else {
			for _, result := range results {
				if !channels.Submit(ctx, outC, result) {
					return nil
				}
			}

			return nil
		}
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if results, err := EnrollOnBehalfOfVersionOne(tx, versionOneTemplates, certTemplates); err != nil {
			return err
		} else {
			for _, result := range results {
				if !channels.Submit(ctx, outC, result) {
					return nil
				}
			}

			return nil
		}
	})

	return nil
}

func EnrollOnBehalfOfVersionTwo(tx graph.Transaction, versionTwoCertTemplates, allCertTemplates []*graph.Node) ([]analysis.CreatePostRelationshipJob, error) {
	results := make([]analysis.CreatePostRelationshipJob, 0)
	for _, certTemplateOne := range allCertTemplates {
		if hasBadEku, err := certTemplateHasEku(certTemplateOne, EkuAnyPurpose); err != nil {
			log.Errorf("error getting ekus for cert template %d: %v", certTemplateOne.ID, err)
		} else if hasBadEku {
			continue
		} else if hasEku, err := certTemplateHasEku(certTemplateOne, EkuCertRequestAgent); err != nil {
			log.Errorf("error getting ekus for cert template %d: %v", certTemplateOne.ID, err)
		} else if !hasEku {
			continue
		} else if domainNode, err := getDomainForCertTemplate(tx, certTemplateOne); err != nil {
			log.Errorf("error getting domain node for cert template %d: %v", certTemplateOne.ID, err)
		} else if isLinked, err := DoesCertTemplateLinkToDomain(tx, certTemplateOne, domainNode); err != nil {
			log.Errorf("error fetching paths from cert template %d to domain: %v", certTemplateOne.ID, err)
		} else if !isLinked {
			continue
		} else {
			for _, certTemplateTwo := range versionTwoCertTemplates {
				if certTemplateOne.ID == certTemplateTwo.ID {
					continue
				} else if authorizedSignatures, err := certTemplateTwo.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
					log.Errorf("Error getting authorized signatures for cert template %d: %v", certTemplateTwo.ID, err)
				} else if authorizedSignatures < 1 {
					continue
				} else if applicationPolicies, err := certTemplateTwo.Properties.Get(ad.ApplicationPolicies.String()).StringSlice(); err != nil {
					log.Errorf("Error getting application policies for cert template %d: %v", certTemplateTwo.ID, err)
				} else if !slices.Contains(applicationPolicies, EkuCertRequestAgent) {
					continue
				} else if isLinked, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					log.Errorf("error fetch paths from cert template %d to domain: %v", certTemplateTwo.ID, err)
				} else if !isLinked {
					continue
				} else {
					results = append(results, analysis.CreatePostRelationshipJob{
						FromID: certTemplateOne.ID,
						ToID:   certTemplateTwo.ID,
						Kind:   ad.EnrollOnBehalfOf,
					})
				}
			}
		}
	}

	return results, nil
}

func EnrollOnBehalfOfVersionOne(tx graph.Transaction, versionOneCertTemplates []*graph.Node, allCertTemplates []*graph.Node) ([]analysis.CreatePostRelationshipJob, error) {
	results := make([]analysis.CreatePostRelationshipJob, 0)

	for _, certTemplateOne := range allCertTemplates {
		//prefilter as much as we can first
		if hasEku, err := certTemplateHasEkuOrAll(certTemplateOne, EkuCertRequestAgent, EkuAnyPurpose); err != nil {
			log.Errorf("Error checking ekus for certtemplate %d: %v", certTemplateOne.ID, err)
		} else if !hasEku {
			continue
		} else if domainNode, err := getDomainForCertTemplate(tx, certTemplateOne); err != nil {
			log.Errorf("Error getting domain node for certtemplate %d: %v", certTemplateOne.ID, err)
		} else if hasPath, err := DoesCertTemplateLinkToDomain(tx, certTemplateOne, domainNode); err != nil {
			log.Errorf("Error fetching paths from certtemplate %d to domain: %v", certTemplateOne.ID, err)
		} else if !hasPath {
			continue
		} else {
			for _, certTemplateTwo := range versionOneCertTemplates {
				if hasPath, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					log.Errorf("Error getting domain node for certtemplate %d: %v", certTemplateTwo.ID, err)
				} else if !hasPath {
					continue
				} else {
					results = append(results, analysis.CreatePostRelationshipJob{
						FromID: certTemplateOne.ID,
						ToID:   certTemplateTwo.ID,
						Kind:   ad.EnrollOnBehalfOf,
					})
				}
			}
		}
	}

	return results, nil
}

func getDomainForCertTemplate(tx graph.Transaction, certTemplate *graph.Node) (*graph.Node, error) {
	if domainSid, err := certTemplate.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return &graph.Node{}, err
	} else if domainNode, err := analysis.FetchNodeByObjectID(tx, domainSid); err != nil {
		return &graph.Node{}, err
	} else {
		return domainNode, nil
	}
}

func certTemplateHasEkuOrAll(certTemplate *graph.Node, targetEkus ...string) (bool, error) {
	if ekus, err := certTemplate.Properties.Get(ad.EKUs.String()).StringSlice(); err != nil {
		return false, err
	} else if len(ekus) == 0 {
		return true, nil
	} else {
		for _, eku := range ekus {
			for _, targetEku := range targetEkus {
				if eku == targetEku {
					return true, nil
				}
			}
		}

		return false, nil
	}
}

func certTemplateHasEku(certTemplate *graph.Node, targetEkus ...string) (bool, error) {
	if ekus, err := certTemplate.Properties.Get(ad.EKUs.String()).StringSlice(); err != nil {
		return false, err
	} else {
		for _, eku := range ekus {
			for _, targetEku := range targetEkus {
				if eku == targetEku {
					return true, nil
				}
			}
		}

		return false, nil
	}
}

func PostADCS(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator, adcsEnabled bool) (*analysis.AtomicPostProcessingStats, error) {
	if !adcsEnabled {
		return &analysis.AtomicPostProcessingStats{}, nil
	}
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing")

	if enterpriseCertAuthorities, err := FetchNodesByKind(ctx, db, ad.EnterpriseCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching enterpriseCA nodes: %w", err)
	} else if rootCertAuthorities, err := FetchNodesByKind(ctx, db, ad.RootCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching rootCA nodes: %w", err)
	} else if certTemplates, err := FetchNodesByKind(ctx, db, ad.CertTemplate); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching cert template nodes: %w", err)
	} else if domains, err := FetchNodesByKind(ctx, db, ad.Domain); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching domain nodes: %w", err)
	} else if step1Stats, err := postADCSPreProcessStep1(ctx, db, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing step 1: %w", err)
	} else if step2Stats, err := postADCSPreProcessStep2(ctx, db, certTemplates); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing step 2: %w", err)
	} else {
		operation.Stats.Merge(step1Stats)
		operation.Stats.Merge(step2Stats)
		var cache = NewADCSCache()
		cache.BuildCache(ctx, db, enterpriseCertAuthorities, certTemplates)

		for _, domain := range domains {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, enterpriseCA := range enterpriseCertAuthorities {
					if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
						if err := PostGoldenCert(ctx, tx, outC, innerDomain, enterpriseCA); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.GoldenCert.String(), err)
						} else {
							return nil
						}
					}
				}
				return nil
			})

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, enterpriseCA := range enterpriseCertAuthorities {
					if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
						if err := PostADCSESC1(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC1.String(), err)
						} else {
							return nil
						}
					}
				}

				return nil
			})

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, enterpriseCA := range enterpriseCertAuthorities {
					if cache.DoesCAChainProperlyToDomain(enterpriseCA, innerDomain) {
						if err := PostADCSESC3(ctx, tx, outC, groupExpansions, enterpriseCA, innerDomain, cache); err != nil {
							log.Errorf("failed post processing for %s: %v", ad.ADCSESC3.String(), err)
						} else {
							return nil
						}
					}
				}

				return nil
			})
		}

		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep1 processes the edges that are not dependent on any other post-processed edges
func postADCSPreProcessStep1(ctx context.Context, db graph.Database, enterpriseCertAuthorities, rootCertAuthorities []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 1")

	if err := PostTrustedForNTAuth(ctx, db, operation); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.TrustedForNTAuth.String(), err)
	} else if err = PostIssuedSignedBy(ctx, db, operation, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.IssuedSignedBy.String(), err)
	} else if err = PostEnterpriseCAFor(ctx, db, operation, enterpriseCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnterpriseCAFor.String(), err)
	} else if err = PostCanAbuseUPNCertMapping(ctx, db, operation, enterpriseCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.CanAbuseUPNCertMapping.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

// postADCSPreProcessStep2 Processes the edges that are dependent on those processed in postADCSPreProcessStep1
func postADCSPreProcessStep2(ctx context.Context, db graph.Database, certTemplates []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing Step 2")

	if err := PostEnrollOnBehalfOf(certTemplates, operation); err != nil {
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

func PostTrustedForNTAuth(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {
	if ntAuthStoreNodes, err := FetchNodesByKind(ctx, db, ad.NTAuthStore); err != nil {
		return err
	} else {
		for _, node := range ntAuthStoreNodes {
			innerNode := node

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if thumbprints, err := innerNode.Properties.Get(ad.CertThumbprints.String()).StringSlice(); err != nil {
					return err
				} else {
					for _, thumbprint := range thumbprints {
						if thumbprint != "" {
							if sourceNodeIDs, err := findNodesByCertThumbprint(thumbprint, tx, ad.EnterpriseCA); err != nil {
								return err
							} else {
								for _, sourceNodeID := range sourceNodeIDs {
									if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
										FromID: sourceNodeID,
										ToID:   innerNode.ID,
										Kind:   ad.TrustedForNTAuth,
									}) {
										return nil
									}
								}
							}
						}
					}
				}
				return nil
			})
		}
	}

	return nil
}

func PostIssuedSignedBy(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node, rootCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range enterpriseCertAuthorities {
			if postRels, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else {
				for _, rel := range postRels {
					if !channels.Submit(ctx, outC, rel) {
						return nil
					}
				}
			}
		}

		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range rootCertAuthorities {
			if postRels, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else {
				for _, rel := range postRels {
					if !channels.Submit(ctx, outC, rel) {
						return nil
					}
				}
			}
		}

		return nil
	})

	return nil
}

func PostEnterpriseCAFor(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, ecaNode := range enterpriseCertAuthorities {
			if thumbprint, err := ecaNode.Properties.Get(ad.CertThumbprint.String()).String(); err != nil {
				return err
			} else if thumbprint != "" {
				if rootCAIDs, err := findNodesByCertThumbprint(thumbprint, tx, ad.RootCA); err != nil {
					return err
				} else {
					for _, rootCANodeID := range rootCAIDs {
						if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
							FromID: ecaNode.ID,
							ToID:   rootCANodeID,
							Kind:   ad.EnterpriseCAFor,
						}) {
							return nil
						}
					}
				}
			}
		}

		return nil
	})

	return nil
}

func processCertChainParent(node *graph.Node, tx graph.Transaction) ([]analysis.CreatePostRelationshipJob, error) {
	if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
		return []analysis.CreatePostRelationshipJob{}, err
	} else if len(certChain) > 1 {
		parentCert := certChain[1]
		if targetNodes, err := findNodesByCertThumbprint(parentCert, tx, ad.EnterpriseCA, ad.RootCA); err != nil {
			return []analysis.CreatePostRelationshipJob{}, err
		} else {
			return slices.Map(targetNodes, func(nodeId graph.ID) analysis.CreatePostRelationshipJob {
				return analysis.CreatePostRelationshipJob{
					FromID: node.ID,
					ToID:   nodeId,
					Kind:   ad.IssuedSignedBy,
				}
			}), nil
		}
	} else {
		return []analysis.CreatePostRelationshipJob{}, ErrNoCertParent
	}
}

func findNodesByCertThumbprint(certThumbprint string, tx graph.Transaction, kinds ...graph.Kind) ([]graph.ID, error) {
	return ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), kinds...),
			query.Equals(
				query.NodeProperty(ad.CertThumbprint.String()),
				certThumbprint,
			),
		)
	}))
}
