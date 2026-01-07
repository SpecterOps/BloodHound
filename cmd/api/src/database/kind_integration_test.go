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

//go:build integration

package database_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
)

// the v7.3.0 migration initializes the kind table with Tag_Tier_Zero, so we're
// simply testing the kind exists
func TestGetKindByName(t *testing.T) {
	type args struct {
		name string
	}
	type want struct {
		err  error
		kind model.Kind
	}
	tests := []struct {
		name  string
		args  args
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: Retrieves Kind Tag_Tier_Zero by name",
			args: args{
				name: "Tag_Tier_Zero",
			},
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			want: want{
				err: nil,
				kind: model.Kind{
					ID:   1,
					Name: "Tag_Tier_Zero",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			kind, err := testSuite.BHDatabase.GetKindByName(testSuite.Context, testCase.args.name)
			if testCase.want.err != nil {
				assert.EqualError(t, testCase.want.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want.kind, kind)
			}
		})
	}
}
