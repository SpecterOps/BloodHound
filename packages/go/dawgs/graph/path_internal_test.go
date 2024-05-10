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
