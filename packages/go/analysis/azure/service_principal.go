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

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func NewServicePrincipalEntityDetails(node *graph.Node) ServicePrincipalDetails {
	return ServicePrincipalDetails{
		Node: FromGraphNode(node),
	}
}

func ServicePrincipalEntityDetails(ctx context.Context, db graph.Database, objectID string, hydrateCounts bool) (ServicePrincipalDetails, error) {
	var details ServicePrincipalDetails

	return details, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewServicePrincipalEntityDetails(node)
			if appID, err := getServicePrincipalAppID(tx, node); err != nil {
				return err
			} else {
				details.Properties[azure.AppID.String()] = appID
			}
			if hydrateCounts {
				details, err = servicePrincipalEntityDetails(tx, node, details)
			}
			return err
		}
	})
}

func getServicePrincipalAppID(tx graph.Transaction, node *graph.Node) (string, error) {
	var appID string
	if servicePrincipalApps, err := FetchServicePrincipalApplications(tx, node); err != nil {
		return appID, err
	} else if servicePrincipalApps.Len() == 0 {
		// Don't want this to break the function, but we'll want to know about it
		log.Warnf("Service principal node %d has no applications attached", node.ID)
	} else {
		app := servicePrincipalApps.Pick()

		if appID, err = app.Properties.Get(common.ObjectID.String()).String(); err != nil {
			log.Errorf("Failed to marshal the object ID of node %d while fetching the service principal ID of application node %d: %v", app.ID, node.ID, err)
		}
	}
	return appID, nil
}

func servicePrincipalEntityDetails(tx graph.Transaction, node *graph.Node, details ServicePrincipalDetails) (ServicePrincipalDetails, error) {

	if roles, err := FetchEntityRoles(tx, node, 0, 0); err != nil {
		return details, err
	} else {
		details.Roles = roles.Len()
	}

	if inboundObjectControl, err := FetchInboundEntityObjectControllers(tx, node, graph.DirectionInbound, 0, 0); err != nil {
		return details, err
	} else {
		details.InboundObjectControl = inboundObjectControl.Len()
	}

	if outboundObjectControl, err := FetchOutboundEntityObjectControl(tx, node, graph.DirectionOutbound, 0, 0); err != nil {
		return details, err
	} else {
		details.OutboundObjectControl = outboundObjectControl.Len()
	}

	if inboundAbusableAppRoleAssignments, err := FetchAbusableAppRoleAssignments(tx, node, graph.DirectionInbound, 0, 0); err != nil {
		return details, err
	} else {
		details.InboundAbusableAppRoleAssignments = inboundAbusableAppRoleAssignments.Len()
	}

	if outboundAbusableAppRoleAssignments, err := FetchAbusableAppRoleAssignments(tx, node, graph.DirectionOutbound, 0, 0); err != nil {
		return details, err
	} else {
		details.OutboundAbusableAppRoleAssignments = outboundAbusableAppRoleAssignments.Len()
	}

	return details, nil
}
