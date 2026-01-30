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

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSupportsETACMiddleware(t *testing.T) {

	var (
		mockCtrl = gomock.NewController(t)
		mockDB   = mocks.NewMockDatabase(mockCtrl)
	)

	defer mockCtrl.Finish()

	tests := []struct {
		name             string
		setupMocks       func()
		bhCtx            ctx.Context
		expectedCode     int
		expectNextHit    bool
		dogTagsOverrides dogtags.TestOverrides
	}{
		{
			name: "Success ETAC disabled",
			setupMocks: func() {
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: false,
				},
			},
		},
		{
			name: "Success All Environments enabled",
			setupMocks: func() {
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:                  true,
						EnvironmentTargetedAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
		},
		{
			name: "Success All Environments disabled and user does have domain in etac list",
			setupMocks: func() {
				mockDB.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				}, nil)
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:                  false,
						EnvironmentTargetedAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
		},
		{
			name: "Error checking for environments on a user",
			setupMocks: func() {
				mockDB.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{{}}, errors.New("an error"))
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:                  false,
						EnvironmentTargetedAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedCode:  http.StatusInternalServerError,
			expectNextHit: false,
		},
		{
			name: "Error All Environments disabled and user does not have domain in etac list",
			setupMocks: func() {
				mockDB.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentTargetedAccessControl{{}}, nil)
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:                  false,
						EnvironmentTargetedAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedCode:  http.StatusForbidden,
			expectNextHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			nextHit := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHit = true
				w.WriteHeader(http.StatusOK)
			})

			handler := SupportsETACMiddleware(mockDB, dogtags.NewTestService(tt.dogTagsOverrides))(next)

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test/12345", nil)
			req = ctx.SetRequestContext(req, &tt.bhCtx)
			req = mux.SetURLVars(req, map[string]string{
				api.URIPathVariableObjectID: "12345",
			})

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectNextHit, nextHit)
		})
	}
}

func TestRequireAllEnvironmentAccessMiddleware(t *testing.T) {

	var (
		mockCtrl = gomock.NewController(t)
	)

	defer mockCtrl.Finish()

	tests := []struct {
		name             string
		setupMocks       func()
		bhCtx            ctx.Context
		expectedCode     int
		expectNextHit    bool
		dogTagsOverrides dogtags.TestOverrides
	}{
		{
			name: "Success ETAC disabled",
			setupMocks: func() {
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: false,
				},
			},
		},
		{
			name: "Success All Environments enabled",
			setupMocks: func() {
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:                  true,
						EnvironmentTargetedAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
		},
		{
			name: "Fail If All Environments is false",
			setupMocks: func() {
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:                  false,
						EnvironmentTargetedAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			expectedCode:  http.StatusForbidden,
			expectNextHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			nextHit := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHit = true
				w.WriteHeader(http.StatusOK)
			})

			handler := RequireAllEnvironmentAccessMiddleware(dogtags.NewTestService(tt.dogTagsOverrides))(next)

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test/12345", nil)
			req = ctx.SetRequestContext(req, &tt.bhCtx)
			req = mux.SetURLVars(req, map[string]string{
				api.URIPathVariableObjectID: "12345",
			})

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectNextHit, nextHit)
		})
	}
}
