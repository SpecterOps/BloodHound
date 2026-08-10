// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	azschema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func GetManageEntraDSEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		source, domainService, err := ops.FetchRelationshipNodes(tx, edge)
		if err != nil {
			return err
		}

		armPaths, err := getManageEntraDSARMComposition(tx, source, domainService)
		if err != nil {
			return err
		} else if armPaths.Len() == 0 {
			return nil
		}

		applicationAdministratorPaths, err := getManageEntraDSRoleComposition(tx, domainService, source, azschema.ApplicationAdministratorRole)
		if err != nil {
			return err
		} else if applicationAdministratorPaths.Len() == 0 {
			return nil
		}

		groupsAdministratorPaths, err := getManageEntraDSRoleComposition(tx, domainService, source, azschema.GroupsAdministratorRole)
		if err != nil {
			return err
		} else if groupsAdministratorPaths.Len() == 0 {
			return nil
		}

		finalPaths.AddPathSet(armPaths)
		finalPaths.AddPathSet(applicationAdministratorPaths)
		finalPaths.AddPathSet(groupsAdministratorPaths)
		return nil
	}); err != nil {
		return graph.NewPathSet(), err
	}

	return finalPaths, nil
}

func getManageEntraDSARMComposition(tx graph.Transaction, source, domainService *graph.Node) (graph.PathSet, error) {
	var (
		finalPaths     = graph.NewPathSet()
		controlTargets = graph.NewNodeSet(domainService)
	)

	ancestorPaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      domainService,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), azschema.Contains)
		},
	})
	if err != nil {
		return nil, err
	}
	controlTargets.AddSet(ancestorPaths.AllNodes())

	controlEdges, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.EndID(), controlTargets.IDs()...),
			query.KindIn(query.Relationship(), azschema.Contributor, azschema.EntraDSContributor),
			query.KindIn(query.Start(), azschema.User, azschema.Group, azschema.ServicePrincipal),
		)
	}))
	if err != nil {
		return nil, err
	}

	for _, controlEdge := range controlEdges {
		controller, scope, err := ops.FetchRelationshipNodes(tx, controlEdge)
		if err != nil {
			return nil, err
		}

		membershipPaths := graph.NewPathSet()
		if controller.ID != source.ID {
			membershipPaths, err = ops.TraversePaths(tx, ops.TraversalPlan{
				Root:      controller,
				Direction: graph.DirectionInbound,
				BranchQuery: func() graph.Criteria {
					return query.Kind(query.Relationship(), azschema.MemberOf)
				},
				ExpansionFilter: func(segment *graph.PathSegment) bool {
					return segment.Node.ID != source.ID
				},
				PathFilter: func(_ *ops.TraversalContext, segment *graph.PathSegment) bool {
					return segment.Node.ID == source.ID
				},
			})
			if err != nil {
				return nil, err
			} else if membershipPaths.Len() == 0 {
				continue
			}
		}

		containmentPaths := graph.NewPathSet()
		if scope.ID != domainService.ID {
			containmentPaths, err = ops.TraversePaths(tx, ops.TraversalPlan{
				Root:      domainService,
				Direction: graph.DirectionInbound,
				BranchQuery: func() graph.Criteria {
					return query.Kind(query.Relationship(), azschema.Contains)
				},
				ExpansionFilter: func(segment *graph.PathSegment) bool {
					return segment.Node.ID != scope.ID
				},
				PathFilter: func(_ *ops.TraversalContext, segment *graph.PathSegment) bool {
					return segment.Node.ID == scope.ID
				},
			})
			if err != nil {
				return nil, err
			} else if containmentPaths.Len() == 0 {
				continue
			}
		}

		controlPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), controller.ID),
			query.Equals(query.EndID(), scope.ID),
			query.Kind(query.Relationship(), controlEdge.Kind),
		)))
		if err != nil {
			return nil, err
		}

		finalPaths.AddPathSet(membershipPaths)
		finalPaths.AddPathSet(controlPaths)
		finalPaths.AddPathSet(containmentPaths)
	}

	return finalPaths, nil
}

func getManageEntraDSRoles(tx graph.Transaction, tenantScopedNode *graph.Node, roleTemplateID string) (graph.NodeSet, error) {
	roles, err := FetchDescendentKindByTenantID(tx, tenantScopedNode, azschema.Role)
	if err != nil {
		return nil, err
	}

	matchingRoles := graph.NewNodeSet()
	for _, role := range roles {
		if templateID, err := role.Properties.Get(azschema.RoleTemplateID.String()).String(); graph.IsErrPropertyNotFound(err) {
			continue
		} else if err != nil {
			return nil, err
		} else if strings.EqualFold(strings.TrimSpace(templateID), strings.TrimSpace(roleTemplateID)) {
			matchingRoles.Add(role)
		}
	}

	return matchingRoles, nil
}

