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
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC7(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	if certTemplates, ok := cache.DomainCertTemplates[domain.ID]; !ok {
		return nil
	} else if firstDegreeCAManagers, err := fetchFirstDegreeNodes(tx, eca, ad.ManageCA); err != nil {
		log.Errorf("Error fetching CA managers for enterprise ca %d: %v", eca.ID, err)
		return nil
	} else {
		var (
			results            = cardinality.NewBitmap32()
			validCertTemplates []*graph.Node
		)

		for _, certTemplate := range certTemplates {
			if valid, err := isCertTemplateValidForESC7a(certTemplate, domain); err != nil {
				log.Warnf("Error validating cert template %d: %v", certTemplate.ID, err)
				continue
			} else if valid {
				validCertTemplates = append(validCertTemplates, certTemplate)
			}
		}

		roleSeparationEnabled, err := eca.Properties.Get(ad.RoleSeparationEnabled.String()).Bool()
		if err != nil || !roleSeparationEnabled {
			if len(validCertTemplates) > 0 {
				for _, principal := range firstDegreeCAManagers {
					results.Add(principal.ID.Uint32())
				}
			}
		} else {
			for _, validCertTemplate := range validCertTemplates {
				for _, enroller := range cache.CertTemplateEnrollers[validCertTemplate.ID] {
					results.Or(CalculateCrossProductNodeSets(groupExpansions, graph.NewNodeSet(enroller).Slice(), firstDegreeCAManagers.Slice()))
				}
			}
		}

		results.Each(func(value uint32) bool {
			channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC7,
			})
			return true
		})
		return nil
	}
	return nil
}

func isCertTemplateValidForESC7a(ct *graph.Node, d *graph.Node) (bool, error) {
	if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if enrolleeSuppliesSubject, err := ct.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
		return false, err
	} else if !enrolleeSuppliesSubject {
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
