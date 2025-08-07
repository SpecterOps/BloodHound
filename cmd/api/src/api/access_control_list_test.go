// Copyright 2025 Specter Ops, Inc.
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

package api_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_CheckAccessToEnvironments(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
	}
	type testData struct {
		name       string
		setupMocks func(t *testing.T, mock *mock)
		input      []string
		expected   bool
	}

	userUuid, err := uuid.NewV4()

	require.NoError(t, err)

	testCases := []testData{
		{
			name: "Positive Test - Exact Match",
			setupMocks: func(t *testing.T, mock *mock) {
				envs := database.EnvironmentAccessList{
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "1",
					},
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "2",
					},
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "3",
					},
				}

				mock.mockDatabase.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), userUuid).Return(envs, nil)
			},
			input:    []string{"1", "2", "3"},
			expected: true,
		},
		{
			name: "Positive Test - Partial Match",
			setupMocks: func(t *testing.T, mock *mock) {
				envs := database.EnvironmentAccessList{
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "1",
					},
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "2",
					},
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "3",
					},
				}

				mock.mockDatabase.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), userUuid).Return(envs, nil)
			},
			input:    []string{"1", "2"},
			expected: true,
		},
		{
			name: "Negative Test - Extra Environment",
			setupMocks: func(t *testing.T, mock *mock) {
				envs := database.EnvironmentAccessList{
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "1",
					},
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "2",
					},
					database.EnvironmentAccess{
						UserID:      "lorem ipsum",
						Environment: "3",
					},
				}

				mock.mockDatabase.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), userUuid).Return(envs, nil)
			},
			input:    []string{"1", "2", "4"},
			expected: false,
		},
		{
			name: "Negative Test - No Allowed Environments",
			setupMocks: func(t *testing.T, mock *mock) {
				envs := database.EnvironmentAccessList{}

				mock.mockDatabase.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), userUuid).Return(envs, nil)
			},
			input:    []string{"1", "2", "4"},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
			}

			testCase.setupMocks(t, mocks)

			actual, err := api.CheckUserAccessToEnvironments(context.Background(), mocks.mockDatabase, userUuid, testCase.input...)

			require.NoError(t, err)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
