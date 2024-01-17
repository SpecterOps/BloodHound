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

func NewApplicationDetails(node *graph.Node) ApplicationDetails {
	return ApplicationDetails{
		Node: FromGraphNode(node),
	}
}

func ApplicationEntityDetails(ctx context.Context, db graph.Database, objectID string, hydrateCounts bool) (ApplicationDetails, error) {
	var details ApplicationDetails

	return details, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewApplicationDetails(node)
			if servicePrincipalID, err := getAppServicePrincipalID(tx, node); err != nil {
				return err
			} else {
				details.Properties[azure.ServicePrincipalID.String()] = servicePrincipalID
			}
			if hydrateCounts {
				if details, err = PopulateApplicationEntityDetailsCounts(tx, node, details); err != nil {
					return err
				}
			}
			return err
		}
	})
}

func getAppServicePrincipalID(tx graph.Transaction, node *graph.Node) (string, error) {
	var servicePrincipalID string
	if appServicePrincipals, err := FetchApplicationServicePrincipals(tx, node); err != nil {
		return "", err
	} else if appServicePrincipals.Len() == 0 {
		// Don't want this to break the function, but we'll want to know about it
		log.Errorf("Application node %d has no service principals attached", node.ID)
	} else {
		servicePrincipal := appServicePrincipals.Pick()

		if servicePrincipalID, err = servicePrincipal.Properties.Get(common.ObjectID.String()).String(); err != nil {
			log.Errorf("Failed to marshal the object ID of node %d while fetching the service principal ID of application node %d: %v", servicePrincipal.ID, node.ID, err)
		}
	}
	return servicePrincipalID, nil
}

func PopulateApplicationEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details ApplicationDetails) (ApplicationDetails, error) {

	if inboundObjectControl, err := FetchInboundEntityObjectControllers(tx, node, graph.DirectionInbound, 0, 0); err != nil {
		return details, err
	} else {
		details.InboundObjectControl = inboundObjectControl.Len()
	}

	return details, nil
}
