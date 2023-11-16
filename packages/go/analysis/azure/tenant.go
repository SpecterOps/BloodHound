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

package azure

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func NewTenantEntityDetails(node *graph.Node) TenantDetails {
	return TenantDetails{
		Node: FromGraphNode(node),
	}
}

func TenantEntityDetails(db graph.Database, objectID string, hydrateCounts bool) (TenantDetails, error) {
	var details TenantDetails

	return details, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewTenantEntityDetails(node)
			if hydrateCounts {
				details, err = PopulateTenantEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
}

func PopulateTenantEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details TenantDetails) (TenantDetails, error) {
	descendentKinds := GetDescendentKinds(azure.Tenant)

	if descendents, err := FetchEntityDescendentCounts(tx, node, 0, 0, descendentKinds...); err != nil {
		return details, err
	} else {
		details.Descendents = descendents
	}

	if inboundObjectControl, err := FetchInboundEntityObjectControllers(tx, node, graph.DirectionInbound, 0, 0); err != nil {
		return details, err
	} else {
		details.InboundObjectControl = inboundObjectControl.Len()
	}

	return details, nil
}

// TenantPrincipals returns the complete set of User, Group, and Service Principal nodes contained by the given Tenant node
func TenantPrincipals(tx graph.Transaction, tenant *graph.Node) (graph.NodeSet, error) {
	if !IsTenantNode(tenant) {
		return nil, fmt.Errorf("node %d must contain kind %s", tenant.ID, azure.Tenant)
	} else {
		return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.StartID(), tenant.ID),
				query.Kind(query.Relationship(), azure.Contains),
				query.KindIn(query.End(), azure.User, azure.Group, azure.ServicePrincipal),
			)
		}))
	}
}

func IsTenantNode(node *graph.Node) bool {
	return node.Kinds.ContainsOneOf(azure.Tenant)
}
