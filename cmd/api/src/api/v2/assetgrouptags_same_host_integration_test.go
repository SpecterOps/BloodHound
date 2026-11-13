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

package v2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/api/dbpool"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	mocks_db "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newSameHostPreviewGraph provisions an isolated PostgreSQL graph so preview
// tests exercise real traversal and schema behavior rather than mocked results.
func newSameHostPreviewGraph(t *testing.T) graph.Database {
	t.Helper()
	var configuration, configurationErr = utils.LoadIntegrationTestConfig()
	require.NoError(t, configurationErr)
	var connectionSettings, parseErr = pgx.ParseConfig(configuration.Database.PostgreSQLConnectionString())
	require.NoError(t, parseErr)
	var postgresConfiguration = pgtestdb.Config{
		DriverName: "pgx", Host: connectionSettings.Host, Port: fmt.Sprint(connectionSettings.Port),
		User: connectionSettings.User, Password: connectionSettings.Password,
		Database: connectionSettings.Database, Options: "sslmode=disable", ForceTerminateConnections: true,
	}
	if strings.HasPrefix(connectionSettings.Host, "/") {
		postgresConfiguration.Host = ""
		postgresConfiguration.Options += "&host=" + url.QueryEscape(connectionSettings.Host)
	}
	var connection = pgtestdb.Custom(t, postgresConfiguration, pgtestdb.NoopMigrator{})
	configuration.Database.Connection = connection.URL()
	var pool, poolErr = dbpool.NewDawgsPool(configuration.Database)
	require.NoError(t, poolErr)
	var database, openErr = dawgs.Open(t.Context(), pg.DriverName, dawgs.Config{ConnectionString: connection.URL(), Pool: pool})
	require.NoError(t, openErr)
	t.Cleanup(func() { require.NoError(t, database.Close(context.Background())) })
	require.NoError(t, migrations.NewGraphMigrator(database).Migrate(t.Context()))
	require.NoError(t, database.AssertSchema(t.Context(), graphschema.DefaultGraphSchema()))
	return database
}

// TestPreviewSelectorsSameHostExpansion verifies the public handler returns the
// same identities and provenance as analysis selection, including the default
// full expansion, cross-platform parents, and preview limits.
func TestPreviewSelectorsSameHostExpansion(t *testing.T) {
	var (
		database            = newSameHostPreviewGraph(t)
		controller          = gomock.NewController(t)
		applicationDatabase = mocks_db.NewMockDatabase(controller)
		resources           = v2.Resources{DB: applicationDatabase, Graph: database}
		nodes               = make(map[string]*graph.Node)
	)
	applicationDatabase.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.AGTParameterKey).Return(appcfg.Parameter{}, nil).AnyTimes()
	applicationDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Return(graphschema.ValidKinds, nil).AnyTimes()
	// The bidirectional chain requires transitive host discovery. Its Azure
	// container makes full/default expansion distinguishable from children-only.
	require.NoError(t, database.WriteTransaction(t.Context(), func(transaction graph.Transaction) error {
		for name, kinds := range map[string]graph.Kinds{
			"computer": {ad.Entity, ad.Computer}, "vm": {azure.Entity, azure.VM},
			"device": {azure.Entity, azure.Device}, "resource-group": {azure.Entity, azure.ResourceGroup},
		} {
			var node, err = transaction.CreateNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: "preview-" + name, common.Name: name}), kinds...)
			if err != nil {
				return err
			}
			nodes[name] = node
		}
		for _, pair := range [][2]string{{"computer", "vm"}, {"vm", "computer"}, {"vm", "device"}, {"device", "vm"}} {
			if _, err := transaction.CreateRelationshipByIDs(nodes[pair[0]].ID, nodes[pair[1]].ID, common.SameHostAs, graph.NewProperties()); err != nil {
				return err
			}
		}
		var _, err = transaction.CreateRelationshipByIDs(nodes["resource-group"].ID, nodes["vm"].ID, azure.Contains, graph.NewProperties())
		return err
	}))
	for _, scenario := range []struct {
		name, seed string
		method     model.AssetGroupExpansionMethod
		useDefault bool
		limit      int
		expected   []string
	}{
		{"computer child expansion", "computer", model.AssetGroupExpansionMethodChildren, false, 10, []string{"computer", "vm", "device"}},
		{"VM child expansion", "vm", model.AssetGroupExpansionMethodChildren, false, 10, []string{"computer", "vm", "device"}},
		{"device child expansion", "device", model.AssetGroupExpansionMethodChildren, false, 10, []string{"computer", "vm", "device"}},
		{"full expansion includes Azure parent", "computer", model.AssetGroupExpansionMethodAll, false, 10, []string{"computer", "vm", "device", "resource-group"}},
		{"default expansion includes Azure parent", "computer", model.AssetGroupExpansionMethodAll, true, 10, []string{"computer", "vm", "device", "resource-group"}},
		{"no expansion", "computer", model.AssetGroupExpansionMethodNone, false, 10, []string{"computer"}},
		{"preview limit", "computer", model.AssetGroupExpansionMethodAll, false, 1, []string{"computer"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var body = v2.PreviewSelectorBody{Seeds: model.SelectorSeeds{{Type: model.SelectorTypeObjectId, Value: "preview-" + scenario.seed}}}
			if !scenario.useDefault {
				body.Expansion = &scenario.method
			}
			var payload, marshalErr = json.Marshal(body)
			require.NoError(t, marshalErr)
			var (
				request  = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v2/asset-group-tags/preview-selectors?limit=%d", scenario.limit), bytes.NewReader(payload)).WithContext(setupUserCtx(setupUser()))
				recorder = httptest.NewRecorder()
				response struct {
					Data v2.GetAssetGroupMembersResponse `json:"data"`
				}
				expected []graph.ID
				actual   []graph.ID
			)
			resources.PreviewSelectors(recorder, request)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			for _, name := range scenario.expected {
				expected = append(expected, nodes[name].ID)
			}
			for _, member := range response.Data.Members {
				actual = append(actual, member.NodeId)
				switch member.NodeId {
				case nodes[scenario.seed].ID:
					require.Equal(t, model.AssetGroupSelectorNodeSourceSeed, member.Source)
				case nodes["resource-group"].ID:
					require.Equal(t, model.AssetGroupSelectorNodeSourceParent, member.Source)
				default:
					require.Equal(t, model.AssetGroupSelectorNodeSourceChild, member.Source)
				}
			}
			require.ElementsMatch(t, expected, actual)
			// The public preview must return the same membership as analysis selection.
			var selected, errs = analysis.FetchNodesFromSeeds(t.Context(), appcfg.AGTParameters{DAWGsWorkerLimit: 2, ExpansionWorkerLimit: 2, SelectorWorkerLimit: 1}, database, body.Seeds, scenario.method, scenario.limit)
			require.Empty(t, errs)
			require.Len(t, selected, len(actual))
			for _, nodeID := range actual {
				require.Contains(t, selected, nodeID)
			}
		})
	}
}
