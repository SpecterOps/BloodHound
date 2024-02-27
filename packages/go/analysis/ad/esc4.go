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
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC4(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	// 1.
	principals := cardinality.NewBitmap32()

	// 2. iterate certtemplates that have an outbound `PublishedTo` edge to eca
	for _, certTemplate := range cache.PublishedTemplateCache[enterpriseCA.ID] {
		// 2c. kick out early if cert template does meet conditions for ESC4
		if valid, err := isCertTemplateValidForESC4(certTemplate); err != nil {
			log.Warnf("error validating cert template %d: %v", certTemplate.ID, err)
			continue
		} else if !valid {
			continue
		} else {
			// 2a.
			principals.Or(CalculateCrossProductNodeSets(
				groupExpansions,
				cache.CertTemplateXXX[certTemplate.ID],
				cache.EnterpriseCAEnrollers[enterpriseCA.ID],
			))
		}
	}
	return nil
}

func isCertTemplateValidForESC4(ct *graph.Node) (bool, error) {
	if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
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
