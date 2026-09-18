// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

func TestService_GetRelationshipKind(t *testing.T) {
	var (
		ctx                = context.Background()
		relationshipKindID = int32(42)
		extension          = services.Extension{
			ID:          7,
			Name:        "TestExtension",
			DisplayName: "Test Extension",
			Namespace:   "TST",
			Version:     "1.0.0",
		}
		baseRelationshipKind = services.RelationshipKind{ID: relationshipKindID, Name: "MemberOf", Extension: extension}
		infos                = []services.KindInfo{{InfoKey: "panel1", Title: "Alpha", Position: 0, RelationshipKindID: &relationshipKindID, Name: "MemberOf"}}
		unexpectedErr        = errors.New("connection refused")
	)

	tests := []struct {
		name       string
		setupMock  func(databaseMock *mocks.MockDatabase)
		wantResult services.RelationshipKind
		wantErr    error
	}{
		{
			name: "attaches infos to relationship kind",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationshipKind(ctx, relationshipKindID).Return(baseRelationshipKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, baseRelationshipKind.Name).Return(infos, nil)
				databaseMock.EXPECT().GetExtension(ctx, extension.ID).Return(extension, nil)
			},
			wantResult: services.RelationshipKind{ID: relationshipKindID, Name: "MemberOf", Info: infos, Extension: extension},
		},
		{
			name: "wraps extension fetch errors",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationshipKind(ctx, relationshipKindID).Return(baseRelationshipKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, baseRelationshipKind.Name).Return(infos, nil)
				databaseMock.EXPECT().GetExtension(ctx, extension.ID).Return(services.Extension{}, unexpectedErr)
			},
			wantResult: services.RelationshipKind{},
			wantErr:    unexpectedErr,
		},
		{
			name: "returns an empty extension when extension is not found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationshipKind(ctx, relationshipKindID).Return(baseRelationshipKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, baseRelationshipKind.Name).Return(infos, nil)
				databaseMock.EXPECT().GetExtension(ctx, extension.ID).Return(services.Extension{}, services.ErrExtensionNotFound)
			},
			wantResult: services.RelationshipKind{ID: relationshipKindID, Name: "MemberOf", Info: infos},
		},
		{
			name: "propagates relationship kind not found",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationshipKind(ctx, relationshipKindID).Return(services.RelationshipKind{}, services.ErrRelationshipKindNotFound)
			},
			wantResult: services.RelationshipKind{},
			wantErr:    services.ErrRelationshipKindNotFound,
		},
		{
			name: "wraps info fetch errors",
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetRelationshipKind(ctx, relationshipKindID).Return(baseRelationshipKind, nil)
				databaseMock.EXPECT().GetKindInfos(ctx, baseRelationshipKind.Name).Return(nil, unexpectedErr)
			},
			wantResult: services.RelationshipKind{},
			wantErr:    unexpectedErr,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			databaseMock := mocks.NewMockDatabase(t)
			testCase.setupMock(databaseMock)
			service := services.NewService(databaseMock)

			result, err := service.GetRelationshipKind(ctx, relationshipKindID)

			assert.Equal(t, testCase.wantResult, result)
			if testCase.wantErr != nil {
				assert.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
