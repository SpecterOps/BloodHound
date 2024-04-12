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
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC13(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca *graph.Node, cache ADCSCache) error {
	if publishedCertTemplates, ok := cache.PublishedTemplateCache[eca.ID]; !ok {
		return nil
	} else {
		for _, template := range publishedCertTemplates {
			if isValid, err := isCertTemplateValidForESC13(template); err != nil {
				log.Errorf("Error checking esc13 cert template: %v", err)
			} else if !isValid {
				continue
			} else if groupNodes, err := getCertTemplateGroupLinks(template, tx); err != nil {
				log.Errorf("Error getting cert template group links: %v", err)
			} else if len(groupNodes) == 0 {
				continue
			} else {
				controlBitmap := CalculateCrossProductNodeSets(groupExpansions, cache.CertTemplateEnrollers[template.ID], cache.EnterpriseCAEnrollers[eca.ID])
				if filtered, err := filterUserDNSResults(tx, controlBitmap, template); err != nil {
					log.Warnf("Error filtering users from victims for esc13: %v", err)
					continue
				} else {
					for _, groupID := range groupNodes.IDs() {
						filtered.Each(func(value uint32) bool {
							channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
								FromID: graph.ID(value),
								ToID:   groupID,
								Kind:   ad.ADCSESC13,
							})
							return true
						})
					}
				}
			}
		}

		return nil
	}
}

func isCertTemplateValidForESC13(ct *graph.Node) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
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

func getCertTemplateGroupLinks(ct *graph.Node, tx graph.Transaction) (graph.NodeSet, error) {
	if policyNodes, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.Start(), ct.ID),
			query.KindIn(query.Relationship(), ad.ExtendedByPolicy),
			query.KindIn(query.End(), ad.IssuancePolicy),
		)
	})); err != nil {
		return graph.NodeSet{}, err
	} else if len(policyNodes) == 0 {
		return graph.NodeSet{}, nil
	} else {
		policyNodeIDs := policyNodes.IDs()
		if groupNodes, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.Start(), policyNodeIDs...),
				query.KindIn(query.Relationship(), ad.OIDGroupLink),
				query.KindIn(query.End(), ad.Group),
			)
		})); err != nil {
			return graph.NodeSet{}, err
		} else {
			return groupNodes, nil
		}
	}
}
