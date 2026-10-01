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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/server/identity/internal/services"
	"github.com/specterops/bloodhound/server/identity/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestService_UserPreferences(t *testing.T) {
	var (
		ctx           = context.Background()
		userID        = uuid.Must(uuid.NewV4())
		database      = mocks.NewMockDatabase(t)
		service       = services.NewService(database)
		storage       = []byte{0, 255, 128, 10}
		databaseError = errors.New("database unavailable")
	)
	database.EXPECT().GetUserPreferences(ctx, userID).Return(storage, nil).Once()
	result, err := service.GetUserPreferences(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, storage, result)

	for _, payload := range [][]byte{nil, storage, make([]byte, services.MaxUserPreferencesBytes)} {
		database.EXPECT().UpsertUserPreferences(ctx, userID, payload).Return(nil).Once()
		require.NoError(t, service.UpsertUserPreferences(ctx, userID, payload))
	}
	require.ErrorIs(t, service.UpsertUserPreferences(ctx, userID, make([]byte, services.MaxUserPreferencesBytes+1)), services.ErrUserPreferencesTooLarge)

	database.EXPECT().GetUserPreferences(ctx, userID).Return(nil, services.ErrUserNotFound).Once()
	_, err = service.GetUserPreferences(ctx, userID)
	require.ErrorIs(t, err, services.ErrUserNotFound)
	database.EXPECT().UpsertUserPreferences(ctx, userID, storage).Return(databaseError).Once()
	require.ErrorIs(t, service.UpsertUserPreferences(ctx, userID, storage), databaseError)
}
