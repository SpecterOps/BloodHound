// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
		extID         = int32(7)
		unexpectedErr = errors.New("connection refused")
		extension     = services.Extension{ID: extID, Name: "TestExtension", DisplayName: "Test Extension", Namespace: "TST", IsBuiltin: true, Version: "1.0.0"}
		baseNodeKind  = services.NodeKind{ID: nodeKindID, Name: "User", DisplayName: "User", SchemaExtensionID: extID}
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
				databaseMock.EXPECT().GetKindInfos(ctx, "User").Return(infos, nil)
				databaseMock.EXPECT().GetExtension(ctx, extID).Return(extension, nil)
			},
			wantResult: services.NodeKind{ID: nodeKindID, Name: "User", DisplayName: "User", SchemaExtensionID: extID, Info: infos, Extension: extension},
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
				databaseMock.EXPECT().GetKindInfos(ctx, "User").Return(nil, unexpectedErr)
			},
			wantResult:      services.NodeKind{},
			wantErr:         unexpectedErr,
			wantErrContains: "fetching kind infos for node kind User",
		},
		{
			name: "error_-_wraps_extension_fetch_error",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNodeKind(ctx, nodeKindID).Return(baseNodeKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, "User").Return(infos, nil)
				databaseMock.EXPECT().GetExtension(ctx, extID).Return(services.Extension{}, unexpectedErr)
			},
			wantResult:      services.NodeKind{},
			wantErr:         unexpectedErr,
			wantErrContains: "fetching extension 7 for node kind 42",
		},
		{
			name: "success_-_attaches_empty_extension_when_extension_not_found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNodeKind(ctx, nodeKindID).Return(baseNodeKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, "User").Return(infos, nil)
				databaseMock.EXPECT().GetExtension(ctx, extID).Return(services.Extension{}, services.ErrExtensionNotFound)
			},
			wantResult: services.NodeKind{ID: nodeKindID, Name: "User", DisplayName: "User", SchemaExtensionID: extID, Info: infos, Extension: services.Extension{}},
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
