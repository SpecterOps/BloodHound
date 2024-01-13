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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/slices"
)

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

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if results, err := EnrollOnBehalfOfSelfControl(tx, versionOneTemplates); err != nil {
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
				if certTemplateOne.ID == certTemplateTwo.ID {
					continue
				} else if authorizedSignatures, err := certTemplateTwo.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
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

func EnrollOnBehalfOfVersionOne(tx graph.Transaction, versionOneCertTemplates []*graph.Node, allCertTemplates []*graph.Node) ([]analysis.CreatePostRelationshipJob, error) {
	results := make([]analysis.CreatePostRelationshipJob, 0)

	for _, certTemplateOne := range allCertTemplates {
		//prefilter as much as we can first
		if slices.Contains(versionOneCertTemplates, certTemplateOne) {
			continue
		} else if hasEku, err := certTemplateHasEkuOrAll(certTemplateOne, EkuCertRequestAgent, EkuAnyPurpose); err != nil {
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

func getDomainForCertTemplate(tx graph.Transaction, certTemplate *graph.Node) (*graph.Node, error) {
	if domainSid, err := certTemplate.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return &graph.Node{}, err
	} else if domainNode, err := analysis.FetchNodeByObjectID(tx, domainSid); err != nil {
		return &graph.Node{}, err
	} else {
		return domainNode, nil
	}
}

func EnrollOnBehalfOfSelfControl(tx graph.Transaction, versionOneCertTemplates []*graph.Node) ([]analysis.CreatePostRelationshipJob, error) {
	results := make([]analysis.CreatePostRelationshipJob, 0)
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
			results = append(results, analysis.CreatePostRelationshipJob{
				FromID: certTemplate.ID,
				ToID:   certTemplate.ID,
				Kind:   ad.EnrollOnBehalfOf,
			})
		}
	}

	return results, nil
}
