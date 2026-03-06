// Copyright 2024 Specter Ops, Inc.
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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	apimocks "github.com/specterops/bloodhound/cmd/api/src/api/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_NewRequestSignature(t *testing.T) {
	t.Run("returns error on context timeout", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		goCtx, cancel := context.WithDeadline(context.Background(), time.Now())
		defer cancel()
		time.Sleep(1 * time.Microsecond)
		_, err = api.NewRequestSignature(goCtx, sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("returns error on empty hmac signature", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		goCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		defer cancel()
		_, err = api.NewRequestSignature(goCtx, nil, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "hasher must not be nil")
	})

	t.Run("success", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		goCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		defer cancel()
		signature, err := api.NewRequestSignature(goCtx, sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.Nil(t, err)
		require.NotEmpty(t, signature)
	})
}

func newTestAuthenticator(t *testing.T, ctrl *gomock.Controller) (api.Authenticator, *dbMocks.MockDatabase, *apimocks.MockAuthExtensions) {
	mockDB := dbMocks.NewMockDatabase(ctrl)
	mockAuthExtensions := apimocks.NewMockAuthExtensions(ctrl)

	cfg := config.Configuration{
		WorkDir: t.TempDir(),
	}

	authenticator := api.NewAuthenticator(cfg, mockDB, mockAuthExtensions)
	return authenticator, mockDB, mockAuthExtensions
}

