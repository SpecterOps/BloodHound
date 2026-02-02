package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware(t *testing.T) {

	// Create the Mock Controller, Database, and Authenticator
	var (
		mockCtrl 			= gomock.NewController(t)
		mockDb 				= dbmocks.NewMockDatabase(mockCtrl)
		//mockAuthenticator	= apiMocks.NewMockAuthenticator(mockCtrl)
	)
	defer mockCtrl.Finish()

	// Create the Configuration
	cfg, err := config.NewDefaultConfiguration()
	require.Nil(t, err)

	// Create the Authenticator (Cfg, Db, AuthExtensions)
	authenticator := api.NewAuthenticator(cfg, mockDb, api.NewAuthExtensions(cfg, mockDb))

	// Create the Authorizer for the Router
	authorizer := auth.NewAuthorizer(mockDb)

	// Create the Router
	router := router.NewRouter(cfg, authorizer, fmt.Sprintf(bootstrap.ContentSecurityPolicy, "", ""))

	// Add AuthMiddleware to the Postrouting
	router.UsePostrouting(
		middleware.AuthMiddleware(authenticator),
	)

	// Create the Return Data for the Request using BHESignature
	mockDb.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(true)

	// Create the Request using BHESignature
	var (
		expectedTime = time.Now()
		expectedID   = must.NewUUIDv4()
		request      = must.NewHTTPRequest(http.MethodGet, "http://example.com/", nil)
	)

	request.Header.Set(headers.Authorization.String(), "bhesignature "+expectedID.String())
	request.Header.Set(headers.RequestDate.String(), expectedTime.Format(time.RFC3339Nano))

	// Create Variable to Obtain the Response
	response := httptest.NewRecorder()

	// Send the Request to the Router
	router.Handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Contains(t, response.Body.String(), "use of API keys has been disabled")
}