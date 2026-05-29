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
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestParseOptionalAssetGroupTagIds(t *testing.T) {
	t.Parallel()

	type args struct {
		tagIdParams []string
	}

	type want struct {
		tagIds []int
		err    bool
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(mockDB *mocks.MockDatabase)
		want      want
	}{
		{
			name:      "empty tagIdParams returns nil",
			args:      args{tagIdParams: []string{}},
			setupMock: func(mockDB *mocks.MockDatabase) {},
			want:      want{tagIds: nil, err: false},
		},
		{
			name:      "non-numeric string returns error",
			args:      args{tagIdParams: []string{"bad"}},
			setupMock: func(mockDB *mocks.MockDatabase) {},
			want:      want{tagIds: nil, err: true},
		},
		{
			name:      "tagIdParam 0 returns hygiene id",
			args:      args{tagIdParams: []string{"0"}},
			setupMock: func(mockDB *mocks.MockDatabase) {},
			want:      want{tagIds: []int{model.AssetGroupTierHygienePlaceholderId}, err: false},
		},
		{
			name: "valid tagIdParam returns that ID",
			args: args{tagIdParams: []string{"5"}},
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{ID: 5}, nil)
			},
			want: want{tagIds: []int{5}, err: false},
		},
		{
			name: "multiple valid tagIdParams returns all IDs",
			args: args{tagIdParams: []string{"5", "7"}},
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{ID: 5}, nil)
				mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 7).Return(model.AssetGroupTag{ID: 7}, nil)
			},
			want: want{tagIds: []int{5, 7}, err: false},
		},
		{
			name: "duplicate tagIdParams are deduplicated",
			args: args{tagIdParams: []string{"5", "5"}},
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 5).Return(model.AssetGroupTag{ID: 5}, nil)
			},
			want: want{tagIds: []int{5}, err: false},
		},
		{
			name: "tagIdParam that does not exist in DB returns error",
			args: args{tagIdParams: []string{"99"}},
			setupMock: func(mockDB *mocks.MockDatabase) {
				mockDB.EXPECT().GetAssetGroupTag(gomock.Any(), 99).Return(model.AssetGroupTag{}, errors.New("not found"))
			},
			want: want{tagIds: nil, err: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockDB := mocks.NewMockDatabase(ctrl)

			tt.setupMock(mockDB)

			tagIds, err := api.ParseOptionalAssetGroupTagIds(context.Background(), mockDB, tt.args.tagIdParams)
			if tt.want.err {
				require.Error(t, err)
				assert.Equal(t, tt.want.tagIds, tagIds)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.tagIds, tagIds)
			}
		})
	}
}
