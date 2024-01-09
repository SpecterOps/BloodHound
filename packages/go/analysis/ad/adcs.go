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

func BuildEsc1Cache(ctx context.Context, db graph.Database, enterpriseCAs, certTemplates []*graph.Node) (map[graph.ID]graph.NodeSet, error) {
	cache := map[graph.ID]graph.NodeSet{}

	return cache, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		for _, ct := range certTemplates {
			if firstDegreePrincipals, err := fetchFirstDegreeNodes(tx, ct, ad.Enroll, ad.GenericAll, ad.AllExtendedRights); err != nil {
				log.Errorf("error fetching enrollers for cert template %d: %w", ct.ID, err)
			} else {
				cache[ct.ID] = firstDegreePrincipals
			}
		}

		for _, eca := range enterpriseCAs {
			if firstDegreeEnrollers, err := fetchFirstDegreeNodes(tx, eca, ad.Enroll); err != nil {
				log.Errorf("error fetching enrollers for enterprise ca %d: %w", eca.ID, err)
			} else {
				cache[eca.ID] = firstDegreeEnrollers
			}
		}

		return nil
	})
}

func fetchFirstDegreeNodes(tx graph.Transaction, targetNode *graph.Node, relKinds ...graph.Kind) (graph.NodeSet, error) {
	return ops.FetchStartNodes(tx.Relationships().Filter(
		query.And(
			query.Kind(query.Start(), ad.Entity),
			query.KindIn(query.Relationship(), relKinds...),
			query.Equals(query.EndID(), targetNode.ID),
		),
	))
}

func PostADCSESC1(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, db graph.Database, expandedGroups impact.PathAggregator, allEnterpriseCAs, allCertTemplates []*graph.Node, enterpriseCA, domain *graph.Node) error {
	results := cardinality.NewBitmap32()

	if enrollCache, err := BuildEsc1Cache(ctx, db, allEnterpriseCAs, allCertTemplates); err != nil {
		return fmt.Errorf("error building cache for esc1: %w", err)
	} else if publishedCertTemplates, err := FetchCertTemplatesPublishedToCA(tx, enterpriseCA); err != nil {
		return fmt.Errorf("error fetching cert templates for ECA %d: %w", enterpriseCA.ID, err)
	} else {
		for _, certTemplate := range publishedCertTemplates.Slice() {
			if validationProperties, err := getValidatePublishedCertTemplateForEsc1PropertyValues(certTemplate); err != nil {
				log.Errorf("error getting published certtemplate validation properties, %w", err)
				continue
			} else if !validatePublishedCertTemplateForEsc1(validationProperties) {
				continue
			} else {
				results.Or(CalculateCrossProductNodeSets(enrollCache[enterpriseCA.ID].Slice(), enrollCache[certTemplate.ID].Slice(), expandedGroups))
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

func PostADCS(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator, adcsEnabled bool) (*analysis.AtomicPostProcessingStats, error) {
	if adcsEnabled {
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

			for _, domain := range domains {
				innerDomain := domain

				operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {

					if enterpriseCAs, err := FetchEnterpriseCAsTrustedForNTAuthToDomain(tx, innerDomain); err != nil {
						return err
					} else {
						for _, enterpriseCA := range enterpriseCAs {
							if validPaths, err := FetchEnterpriseCAsCertChainPathToDomain(tx, enterpriseCA, innerDomain); err != nil {
								log.Errorf("error fetching paths from enterprise ca %d to domain %d: %w", enterpriseCA.ID, innerDomain.ID, err)
							} else if validPaths.Len() == 0 {
								continue
							} else {
								if err := PostGoldenCert(ctx, tx, outC, innerDomain, enterpriseCA); err != nil {
									log.Errorf("failed post processing for %s: %w", ad.GoldenCert.String(), err)
								} else if err := PostADCSESC1(ctx, tx, outC, db, groupExpansions, enterpriseCertAuthorities, certTemplates, enterpriseCA, innerDomain); err != nil {
									log.Errorf("failed post processing for %s: %w", ad.ADCSESC1.String(), err)
								}
							}
						}
					}
					return nil
				})
			}

			return &operation.Stats, operation.Done()
		}
	} else {
		return &analysis.AtomicPostProcessingStats{}, nil
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
		log.Errorf("error fetching host ca computer for enterprise ca %d: %w", enterpriseCA.ID, err)
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
