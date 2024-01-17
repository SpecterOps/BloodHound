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
	analysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func TierZeroWellKnownSIDSuffixes() []string {
	return []string{
		analysis.EnterpriseDomainControllersGroupSIDSuffix,
		analysis.AdministratorAccountSIDSuffix,
		analysis.DomainAdminsGroupSIDSuffix,
		analysis.DomainControllersGroupSIDSuffix,
		analysis.SchemaAdminsGroupSIDSuffix,
		analysis.EnterpriseAdminsGroupSIDSuffix,
		analysis.KeyAdminsGroupSIDSuffix,
		analysis.EnterpriseKeyAdminsGroupSIDSuffix,
		analysis.BackupOperatorsGroupSIDSuffix,
		analysis.AdministratorsGroupSIDSuffix,
	}
}

func FetchWellKnownTierZeroEntities(tx graph.Transaction, domainSID string) (graph.NodeSet, error) {
	nodes := graph.NewNodeSet()

	for _, wellKnownSIDSuffix := range TierZeroWellKnownSIDSuffixes() {
		if err := tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				// Make sure we have the Group or User label. This should cover the case for URA as well as filter out all the other localgroups
				query.KindIn(query.Node(), ad.Group, ad.User),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellKnownSIDSuffix),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			)
		}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				nodes.Add(node)
			}

			return cursor.Error()
		}); err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

func FetchAllGroupMembers(tx graph.Transaction, targets graph.NodeSet) (graph.NodeSet, error) {
	allGroupMembers := graph.NewNodeSet()

	for _, target := range targets {
		if target.Kinds.ContainsOneOf(ad.Group) {
			if groupMembers, err := analysis.FetchGroupMembers(tx, target, 0, 0); err != nil {
				return nil, err
			} else {
				allGroupMembers.AddSet(groupMembers)
			}
		}
	}

	return allGroupMembers, nil
}

func FetchDomainTierZeroAssets(tx graph.Transaction, domain *graph.Node) (graph.NodeSet, error) {
	domainSID, _ := domain.Properties.GetOrDefault(ad.DomainSID.String(), "").String()

	return ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Entity),
			query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero),
		)
	}))
}