func TestValidateRequestSignature(t *testing.T) {

	enabled, err := types.NewJSONBObject(map[string]any{"enabled": true})
	require.NoError(t, err)

	enableApiKeyParameter := appcfg.Parameter{
		Key:         appcfg.APITokens,
		Name:        "",
		Description: "",
		Value:       enabled,
	}

	t.Run("should return 400 error on missing request date header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, _, _ := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "no request date header")
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should return 400 error on malformed request date header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, _, _ := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), "mwahahahahaha!!!")

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "malformed request date")
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should return 400 error on missing signature header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, _, _ := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "no signature header")
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should return 400 error on malformed signature header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, _ := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		req.Header.Add(headers.Signature.String(), "I'm a bad signature")

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "malformed signature header")
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should return 500 error on failure to retrieve auth token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, _ := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, fmt.Errorf("all your base are belong to us"))

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "internal error")
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("should return 500 error on failure to initialize user auth context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, mockAuthExtensions := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, fmt.Errorf("somebody set up us the bomb"))

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "internal error")
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("should return 403 when user is disabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, mockAuthExtensions := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{
			Owner: model.User{
				IsDisabled: true,
			},
		}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "user disabled")
		require.Equal(t, http.StatusForbidden, status)
	})

	t.Run("should return 401 when Request-Date header time is too skewed from server", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, mockAuthExtensions := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		badRequestDate := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		req.Header.Add(headers.RequestDate.String(), badRequestDate)
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "signature too far behind")
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("should return 200 and stream request body to disk on payloads that exceed ThresholdLargePayload", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		mockAuthExtensions := apimocks.NewMockAuthExtensions(ctrl)

		cfg := config.Configuration{
			WorkDir: t.TempDir(),
		}
		os.Mkdir(cfg.TempDirectory(), 0o755)

		authenticator := api.NewAuthenticator(cfg, mockDB, mockAuthExtensions)

		payload := make([]byte, api.ThresholdLargePayload+1)
		req, err := http.NewRequest(http.MethodPost, "http://teapotsrus.dev", bytes.NewBuffer(payload))
		require.NoError(t, err)

		req.ContentLength = int64(len(payload))
		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{Key: "token"}, nil)
		mockDB.EXPECT().UpdateAuthToken(gomock.Any(), gomock.Any()).Return(nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		tmpFiles, err := os.ReadDir(cfg.TempDirectory())
		assert.NoError(t, err)
		assert.Len(t, slicesext.Filter(tmpFiles, func(file fs.DirEntry) bool {
			return strings.HasPrefix(file.Name(), "bh-request-")
		}), 1)

		// Closing the body should remove the tmp file
		req.Body.Close()
		tmpFiles, err = os.ReadDir(cfg.TempDirectory())
		assert.NoError(t, err)
		assert.Len(t, slicesext.Filter(tmpFiles, func(file fs.DirEntry) bool {
			return strings.HasPrefix(file.Name(), "bh-request-")
		}), 0)
	})

	t.Run("test handling of 'small' payloads within ThresholdLargePayload", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, mockAuthExtensions := newTestAuthenticator(t, ctrl)

		payload := make([]byte, api.ThresholdLargePayload)
		req, err := http.NewRequest(http.MethodPost, "http://teapotsrus.dev", bytes.NewBuffer(payload))
		require.NoError(t, err)
		defer req.Body.Close()

		req.ContentLength = int64(len(payload) - 1)
		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{Key: "token"}, nil)
		mockDB.EXPECT().UpdateAuthToken(gomock.Any(), gomock.Any()).Return(nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		// "small" payloads should not create a tmp file
		tmpFiles, err := os.ReadDir(t.TempDir())
		assert.NoError(t, err)
		assert.Len(t, slicesext.Filter(tmpFiles, func(file fs.DirEntry) bool {
			return strings.HasPrefix(file.Name(), "bh-request-")
		}), 0)
	})

	t.Run("test signature digest mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, mockAuthExtensions := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		datetime := time.Now().Format(time.RFC3339)
		req.Header.Add(headers.RequestDate.String(), datetime)
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "badtoken", datetime, http.MethodGet, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
			Key: "token",
		}, nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "signature digest mismatch")
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("test successful signature validation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator, mockDB, mockAuthExtensions := newTestAuthenticator(t, ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		datetime := time.Now().Format(time.RFC3339)
		req.Header.Add(headers.RequestDate.String(), datetime)
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", datetime, req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)
		mockDB.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
			Key: "token",
		}, nil)
		mockDB.EXPECT().UpdateAuthToken(gomock.Any(), gomock.Any()).Return(nil)
		mockAuthExtensions.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
	})

	t.Run("test bhesignature attempt with disabled api keys", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		authenticator, mockDB, _ := newTestAuthenticator(t, ctrl)

		request, err := http.NewRequest(http.MethodGet, "http://teapotrus.dev", nil)
		require.NoError(t, err)

		datetime := time.Now().Format(time.RFC3339)
		request.Header.Add(headers.RequestDate.String(), datetime)
		signature, err := api.NewRequestSignature(context.Background(), sha256.New, "token", datetime, request.Method, request.RequestURI, nil)
		require.NoError(t, err)
		request.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		disabled, err := types.NewJSONBObject(map[string]any{"enabled": false})
		require.NoError(t, err)

		enableApiKeyParameter = appcfg.Parameter{
			Key:         appcfg.APITokens,
			Name:        "",
			Description: "",
			Value:       disabled,
		}

		mockDB.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.APITokens).Return(enableApiKeyParameter, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, request, time.Now())
		require.ErrorIs(t, api.ErrApiKeysDisabled, err)
		require.Equal(t, http.StatusUnauthorized, status)
	})
}

