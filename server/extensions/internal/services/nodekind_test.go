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
package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/specterops/bloodhound/server/extensions/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetNodeKind(t *testing.T) {
	var (
		ctx           = context.Background()
		nodeKindID    = int32(42)
		unexpectedErr = errors.New("connection refused")
		baseNodeKind  = services.NodeKind{ID: nodeKindID, Name: "User", DisplayName: "User"}
		infos         = []services.KindInfo{{InfoKey: "panel1", Title: "Alpha", Position: 0, NodeKindID: &nodeKindID, Name: "User"}}
	)

	tests := []struct {
		name            string
		setupMock       func(databaseMock *mocks.MockDatabase)
		wantResult      services.NodeKind
		wantErr         error
		wantErrContains string
	}{
		{
			name: "success_-_attaches_infos_to_node_kind",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNodeKind(ctx, nodeKindID).Return(baseNodeKind, nil)
				databaseMock.EXPECT().GetKindInfosByNodeKindID(ctx, nodeKindID).Return(infos, nil)
			},
			wantResult: services.NodeKind{ID: nodeKindID, Name: "User", DisplayName: "User", Info: infos},
		},
		{
			name: "error_-_propagates_node_kind_not_found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNodeKind(ctx, nodeKindID).Return(services.NodeKind{}, services.ErrNodeKindNotFound)
			},
			wantResult: services.NodeKind{},
			wantErr:    services.ErrNodeKindNotFound,
		},
		{
			name: "error_-_wraps_info_fetch_error",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNodeKind(ctx, nodeKindID).Return(baseNodeKind, nil)
				databaseMock.EXPECT().GetKindInfosByNodeKindID(ctx, nodeKindID).Return(nil, unexpectedErr)
			},
			wantResult:      services.NodeKind{},
			wantErr:         unexpectedErr,
			wantErrContains: "fetching kind infos for node kind 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			databaseMock := mocks.NewMockDatabase(t)
			tt.setupMock(databaseMock)
			svc := services.NewService(databaseMock)

			result, err := svc.GetNodeKind(ctx, nodeKindID)
			assert.Equal(t, tt.wantResult, result)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
