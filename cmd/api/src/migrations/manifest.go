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

package migrations

import (
	"context"
	"strings"

	"github.com/specterops/bloodhound/src/version"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func Version_277_Migration(db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Migrating node property casing")()

	return db.BatchOperation(context.Background(), func(batch graph.Batch) error {
		if err := batch.Nodes().Filterf(func() graph.Criteria {
			return query.KindIn(query.Node(), ad.Entity, azure.Entity)
		}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			var count = 0

			for node := range cursor.Chan() {
				var dirty = false

				if objectId, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
					log.Errorf("error getting objectid for node %d: %v", node.ID, err)
					continue
				} else if objectId != strings.ToUpper(objectId) {
					dirty = true
					node.Properties.Set(common.ObjectID.String(), strings.ToUpper(objectId))
				}

				//We may not always get these properties, so ignore errors
				if operatingSystem, err := node.Properties.Get(common.OperatingSystem.String()).String(); err == nil && operatingSystem != "" && operatingSystem != strings.ToUpper(operatingSystem) {
					dirty = true
					node.Properties.Set(common.OperatingSystem.String(), strings.ToUpper(operatingSystem))
				}

				if distinguishedName, err := node.Properties.Get(ad.DistinguishedName.String()).String(); err == nil && distinguishedName != strings.ToUpper(distinguishedName) {
					dirty = true
					node.Properties.Set(ad.DistinguishedName.String(), strings.ToUpper(distinguishedName))
				}

				if name, err := node.Properties.Get(common.Name.String()).String(); err == nil && name != strings.ToUpper(name) {
					dirty = true
					node.Properties.Set(common.Name.String(), strings.ToUpper(name))
				}

				if dirty {
					var identityKind graph.Kind

					if node.Kinds.ContainsOneOf(ad.Entity) {
						identityKind = ad.Entity
					} else if node.Kinds.ContainsOneOf(azure.Entity) {
						identityKind = azure.Entity
					} else {
						log.Errorf("Unable to figure out base kind of node %d", node.ID)
					}

					if identityKind != nil {
						if err := batch.UpdateNodeBy(graph.NodeUpdate{
							Node:               node,
							IdentityKind:       identityKind,
							IdentityProperties: []string{common.ObjectID.String()},
						}); err != nil {
							log.Errorf("Error updating node %d: %v", node.ID, err)
						}
					}
				}

				if count++; count%10000 == 0 {
					log.Infof("Completed %d nodes in migration", count)
				}
			}

			log.Infof("Completed %d nodes in migration", count)
			return cursor.Error()
		}); err != nil {
			return err
		} else {
			return nil
		}
	})
}

var Manifest = []Migration{
	{
		Version: version.Version{Major: 2, Minor: 3, Patch: 0},
		Execute: func(db graph.Database) error {
			defer log.Measure(log.LevelInfo, "Deleting all existing role nodes")()

			return db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
				return tx.Nodes().Filterf(func() graph.Criteria {
					return query.Kind(query.Node(), azure.Role)
				}).Delete()
			})
		},
	},
	{
		Version: version.Version{Major: 2, Minor: 6, Patch: 3},
		Execute: func(db graph.Database) error {
			defer log.Measure(log.LevelInfo, "Deleting all LocalToComputer/RemoteInteractiveLogin edges and ADLocalGroup labels")()

			return db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
				//Remove ADLocalGroup label from all nodes that also have the group label
				if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Node(), ad.LocalGroup),
						query.Kind(query.Node(), ad.Group),
					)
				})); err != nil {
					return err
				} else {
					for _, node := range nodes {
						node.DeleteKinds(ad.LocalGroup)
						if err := tx.UpdateNode(node); err != nil {
							return err
						}
					}
				}

				//Delete all local group edges
				if err := tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Or(
							query.Kind(query.Relationship(), ad.RemoteInteractiveLogonPrivilege),
							query.Kind(query.Relationship(), ad.LocalToComputer),
							query.Kind(query.Relationship(), ad.MemberOfLocalGroup),
						),
						query.Kind(query.Start(), ad.Entity),
					)
				}).Delete(); err != nil {
					return err
				}

				return nil
			})
		},
	},
	{
		Version: version.Version{Major: 2, Minor: 7, Patch: 7},
		Execute: Version_277_Migration,
	},
}
