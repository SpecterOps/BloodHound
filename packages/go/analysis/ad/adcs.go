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
	"github.com/specterops/bloodhound/analysis"
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

func PostEnrollOnBehalfOf(ctx context.Context, db graph.Database, certTemplates []graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "EnrollOnBehalfOf Post Processing")

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

	return &operation.Stats, operation.Done()
}

func enrollOnBehalfOfVersionTwo(tx graph.Transaction, versionTwoCertTemplates, allCertTemplates []graph.Node, outC chan<- analysis.CreatePostRelationshipJob) error {
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
				if authorizedSignatures, err := certTemplateTwo.Properties.Get(ad.AuthorizedSignatures.String()).Int(); err != nil {
					log.Errorf("Error getting authorized signatures for cert template %d: %v", certTemplateTwo.ID, err)
				} else if authorizedSignatures < 1 {
					continue
				} else if applicationPolicies, err := certTemplateTwo.Properties.Get(ad.ApplicationPolicies.String()).StringSlice(); err != nil {
					log.Errorf("Error getting applicaiton policies for cert template %d: %v", certTemplateTwo.ID, err)
				} else if !slices.Contains(applicationPolicies, EkuCertRequestAgent) {
					continue
				} else if isLinked, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					log.Errorf("error fetch paths from cert template %d to domain: %v", certTemplateTwo.ID, err)
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
				if certTemplateTwo.ID == certTemplateOne.ID {
					continue
				} else if hasPath, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					log.Errorf("Error getting domain node for certtemplate %d: %v", certTemplateTwo.ID, err)
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
			log.Errorf("Error checking ekus for certtemplate %d: %v", certTemplate.ID, err)
		} else if !hasEku {
			continue
		} else if subjectRequireUpn, err := certTemplate.Properties.Get(ad.SubjectAltRequireUPN.String()).Bool(); err != nil {
			log.Errorf("Error getting subjectAltRequireUPN for certtemplate %d: %v", certTemplate.ID, err)
		} else if !subjectRequireUpn {
			continue
		} else if domainNode, err := getDomainForCertTemplate(tx, certTemplate); err != nil {
			log.Errorf("Error getting domain for certtemplate %d: %v", certTemplate.ID, err)
		} else if doesLink, err := DoesCertTemplateLinkToDomain(tx, certTemplate, domainNode); err != nil {
			log.Errorf("Error fetching paths from certtemplate %d to domain: %v", certTemplate.ID, err)
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
	if ekus, err := certTemplate.Properties.Get(ad.EffectiveEKUs.String()).StringSlice(); err != nil {
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
	if ekus, err := certTemplate.Properties.Get(ad.EffectiveEKUs.String()).StringSlice(); err != nil {
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

func PostTrustedForNTAuth(ctx context.Context, db graph.Database, ntAuthStoreNodes []graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "TrustedForNTAuth Post Processing")

	for _, node := range ntAuthStoreNodes {
		innerNode := node

		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if thumbprints, err := node.Properties.Get(ad.CertThumbprint.String()).StringSlice(); err != nil {
				return err
			} else {
				for _, thumbprint := range thumbprints {
					if sourceNodeIDs, err := findMatchingCertChainIDs(thumbprint, tx, ad.EnterpriseCA); err != nil {
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
			return nil
		})
	}

	return &operation.Stats, operation.Done()
}

func PostIssuedSignedBy(ctx context.Context, db graph.Database, enterpriseCertAuthorities []graph.Node, rootCertAuthorities []graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "IssuedSignBy Post Processing")

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

	return &operation.Stats, operation.Done()
}

func processCertChainParent(node graph.Node, tx graph.Transaction) ([]analysis.CreatePostRelationshipJob, error) {
	if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
		return []analysis.CreatePostRelationshipJob{}, err
	} else if len(certChain) > 1 {
		parentCert := certChain[1]
		if targetNodes, err := findMatchingCertChainIDs(parentCert, tx, ad.EnterpriseCA, ad.RootCA); err != nil {
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

func findMatchingCertChainIDs(certThumbprint string, tx graph.Transaction, kinds ...graph.Kind) ([]graph.ID, error) {
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
