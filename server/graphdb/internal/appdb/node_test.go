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
//go:generate go run go.uber.org/mock/mockgen -destination=./mocks/graph_reader.go -package=mocks -source=store.go graphReader

package appdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb"
	"github.com/specterops/bloodhound/server/graphdb/internal/appdb/mocks"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStore_GetNode(t *testing.T) {
	t.Parallel()

	var (
		ctx    = context.Background()
		nodeID = int64(12345)
		txErr  = errors.New("connection refused")
	)

	tests := []struct {
		name      string
		setupMock func(mockGraph *mocks.MockgraphReader)
		wantErr   error
	}{
		{
			name: "error_-_maps_graph_not_found_to_ErrNodeNotFound",
			setupMock: func(mockGraph *mocks.MockgraphReader) {
				mockGraph.EXPECT().
					ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(graph.ErrNoResultsFound)
			},
			wantErr: services.ErrNodeNotFound,
		},
		{
			name: "error_-_propagates_transaction_error",
			setupMock: func(mockGraph *mocks.MockgraphReader) {
				mockGraph.EXPECT().
					ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(txErr)
			},
			wantErr: txErr,
		},
		{
			name: "error_-_returns_ErrNodeNotFound_when_node_is_nil",
			setupMock: func(mockGraph *mocks.MockgraphReader) {
				mockGraph.EXPECT().
					ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: services.ErrNodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGraph := mocks.NewMockgraphReader(ctrl)
			tt.setupMock(mockGraph)

			store := appdb.NewStore(mockGraph, nil)
			_, err := store.GetNode(ctx, nodeID)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestStore_GetNodeKindsByNames(t *testing.T) {
	t.Parallel()

	var (
		ctx         = context.Background()
		kindNames   = []string{"User", "Group"}
		expectedSQL = `SELECT nk.id, k.name FROM schema_node_kinds AS nk JOIN kind AS k ON nk.kind_id = k.id WHERE k.name IN ($1, $2)`
		dbErr       = errors.New("db error")
	)

	tests := []struct {
		name        string
		inputNames  []string
		setupMock   func(pool pgxmock.PgxPoolIface)
		wantKinds   int
		wantErr     error
		checkResult func(t *testing.T, kinds []services.Kind)
	}{
		{
			name:       "success_-_returns_all_kinds",
			inputNames: kindNames,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				var (
					id1 = int32(1)
					id2 = int32(2)
				)
				pool.ExpectQuery(expectedSQL).
					WithArgs("User", "Group").
					WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
						AddRow(&id1, "User").
						AddRow(&id2, "Group"))
			},
			wantKinds: 2,
			checkResult: func(t *testing.T, kinds []services.Kind) {
				t.Helper()
				require.NotNil(t, kinds[0].ID)
				assert.Equal(t, int32(1), *kinds[0].ID)
				assert.Equal(t, "User", kinds[0].Name)
				require.NotNil(t, kinds[1].ID)
				assert.Equal(t, int32(2), *kinds[1].ID)
				assert.Equal(t, "Group", kinds[1].Name)
			},
		},
		{
			name:       "success_-_returns_unregistered_kinds_with_nil_id",
			inputNames: kindNames,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				id1 := int32(1)
				pool.ExpectQuery(expectedSQL).
					WithArgs("User", "Group").
					WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
						AddRow(&id1, "User"))
			},
			wantKinds: 2,
			checkResult: func(t *testing.T, kinds []services.Kind) {
				t.Helper()
				require.NotNil(t, kinds[0].ID)
				assert.Equal(t, int32(1), *kinds[0].ID)
				assert.Equal(t, "User", kinds[0].Name)
				assert.Nil(t, kinds[1].ID)
				assert.Equal(t, "Group", kinds[1].Name)
			},
		},
		{
			name:       "error_-_propagates_database_error",
			inputNames: kindNames,
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSQL).
					WithArgs("User", "Group").
					WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:       "success_-_returns_empty_slice_for_empty_input",
			inputNames: []string{},
			setupMock:  func(pool pgxmock.PgxPoolIface) {},
			wantKinds:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer pool.Close()

			tt.setupMock(pool)

			store := appdb.NewStore(nil, pool)
			kinds, err := store.GetNodeKindsByNames(ctx, tt.inputNames)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Len(t, kinds, tt.wantKinds)
				if tt.checkResult != nil {
					tt.checkResult(t, kinds)
					pool.ExpectationsWereMet()
				}
			}
		})
	}
}
