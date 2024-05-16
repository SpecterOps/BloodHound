// Copyright 2024 Specter Ops, Inc.
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

package graph

import (
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
)

var (
	groupKind      = StringKind("group")
	domainKind     = StringKind("domain")
	userKind       = StringKind("user")
	computerKind   = StringKind("computer")
	permissionKind = StringKind("permission")
	membershipKind = StringKind("member")
)

func Test_ComputeAndSetSize(t *testing.T) {
	var (
		idSequence = int64(0)

		domainNode = NewNode(ID(1), NewProperties(), domainKind)
		groupNode  = NewNode(ID(2), NewProperties(), groupKind)
		userNode   = NewNode(ID(3), NewProperties(), userKind)

		domainSegment = NewRootPathSegment(domainNode)
		originalSize  = int64(domainSegment.SizeOf())
	)

	// Add a Group segment
	edge := NewRelationship(ID(atomic.AddInt64(&idSequence, 1)), groupNode.ID, domainNode.ID, NewProperties(), permissionKind)
	groupSegment := &PathSegment{
		Node:  groupNode,
		Trunk: domainSegment,
		Edge:  edge,
	}
	groupSegment.computeAndSetSize()

	// Appending the branch and calling the function should update size
	domainSegment.Branches = append(domainSegment.Branches, groupSegment)
	domainSegment.computeAndSetSize()

	sizeWithOneBranch := int64(domainSegment.SizeOf())
	require.Greater(t, sizeWithOneBranch, originalSize)

	// Add a User segment
	edge = NewRelationship(ID(atomic.AddInt64(&idSequence, 1)), userNode.ID, domainNode.ID, NewProperties(), permissionKind)
	userSegment := &PathSegment{
		Node:  userNode,
		Trunk: domainSegment,
		Edge:  edge,
	}
	userSegment.computeAndSetSize()

	domainSegment.Branches = append(domainSegment.Branches, userSegment)
	domainSegment.computeAndSetSize()

	// Appending the branch and calling the function should update size
	sizeWithTwoBranches := int64(domainSegment.SizeOf())
	require.Greater(t, sizeWithTwoBranches, sizeWithOneBranch)

	// Remove one branch and call the function, to ensure the size reduces accordingly
	domainSegment.Branches = []*PathSegment{groupSegment}
	domainSegment.computeAndSetSize()
	require.Less(t, int64(domainSegment.size), sizeWithTwoBranches)
}
