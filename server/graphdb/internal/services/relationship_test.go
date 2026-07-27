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
	"encoding/json"
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
		name            string
		includeKindInfo bool
		setupMock       func(databaseMock *mocks.MockDatabase)
		wantResult      services.Relationship
		wantErr         error
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
			name:            "fetches filters and sorts relationship kind infos when requested",
			includeKindInfo: true,
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(baseRelationship, nil)
				databaseMock.EXPECT().GetKindByName(ctx, kindName).Return(resolvedKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, kindName).Return([]services.KindInfo{
					{
						InfoKey:            "later",
						Title:              "Beta",
						Position:           1,
						RelationshipKindID: int32Ptr(42),
						Content:            json.RawMessage(`{"markdown":{"content":"later"}}`),
					},
					{
						InfoKey:    "node_info",
						Title:      "Node Info",
						Position:   0,
						NodeKindID: int32Ptr(99),
						Content:    json.RawMessage(`{"markdown":{"content":"node"}}`),
					},
					{
						InfoKey:            "first",
						Title:              "Gamma",
						Position:           0,
						RelationshipKindID: int32Ptr(42),
						Content:            json.RawMessage(`{"markdown":{"content":"first"}}`),
					},
					{
						InfoKey:            "middle",
						Title:              "Alpha",
						Position:           1,
						RelationshipKindID: int32Ptr(42),
						Content:            json.RawMessage(`{"markdown":{"content":"middle"}}`),
					},
				}, nil)
			},
			wantResult: services.Relationship{
				ID:           relationshipID,
				SourceNodeID: 100,
				TargetNodeID: 200,
				Kind:         resolvedKind,
				Properties:   map[string]any{"foo": "bar"},
				KindInfos: []services.KindInfo{
					{
						InfoKey:            "first",
						Title:              "Gamma",
						Position:           0,
						RelationshipKindID: int32Ptr(42),
						Content:            json.RawMessage(`{"markdown":{"content":"first"}}`),
					},
					{
						InfoKey:            "middle",
						Title:              "Alpha",
						Position:           1,
						RelationshipKindID: int32Ptr(42),
						Content:            json.RawMessage(`{"markdown":{"content":"middle"}}`),
					},
					{
						InfoKey:            "later",
						Title:              "Beta",
						Position:           1,
						RelationshipKindID: int32Ptr(42),
						Content:            json.RawMessage(`{"markdown":{"content":"later"}}`),
					},
				},
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
			name:            "does not fetch relationship kind info when the resolved kind id is nil",
			includeKindInfo: true,
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
			name: "returns relationship with unresolved kind when kind not found in schema",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(baseRelationship, nil)
				databaseMock.EXPECT().GetKindByName(ctx, kindName).Return(services.Kind{}, services.ErrKindNotFound)
			},
			wantResult: services.Relationship{
				ID:           relationshipID,
				SourceNodeID: 100,
				TargetNodeID: 200,
				Kind:         services.Kind{Name: kindName, ID: nil},
				Properties:   map[string]any{"foo": "bar"},
			},
		},
		{
			name: "propagates unexpected database errors",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(services.Relationship{}, unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
		{
			name:            "propagates relationship kind info errors",
			includeKindInfo: true,
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationship(ctx, relationshipID).Return(baseRelationship, nil)
				databaseMock.EXPECT().GetKindByName(ctx, kindName).Return(resolvedKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, kindName).Return(nil, unexpectedErr)
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

			result, err := svc.GetRelationship(ctx, relationshipID, tt.includeKindInfo)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
