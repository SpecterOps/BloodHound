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
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

func FilterEntityActiveAssignments() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.HasRole, azure.MemberOf)
}

func FilterEntityPIMAssignments() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.Grant, azure.GrantSelf, azure.MemberOf)
}

func FilterRoleApprovers() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.AZRoleApprover, azure.MemberOf)
}

func FilterEntityEligibleRoles() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.AZRoleEligible, azure.MemberOf)
}

func FilterExecutionPrivileges() graph.Criteria {
	return query.KindIn(query.Relationship(), append(azure.ExecutionPrivileges(), azure.MemberOf)...)
}

func FilterKeyReaders() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.MemberOf, azure.Contributor, azure.Owner, azure.GetKeys)
}

func FilterCertificateReaders() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.MemberOf, azure.Contributor, azure.Owner, azure.GetCertificates)
}

func FilterSecretReaders() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.MemberOf, azure.Contributor, azure.Owner, azure.GetSecrets)
}

func FilterControlsRelationships() graph.Criteria {
	return query.KindIn(query.Relationship(), append(azure.ControlRelationships(), azure.MemberOf, azure.Contains)...)
}

func FilterAppRoleAssignmentTransitRelationships() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.AppRoleTransitRelationshipKinds()...)
}

func FilterAbusableAppRoleAssignmentRelationships() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.AbusableAppRoleRelationshipKinds()...)
}

func FilterGroupMembership() graph.Criteria {
	return query.Kind(query.Relationship(), azure.MemberOf)
}

func FilterGroupMembers() graph.Criteria {
	return query.And(
		query.Kind(query.Relationship(), azure.MemberOf),
		query.Kind(query.Start(), azure.Entity),
	)
}

// FilterRoleAssignableGroupMembersUsers matches user membership in role assignable groups. The property criteria must
// stay consistent with isRoleAssignableGroup.
func FilterRoleAssignableGroupMembersUsers() graph.Criteria {
	return query.And(
		query.Kind(query.Relationship(), azure.MemberOf),
		query.Kind(query.Start(), azure.User),
		// Current ingestion writes a boolean. Match legacy string "true" as well so
		// both representations populate the membership set used to exclude targets.
		query.Or(
			query.Equals(query.EndProperty(azure.IsAssignableToRole.String()), true),
			query.Equals(query.EndProperty(azure.IsAssignableToRole.String()), "true"),
		),
	)
}

func FilterContains() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.Contains)
}

// isRoleAssignableGroup reports whether the given group node is marked as role assignable. Current ingestion stores a
// boolean, but older data can contain the string "true"; both are accepted. False, missing, or invalid values are not.
// This must stay consistent with the property criteria in FilterRoleAssignableGroupMembersUsers.
func isRoleAssignableGroup(group *graph.Node) bool {
	switch group.Properties.Get(azure.IsAssignableToRole.String()).Any() {
	case true, "true":
		return true
	default:
		return false
	}
}

func roleDescentFilter(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
	var (
		groupVisited     = false
		acceptDescendent = true
	)

	segment.Path().WalkReverse(func(start, end *graph.Node, relationship *graph.Relationship) bool {
		if end.Kinds.ContainsOneOf(azure.Group) {
			// If this is the second group in the path then we do not inherit the terminal role
			if groupVisited || start.Kinds.ContainsOneOf(azure.Group) {
				acceptDescendent = false
				return false
			} else {
				groupVisited = true
			}

			// False, missing, or invalid values must stop traversal through this group.
			if !isRoleAssignableGroup(end) {
				if end.Properties.Get(azure.IsAssignableToRole.String()).IsNil() {
					slog.Warn(
						"Node is missing property isassignabletorole",
						slog.Uint64("node_id", uint64(end.ID)),
					)
				}
				acceptDescendent = false
				return false
			}
		}

		return true
	})

	return acceptDescendent
}
