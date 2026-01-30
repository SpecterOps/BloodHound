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
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func RequiresMigration(ctx context.Context, db graph.Database) (bool, error) {
	if currentMigration, err := GetMigrationData(ctx, db); err != nil {
		if errors.Is(err, graph.ErrNoResultsFound) || errors.Is(err, ErrNoMigrationData) {
			return true, nil
		} else {
			return false, fmt.Errorf("unable to get graph db migration data: %w", err)
		}
	} else {
		return version.GetVersion().GreaterThan(currentMigration), nil
	}
}

func Version_860_Migration(ctx context.Context, db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to remove AZOwner edges to AZDevice")()

	targetCriteria := query.And(
		query.Kind(query.Start(), azure.Entity),
		query.Kind(query.Relationship(), azure.Owns),
		query.Kind(query.End(), azure.Device),
	)

	return db.BatchOperation(ctx, func(batch graph.Batch) error {
		rels, err := ops.FetchRelationships(batch.Relationships().Filter(targetCriteria))

		if err != nil {
			return err
		}

		for _, rel := range rels {
			if err := batch.DeleteRelationship(rel.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

func Version_852_Migration(ctx context.Context, db graph.Database) error {
	const removalQuery = `
		delete from edge e
		where e.kind_id = any(
			select id from kind k where k.name = any(array['AdminTo', 'ExecuteDCOM', 'CanPSRemote', 'CanRDP'])
		);`

	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to cleanup bad LocalGroup edges from 8.5.1")()

	// The incorrectly assigned edges were only created in PG. Due to the way translation
	// works, edge queries match on nodes and will not be able to address edges created
	// with non-existent node IDs.
	if pg.IsPostgreSQLGraph(db) {
		return db.Run(ctx, removalQuery, nil)
	}

	return nil
}

func Version_830_Migration(ctx context.Context, db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to cleanup bad `lastseen` properties from 7.4.0")()

	// This is a bit gross, but we can't use `query.Equals` here because we need to
	// force a string comparison, otherwise there is no value we can pass that will
	// correctly map to the "{}" string value we need.
	targetCriteria := query.And(
		query.StringStartsWith(query.RelationshipProperty(common.LastSeen.String()), "{}"),
		query.StringEndsWith(query.RelationshipProperty(common.LastSeen.String()), "{}"),
	)

	return db.BatchOperation(ctx, func(batch graph.Batch) error {
		rels, err := ops.FetchRelationships(batch.Relationships().Filter(targetCriteria))
		if err != nil {
			return err
		}

		for _, rel := range rels {
			if err := batch.DeleteRelationship(rel.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

func Version_813_Migration(ctx context.Context, db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to revert MemberOf between well known groups")()

	targetCriteria := query.And(
		query.Kind(query.Start(), ad.Group),
		query.Kind(query.End(), ad.Group),
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Or(
			query.And(
				query.StringEndsWith(query.StartProperty(common.ObjectID.String()), wellknown.AuthenticatedUsersSIDSuffix.String()),
				query.StringEndsWith(query.EndProperty(common.ObjectID.String()), wellknown.AuthenticatedUsersSIDSuffix.String()),
			),
			query.And(
				query.StringEndsWith(query.StartProperty(common.ObjectID.String()), wellknown.EveryoneSIDSuffix.String()),
				query.StringEndsWith(query.EndProperty(common.ObjectID.String()), wellknown.EveryoneSIDSuffix.String()),
			),
		),
	)

	return db.BatchOperation(ctx, func(batch graph.Batch) error {
		rels, err := ops.FetchRelationships(batch.Relationships().Filter(targetCriteria))
		if err != nil {
			return err
		}

		for _, rel := range rels {
			if err := batch.DeleteRelationship(rel.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

// Version_740_Migration is intended to split the TrustedBy edge to into SameForestTrust and CrossForestTrust edges
func Version_740_Migration(ctx context.Context, db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to split the TrustedBy edges")()

	targetCriteria := query.And(
		query.Kind(query.Start(), ad.Domain),
		query.Kind(query.End(), ad.Domain),
		query.Kind(query.Relationship(), graph.StringKind("TrustedBy")),
	)

	return db.BatchOperation(ctx, func(batch graph.Batch) error {
		rels, err := ops.FetchRelationships(batch.Relationships().Filter(targetCriteria))
		if err != nil {
			return err
		}

		for _, rel := range rels {
			// Determine new edge kind
			trustType, _ := rel.Properties.Get(ad.TrustType.String()).String()
			newEdgeKind := ad.CrossForestTrust
			if trustType == "ParentChild" {
				newEdgeKind = ad.SameForestTrust
			}

			edgeProperties := graph.NewProperties()
			edgeProperties.Set(ad.IsACL.String(), false)
			edgeProperties.Set(ad.TrustType.String(), trustType)
			edgeProperties.Set(common.LastSeen.String(), rel.Properties.Get(common.LastSeen.String()).Any())

			// Create new edge in opposite direction
			if err := batch.CreateRelationship(&graph.Relationship{
				StartID:    rel.EndID,
				EndID:      rel.StartID,
				Kind:       newEdgeKind,
				Properties: edgeProperties,
			}); err != nil {
				return err
			}

			// Delete old edge
			if err := batch.DeleteRelationship(rel.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

// Version_730_Migration removes the leftover 'adminrightscount' property from User nodes
func Version_730_Migration(ctx context.Context, db graph.Database) error {
	const adminRightsCount = "adminrightscount"

	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to remove admin_rights_count property from user nodes and smbsigning from computer nodes")

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		// MATCH(n:User) WHERE n.adminrightscount <> null
		if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.User),
				query.IsNotNull(query.NodeProperty(adminRightsCount)),
			)
		})); err != nil {
			return err
		} else {
			for _, node := range nodes {
				node.Properties.Delete(adminRightsCount)
				if err := tx.UpdateNode(node); err != nil {
					return err
				}
			}
		}

		if nodes, err := ops.FetchNodes(tx.Nodes().Filter(query.And(
			query.Kind(query.Node(), ad.Computer),
			query.IsNotNull(query.NodeProperty(ad.SMBSigning.String())),
			query.Equals(query.NodeProperty(ad.SMBSigning.String()), false),
		))); err != nil {
			return err
		} else {
			for _, node := range nodes {
				node.Properties.Delete(ad.SMBSigning.String())
				if err := tx.UpdateNode(node); err != nil {
					return err
				}
			}

			return nil
		}
	})
}

// Version_620_Migration is intended to rename the RemoteInteractiveLogonPrivilege edge to RemoteInteractiveLogonRight
// See: https://specterops.atlassian.net/browse/BED-4428
func Version_620_Migration(ctx context.Context, db graph.Database) error {
	defer measure.LogAndMeasure(slog.LevelInfo, "Migration to rename RemoteInteractiveLogonPrivilege edges")()

	// MATCH p=(n:Base)-[:RemoteInteractiveLogonPrivilege]->(m:Base) RETURN p
	targetCriteria := query.And(
		query.Kind(query.Start(), ad.Entity),
		query.Kind(query.Relationship(), graph.StringKind("RemoteInteractiveLogonPrivilege")),
		query.Kind(query.End(), ad.Entity),
	)

	edgeProperties := graph.NewProperties()
	edgeProperties.Set(common.LastSeen.String(), time.Now().UTC())

	// Get all RemoteInteractiveLogonPrivilege edges, use the start/end ids to insert new edges, and delete the old ones
	return db.BatchOperation(ctx, func(batch graph.Batch) error {
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
func Version_513_Migration(ctx context.Context, db graph.Database) error {
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

	if nodes, err := ops.ParallelFetchNodes(ctx, db, targetCriteria, analysis.MaximumDatabaseParallelWorkers); err != nil {
		return err
	} else if err := db.BatchOperation(ctx, func(batch graph.Batch) error {
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

		slog.Info(fmt.Sprintf("Migration removed all non-entity kinds from %d incorrectly labeled nodes", nodes.Len()))
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func Version_508_Migration(ctx context.Context, db graph.Database) error {
	defer measure.Measure(slog.LevelInfo, "Migrating Azure Owns to Owner")()

	return db.BatchOperation(ctx, func(batch graph.Batch) error {
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

func Version_277_Migration(ctx context.Context, db graph.Database) error {
	defer measure.Measure(slog.LevelInfo, "Migrating node property casing")()

	return db.BatchOperation(ctx, func(batch graph.Batch) error {
		if err := batch.Nodes().Filterf(func() graph.Criteria {
			return query.KindIn(query.Node(), ad.Entity, azure.Entity)
		}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			var count = 0

			for node := range cursor.Chan() {
				var dirty = false

				if objectId, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
					slog.Error(fmt.Sprintf("Error getting objectid for node %d: %v", node.ID, err))
					continue
				} else if objectId != strings.ToUpper(objectId) {
					dirty = true
					node.Properties.Set(common.ObjectID.String(), strings.ToUpper(objectId))
				}

				// We may not always get these properties, so ignore errors
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
						slog.Error(fmt.Sprintf("Unable to figure out base kind of node %d", node.ID))
					}

					if identityKind != nil {
						if err := batch.UpdateNodeBy(graph.NodeUpdate{
							Node:               node,
							IdentityKind:       identityKind,
							IdentityProperties: []string{common.ObjectID.String()},
						}); err != nil {
							slog.Error(fmt.Sprintf("Error updating node %d: %v", node.ID, err))
						}
					}
				}

				if count++; count%10000 == 0 {
					slog.Info(fmt.Sprintf("Completed %d nodes in migration", count))
				}
			}

			slog.Info(fmt.Sprintf("Completed %d nodes in migration", count))
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
		Execute: func(ctx context.Context, db graph.Database) error {
			defer measure.Measure(slog.LevelInfo, "Deleting all existing role nodes")()

			return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
				return tx.Nodes().Filterf(func() graph.Criteria {
					return query.Kind(query.Node(), azure.Role)
				}).Delete()
			})
		},
	},
	{
		Version: version.Version{Major: 2, Minor: 6, Patch: 3},
		Execute: func(ctx context.Context, db graph.Database) error {
			defer measure.Measure(slog.LevelInfo, "Deleting all LocalToComputer/RemoteInteractiveLogin edges and ADLocalGroup labels")()

			// This kind has long since gone missing from our schemas but the assert below reintroduces it for the
			// purposes of running this migration
			rilpKind := graph.StringKind("RemoteInteractiveLogonPrivilege")

			if err := db.AssertSchema(ctx, graph.Schema{
				Graphs: []graph.Graph{{
					Edges: graph.Kinds{rilpKind},
				}},
			}); err != nil {
				return err
			}

			return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
				// Remove ADLocalGroup label from all nodes that also have the group label
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

				// Delete all local group edges
				if err := tx.Relationships().Filterf(func() graph.Criteria {
					return query.And(
						query.Or(
							query.Kind(query.Relationship(), rilpKind),
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
	{
		Version: version.Version{Major: 7, Minor: 3, Patch: 0},
		Execute: Version_730_Migration,
	},
	{
		Version: version.Version{Major: 7, Minor: 4, Patch: 0},
		Execute: Version_740_Migration,
	},
	{
		Version: version.Version{Major: 8, Minor: 1, Patch: 3},
		Execute: Version_813_Migration,
	},
	{
		Version: version.Version{Major: 8, Minor: 3, Patch: 0},
		Execute: Version_830_Migration,
	},
	{
		Version: version.Version{Major: 8, Minor: 5, Patch: 2},
		Execute: Version_852_Migration,
	},
	{
		Version: version.Version{Major: 8, Minor: 6, Patch: 0},
		Execute: Version_860_Migration,
	},
}
