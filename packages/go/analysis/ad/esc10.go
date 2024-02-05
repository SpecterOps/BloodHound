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

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC10a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	if canAbuseUPNRels, err := FetchCanAbuseUPNCertMappingRels(tx, eca); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseUPNRels) == 0 {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[eca.ID]; !ok {
		return nil
	} else if ecaControllers, ok := cache.EnterpriseCAEnrollers[eca.ID]; !ok {
		return nil
	} else {
		results := cardinality.NewBitmap32()

		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC10(template, false); err != nil {
				log.Warnf("error validating cert template %d: %v", template.ID, err)
				continue
			} else if !valid {
				continue
			} else if certTemplateControllers, ok := cache.CertTemplateControllers[template.ID]; !ok {
				log.Debugf("Failed to retrieve controllers for cert template %d from cache", template.ID)
				continue
			} else {
				victimBitmap := getVictimBitmap(groupExpansions, certTemplateControllers, ecaControllers)

				if filteredVictims, err := filterUserDNSResults(tx, victimBitmap, template); err != nil {
					log.Warnf("error filtering users from victims for esc9a: %v", err)
					continue
				} else if attackers, err := FetchAttackersForEscalations9and10(tx, filteredVictims, false); err != nil {
					log.Warnf("Error getting start nodes for esc10a attacker nodes: %v", err)
					continue
				} else {
					results.Or(cardinality.NodeIDsToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint32) bool {
			channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC10a,
			})
			return true
		})
	}
	return nil
}

func PostADCSESC10b(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	if canAbuseUPNRels, err := FetchCanAbuseUPNCertMappingRels(tx, enterpriseCA); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseUPNRels) == 0 {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		return nil
	} else if ecaControllers, ok := cache.EnterpriseCAEnrollers[enterpriseCA.ID]; !ok {
		return nil
	} else {
		results := cardinality.NewBitmap32()

		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC10(template, true); err != nil {
				log.Warnf("error validating cert template %d: %v", template.ID, err)
				continue
			} else if !valid {
				continue
			} else if certTemplateControllers, ok := cache.CertTemplateControllers[template.ID]; !ok {
				log.Debugf("Failed to retrieve controllers for cert template %d from cache", template.ID)
				continue
			} else {
				victimBitmap := getVictimBitmap(groupExpansions, certTemplateControllers, ecaControllers)

				if attackers, err := FetchAttackersForEscalations9and10(tx, victimBitmap, true); err != nil {
					log.Warnf("Error getting start nodes for esc10b attacker nodes: %v", err)
					continue
				} else {
					results.Or(cardinality.NodeIDsToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint32) bool {
			channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC10b,
			})
			return true
		})
	}
	return nil
}

func isCertTemplateValidForESC10(ct *graph.Node, scenarioB bool) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if enrolleeSuppliesSubject, err := ct.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
		return false, err
	} else if enrolleeSuppliesSubject {
		return false, nil
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return false, err
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else if !scenarioB {
		if subjectAltRequireUPN, err := ct.Properties.Get(ad.SubjectAltRequireUPN.String()).Bool(); err != nil {
			return false, err
		} else if subjectAltRequireSPN, err := ct.Properties.Get(ad.SubjectAltRequireSPN.String()).Bool(); err != nil {
			return false, err
		} else if subjectAltRequireSPN || subjectAltRequireUPN {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		if subjectAltRequireDNS, err := ct.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
			return false, err
		} else if subjectAltRequireDNS {
			return true, nil
		} else {
			return false, nil
		}
	}
}
