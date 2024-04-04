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
	"github.com/specterops/bloodhound/log"

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

	if inboundObjectControl, err := FetchInboundEntityObjectControllers(tx, node, 0, 0); err != nil {
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

func FetchTenants(ctx context.Context, db graph.Database) (graph.NodeSet, error) {
	var nodeSet graph.NodeSet
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if nodeSet, err = ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), azure.Tenant)
		})); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	} else {
		return nodeSet, nil
	}
}

// TenantRoles returns the NodeSet of roles for a given tenant that match one of the given role template IDs. If no role template ID is provided, then all of the tenant role nodes are returned in the NodeSet.
func TenantRoles(tx graph.Transaction, tenant *graph.Node, roleTemplateIDs ...string) (graph.NodeSet, error) {
	defer log.LogAndMeasure(log.LevelInfo, "Tenant %d TenantRoles", tenant.ID)()

	if !IsTenantNode(tenant) {
		return nil, fmt.Errorf("cannot fetch tenant roles - node %d must be of kind %s", tenant.ID, azure.Tenant)
	}

	conditions := []graph.Criteria{
		query.Equals(query.StartID(), tenant.ID),
		query.Kind(query.Relationship(), azure.Contains),
		query.Kind(query.End(), azure.Role),
	}

	if len(roleTemplateIDs) > 0 {
		conditions = append(conditions, query.In(query.EndProperty(azure.RoleTemplateID.String()), roleTemplateIDs))
	}

	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(conditions...)
	}))
}

// TenantApplicationsAndServicePrincipals returns the complete set of application and service principal nodes contained by the given Tenant node
func TenantApplicationsAndServicePrincipals(tx graph.Transaction, tenant *graph.Node) (graph.NodeSet, error) {
	if !IsTenantNode(tenant) {
		return nil, fmt.Errorf("node %d must contain kind %s", tenant.ID, azure.Tenant)
	} else {
		return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.StartID(), tenant.ID),
				query.Kind(query.Relationship(), azure.Contains),
				query.KindIn(query.End(), azure.App, azure.ServicePrincipal),
			)
		}))
	}
}

func IsTenantNode(node *graph.Node) bool {
	return node.Kinds.ContainsOneOf(azure.Tenant)
}
