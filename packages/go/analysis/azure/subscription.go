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
)

func NewSubscriptionEntityDetails(node *graph.Node) SubscriptionDetails {
	return SubscriptionDetails{
		Node: FromGraphNode(node),
	}
}

func SubscriptionEntityDetails(ctx context.Context, db graph.Database, objectID string, hydrateCounts bool) (SubscriptionDetails, error) {
	var details SubscriptionDetails

	return details, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewSubscriptionEntityDetails(node)
			if hydrateCounts {
				details, err = PopulateSubscriptionEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
}

func PopulateSubscriptionEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details SubscriptionDetails) (SubscriptionDetails, error) {
	var descendentKinds = GetDescendentKinds(azure.Subscription)

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
