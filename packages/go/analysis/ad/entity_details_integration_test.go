// Copyright 2026 Specter Ops, Inc.
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

package ad_test

import (
	"testing"

	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestEntityDetails(t *testing.T) {
	t.Parallel()

	var (
		suite              = setupIntegrationTestSuite(t)
		domainSID          = RandomDomainSID()
		computerObjectID   = newRandomObjectID(t)
		siteServerObjectID = newRandomObjectID(t)
		computerName       = "COMPUTER.TEST.LOCAL"
		siteServerName     = "CN=COMPUTER,CN=SERVERS,CN=TEST-SITE,CN=SITES,CN=CONFIGURATION,DC=TEST,DC=LOCAL"
		computer           = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:     computerName,
			common.ObjectID: computerObjectID,
			ad.DomainSID:    domainSID,
		}), ad.Entity, ad.Computer)
		siteServer = NewNode(t, &suite, graph.AsProperties(graph.PropertyMap{
			common.Name:     siteServerName,
			common.ObjectID: siteServerObjectID,
			ad.DomainSID:    domainSID,
		}), ad.Entity, ad.SiteServer)
	)
	defer teardownIntegrationTestSuite(t, &suite)

	NewRelationship(t, &suite, siteServer, computer, ad.ServerIs)

	t.Run("computer includes its linked site server", func(t *testing.T) {
		details, err := adAnalysis.ComputerEntityDetails(suite.Context, suite.GraphDB, computerObjectID)
		require.NoError(t, err)
		require.Equal(t, siteServerObjectID, details.Properties.Map["siteservernode"])
		require.Equal(t, siteServerName, details.Properties.Map["siteservernodename"])
	})

	t.Run("site server includes its referenced computer", func(t *testing.T) {
		details, err := adAnalysis.SiteServerEntityDetails(suite.Context, suite.GraphDB, siteServerObjectID)
		require.NoError(t, err)
		require.Equal(t, computerObjectID, details.Properties.Map["serverreferencecomputer"])
		require.Equal(t, computerName, details.Properties.Map["serverreferencecomputername"])
	})
}
