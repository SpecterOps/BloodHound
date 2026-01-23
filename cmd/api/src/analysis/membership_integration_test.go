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

//go:build integration

package analysis_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	analysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/algo"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestRealizeNodeKindDuplexMap(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RootADHarness.Setup(testContext)
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		var (
			domainNode                    = testContext.FindNode(query.Equals(query.NodeProperty("name"), "DomainA"))
			impactMap, impactErr          = analysis.FetchPathMembers(context.Background(), db, domainNode.ID, graph.DirectionInbound)
			impactKindMap, realizationErr = analysis.NodeDuplexByKinds(context.Background(), db, impactMap)
		)

		require.Nil(t, impactErr)
		require.Nil(t, realizationErr)
		require.Equalf(t, 9, int(impactMap.Cardinality()), "Failed to collect expected nodes. Saw IDs: %+v", impactMap.Slice())

		require.Equal(t, 3, int(impactKindMap.Get(ad.Domain).Cardinality()))
		require.Equal(t, 2, int(impactKindMap.Get(ad.Group).Cardinality()))
		require.Equal(t, 3, int(impactKindMap.Get(ad.User).Cardinality()))
		require.Equal(t, 1, int(impactKindMap.Get(ad.GPO).Cardinality()))
	})
}

func TestAnalyzeExposure(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RootADHarness.Setup(testContext)
		harness.TrustDCSync.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		var (
			domainNode     = testContext.FindNode(query.Equals(query.NodeProperty("name"), "DomainA"))
			impactMap, err = analysis.FetchPathMembers(context.Background(), db, domainNode.ID, graph.DirectionInbound)
		)

		require.Nil(t, err)
		require.Equalf(t, 9, int(impactMap.Cardinality()), "Failed to collect expected nodes. Saw IDs: %+v", impactMap.Slice())
	})
}

func TestResolveReachOfGroupMembershipComponents(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDP.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		memberships, err := algo.FetchReachabilityCache(context.Background(), db, query.KindIn(query.Relationship(), ad.MemberOf, ad.MemberOfLocalGroup))

		test.RequireNilErr(t, err)

		// Because the algorithm uses a condensed (SCC) version of the directed graph, component membership
		// will always include the origin member that reach was computed from. Typically, downstream users
		// of this function will remove the ID from their merged bitmap after reachability is computed.
		require.Equal(t, 4, int(memberships.ReachOfComponentContainingMember(harness.RDP.DomainGroupA.ID.Uint64(), graph.DirectionInbound).Cardinality()))
		require.Equal(t, 2, int(memberships.ReachOfComponentContainingMember(harness.RDP.DomainGroupB.ID.Uint64(), graph.DirectionInbound).Cardinality()))
		require.Equal(t, 2, int(memberships.ReachOfComponentContainingMember(harness.RDP.DomainGroupC.ID.Uint64(), graph.DirectionInbound).Cardinality()))
		require.Equal(t, 2, int(memberships.ReachOfComponentContainingMember(harness.RDP.DomainGroupD.ID.Uint64(), graph.DirectionInbound).Cardinality()))
		require.Equal(t, 3, int(memberships.ReachOfComponentContainingMember(harness.RDP.DomainGroupE.ID.Uint64(), graph.DirectionInbound).Cardinality()))
	})
}
