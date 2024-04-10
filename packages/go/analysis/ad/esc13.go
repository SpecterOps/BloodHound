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
)

func PostADCSESC13(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return err
	} else if allIssuancePolicies, err := fetchAllIssuancePolicies(tx, domainSid); err != nil {
		return err
	} else {
		certTemplates := cache.PublishedTemplateCache[enterpriseCA.ID]

		// Get an O(1) lookup of Issuance Policies keyed by CertificatePolicyOID
		certTemplateOIDToIssuancePolicyMap := getIssuancePolicyCertOIDMap(allIssuancePolicies)

		// For each certTemplate, find all issuance policies within its CertificatePolicy property array
		// such that IssuancePolicy.CertificatePolicyOID is in CertificateTemplate.CertificatePolicy
		// and shares its domain
		for _, certTemplate := range certTemplates {
			if certPolicies, err := certTemplate.Properties.Get(ad.CertificatePolicy.String()).StringSlice(); err != nil {
				continue
			} else {
				for _, policy := range certPolicies {
					for _, issuancePolicy := range certTemplateOIDToIssuancePolicyMap[policy] {
						// Create ExtendedByPolicy edge
						channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
							FromID: issuancePolicy.ID,
							ToID:   certTemplate.ID,
							Kind:   ad.ExtendedByPolicy,
						})
					}
				}
			}
		}

		return nil
	}
}

func fetchAllIssuancePolicies(tx graph.Transaction, domainSid string) (graph.NodeSet, error) {
	if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(
		func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.IssuancePolicy),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
			)
		},
	)); err != nil {
		return nil, err
	} else {
		set := make(graph.NodeSet)
		set.Add(nodes...)
		return set, nil
	}
}

func getIssuancePolicyCertOIDMap(issuancePolicies graph.NodeSet) map[string][]graph.Node {
	oidMap := make(map[string][]graph.Node)

	for _, policy := range issuancePolicies {
		if certTemplateOID, err := policy.Properties.Get(ad.CertTemplateOID.String()).String(); err != nil {
			continue
		} else {
			oidMap[certTemplateOID] = append(oidMap[certTemplateOID], *policy)
		}
	}

	return oidMap
}
