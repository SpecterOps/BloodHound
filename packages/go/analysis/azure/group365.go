// Copyright 2025 Specter Ops, Inc.
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

func NewGroup365EntityDetails(node *graph.Node) Group365Details {

	return Group365Details{
		Node: FromGraphNode(node),
	}
}

func Group365EntityDetails(db graph.Database, objectID string, hydrateCounts bool) (Group365Details, error) {

	var details Group365Details
	return details, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {

		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err

		} else {
			details = NewGroup365EntityDetails(node)

			if hydrateCounts {
				details, err = PopulateGroup365EntityDetailsCounts(tx, node, details)
			}

			return err
		}
	})
}

func PopulateGroup365EntityDetailsCounts(tx graph.Transaction, node *graph.Node, details Group365Details) (Group365Details, error) {

	if inboundObjectControl, err := FetchInboundEntityObjectControllers(tx, node, 0, 0); err != nil {
		return details, err

	} else {
		details.InboundObjectControl = inboundObjectControl.Len()
	}

	return details, nil
}