func TestAuthExtensions(t *testing.T) {
	t.Parallel()

	validSignatureByteString := "D9GUrNzL6b9l4wqHOkLPgEr7VHhZ/LPvvgfsHlUdPETiHw0IkQ2KuMLg5Q+aRclZYUD97PH95XMtfZy0rPBhEQ=="
	cfg := config.Configuration{
		Crypto: config.CryptoConfiguration{
			JWT: config.JWTConfiguration{
				SigningKey: validSignatureByteString,
			},
		},
	}
	validUserID := uuid.Must(uuid.NewV4())
	validClientID := uuid.Must(uuid.NewV4())
	testError := errors.New("user select error was not found due to being a clientId")

	t.Run("test InitContextFromClaims returns default", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		extensions := api.NewAuthExtensions(cfg, mockDB)

		claims := &jwt.RegisteredClaims{}
		authCtx, err := extensions.InitContextFromClaims(context.Background(), claims)
		require.NoError(t, err)
		require.NotNil(t, authCtx)
	})

	t.Run("test InitContextFromToken returns error if UserId is invalid", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		extensions := api.NewAuthExtensions(cfg, mockDB)

		token := model.AuthToken{
			UserID: uuid.NullUUID{
				Valid: false,
			},
		}
		authCtx, err := extensions.InitContextFromToken(context.Background(), token)
		require.NotNil(t, authCtx)
		require.Error(t, err)
		require.Equal(t, err, database.ErrNotFound)
	})

	t.Run("test InitContextFromToken calls db if UserId is valid", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		mockDB.EXPECT().GetUser(gomock.Any(), validUserID).Return(model.User{}, nil)

		extensions := api.NewAuthExtensions(cfg, mockDB)

		token := model.AuthToken{
			UserID: uuid.NullUUID{
				Valid: true,
				UUID:  validUserID,
			},
		}
		authCtx, err := extensions.InitContextFromToken(context.Background(), token)
		require.NoError(t, err)
		require.NotNil(t, authCtx)
		require.NotNil(t, authCtx.Owner)
	})

	t.Run("test InitContextFromToken calls db if UserId is valid, if errors, error returned", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		mockDB.EXPECT().GetUser(gomock.Any(), validClientID).Return(model.User{}, testError)

		extensions := api.NewAuthExtensions(cfg, mockDB)

		token := model.AuthToken{
			UserID: uuid.NullUUID{
				Valid: true,
				UUID:  validClientID,
			},
		}
		authCtx, err := extensions.InitContextFromToken(context.Background(), token)
		require.Error(t, err)
		require.Equal(t, err, testError)
		require.NotNil(t, authCtx)
	})

	t.Run("test ParseClaimsAndVerifySignature returns claims with valid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		extensions := api.NewAuthExtensions(cfg, mockDB)

		signingKey, err := base64.StdEncoding.DecodeString(validSignatureByteString)
		require.NoError(t, err)
		presignedClaims := &jwt.RegisteredClaims{
			Issuer:    "meeeeee",
			Subject:   validUserID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, presignedClaims)
		signedToken, err := token.SignedString(signingKey)
		require.NoError(t, err)
		responseClaims, err := extensions.ParseClaimsAndVerifySignature(context.Background(), signedToken)
		require.NoError(t, err)
		require.Equal(t, responseClaims, presignedClaims)
	})

	t.Run("test ParseClaimsAndVerifySignature returns claims and error with invalid token", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		extensions := api.NewAuthExtensions(cfg, mockDB)

		signingKey, err := base64.StdEncoding.DecodeString(validSignatureByteString)
		require.NoError(t, err)
		presignedClaims := &jwt.RegisteredClaims{
			Issuer:    "meeeeee",
			Subject:   validUserID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute * 5)),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, presignedClaims)
		signedToken, err := token.SignedString(signingKey)
		require.NoError(t, err)
		responseClaims, err := extensions.ParseClaimsAndVerifySignature(context.Background(), signedToken)
		require.Error(t, err)
		require.Equal(t, responseClaims, presignedClaims)
		require.IsType(t, &jwt.ValidationError{}, err)
		require.Equal(t, jwt.ValidationErrorExpired, jwt.ValidationErrorExpired&err.(*jwt.ValidationError).Errors)
	})

	t.Run("test ParseClaimsAndVerifySignature returns invalid Auth with invalid signature", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := dbMocks.NewMockDatabase(ctrl)
		extensions := api.NewAuthExtensions(cfg, mockDB)

		invalidSignatureByteString := "9vnEqkOm1LQP1ntaR0ItGeJcLbZGfem7oLotY4rn61OhXS3SBorlTDYT2CJ6abdWQb7LULJKXLHDrl+6aAKf9Q=="
		signingKey, err := base64.StdEncoding.DecodeString(invalidSignatureByteString)
		require.NoError(t, err)
		presignedClaims := &jwt.RegisteredClaims{
			Issuer:    "meeeeee",
			Subject:   validUserID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, presignedClaims)
		signedToken, err := token.SignedString(signingKey)
		require.NoError(t, err)
		responseClaims, err := extensions.ParseClaimsAndVerifySignature(context.Background(), signedToken)
		require.Error(t, err)
		require.Equal(t, responseClaims, presignedClaims)
		require.Equal(t, api.ErrInvalidAuth, err)
	})
}
