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

func NewFunctionAppEntityDetails(node *graph.Node) FunctionAppDetails {
	return FunctionAppDetails{
		Node: FromGraphNode(node),
	}
}

func FunctionAppEntityDetails(ctx context.Context, db graph.Database, objectID string, hydrateCounts bool) (FunctionAppDetails, error) {
	var details FunctionAppDetails

	return details, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details = NewFunctionAppEntityDetails(node)
			if hydrateCounts {
				details, err = PopulateFunctionAppEntityDetailsCounts(tx, node, details)
			}
			return err
		}
	})
}

func PopulateFunctionAppEntityDetailsCounts(tx graph.Transaction, node *graph.Node, details FunctionAppDetails) (FunctionAppDetails, error) {

	if inboundObjectControl, err := FetchInboundEntityObjectControllers(tx, node, graph.DirectionInbound, 0, 0); err != nil {
		return details, err
	} else {
		details.InboundObjectControl = inboundObjectControl.Len()
	}

	return details, nil
}
