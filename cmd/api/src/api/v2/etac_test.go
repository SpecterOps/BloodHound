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

package v2_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_CheckAccessToEnvironments(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockDatabase *mocks.MockDatabase
		mockUser     model.User
	}
	type input struct {
		environments []string
		user         model.User
	}
	type testData struct {
		name       string
		setupMocks func(t *testing.T, mock *mock)
		input      input
		expected   bool
	}

	userUuid, err := uuid.NewV4()

	require.NoError(t, err)

	testCases := []testData{
		{
			name: "Positive Test - Exact Match",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				envs := []model.EnvironmentTargetedAccessControl{
					{
						UserID:        userUuid.String(),
						EnvironmentID: "1",
					},
					{
						UserID:        userUuid.String(),
						EnvironmentID: "2",
					},
					{
						UserID:        userUuid.String(),
						EnvironmentID: "3",
					},
				}

				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), mock.mockUser).Return(envs, nil)
			},
			input: input{
				environments: []string{"1", "2", "3"},
				user: model.User{
					Unique: model.Unique{
						ID: userUuid,
					},
					AllEnvironments: false,
				},
			},
			expected: true,
		},
		{
			name: "Positive Test - Partial Match",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				envs := []model.EnvironmentTargetedAccessControl{
					{
						UserID:        userUuid.String(),
						EnvironmentID: "1",
					},
					{
						UserID:        userUuid.String(),
						EnvironmentID: "2",
					},
					{
						UserID:        userUuid.String(),
						EnvironmentID: "3",
					},
				}

				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), mock.mockUser).Return(envs, nil)
			},
			input: input{
				environments: []string{"1", "2"},
				user: model.User{
					Unique: model.Unique{
						ID: userUuid,
					},
					AllEnvironments: false,
				},
			}, expected: true,
		},
		{
			name: "Negative Test - Extra Environment",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				envs := []model.EnvironmentTargetedAccessControl{
					{
						UserID:        userUuid.String(),
						EnvironmentID: "1",
					},
					{
						UserID:        userUuid.String(),
						EnvironmentID: "2",
					},
					{
						UserID:        userUuid.String(),
						EnvironmentID: "3",
					},
				}

				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), mock.mockUser).Return(envs, nil)
			},
			input: input{
				environments: []string{"1", "2", "4"},
				user: model.User{
					Unique: model.Unique{
						ID: userUuid,
					},
					AllEnvironments: false,
				},
			},
			expected: false,
		},
		{
			name: "Negative Test - No Allowed Environments",
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()

				envs := []model.EnvironmentTargetedAccessControl{}

				mock.mockDatabase.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), mock.mockUser).Return(envs, nil)
			},
			input: input{
				environments: []string{"1", "2", "4"},
				user: model.User{
					Unique: model.Unique{
						ID: userUuid,
					},
					AllEnvironments: false,
				},
			},
			expected: false,
		},
		{
			name:       "Positive Test - All Environments True",
			setupMocks: func(t *testing.T, mock *mock) {},
			input: input{
				environments: []string{"1", "2", "4"},
				user: model.User{
					Unique: model.Unique{
						ID: userUuid,
					},
					AllEnvironments: true,
				},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase: mocks.NewMockDatabase(ctrl),
				mockUser:     testCase.input.user,
			}

			testCase.setupMocks(t, mocks)

			actual, err := v2.CheckUserAccessToEnvironments(context.Background(), mocks.mockDatabase, testCase.input.user, testCase.input.environments...)

			require.NoError(t, err)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
