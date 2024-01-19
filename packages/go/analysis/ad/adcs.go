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
			var (
				reqManagerApproval, _      = certTemplate.Properties.Get(ad.RequiresManagerApproval.String()).Bool()
				authenticationEnabled, _   = certTemplate.Properties.Get(ad.AuthenticationEnabled.String()).Bool()
				enrolleeSuppliesSubject, _ = certTemplate.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool()
				schemaVersion, _           = certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64()
				authorizedSignatures, _    = certTemplate.Properties.Get(ad.AuthorizedSignatures.String()).Float64()
			)
			if !isCertTemplateValidForEsc1(reqManagerApproval, authenticationEnabled, enrolleeSuppliesSubject, schemaVersion, authorizedSignatures) {
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

func PostADCSESC6a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	var (
		results                      = cardinality.NewBitmap32()
		isUserSpecifiesSanEnabled, _ = enterpriseCA.Properties.Get(ad.IsUserSpecifiesSanEnabled.String()).Bool()
	)

	if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		return nil
	} else if !isUserSpecifiesSanEnabled {
		return nil
	} else if canAbuseWeakCertBindingRels, err := FetchCanAbuseWeakCertBindingRels(tx, enterpriseCA); err != nil {
		log.Errorf("error getting canAbuseWeakCertBindingRels for enterpriseCA %d: %v", enterpriseCA.ID, err)
	} else if len(canAbuseWeakCertBindingRels) == 0 {
		return nil
	} else {
		var (
			tempResults     = cardinality.NewBitmap32()
			tempUserResults = cardinality.NewBitmap32()
			exception       = false
		)
		for _, publishedCertTemplate := range publishedCertTemplates {
			var (
				reqManagerApproval, _         = publishedCertTemplate.Properties.Get(ad.RequiresManagerApproval.String()).Bool()
				authenticationEnabled, _      = publishedCertTemplate.Properties.Get(ad.AuthenticationEnabled.String()).Bool()
				noSecurityExtension, _        = publishedCertTemplate.Properties.Get(ad.NoSecurityExtension.String()).Bool()
				schemaVersion, _              = publishedCertTemplate.Properties.Get(ad.SchemaVersion.String()).Float64()
				authorizedSignatures, _       = publishedCertTemplate.Properties.Get(ad.AuthorizedSignatures.String()).Float64()
				subjectAltRequireEmail, _     = publishedCertTemplate.Properties.Get("subjectaltrequireemail").Bool()
				subjectRequireEmail, _        = publishedCertTemplate.Properties.Get("subjectrequireemail").Bool()
				subjectAltRequireDNS, _       = publishedCertTemplate.Properties.Get("subjectaltrequiredns").Bool()
				subjectAltRequireDomainDNS, _ = publishedCertTemplate.Properties.Get("subjectaltrequiredomaindns").Bool()
			)

			if !isCertTemplateValidForEsc6a(reqManagerApproval, authenticationEnabled, noSecurityExtension, schemaVersion, authorizedSignatures) {
				continue
			} else {

				if (!subjectAltRequireEmail && !subjectRequireEmail) || schemaVersion == 1 {
					exception = true
				}

				if !subjectAltRequireDNS && !subjectAltRequireDomainDNS {
					tempUserResults.Or(CalculateCrossProductNodeSets(groupExpansions, cache.CertTemplateControllers[publishedCertTemplate.ID], cache.EnterpriseCAEnrollers[enterpriseCA.ID]))
				}

				tempResults.Or(CalculateCrossProductNodeSets(groupExpansions, cache.CertTemplateControllers[publishedCertTemplate.ID], cache.EnterpriseCAEnrollers[enterpriseCA.ID]))
			}
		}

		tempResults.Each(func(value uint32) (bool, error) {
			sourceID := graph.ID(value)

			if resultNode, err := tx.Nodes().Filter(query.Equals(query.NodeID(), sourceID)).First(); err != nil {
				return true, nil
			} else {
				if resultNode.Kinds.ContainsOneOf(ad.Group) {
					results.Add(value)
				} else if resultNode.Kinds.ContainsOneOf(ad.Computer) {
					email, _ := resultNode.Properties.Get(common.Email.String()).String()

					if email != "" || exception {
						results.Add(value)
					}

				}
			}
			return true, nil
		})

		tempUserResults.Each(func(value uint32) (bool, error) {
			sourceID := graph.ID(value)

			if resultNode, err := tx.Nodes().Filter(query.Equals(query.NodeID(), sourceID)).First(); err != nil {
				return true, nil
			} else {
				if resultNode.Kinds.ContainsOneOf(ad.User) {
					email, _ := resultNode.Properties.Get(common.Email.String()).String()
					if email != "" || exception {
						results.CheckedAdd(value)
					}
				}
			}
			return true, nil
		})

		results.Each(func(value uint32) (bool, error) {
			if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC6a,
			}) {
				return false, nil
			} else {
				return true, nil
			}
		})
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
						} else {
							return nil
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
