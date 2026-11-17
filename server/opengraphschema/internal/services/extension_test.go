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

	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ListExtensions(t *testing.T) {
	t.Parallel()
	var (
		ctx           = context.Background()
		databaseError = errors.New("listing failed")
		extensions    = []services.Extension{{ID: 1, Name: "AD", DisplayName: "Active Directory", Version: "v1", Namespace: "AD", IsBuiltin: true}}
		tests         = []struct {
			name      string
			setupMock func(*mocks.MockDatabase)
			want      []services.Extension
			wantError error
		}{
			{name: "extensions", setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().ListExtensions(ctx).Return(extensions, nil).Once()
			}, want: extensions},
			{name: "empty", setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().ListExtensions(ctx).Return([]services.Extension{}, nil).Once()
			}, want: []services.Extension{}},
			{name: "database error", setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().ListExtensions(ctx).Return(nil, databaseError).Once()
			}, wantError: databaseError},
		}
	)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				databaseMock = mocks.NewMockDatabase(t)
				service      = services.NewService(databaseMock)
			)
			testCase.setupMock(databaseMock)
			result, err := service.ListExtensions(ctx)
			assert.Equal(t, testCase.want, result)
			if testCase.wantError != nil {
				assert.ErrorIs(t, err, testCase.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
