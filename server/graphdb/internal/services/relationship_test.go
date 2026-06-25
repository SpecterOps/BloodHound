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

func TestService_GetRelationship(t *testing.T) {
	var (
		ctx              = context.Background()
		relationshipID   = int64(1234567890)
		kindName         = "MemberOf"
		kindID           = int32(42)
		unexpectedErr    = errors.New("connection refused")
		baseRelationship = services.Relationship{
			ID:           relationshipID,
			SourceNodeID: 100,
			TargetNodeID: 200,
			Kind:         services.Kind{Name: kindName},
			Properties:   map[string]any{"foo": "bar"},
		}
		resolvedKind = services.Kind{ID: &kindID, Name: kindName}
		nilIDKind    = services.Kind{Name: kindName}
	)

	tests := []struct {
		name       string
		setupMock  func(databaseMock *mocks.MockDatabase)
		wantResult services.Relationship
		wantErr    error
	}{
		{
			name: "resolves the kind id and preserves endpoint node ids on success",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(baseRelationship, nil)
				databaseMock.EXPECT().GetKindByName(ctx, kindName).Return(resolvedKind, nil)
			},
			wantResult: services.Relationship{
				ID:           relationshipID,
				SourceNodeID: 100,
				TargetNodeID: 200,
				Kind:         resolvedKind,
				Properties:   map[string]any{"foo": "bar"},
			},
		},
		{
			name: "preserves a nil kind id when the kind has no schema_relationship_kinds entry",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(baseRelationship, nil)
				databaseMock.EXPECT().GetKindByName(ctx, kindName).Return(nilIDKind, nil)
			},
			wantResult: services.Relationship{
				ID:           relationshipID,
				SourceNodeID: 100,
				TargetNodeID: 200,
				Kind:         nilIDKind,
				Properties:   map[string]any{"foo": "bar"},
			},
		},
		{
			name: "propagates relationship not found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(services.Relationship{}, services.ErrRelationshipNotFound)
			},
			wantErr: services.ErrRelationshipNotFound,
		},
		{
			name: "propagates kind not found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(baseRelationship, nil)
				databaseMock.EXPECT().GetKindByName(ctx, kindName).Return(services.Kind{}, services.ErrKindNotFound)
			},
			wantErr: services.ErrKindNotFound,
		},
		{
			name: "propagates unexpected database errors",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(services.Relationship{}, unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			tt.setupMock(databaseMock)

			result, err := svc.GetRelationship(ctx, relationshipID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
