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
)

func NewUserEntityDetails(node *graph.Node) UserDetails {
	return UserDetails{
		Node: FromGraphNode(node),
	}
}

func UserEntityDetails(db graph.Database, objectID string, hydrateCounts bool) (UserDetails, error) {
	var details UserDetails

	return details, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewUserEntityDetails(node)
			if hydrateCounts {
				details, err = PopulateUserEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
}

func PopulateUserEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details UserDetails) (UserDetails, error) {

	if roles, err := FetchEntityRoles(tx, node, 0, 0); err != nil {
		return details, err
	} else {
		details.Roles = roles.Len()
	}

	if groupMembership, err := FetchEntityGroupMembershipPaths(tx, node); err != nil {
		return details, err
	} else {
		details.GroupMembership = groupMembership.Len()
	}

	if entityExecutionPrivileges, err := FetchOutboundEntityExecutionPrivileges(tx, node, graph.DirectionOutbound, 0, 0); err != nil {
		return details, err
	} else {
		details.ExecutionPrivileges = entityExecutionPrivileges.Len()
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

	return details, nil
}
