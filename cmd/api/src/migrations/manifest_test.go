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

package migrations_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	graphMocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func expectMigrationDataLookup(mockCtrl *gomock.Controller, mockGraphDB *graphMocks.MockDatabase, node *graph.Node, lookupErr error) {
	var (
		mockTransaction = graphMocks.NewMockTransaction(mockCtrl)
		mockNodeQuery   = graphMocks.NewMockNodeQuery(mockCtrl)
	)

	mockTransaction.EXPECT().Nodes().Return(mockNodeQuery)
	mockNodeQuery.EXPECT().Filterf(gomock.Any()).Return(mockNodeQuery)
	mockNodeQuery.EXPECT().First().Return(node, lookupErr)

	mockGraphDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
			return txDelegate(mockTransaction)
		},
	)
}

func newMigrationDataNode(migrationVersion version.Version) *graph.Node {
	return graph.NewNode(1, graph.AsProperties(map[string]any{
		"Major": migrationVersion.Major,
		"Minor": migrationVersion.Minor,
		"Patch": migrationVersion.Patch,
	}), common.MigrationData)
}

func TestRequiresMigration(t *testing.T) {
	t.Run("no migration data requires migration", func(t *testing.T) {
		var (
			mockCtrl    = gomock.NewController(t)
			mockGraphDB = graphMocks.NewMockDatabase(mockCtrl)
		)

		expectMigrationDataLookup(mockCtrl, mockGraphDB, nil, graph.ErrNoResultsFound)

		requiresMigration, err := migrations.RequiresMigration(context.Background(), mockGraphDB)
		require.NoError(t, err)
		require.True(t, requiresMigration)
	})

	t.Run("migration data at current version does not require migration", func(t *testing.T) {
		var (
			mockCtrl    = gomock.NewController(t)
			mockGraphDB = graphMocks.NewMockDatabase(mockCtrl)
		)

		expectMigrationDataLookup(mockCtrl, mockGraphDB, newMigrationDataNode(version.GetVersion()), nil)

		requiresMigration, err := migrations.RequiresMigration(context.Background(), mockGraphDB)
		require.NoError(t, err)
		require.False(t, requiresMigration)
	})
}
