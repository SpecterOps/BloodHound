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
	"github.com/specterops/bloodhound/slices"
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

func PostADCSESC1Domain(tx graph.Transaction, domain *graph.Node, groupExpansions impact.PathAggregator, enrollCache map[graph.ID]graph.NodeSet) (cardinality.Duplex[uint32], error) {
	results := cardinality.NewBitmap32()
	if enterpriseCAs, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      domain,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.TrustedForNTAuth, ad.NTAuthStoreFor)
		},
		DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			depth := segment.Depth()
			if depth == 1 && !segment.Edge.Kind.Is(ad.NTAuthStoreFor) {
				return false
			} else if depth == 2 && !segment.Edge.Kind.Is(ad.TrustedForNTAuth) {
				return false
			} else {
				return true
			}
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
		},
	}); err != nil {
		return results, err
	} else {
		for _, enterpriseCA := range enterpriseCAs {
			if validPaths, err := FetchEnterpriseCAPathToDomain(tx, enterpriseCA, domain); err != nil {
				log.Errorf("error fetching paths from enterprise ca %d to domain %d: %w", enterpriseCA.ID, domain.ID, err)
			} else if validPaths.Len() == 0 {
				continue
			} else if publishedCertTemplates, err := FetchCertTemplatesPublishedToCA(tx, enterpriseCA); err != nil {
				log.Errorf("error fetching cert templates for ECA %d: %w", enterpriseCA.ID, err)
			} else {
				for _, certTemplate := range publishedCertTemplates.Slice() {
					if validationProperties, err := getValidatePublishedCertTemplateForEsc1PropertyValues(certTemplate); err != nil {
						log.Errorf("error getting published certtemplate validation properties, %w", err)
						continue
					} else if !validatePublishedCertTemplateForEsc1(validationProperties) {
						continue
					} else {
						results.Or(CalculateCrossProductNodeSets(enrollCache[enterpriseCA.ID].Slice(), enrollCache[certTemplate.ID].Slice(), groupExpansions))
					}
				}
			}
		}
	}
	return results, nil
}

func PostADCSESC1(ctx context.Context, db graph.Database, enterpriseCas, certTemplates []*graph.Node, expandedGroups impact.PathAggregator, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {
	if enrollCache, err := BuildEsc1Cache(ctx, db, enterpriseCas, certTemplates); err != nil {
		return fmt.Errorf("error building cache for esc1: %w", err)
	} else if domains, err := FetchNodesByKind(ctx, db, ad.Domain); err != nil {
		return fmt.Errorf("error getting domains for esc1: %w", err)
	} else {
		for _, domain := range domains {
			innerDomain := domain
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if results, err := PostADCSESC1Domain(tx, innerDomain, expandedGroups, enrollCache); err != nil {
					return err
				} else {
					results.Each(func(value uint32) (bool, error) {
						outC <- analysis.CreatePostRelationshipJob{
							FromID: graph.ID(value),
							ToID:   innerDomain.ID,
							Kind:   ad.ADCSESC1,
						}

						return true, nil
					})
				}

				return nil
			})
		}
	}

	return nil
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