func getManageEntraDSRolePrincipals(tx graph.Transaction, tenantScopedNode *graph.Node, roleTemplateID string) (graph.NodeSet, error) {
	roles, err := getManageEntraDSRoles(tx, tenantScopedNode, roleTemplateID)
	if err != nil {
		return nil, err
	}

	principals, err := roleMembers(tx, roles)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		principals.Remove(role.ID)
	}

	return principals, nil
}

func getManageEntraDSRoleComposition(tx graph.Transaction, tenantScopedNode, source *graph.Node, roleTemplateID string) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()
	roles, err := getManageEntraDSRoles(tx, tenantScopedNode, roleTemplateID)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		rolePaths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      role,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), azschema.MemberOf, azschema.HasRole)
			},
			DescentFilter: roleDescentFilter,
			ExpansionFilter: func(segment *graph.PathSegment) bool {
				return segment.Node.ID != source.ID
			},
			PathFilter: func(_ *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.ID == source.ID
			},
		})
		if err != nil {
			return nil, err
		} else if rolePaths.Len() == 0 {
			continue
		}

		finalPaths.AddPathSet(rolePaths)
	}

	return finalPaths, nil
}

// ManageEntraDS creates the traversable AZManageEntraDS relationship only when the same
// effective principal has all permissions observed in validation: Contributor or Domain Services
// Contributor over the managed-domain resource, Application Administrator, and Groups Administrator.
// Directory roles are scoped by their tenantid property; tenant-to-role AZContains is not required.
// The source ARM assignment may be direct, inherited through AZContains, or effective through nested
// AZMemberOf membership.
func ManageEntraDS(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing Entra Domain Services contributors",
		attr.Namespace("analysis"),
		attr.Function("ManageEntraDS"),
		attr.Scope("process"),
	)()

	tenants, err := FetchTenants(ctx, db)
	if err != nil {
		return &post.AtomicPostProcessingStats{}, err
	}

	operation := post.NewPostRelationshipOperation(ctx, db, "AZManageEntraDS Post Processing")
	for _, tenant := range tenants {
		tenant := tenant
		if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
			applicationAdministrators, err := getManageEntraDSRolePrincipals(tx, tenant, azschema.ApplicationAdministratorRole)
			if err != nil {
				return err
			}
			groupsAdministrators, err := getManageEntraDSRolePrincipals(tx, tenant, azschema.GroupsAdministratorRole)
			if err != nil {
				return err
			}

			domainServices, err := FetchDescendentKindByTenantID(tx, tenant, azschema.EntraDS)
			if err != nil {
				return err
			}

			for _, domainService := range domainServices {
				controllers, err := effectiveEntraDSResourceControllers(tx, domainService)
				if err != nil {
					return err
				}

				for _, controller := range controllers {
					if applicationAdministrators.ContainsID(controller.ID) && groupsAdministrators.ContainsID(controller.ID) {
						if !channels.Submit(ctx, outC, post.EnsureRelationshipJob{
							FromID: controller.ID,
							ToID:   domainService.ID,
							Kind:   azschema.ManageEntraDS,
						}) {
							return nil
						}
					}
				}
			}

			return nil
		}); err != nil {
			_ = operation.Done()
			return &operation.Stats, err
		}
	}

	return &operation.Stats, operation.Done()
}

func effectiveEntraDSResourceControllers(tx graph.Transaction, domainService *graph.Node) (graph.NodeSet, error) {
	controlTargets := graph.NewNodeSet(domainService)
	if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
		Root:      domainService,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.Kind(query.Relationship(), azschema.Contains)
		},
	}); err != nil {
		return nil, err
	} else {
		controlTargets.AddSet(paths.AllNodes())
	}

	controllers, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.EndID(), controlTargets.IDs()...),
			query.KindIn(query.Relationship(), azschema.Contributor, azschema.EntraDSContributor),
			query.KindIn(query.Start(), azschema.User, azschema.Group, azschema.ServicePrincipal),
		)
	}))
	if err != nil {
		return nil, err
	}

	effectiveControllers := graph.NewNodeSet()
	for _, controller := range controllers {
		effectiveControllers.Add(controller)
		if !controller.Kinds.ContainsOneOf(azschema.Group) {
			continue
		}

		if paths, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      controller,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.Kind(query.Relationship(), azschema.MemberOf)
			},
			PathFilter: func(_ *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(azschema.User, azschema.Group, azschema.ServicePrincipal)
			},
		}); err != nil {
			return nil, err
		} else {
			effectiveControllers.AddSet(paths.AllNodes())
		}
	}

	return effectiveControllers, nil
}
