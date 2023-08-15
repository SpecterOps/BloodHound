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
// +build integration

package analysis_test

import (
	"context"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/src/test"
	"testing"

	analysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test/integration"
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

func TestResolveAllGroupMemberships(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.RDP.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		memberships, err := analysis.ResolveAllGroupMemberships(context.Background(), db)

		test.RequireNilErr(t, err)

		require.Equal(t, 3, int(memberships.Cardinality(harness.RDP.DomainGroupA.ID.Uint32()).Cardinality()))
		require.Equal(t, 2, int(memberships.Cardinality(harness.RDP.DomainGroupB.ID.Uint32()).Cardinality()))
		require.Equal(t, 1, int(memberships.Cardinality(harness.RDP.DomainGroupC.ID.Uint32()).Cardinality()))
		require.Equal(t, 1, int(memberships.Cardinality(harness.RDP.DomainGroupD.ID.Uint32()).Cardinality()))
		require.Equal(t, 2, int(memberships.Cardinality(harness.RDP.DomainGroupE.ID.Uint32()).Cardinality()))
	})
}
