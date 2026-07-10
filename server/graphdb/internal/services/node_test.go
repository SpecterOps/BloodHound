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

	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/specterops/bloodhound/server/graphdb/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// int32Ptr is a helper function that returns a pointer to an int32 value.
func int32Ptr(v int32) *int32 {
	return &v
}

func TestService_GetNode(t *testing.T) {
	var (
		ctx           = context.Background()
		nodeID        = int64(9876543210)
		unexpectedErr = errors.New("connection refused")
		baseNode      = services.Node{
			ID: nodeID,
			Kinds: []services.Kind{
				{Name: "User"},
				{Name: "Group"},
			},
			Properties: map[string]any{"name": "admin"},
		}
		resolvedKinds = []services.Kind{
			{ID: int32Ptr(1), Name: "User"},
			{ID: int32Ptr(2), Name: "Group"},
		}
	)

	tests := []struct {
		name       string
		setupMock  func(databaseMock *mocks.MockDatabase)
		wantResult services.Node
		wantErr    error
	}{
		{
			name: "success_-_resolves_all_kind_ids",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNode(ctx, nodeID).Return(baseNode, nil)
				databaseMock.EXPECT().GetNodeKindsByNames(ctx, []string{"User", "Group"}).Return(resolvedKinds, nil)
			},
			wantResult: services.Node{
				ID:         nodeID,
				Kinds:      resolvedKinds,
				Properties: map[string]any{"name": "admin"},
			},
		},
		{
			name: "error_-_propagates_node_not_found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNode(ctx, nodeID).Return(services.Node{}, services.ErrNodeNotFound)
			},
			wantErr: services.ErrNodeNotFound,
		},
		{
			name: "error_-_propagates_unexpected_database_errors",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetNode(ctx, nodeID).Return(services.Node{}, unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
		{
			name: "success_-_handles_node_with_single_kind",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				singleKindNode := services.Node{
					ID:         nodeID,
					Kinds:      []services.Kind{{Name: "User"}},
					Properties: map[string]any{"name": "admin"},
				}
				databaseMock.EXPECT().GetNode(ctx, nodeID).Return(singleKindNode, nil)
				databaseMock.EXPECT().GetNodeKindsByNames(ctx, []string{"User"}).Return([]services.Kind{{ID: int32Ptr(1), Name: "User"}}, nil)
			},
			wantResult: services.Node{
				ID:         nodeID,
				Kinds:      []services.Kind{{ID: int32Ptr(1), Name: "User"}},
				Properties: map[string]any{"name": "admin"},
			},
		},
		{
			name: "success_-_handles_mixed_resolved_and_unresolved_kinds",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				mixedKindNode := services.Node{
					ID:         nodeID,
					Kinds:      []services.Kind{{Name: "User"}, {Name: "CustomKind"}},
					Properties: map[string]any{"name": "admin"},
				}
				mixedResolvedKinds := []services.Kind{
					{ID: int32Ptr(1), Name: "User"},
					{ID: nil, Name: "CustomKind"},
				}
				databaseMock.EXPECT().GetNode(ctx, nodeID).Return(mixedKindNode, nil)
				databaseMock.EXPECT().GetNodeKindsByNames(ctx, []string{"User", "CustomKind"}).Return(mixedResolvedKinds, nil)
			},
			wantResult: services.Node{
				ID: nodeID,
				Kinds: []services.Kind{
					{ID: int32Ptr(1), Name: "User"},
					{ID: nil, Name: "CustomKind"},
				},
				Properties: map[string]any{"name": "admin"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			tt.setupMock(databaseMock)

			result, err := svc.GetNode(ctx, nodeID, false)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
