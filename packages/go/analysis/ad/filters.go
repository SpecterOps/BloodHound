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

package ad

import (
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func FilterSessions() graph.Criteria {
	return query.Kind(query.Relationship(), ad.HasSession)
}

func FilterGroupMembership() graph.Criteria {
	return query.Kind(query.Relationship(), ad.MemberOf)
}

func FilterGroupMembers() graph.Criteria {
	return query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.Start(), ad.Entity),
	)
}

func IsValidInboundControllerPath(next *graph.PathSegment) bool {
	// First degree edge may not be MemberOf
	if next.Depth() == 1 {
		return !next.Edge.Kind.Is(ad.MemberOf)
	}

	// All 1+n degree edges must be MemberOf
	return next.Edge.Kind.Is(ad.MemberOf)
}

func InboundControllerPaths() traversal.SegmentFilter {
	return traversal.UniquePathSegmentFilter(IsValidInboundControllerPath)
}

func InboundControllerNodes(nodes *traversal.NodeCollector, skip, limit int) traversal.SegmentFilter {
	return traversal.AcyclicNodeFilter(
		traversal.FilteredSkipLimit(
			func(next *graph.PathSegment) (bool, bool) {
				// Controllers collect and descend only on valid controller path segments
				isValidInboundController := IsValidInboundControllerPath(next)
				return isValidInboundController, isValidInboundController
			},
			nodes.Collect,
			skip,
			limit,
		),
	)
}

func OutboundControlledPaths(paths *traversal.PathCollector) traversal.SegmentFilter {
	return traversal.UniquePathSegmentFilter(func(next *graph.PathSegment) bool {
		// Is this a controlling edge?
		if next.Edge.Kind.Is(ad.ACLRelationships()...) {
			// If so, then this path is terminal. Mark it as a result and do not descend into it.
			paths.Add(next.Path())
			return false
		}

		// Otherwise descend past this node in traversal
		return true
	})
}

func OutboundControlledNodes(nodes *traversal.NodeCollector, skip, limit int) traversal.SegmentFilter {
	return traversal.AcyclicNodeFilter(
		traversal.FilteredSkipLimit(
			func(next *graph.PathSegment) (bool, bool) {
				// Collect path segments that end in an ACL kind and continue descent only if the segment is MemberOf
				return next.Edge.Kind.Is(ad.ACLRelationships()...), next.Edge.Kind.Is(ad.MemberOf)
			},
			nodes.Collect,
			skip,
			limit,
		),
	)
}

func OutboundControlDescentFilter(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
	var (
		shouldDescend = true
		sawControlRel = false
	)

	// We want to ensure that MemberOf is expanded as well as controls relationships but with one exception: we do not
	// want to traverse more than one degree of control relationships. The question being answered for this entity query
	// is, "what does this entity have direct control of, including the entity's group memberships." We also do not want to
	// traverse a MemberOf after we traverse a control relationship
	segment.Path().Walk(func(_, _ *graph.Node, relationship *graph.Relationship) bool {
		if relationship.Kind.Is(ad.ACLRelationships()...) {
			if !sawControlRel {
				sawControlRel = true
			} else {
				// Reaching this condition means that this descent would result in a second control
				// relationship in this path, making this descendent ineligible for further traversal
				shouldDescend = false
				return false
			}
		} else if relationship.Kind.Is(ad.MemberOf) && sawControlRel {
			//If we've already seen a control rel, and we get to a MemberOf, we need to prevent a descent as well
			shouldDescend = false
			return false
		}

		return true
	})

	return shouldDescend
}

func FilterContainsRelationship() graph.Criteria {
	return query.Kind(query.Relationship(), ad.Contains)
}

func BlocksInheritanceDescentFilter(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
	if !segment.Node.Kinds.ContainsOneOf(ad.OU) {
		return true
	} else if previousNode := segment.Trunk.Node; !previousNode.Kinds.ContainsOneOf(ad.OU) {
		//If our previous node is not an OU, continue descent
		return true
	} else if blocksInheritance, err := previousNode.Properties.Get(ad.BlocksInheritance.String()).Bool(); err != nil {
		//If we get an error, we'll just default to unenforced
		return true
	} else if blocksInheritance {
		//If our previous node blocks inheritance, we don't want to descend further, but we still want this node
		return false
	} else {
		return true
	}
}

func SelectUsersCandidateFilter(node *graph.Node) bool {
	return node.Kinds.ContainsOneOf(ad.User)
}

func SelectComputersCandidateFilter(node *graph.Node) bool {
	return node.Kinds.ContainsOneOf(ad.Computer)
}

func SelectGPOTierZeroCandidateFilter(node *graph.Node) bool {
	if tags, err := node.Properties.Get(common.SystemTags.String()).String(); err != nil {
		return false
	} else if node.Kinds.ContainsOneOf(ad.Group) {
		// GPOs don’t apply to groups.
		return false
	} else {
		return strings.Contains(tags, ad.AdminTierZero)
	}
}

func SelectGPOContainerCandidateFilter(node *graph.Node) bool {
	return node.Kinds.ContainsOneOf(ad.OU, ad.Domain)
}

func FilterEnrollers(node graph.Node) graph.Criteria {
	return query.And(
		query.Equals(query.EndID(), node.ID),
		query.Kind(query.Relationship(), ad.Enroll),
		query.Kind(query.Start(), ad.Entity),
	)
}

func FilterPublishedCAs(certTemplate *graph.Node) graph.Criteria {
	return query.And(
		query.Equals(query.StartID(), certTemplate.ID),
		query.KindIn(query.End(), ad.EnterpriseCA),
		query.KindIn(query.Relationship(), ad.PublishedTo),
	)
}
