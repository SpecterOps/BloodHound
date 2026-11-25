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

	"github.com/specterops/bloodhound/server/savedqueries/internal/services"
	"github.com/specterops/bloodhound/server/savedqueries/internal/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_GetSavedQuery(t *testing.T) {
	type mock struct {
		database *mocks.MockDatabase
	}

	type expected struct {
		savedQuery services.SavedQuery
		err        error
	}

	type testData struct {
		name       string
		actor      services.Actor
		setupMocks func(mock)
		expected   expected
	}

	var (
		ctx           = context.Background()
		savedQueryID  = int64(42)
		ownerID       = "11111111-1111-1111-1111-111111111111"
		otherUserID   = "22222222-2222-2222-2222-222222222222"
		unexpectedErr = errors.New("database unavailable")
		savedQuery    = services.SavedQuery{ID: savedQueryID, UserID: ownerID, Name: "Owned query"}
	)

	tt := []testData{
		{
			name:  "Success: owner can read a private query - 200",
			actor: services.Actor{UserID: ownerID},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetSavedQuery(ctx, savedQueryID, ownerID).Return(savedQuery, services.SavedQueryAccess{}, nil)
			},
			expected: expected{savedQuery: savedQuery},
		},
		{
			name:  "Success: administrator can read another user's private query - 200",
			actor: services.Actor{UserID: otherUserID, IsAdministrator: true},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetSavedQuery(ctx, savedQueryID, otherUserID).Return(savedQuery, services.SavedQueryAccess{}, nil)
			},
			expected: expected{savedQuery: savedQuery},
		},
		{
			name:  "Success: user can read a public or shared query - 200",
			actor: services.Actor{UserID: otherUserID},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetSavedQuery(ctx, savedQueryID, otherUserID).Return(savedQuery, services.SavedQueryAccess{PublicOrShared: true}, nil)
			},
			expected: expected{savedQuery: savedQuery},
		},
		{
			name:  "Error: inaccessible private query is hidden - 404",
			actor: services.Actor{UserID: otherUserID},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetSavedQuery(ctx, savedQueryID, otherUserID).Return(savedQuery, services.SavedQueryAccess{}, nil)
			},
			expected: expected{err: services.ErrSavedQueryNotFound},
		},
		{
			name:  "Error: missing query is propagated - 404",
			actor: services.Actor{UserID: otherUserID},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetSavedQuery(ctx, savedQueryID, otherUserID).Return(services.SavedQuery{}, services.SavedQueryAccess{}, services.ErrSavedQueryNotFound)
			},
			expected: expected{err: services.ErrSavedQueryNotFound},
		},
		{
			name:  "Error: database failure is propagated - 500",
			actor: services.Actor{UserID: otherUserID},
			setupMocks: func(mock mock) {
				mock.database.EXPECT().GetSavedQuery(ctx, savedQueryID, otherUserID).Return(services.SavedQuery{}, services.SavedQueryAccess{}, unexpectedErr)
			},
			expected: expected{err: unexpectedErr},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				databaseMock = mocks.NewMockDatabase(t)
				service      = services.NewService(databaseMock)
			)

			testCase.setupMocks(mock{database: databaseMock})
			actual, err := service.GetSavedQuery(ctx, savedQueryID, testCase.actor)

			assert.Equal(t, testCase.expected.savedQuery, actual)
			assert.ErrorIs(t, err, testCase.expected.err)
			if testCase.expected.err == nil {
				assert.NoError(t, err)
			}
		})
	}
}
