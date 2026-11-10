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

type user struct{}

func (s user) GetID() uuid.UUID {
	return uuid.Nil
}

func (s user) HasAllEnvironments() bool {
	return false
}

type administratorUser struct {
	user
	administrator bool
}

func (s administratorUser) IsAdministrator() bool {
	return s.administrator
}

func TestRequestContextResolver_Resolve(t *testing.T) {
	type expected struct {
		principal users.AuthenticatedUser
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
					principal: users.AuthenticatedUser{
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
				expected: expected{principal: users.AuthenticatedUser{}, isUser: false},
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

func TestIsAdministrator(t *testing.T) {
	type testData struct {
		name     string
		user     users.User
		expected bool
	}

	var tt = []testData{
		{
			name:     "Success: administrator capability returns true",
			user:     administratorUser{administrator: true},
			expected: true,
		},
		{
			name:     "Success: non-administrator capability returns false",
			user:     administratorUser{administrator: false},
			expected: false,
		},
		{
			name:     "Success: user without administrator capability returns false",
			user:     user{},
			expected: false,
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, users.IsAdministrator(testCase.user))
		})
	}
}