func PostEnrollOnBehalfOf(certTemplates []graph.Node, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {
	versionOneTemplates := make([]graph.Node, 0)
	versionTwoTemplates := make([]graph.Node, 0)

	for _, node := range certTemplates {
		if version, err := node.Properties.Get(ad.SchemaVersion.String()).Int(); err != nil {
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
		return enrollOnBehalfOfVersionTwo(tx, versionTwoTemplates, certTemplates, outC)
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		return enrollOnBehalfOfVersionOne(tx, versionOneTemplates, certTemplates, outC)
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		return enrollOnBehalfOfSelfControl(tx, versionOneTemplates, outC)
	})

	return nil
}

func enrollOnBehalfOfVersionTwo(tx graph.Transaction, versionTwoCertTemplates, allCertTemplates []graph.Node, outC chan<- analysis.CreatePostRelationshipJob) error {
	for _, certTemplateOne := range allCertTemplates {
		if hasBadEku, err := certTemplateHasEku(certTemplateOne, EkuAnyPurpose); err != nil {
			log.Errorf("error getting ekus for cert template %d: %w", certTemplateOne.ID, err)
		} else if hasBadEku {
			continue
		} else if hasEku, err := certTemplateHasEku(certTemplateOne, EkuCertRequestAgent); err != nil {
			log.Errorf("error getting ekus for cert template %d: %w", certTemplateOne.ID, err)
		} else if !hasEku {
			continue
		} else if domainNode, err := getDomainForCertTemplate(tx, certTemplateOne); err != nil {
			log.Errorf("error getting domain node for cert template %d: %w", certTemplateOne.ID, err)
		} else if isLinked, err := DoesCertTemplateLinkToDomain(tx, certTemplateOne, domainNode); err != nil {
			log.Errorf("error fetching paths from cert template %d to domain: %w", certTemplateOne.ID, err)
		} else if !isLinked {
			continue
		} else {
			for _, certTemplateTwo := range versionTwoCertTemplates {
				if authorizedSignatures, err := certTemplateTwo.Properties.Get(ad.AuthorizedSignatures.String()).Int(); err != nil {
					log.Errorf("Error getting authorized signatures for cert template %d: %w", certTemplateTwo.ID, err)
				} else if authorizedSignatures < 1 {
					continue
				} else if applicationPolicies, err := certTemplateTwo.Properties.Get(ad.ApplicationPolicies.String()).StringSlice(); err != nil {
					log.Errorf("Error getting application policies for cert template %d: %w", certTemplateTwo.ID, err)
				} else if !slices.Contains(applicationPolicies, EkuCertRequestAgent) {
					continue
				} else if isLinked, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					log.Errorf("error fetch paths from cert template %d to domain: %w", certTemplateTwo.ID, err)
				} else if !isLinked {
					continue
				} else {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: certTemplateOne.ID,
						ToID:   certTemplateTwo.ID,
						Kind:   ad.EnrollOnBehalfOf,
					}
				}
			}
		}
	}

	return nil
}

func enrollOnBehalfOfVersionOne(tx graph.Transaction, versionOneCertTemplates []graph.Node, allCertTemplates []graph.Node, outC chan<- analysis.CreatePostRelationshipJob) error {
	for _, certTemplateOne := range allCertTemplates {
		//prefilter as much as we can first
		if hasEku, err := certTemplateHasEkuOrAll(certTemplateOne, EkuCertRequestAgent, EkuAnyPurpose); err != nil {
			log.Errorf("Error checking ekus for certtemplate %d: %w", certTemplateOne.ID, err)
		} else if !hasEku {
			continue
		} else if domainNode, err := getDomainForCertTemplate(tx, certTemplateOne); err != nil {
			log.Errorf("Error getting domain node for certtemplate %d: %w", certTemplateOne.ID, err)
		} else if hasPath, err := DoesCertTemplateLinkToDomain(tx, certTemplateOne, domainNode); err != nil {
			log.Errorf("Error fetching paths from certtemplate %d to domain: %w", certTemplateOne.ID, err)
		} else if !hasPath {
			continue
		} else {
			for _, certTemplateTwo := range versionOneCertTemplates {
				if certTemplateTwo.ID == certTemplateOne.ID {
					continue
				} else if hasPath, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					log.Errorf("Error getting domain node for certtemplate %d: %w", certTemplateTwo.ID, err)
				} else if !hasPath {
					continue
				} else {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: 0,
						ToID:   0,
						Kind:   nil,
					}
				}
			}
		}
	}

	return nil
}

