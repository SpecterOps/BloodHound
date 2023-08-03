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
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/log"
)

func FilterEntityActiveAssignments() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.HasRole, azure.MemberOf)
}

func FilterEntityPIMAssignments() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.Grant, azure.GrantSelf, azure.MemberOf)
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
	return query.KindIn(query.Relationship(), append(azure.ControlRelationships(), azure.MemberOf)...)
}

func FilterAppRoleAssignmentTransitRelationships() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.AppRoleTransitRelationshipKinds()...)
}

func FilterAbusableAppRoleAssignmentRelationships() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.AbusableAppRoleRelationshipKinds()...)
}

func FilterDescendents(kinds ...graph.Kind) graph.CriteriaProvider {
	return func() graph.Criteria {
		return query.And(
			query.Kind(query.Relationship(), azure.Contains),
			query.KindIn(query.End(), kinds...),
		)
	}
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

func FilterContains() graph.Criteria {
	return query.KindIn(query.Relationship(), azure.Contains)
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

			// If the group does not allow role inheritance then we do not inherit the terminal role
			if isRoleAssignable, err := end.Properties.Get(azure.IsAssignableToRole.String()).Bool(); err != nil || !isRoleAssignable {
				if graph.IsErrPropertyNotFound(err) {
					log.Errorf("Node %d is missing property %s", end.ID, azure.IsAssignableToRole)
				}

				acceptDescendent = false
				return false
			}
		}

		return true
	})

	return acceptDescendent
}
