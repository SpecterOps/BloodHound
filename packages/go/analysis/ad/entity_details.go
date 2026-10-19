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

package ad

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

const (
	serverReferenceComputerNameProperty = "serverreferencecomputername"
	serverReferenceComputerProperty     = "serverreferencecomputer"
	siteServerNodeNameProperty          = "siteservernodename"
	siteServerNodeProperty              = "siteservernode"
)

// ComputerEntityDetails fetches a Computer and decorates it with its linked SiteServer properties.
func ComputerEntityDetails(ctx context.Context, database graph.Database, objectID string) (*graph.Node, error) {
	var (
		computer           *graph.Node
		err                error
		linkedSiteServers  graph.NodeSet
		siteServer         *graph.Node
		siteServerName     string
		siteServerObjectID string
	)

	if err = database.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if computer, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.NodeProperty(common.ObjectID.String()), objectID),
				query.Kind(query.Node(), ad.Computer),
			)
		}).First(); err != nil {
			return err
		}

		if linkedSiteServers, err = ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), ad.SiteServer),
				query.Kind(query.Relationship(), ad.ServerIs),
				query.Equals(query.EndID(), computer.ID),
			)
		}).Limit(1)); err != nil {
			return err
		} else if linkedSiteServers.Len() == 0 {
			return nil
		}

		siteServer = linkedSiteServers.Pick()
		if siteServerObjectID, err = siteServer.Properties.Get(common.ObjectID.String()).String(); err != nil {
			return fmt.Errorf("reading linked site server object ID: %w", err)
		}

		computer.Properties.Set(siteServerNodeProperty, siteServerObjectID)
		if siteServerName, err = siteServer.Properties.Get(common.Name.String()).String(); err == nil {
			computer.Properties.Set(siteServerNodeNameProperty, siteServerName)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return computer, nil
}

// SiteServerEntityDetails fetches a SiteServer and decorates it with its linked Computer properties.
func SiteServerEntityDetails(ctx context.Context, database graph.Database, objectID string) (*graph.Node, error) {
	var (
		computer         *graph.Node
		computerName     string
		computerObjectID string
		err              error
		linkedComputers  graph.NodeSet
		siteServer       *graph.Node
	)

	if err = database.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if siteServer, err = tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.NodeProperty(common.ObjectID.String()), objectID),
				query.Kind(query.Node(), ad.SiteServer),
			)
		}).First(); err != nil {
			return err
		}

		if linkedComputers, err = ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Equals(query.StartID(), siteServer.ID),
				query.Kind(query.Relationship(), ad.ServerIs),
				query.Kind(query.End(), ad.Computer),
			)
		}).Limit(1)); err != nil {
			return err
		} else if linkedComputers.Len() == 0 {
			return nil
		}

		computer = linkedComputers.Pick()
		if computerObjectID, err = computer.Properties.Get(common.ObjectID.String()).String(); err != nil {
			return fmt.Errorf("reading linked computer object ID: %w", err)
		}

		siteServer.Properties.Set(serverReferenceComputerProperty, computerObjectID)
		if computerName, err = computer.Properties.Get(common.Name.String()).String(); err == nil {
			siteServer.Properties.Set(serverReferenceComputerNameProperty, computerName)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return siteServer, nil
}
