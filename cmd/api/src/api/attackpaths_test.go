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
	"strconv"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/stretchr/testify/assert"
)

func TestParseAssetGroupTagIdParams(t *testing.T) {
	t.Parallel()

	type want struct {
		res []int
		err error
	}

	tests := []struct {
		name   string
		params []string
		want   want
	}{
		{
			name:   "Success: empty params returns empty slice",
			params: []string{},
			want:   want{res: []int{}, err: nil},
		},
		{
			name:   "Success: single valid integer",
			params: []string{"5"},
			want:   want{res: []int{5}, err: nil},
		},
		{
			name:   "Success: multiple valid integers",
			params: []string{"5", "7"},
			want:   want{res: []int{5, 7}, err: nil},
		},
		{
			name:   "Error: non-numeric string returns error",
			params: []string{"bad"},
			want:   want{res: nil, err: strconv.ErrSyntax},
		},
		{
			name:   "Error: first valid then non-numeric returns error",
			params: []string{"5", "bad"},
			want:   want{res: nil, err: strconv.ErrSyntax},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tagIds, err := api.ParseAssetGroupTagIdParams(tt.params)
			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.res, tagIds)
		})
	}
}


