// Copyright 2026 Specter Ops, Inc.
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

package queries

import (
	"context"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

const (
	serverReferenceComputerNameProperty = "serverreferencecomputername"
	serverReferenceComputerProperty     = "serverreferencecomputer"
	siteServerNodeNameProperty          = "siteservernodename"
	siteServerNodeProperty              = "siteservernode"
)

// GetADEntityDetails fetches an AD entity and decorates it with properties from its linked ServerIs entity.
func (s *GraphQuery) GetADEntityDetails(ctx context.Context, objectID string, entityType graph.Kind) (*graph.Node, error) {
	var (
		err  error
		node *graph.Node
	)

	if err = s.Graph.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err = getEntityByObjectID(tx, objectID, entityType); err != nil {
			return err
		}

		return addServerIsLinkedProperties(tx, node, entityType)
	}); err != nil {
		return nil, err
	}

	return node, nil
}

func addServerIsLinkedProperties(tx graph.Transaction, node *graph.Node, entityType graph.Kind) error {
	if entityType.Is(ad.SiteServer) {
		if linkedComputer, err := getServerIsLinkedEntity(tx, node, graph.DirectionOutbound, ad.Computer); err != nil {
			return err
		} else if linkedComputer != nil {
			node.Properties.Set(serverReferenceComputerProperty, linkedComputer.Properties.Get(common.ObjectID.String()).Any())
			node.Properties.Set(serverReferenceComputerNameProperty, linkedComputer.Properties.Get(common.Name.String()).Any())
		}
	} else if entityType.Is(ad.Computer) {
		if linkedSiteServer, err := getServerIsLinkedEntity(tx, node, graph.DirectionInbound, ad.SiteServer); err != nil {
			return err
		} else if linkedSiteServer != nil {
			node.Properties.Set(siteServerNodeProperty, linkedSiteServer.Properties.Get(common.ObjectID.String()).Any())
			node.Properties.Set(siteServerNodeNameProperty, linkedSiteServer.Properties.Get(common.Name.String()).Any())
		}
	}

	return nil
}

func getServerIsLinkedEntity(tx graph.Transaction, node *graph.Node, direction graph.Direction, relatedKind graph.Kind) (*graph.Node, error) {
	var (
		err               error
		linkedNode        *graph.Node
		relationshipQuery graph.RelationshipQuery
	)

	relationshipQuery = tx.Relationships().Filterf(func() graph.Criteria {
		if direction == graph.DirectionInbound {
			return query.And(
				query.Kind(query.Start(), relatedKind),
				query.Kind(query.Relationship(), ad.ServerIs),
				query.Equals(query.EndID(), node.ID),
			)
		}

		return query.And(
			query.Equals(query.StartID(), node.ID),
			query.Kind(query.Relationship(), ad.ServerIs),
			query.Kind(query.End(), relatedKind),
		)
	})

	err = relationshipQuery.Limit(1).FetchDirection(direction.Reverse(), func(cursor graph.Cursor[graph.DirectionalResult]) error {
		for result := range cursor.Chan() {
			linkedNode = result.Node
		}

		return cursor.Error()
	})

	return linkedNode, err
}
