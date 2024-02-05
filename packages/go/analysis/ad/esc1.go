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

func PostADCSESC1(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, expandedGroups impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap32()
	if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		return nil
	} else {
		for _, certTemplate := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForEsc1(certTemplate); err != nil {
				log.Warnf("error validating cert template %d: %v", certTemplate.ID, err)
				continue
			} else if !valid {
				continue
			} else {
				results.Or(CalculateCrossProductNodeSets(expandedGroups, cache.CertTemplateControllers[certTemplate.ID], cache.EnterpriseCAEnrollers[enterpriseCA.ID]))
			}
		}
	}

	results.Each(func(value uint32) bool {
		channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: graph.ID(value),
			ToID:   domain.ID,
			Kind:   ad.ADCSESC1,
		})
		return true
	})
	return nil
}

func isCertTemplateValidForEsc1(ct *graph.Node) (bool, error) {
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
