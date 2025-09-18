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
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSupportsETACMiddleware(t *testing.T) {

	var (
		mockCtrl = gomock.NewController(t)
		mockDB   = dbmocks.NewMockDatabase(mockCtrl)
	)

	defer mockCtrl.Finish()

	tests := []struct {
		name          string
		setupMocks    func()
		bhCtx         ctx.Context
		expectedCode  int
		expectNextHit bool
	}{
		{
			name: "Success feature flag disabled",
			setupMocks: func() {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureEnvironmentAccessControl).
					Return(appcfg.FeatureFlag{Enabled: false}, nil)
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
		},
		{
			name: "Success All Environments enabled",
			setupMocks: func() {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureEnvironmentAccessControl).
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:          true,
						EnvironmentAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
		},
		{
			name: "Success All Environments disabled and user does have domain in etac list",
			setupMocks: func() {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureEnvironmentAccessControl).
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentAccess{
					{
						EnvironmentID: "12345",
					},
				}, nil)
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:          false,
						EnvironmentAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			expectedCode:  http.StatusOK,
			expectNextHit: true,
		},
		{
			name: "Error getting feature flag",
			setupMocks: func() {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureEnvironmentAccessControl).
					Return(appcfg.FeatureFlag{}, errors.New("db failure"))
			},
			expectedCode:  http.StatusInternalServerError, // whatever HandleDatabaseError writes
			expectNextHit: false,
		},
		{
			name: "Error checking for environments on a user",
			setupMocks: func() {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureEnvironmentAccessControl).
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentAccess{{}}, errors.New("an error"))
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:          false,
						EnvironmentAccessControl: nil,
					},
					Session: model.UserSession{},
				},
			},
			expectedCode:  http.StatusInternalServerError,
			expectNextHit: false,
		},
		{
			name: "Error All Environments disabled and user does not have domain in etac list",
			setupMocks: func() {
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), appcfg.FeatureEnvironmentAccessControl).
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetEnvironmentAccessListForUser(gomock.Any(), gomock.Any()).Return([]model.EnvironmentAccess{{}}, nil)
			},
			bhCtx: ctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner: model.User{
						AllEnvironments:          false,
						EnvironmentAccessControl: nil,
					},
					Session: model.UserSession{},
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

			handler := SupportsETACMiddleware(mockDB)(next)

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
