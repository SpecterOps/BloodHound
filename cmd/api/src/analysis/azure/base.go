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

	"github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func NewBaseEntityDetails(node *graph.Node) azure.BaseDetails {
	return azure.BaseDetails{
		Node: azure.FromGraphNode(node),
	}
}

func BaseEntityDetails(db graph.Database, objectID string, hydrateCounts bool) (azure.BaseDetails, error) {
	var details azure.BaseDetails

	return details, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if node, err := azure.FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewBaseEntityDetails(node)
			if hydrateCounts {
				details, err = PopulateBaseEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
}

func PopulateBaseEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details azure.BaseDetails) (azure.BaseDetails, error) {

	if outboundObjectControl, err := azure.FetchOutboundEntityObjectControl(tx, node, graph.DirectionOutbound, 0, 0); err != nil {
		return details, err
	} else {
		details.OutboundObjectControl = outboundObjectControl.Len()
	}

	return details, nil
}
