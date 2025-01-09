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
	"errors"
	"fmt"
	"github.com/specterops/bloodhound/log/measure"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/analysis"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/version"
)

func RequiresMigration(ctx context.Context, db graph.Database) (bool, error) {
	if currentMigration, err := GetMigrationData(ctx, db); err != nil {
		if errors.Is(err, graph.ErrNoResultsFound) || errors.Is(err, ErrNoMigrationData) {
			return true, nil
		} else {
			return false, fmt.Errorf("unable to get graph db migration data: %w", err)
		}
	} else {
		return LatestGraphMigrationVersion().GreaterThan(currentMigration), nil
	}
}

// Version_620_Migration is intended to rename the RemoteInteractiveLogonPrivilege edge to RemoteInteractiveLogonRight
// See: https://specterops.atlassian.net/browse/BED-4428
func Version_620_Migration(db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to rename RemoteInteractiveLogonPrivilege edges")()

	// MATCH p=(n:Base)-[:RemoteInteractiveLogonPrivilege]->(m:Base) RETURN p
	targetCriteria := query.And(
		query.Kind(query.Start(), ad.Entity),
		query.Kind(query.Relationship(), graph.StringKind("RemoteInteractiveLogonPrivilege")),
		query.Kind(query.End(), ad.Entity),
	)

	edgeProperties := graph.NewProperties()
	edgeProperties.Set(common.LastSeen.String(), time.Now().UTC())

	//Get all RemoteInteractiveLogonPrivilege edges, use the start/end ids to insert new edges, and delete the old ones
	return db.BatchOperation(context.Background(), func(batch graph.Batch) error {
		rels, err := ops.FetchRelationships(batch.Relationships().Filter(targetCriteria))
		if err != nil {
			return err
		}

		for _, rel := range rels {
			if err := batch.CreateRelationshipByIDs(rel.StartID, rel.EndID, ad.RemoteInteractiveLogonRight, edgeProperties); err != nil {
				return err
			} else if err := batch.DeleteRelationship(rel.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

// Version_513_Migration covers a bug discovered in ingest that would result in nodes containing more than one valid
// kind. For example, the following query should return no results: (n:Base) where (n:User and n:Computer) return n
// but, at time of writing, due to the ingest defect several environments will return results for this query.
//
// For instances where nodes are discovered that contain more than one valid kind assignment, the node's kinds are
// reset to the matching base entity kind such that:
//
// node.Kinds = Kinds{ad.Entity, ad.User, ad.Computer} must be reset to:
// node.Kinds = Kinds{ad.Entity}
func Version_513_Migration(db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to remove incorrectly ingested labels")()

	// Cypher for the below filter is: size(labels(n)) > 2 and not (n:Group and n:ADLocalGroup) or size(labels(n)) > 3 and (n:Group and n:ADLocalGroup)
	targetCriteria := query.Or(
		query.And(
			query.GreaterThan(query.Size(query.KindsOf(query.Node())), 2),
			query.Not(query.Kind(query.Node(), ad.Entity, ad.Group, ad.LocalGroup)),
		),
		query.And(
			query.GreaterThan(query.Size(query.KindsOf(query.Node())), 3),
			query.Kind(query.Node(), ad.Entity, ad.Group, ad.LocalGroup),
		),
	)

	if nodes, err := ops.ParallelFetchNodes(context.Background(), db, targetCriteria, analysis.MaximumDatabaseParallelWorkers); err != nil {
		return err
	} else if err := db.BatchOperation(context.Background(), func(batch graph.Batch) error {
		for _, node := range nodes {
			if node.Kinds.ContainsOneOf(ad.Entity) {
				// Nodes are designed to track additions and deletions. By making this assignment, the update logic
				// will append removals for all kinds not the base kind such that only the base kind remains on the
				// node in the database.
				node.DeletedKinds = node.Kinds.Remove(ad.Entity)

				if err := batch.UpdateNodeBy(graph.NodeUpdate{
					Node:               node,
					IdentityKind:       ad.Entity,
					IdentityProperties: []string{common.ObjectID.String()},
				}); err != nil {
					return err
				}
			} else if node.Kinds.ContainsOneOf(azure.Entity) {
				// Nodes are designed to track additions and deletions. By making this assignment, the update logic
				// will append removals for all kinds not the base kind such that only the base kind remains on the
				// node in the database.
				node.DeletedKinds = node.Kinds.Remove(azure.Entity)

				if err := batch.UpdateNodeBy(graph.NodeUpdate{
					Node:               node,
					IdentityKind:       azure.Entity,
					IdentityProperties: []string{common.ObjectID.String()},
				}); err != nil {
					return err
				}
			}
		}

		log.Infof(fmt.Sprintf("Migration removed all non-entity kinds from %d incorrectly labeled nodes", nodes.Len()))
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func Version_508_Migration(db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Migrating Azure Owns to Owner")()

	return db.BatchOperation(context.Background(), func(batch graph.Batch) error {
		return batch.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Start(), azure.Entity),
				// Not all of these node types are being changed, but there's no harm in adding them to the migration
				query.KindIn(query.End(), azure.ManagementGroup, azure.ResourceGroup, azure.Subscription, azure.KeyVault, azure.AutomationAccount, azure.ContainerRegistry, azure.LogicApp, azure.VMScaleSet, azure.WebApp, azure.FunctionApp, azure.ManagedCluster, azure.VM),
				query.Kind(query.Relationship(), azure.Owns),
			)
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for rel := range cursor.Chan() {
				startId, endId := rel.StartID, rel.EndID
				newProperties := graph.NewProperties()
				if lastSeen, err := rel.Properties.Get(common.LastSeen.String()).Time(); err != nil {
					newProperties.Set(common.LastSeen.String(), time.Now().UTC())
				} else {
					newProperties.Set(common.LastSeen.String(), lastSeen)
				}

				if err := batch.CreateRelationshipByIDs(startId, endId, azure.Owner, newProperties); err != nil {
					return err
				} else if err := batch.DeleteRelationship(rel.ID); err != nil {
					return err
				}
			}

			return nil
		})
	})
}

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
					log.Errorf(fmt.Sprintf("Error getting objectid for node %d: %v", node.ID, err))
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
						log.Errorf(fmt.Sprintf("Unable to figure out base kind of node %d", node.ID))
					}

					if identityKind != nil {
						if err := batch.UpdateNodeBy(graph.NodeUpdate{
							Node:               node,
							IdentityKind:       identityKind,
							IdentityProperties: []string{common.ObjectID.String()},
						}); err != nil {
							log.Errorf(fmt.Sprintf("Error updating node %d: %v", node.ID, err))
						}
					}
				}

				if count++; count%10000 == 0 {
					log.Infof(fmt.Sprintf("Completed %d nodes in migration", count))
				}
			}

			log.Infof(fmt.Sprintf("Completed %d nodes in migration", count))
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
							query.Kind(query.Relationship(), graph.StringKind("RemoteInteractiveLogonPrivilege")),
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
	{
		Version: version.Version{Major: 5, Minor: 0, Patch: 8},
		Execute: Version_508_Migration,
	},
	{
		Version: version.Version{Major: 5, Minor: 13, Patch: 0},
		Execute: Version_513_Migration,
	},
	{
		Version: version.Version{Major: 6, Minor: 2, Patch: 0},
		Execute: Version_620_Migration,
	},
}

func LatestGraphMigrationVersion() version.Version {
	var latestVersion version.Version

	for _, migration := range Manifest {
		if migration.Version.GreaterThan(latestVersion) {
			latestVersion = migration.Version
		}
	}

	return latestVersion
}
