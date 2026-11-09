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

package hybrid_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/hybrid"
	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func newHostCorrelationDatabase(t *testing.T) graph.Database {
	t.Helper()
	var configuration, err = utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	var connectionSettings, parseErr = pgx.ParseConfig(configuration.Database.PostgreSQLConnectionString())
	require.NoError(t, parseErr)
	var postgresConfiguration = pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      connectionSettings.Host,
		Port:                      fmt.Sprint(connectionSettings.Port),
		User:                      connectionSettings.User,
		Password:                  connectionSettings.Password,
		Database:                  connectionSettings.Database,
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
	if strings.HasPrefix(connectionSettings.Host, "/") {
		postgresConfiguration.Host = ""
		postgresConfiguration.Options += "&host=" + url.QueryEscape(connectionSettings.Host)
	}
	var connection = pgtestdb.Custom(t, postgresConfiguration, pgtestdb.NoopMigrator{})
	configuration.Database.Connection = connection.URL()
	var pool, poolErr = dbpool.NewDawgsPool(configuration.Database)
	require.NoError(t, poolErr)

	var database, openErr = dawgs.Open(t.Context(), pg.DriverName, dawgs.Config{
		ConnectionString: connection.URL(),
		Pool:             pool,
	})
	require.NoError(t, openErr)
	t.Cleanup(func() { require.NoError(t, database.Close(context.Background())) })
	require.NoError(t, migrations.NewGraphMigrator(database).Migrate(t.Context()))
	require.NoError(t, database.AssertSchema(t.Context(), graphschema.DefaultGraphSchema()))
	return database
}

func ingestHostCorrelationFixture(t *testing.T, database graph.Database, indexes []int, useRawObjectIDs bool) []ein.GenericNode {
	t.Helper()
	var (
		fixture struct {
			Graph struct {
				Nodes []ein.GenericNode `json:"nodes"`
			} `json:"graph"`
		}
		selected []ein.GenericNode
	)
	var contents, err = os.ReadFile("testdata/hosts.json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(contents, &fixture))
	for _, index := range indexes {
		selected = append(selected, fixture.Graph.Nodes[index])
	}
	fixture.Graph.Nodes = selected
	contents, err = json.Marshal(fixture)
	require.NoError(t, err)
	var schema, schemaErr = payload.LoadSchema()
	require.NoError(t, schemaErr)

	require.NoError(t, database.BatchOperation(t.Context(), func(batch graph.Batch) error {
		var ingestContext = graphify.NewIngestContext(t.Context(),
			graphify.WithBatchUpdater(batch),
			graphify.WithEndpointResolver(endpoint.NewResolver(database)),
			graphify.WithUseRawObjectIDs(useRawObjectIDs),
		)
		var _, ingestErr = graphify.ReadFileForIngest(ingestContext, bytes.NewReader(contents), graphify.ReadOptions{
			IngestSchema:       schema,
			FileType:           model.FileTypeJson,
			RegisterSourceKind: func(graph.Kind) error { return nil }, // Fixture uses only built-in source kinds.
		})
		return ingestErr
	}))
	return selected
}

func fetchHostCorrelationEdges(t *testing.T, database graph.Database) map[[2]string]graph.ID {
	t.Helper()
	var edges = make(map[[2]string]graph.ID)
	require.NoError(t, database.ReadTransaction(t.Context(), func(tx graph.Transaction) error {
		var relationships, err = ops.FetchRelationships(tx.Relationships().Filter(query.Kind(query.Relationship(), common.SameHost)))
		if err != nil {
			return err
		}
		for _, relationship := range relationships {
			var start, end, fetchErr = ops.FetchRelationshipNodes(tx, relationship)
			if fetchErr != nil {
				return fetchErr
			}
			var (
				startObjectID, startErr = start.Properties.Get(common.ObjectID.String()).String()
				endObjectID, endErr     = end.Properties.Get(common.ObjectID.String()).String()
				key                     = [2]string{strings.ToUpper(startObjectID), strings.ToUpper(endObjectID)}
			)
			require.NoError(t, startErr)
			require.NoError(t, endErr)
			require.NotContains(t, edges, key, "duplicate correlation edge")
			require.False(t, relationship.Properties.Get(common.FirstSeen.String()).IsNil())
			edges[key] = relationship.ID
		}
		return nil
	}))
	return edges
}

func TestHostCorrelation(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		indexes []int
		pairs   [][2]int
	}{
		{name: "computer and VM", indexes: []int{0, 1}, pairs: [][2]int{{0, 1}}},
		{name: "VM and device", indexes: []int{1, 2}, pairs: [][2]int{{0, 1}}},
		{name: "device and computer", indexes: []int{2, 0}, pairs: [][2]int{{0, 1}}},
		{name: "all three", indexes: []int{0, 1, 2}, pairs: [][2]int{{0, 1}, {0, 2}, {1, 2}}},
		{name: "unmatched computer", indexes: []int{0}},
		{name: "unmatched VM", indexes: []int{1}},
		{name: "unmatched device", indexes: []int{2}},
	} {
		for _, useRawObjectIDs := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/rawObjectIDs=%t", testCase.name, useRawObjectIDs), func(t *testing.T) {
				var (
					database = newHostCorrelationDatabase(t)
					nodes    = ingestHostCorrelationFixture(t, database, testCase.indexes, useRawObjectIDs)
				)
				var stats, err = hybrid.PostHybrid(t.Context(), database)
				require.NoError(t, err)
				if len(testCase.pairs) > 0 {
					require.Contains(t, stats.RelationshipsCreated, common.SameHost)
					require.Equal(t, int32(len(testCase.pairs)*2), *stats.RelationshipsCreated[common.SameHost], "new edges must be submitted only once per run")
				}
				var firstRun = fetchHostCorrelationEdges(t, database)
				require.Len(t, firstRun, len(testCase.pairs)*2)
				for _, pair := range testCase.pairs {
					var (
						start = strings.ToUpper(nodes[pair[0]].ID)
						end   = strings.ToUpper(nodes[pair[1]].ID)
					)
					require.Contains(t, firstRun, [2]string{start, end})
					require.Contains(t, firstRun, [2]string{end, start})
				}
				_, err = hybrid.PostHybrid(t.Context(), database)
				require.NoError(t, err)
				require.Equal(t, firstRun, fetchHostCorrelationEdges(t, database), "repeat runs must preserve relationship IDs")
			})
		}
	}
}

func TestHostCorrelationRemovesStaleEdges(t *testing.T) {
	var database = newHostCorrelationDatabase(t)
	ingestHostCorrelationFixture(t, database, []int{0, 1, 2}, false)
	var _, err = hybrid.PostHybrid(t.Context(), database)
	require.NoError(t, err)
	require.Len(t, fetchHostCorrelationEdges(t, database), 6)

	require.NoError(t, database.WriteTransaction(t.Context(), func(tx graph.Transaction) error {
		var device, fetchErr = tx.Nodes().Filter(query.Kind(query.Node(), azure.Device)).First()
		if fetchErr != nil {
			return fetchErr
		}
		device.Properties.Set(azure.OnPremID.String(), "")
		device.Properties.Set("physical_ids", []string{})
		return tx.UpdateNode(device)
	}))
	_, err = hybrid.PostHybrid(t.Context(), database)
	require.NoError(t, err)
	var remaining = fetchHostCorrelationEdges(t, database)
	require.Len(t, remaining, 2, "computer/VM correlation survives removal of device identifiers")
	for pair := range remaining {
		require.NotContains(t, pair, "01234567-89AB-CDEF-0123-456789ABCDEF")
	}
}
