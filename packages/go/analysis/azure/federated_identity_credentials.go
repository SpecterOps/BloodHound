// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func FederatedIdentityCredentialEntityDetails(ctx context.Context, db graph.Database, validPrimaryKinds graphschema.ValidPrimaryKinds, objectID string, hydrateCounts bool) (FederatedIdentityCredentialDetails, error) {
	var details FederatedIdentityCredentialDetails

	return details, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := FetchEntityByObjectID(tx, objectID); err != nil {
			return err
		} else {
			details.Node = FromGraphNode(validPrimaryKinds, node)
			if appID, err := getFICAppID(tx, node); err != nil {
				return err
			} else {
				details.Properties[azure.FederatedIdentityCredentialAppID.String()] = appID
			}

			return err
		}
	})
}

func getFICAppID(tx graph.Transaction, node *graph.Node) (string, error) {
	var appID string
	if ficApp, err := fetchFICApp(tx, node); err != nil {
		return appID, err
	} else if ficApp.Len() == 0 {
		slog.Warn(fmt.Sprintf("Federated identity credential node %d has no applications attached", node.ID))
	} else {
		app := ficApp.Pick()

		if appID, err = app.Properties.Get(common.ObjectID.String()).String(); err != nil {
			slog.Error(fmt.Sprintf("Failed to marshal the object ID of node %d while fetching the app ID of federated identity credential node %d: %v", app.ID, node.ID, err))
		}
	}

	return appID, nil
}

func fetchFICApp(tx graph.Transaction, federatedIdentityCredential *graph.Node) (graph.NodeSet, error) {
	return ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), federatedIdentityCredential.ID),
			query.Kind(query.Relationship(), azure.AZAuthenticatesTo),
			query.Kind(query.End(), azure.App),
		)
	}))
}
