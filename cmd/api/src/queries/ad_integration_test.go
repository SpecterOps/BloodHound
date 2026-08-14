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

package queries_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/cache"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestGraphQuery_GetADEntityDetails(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		var (
			computer = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "COMPUTER-NAME",
				common.ObjectID: "COMPUTER-1",
			}), ad.Entity, ad.Computer)
			siteServer = testContext.NewNode(graph.AsProperties(graph.PropertyMap{
				common.Name:     "SITE-SERVER-NAME",
				common.ObjectID: "SITE-SERVER-1",
			}), ad.Entity, ad.SiteServer)
		)

		testContext.NewNode(graph.AsProperties(graph.PropertyMap{
			common.Name:     "UNLINKED-COMPUTER-NAME",
			common.ObjectID: "UNLINKED-COMPUTER-1",
		}), ad.Entity, ad.Computer)
		testContext.NewRelationship(siteServer, computer, ad.ServerIs)

		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		graphQuery := queries.NewGraphQuery(db, cache.Cache{}, config.Configuration{})

		t.Run("computer includes linked site server properties", func(t *testing.T) {
			computer, err := graphQuery.GetADEntityDetails(context.Background(), "COMPUTER-1", ad.Computer)
			require.NoError(t, err)
			require.Equal(t, "SITE-SERVER-1", computer.Properties.Map["siteservernode"])
			require.Equal(t, "SITE-SERVER-NAME", computer.Properties.Map["siteservernodename"])
		})

		t.Run("site server includes linked computer properties", func(t *testing.T) {
			siteServer, err := graphQuery.GetADEntityDetails(context.Background(), "SITE-SERVER-1", ad.SiteServer)
			require.NoError(t, err)
			require.Equal(t, "COMPUTER-1", siteServer.Properties.Map["serverreferencecomputer"])
			require.Equal(t, "COMPUTER-NAME", siteServer.Properties.Map["serverreferencecomputername"])
		})

		t.Run("unlinked entity remains undecorated", func(t *testing.T) {
			computer, err := graphQuery.GetADEntityDetails(context.Background(), "UNLINKED-COMPUTER-1", ad.Computer)
			require.NoError(t, err)
			_, hasSiteServerNode := computer.Properties.Map["siteservernode"]
			require.False(t, hasSiteServerNode)
		})
	})
}
