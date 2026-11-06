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

package users_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/users"
	"github.com/stretchr/testify/assert"
)

func TestRequestContextResolver_Resolve(t *testing.T) {
	type expected struct {
		principal users.Principal
		isUser    bool
	}

	type testData struct {
		name     string
		buildCtx func() context.Context
		expected expected
	}

	var (
		userID = uuid.Must(uuid.NewV4())
		tt     = []testData{
			{
				name: "Success: authenticated administrator is translated to a principal",
				buildCtx: func() context.Context {
					return bhctx.Set(context.Background(), &bhctx.Context{AuthCtx: auth.Context{Owner: model.User{
						Unique:          model.Unique{ID: userID},
						PrincipalName:   "admin@example.com",
						EmailAddress:    null.StringFrom("admin@example.com"),
						Roles:           model.Roles{{Name: auth.RoleAdministrator}},
						AllEnvironments: true,
					}}})
				},
				expected: expected{
					principal: users.Principal{
						ID:              userID,
						PrincipalName:   "admin@example.com",
						EmailAddress:    "admin@example.com",
						Administrator:   true,
						AllEnvironments: true,
					},
					isUser: true,
				},
			},
			{
				name: "Error: context without an authenticated user is rejected",
				buildCtx: func() context.Context {
					return context.Background()
				},
				expected: expected{principal: users.Principal{}, isUser: false},
			},
		}
	)

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			principal, isUser := users.NewResolver().Resolve(testCase.buildCtx())

			assert.Equal(t, testCase.expected.principal, principal)
			assert.Equal(t, testCase.expected.isUser, isUser)
		})
	}
}