func getDomainForCertTemplate(tx graph.Transaction, certTemplate graph.Node) (graph.Node, error) {
	if domainSid, err := certTemplate.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return graph.Node{}, err
	} else if domainNode, err := analysis.FetchNodeByObjectID(tx, domainSid); err != nil {
		return graph.Node{}, err
	} else {
		return *domainNode, nil
	}
}

func enrollOnBehalfOfSelfControl(tx graph.Transaction, versionOneCertTemplates []graph.Node, outC chan<- analysis.CreatePostRelationshipJob) error {
	for _, certTemplate := range versionOneCertTemplates {
		if hasEku, err := certTemplateHasEkuOrAll(certTemplate, EkuAnyPurpose); err != nil {
			log.Errorf("Error checking ekus for certtemplate %d: %w", certTemplate.ID, err)
		} else if !hasEku {
			continue
		} else if subjectRequireUpn, err := certTemplate.Properties.Get(ad.SubjectAltRequireUPN.String()).Bool(); err != nil {
			log.Errorf("Error getting subjectAltRequireUPN for certtemplate %d: %w", certTemplate.ID, err)
		} else if !subjectRequireUpn {
			continue
		} else if domainNode, err := getDomainForCertTemplate(tx, certTemplate); err != nil {
			log.Errorf("Error getting domain for certtemplate %d: %w", certTemplate.ID, err)
		} else if doesLink, err := DoesCertTemplateLinkToDomain(tx, certTemplate, domainNode); err != nil {
			log.Errorf("Error fetching paths from certtemplate %d to domain: %w", certTemplate.ID, err)
		} else if !doesLink {
			continue
		} else {
			outC <- analysis.CreatePostRelationshipJob{
				FromID: certTemplate.ID,
				ToID:   certTemplate.ID,
				Kind:   ad.EnrollOnBehalfOf,
			}
		}
	}

	return nil
}

func certTemplateHasEkuOrAll(certTemplate graph.Node, targetEkus ...string) (bool, error) {
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

func certTemplateHasEku(certTemplate graph.Node, targetEkus ...string) (bool, error) {
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

func PostADCS(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing")

	if enterpriseCertAuthorities, err := FetchNodesByKind(ctx, db, ad.EnterpriseCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching enterpriseCA nodes: %w", err)
	} else if rootCertAuthorities, err := FetchNodesByKind(ctx, db, ad.RootCA); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching rootCA nodes: %w", err)
	} else if certTemplates, err := FetchNodesByKind(ctx, db, ad.CertTemplate); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching cert template nodes: %w", err)
	} else if preProcessStats, err := postADCSPreProcess(ctx, db, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed adcs pre-processing: %w", err)
	} else {
		operation.Stats.Merge(preProcessStats)
		if err := PostADCSESC1(ctx, db, enterpriseCertAuthorities, certTemplates, groupExpansions, operation); err != nil {
			return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.ADCSESC1.String(), err)
		}

		return &operation.Stats, operation.Done()
	}
}

func postADCSPreProcess(ctx context.Context, db graph.Database, enterpriseCertAuthorities, rootCertAuthorities []*graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "ADCS Post Processing - No Dependencies")
	if err := PostTrustedForNTAuth(ctx, db, operation); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.TrustedForNTAuth.String(), err)
	} else if err := PostIssuedSignedBy(ctx, db, operation, enterpriseCertAuthorities, rootCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.IssuedSignedBy.String(), err)
	} else if err := PostEnterpriseCAFor(ctx, db, operation, enterpriseCertAuthorities); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed post processing for %s: %w", ad.EnterpriseCAFor.String(), err)
	} else {
		return &operation.Stats, operation.Done()
	}
}

func PostTrustedForNTAuth(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {

	if ntAuthStoreNodes, err := FetchNodesByKind(ctx, db, ad.NTAuthStore); err != nil {
		return err
	} else {
		for _, node := range ntAuthStoreNodes {
			innerNode := node

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if thumbprints, err := node.Properties.Get(ad.CertThumbprints.String()).StringSlice(); err != nil {
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
