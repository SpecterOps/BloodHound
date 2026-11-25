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

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/handlers"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlers_ListExtensions(t *testing.T) {
	t.Parallel()
	var (
		readPermissions = model.Permissions{auth.Permissions().OpenGraphRead}
		tests           = []struct {
			name             string
			permissions      model.Permissions
			unauthenticated  bool
			setupMock        func(context.Context, *mocks.MockOpenGraphSchema)
			wantStatus       int
			wantBody         string
			wantErrorMessage string
		}{
			{
				name:        "extensions preserve response shape and display names",
				permissions: readPermissions,
				setupMock: func(ctx context.Context, serviceMock *mocks.MockOpenGraphSchema) {
					serviceMock.EXPECT().ListExtensions(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]services.Extension{
						{ID: 1, Name: "InternalAD", DisplayName: "Active Directory", Namespace: "AD", Version: "v1", IsBuiltin: true, CreatedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)},
						{ID: 2, Name: "InternalCustom", DisplayName: "Custom", Namespace: "CUSTOM", Version: "v2", CreatedAt: time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, time.January, 4, 0, 0, 0, 0, time.UTC)},
					}, 2, nil).Once()
				},
				wantStatus: http.StatusOK,
				wantBody:   `{"data":{"extensions":[{"id":1,"name":"Active Directory","namespace":"AD","version":"v1","is_builtin":true,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z"},{"id":2,"name":"Custom","namespace":"CUSTOM","version":"v2","is_builtin":false,"created_at":"2026-01-03T00:00:00Z","updated_at":"2026-01-04T00:00:00Z"}]}}`,
			},
			{
				name:        "OpenGraph read permission allows empty results",
				permissions: readPermissions,
				setupMock: func(ctx context.Context, serviceMock *mocks.MockOpenGraphSchema) {
					serviceMock.EXPECT().ListExtensions(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]services.Extension{}, 0, nil).Once()
				},
				wantStatus: http.StatusOK,
				wantBody:   `{"data":{"extensions":[]}}`,
			},
			{
				name:        "nil results encode empty array",
				permissions: readPermissions,
				setupMock: func(ctx context.Context, serviceMock *mocks.MockOpenGraphSchema) {
					serviceMock.EXPECT().ListExtensions(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, nil).Once()
				},
				wantStatus: http.StatusOK,
				wantBody:   `{"data":{"extensions":[]}}`,
			},
			{
				name:        "service error",
				permissions: readPermissions,
				setupMock: func(ctx context.Context, serviceMock *mocks.MockOpenGraphSchema) {
					serviceMock.EXPECT().ListExtensions(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, errors.New("private database detail")).Once()
				},
				wantStatus:       http.StatusInternalServerError,
				wantErrorMessage: "an internal error has occurred that is preventing the service from servicing this request",
			},
			{
				name:        "other read permission is forbidden",
				permissions: model.Permissions{auth.Permissions().GraphDBRead},
				setupMock: func(ctx context.Context, serviceMock *mocks.MockOpenGraphSchema) {
				},
				wantStatus:       http.StatusForbidden,
				wantErrorMessage: "not authorized",
			},
			{
				name:            "unauthenticated request",
				unauthenticated: true,
				setupMock: func(ctx context.Context, serviceMock *mocks.MockOpenGraphSchema) {
				},
				wantStatus:       http.StatusUnauthorized,
				wantErrorMessage: "not authenticated",
			},
		}
	)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				serviceMock    = mocks.NewMockOpenGraphSchema(t)
				handlerSet     = handlers.NewHandlersContainer(serviceMock)
				recorder       = httptest.NewRecorder()
				requestContext = &bhctx.Context{}
			)
			if !testCase.unauthenticated {
				requestContext.AuthCtx = auth.Context{Owner: model.User{
					Unique: model.Unique{ID: uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")},
					Roles:  model.Roles{{Permissions: testCase.permissions}},
				}}
			}
			var (
				request         = bhctx.SetRequestContext(httptest.NewRequest(http.MethodGet, "/api/v2/extensions", nil), requestContext)
				permissionCheck = middleware.PermissionsCheckAll(auth.NewAuthorizer(nil), auth.Permissions().OpenGraphRead)
				handler         = permissionCheck(http.HandlerFunc(handlerSet.ListExtensions))
			)
			testCase.setupMock(request.Context(), serviceMock)
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, testCase.wantStatus, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			if testCase.wantBody != "" {
				assert.JSONEq(t, testCase.wantBody, recorder.Body.String())
			} else {
				var body responses.ErrorWrapper
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
				assert.Equal(t, testCase.wantStatus, body.HTTPStatus)
				require.Len(t, body.Errors, 1)
				assert.Empty(t, body.Errors[0].Context)
				assert.Equal(t, testCase.wantErrorMessage, body.Errors[0].Message)
				assert.NotContains(t, recorder.Body.String(), "private database detail")
			}
		})
	}
}
