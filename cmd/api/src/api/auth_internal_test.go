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

package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	cryptoMocks "github.com/specterops/bloodhound/crypto/mocks"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/slicesext"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	commitId    = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
	testyUserId = uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222")
	testyUser   = model.User{
		Unique: model.Unique{
			ID: testyUserId,
		},
		PrincipalName: "testy",
		EmailAddress:  null.StringFrom("test@email.com"),
	}
)

func setupRequest(user model.User) (context.Context, LoginRequest) {
	bhCtx := ctx.Context{
		RequestID: "12345",
		RequestIP: "1.2.3.4",
	}
	testCtx := context.Background()
	testCtx = ctx.Set(testCtx, &bhCtx)

	var loginRequest LoginRequest
	if user.PrincipalName == "" {
		loginRequest.Username = "nonExistentUser"
	} else {
		loginRequest.Username = user.PrincipalName
	}

	return testCtx, loginRequest
}

func buildAuditLog(testCtx context.Context, user model.User, loginRequest LoginRequest, status model.AuditLogEntryStatus, loginError error) model.AuditLog {
	bhCtx := ctx.Get(testCtx)

	auditLog := model.AuditLog{
		Action:          model.AuditLogActionLoginAttempt,
		ActorName:       user.PrincipalName,
		ActorEmail:      user.EmailAddress.ValueOrZero(),
		Fields:          types.JSONUntypedObject{"username": loginRequest.Username},
		RequestID:       bhCtx.RequestID,
		SourceIpAddress: bhCtx.RequestIP,
		Status:          status,
		CommitID:        commitId,
	}

	if user.ID.String() != "00000000-0000-0000-0000-000000000000" {
		auditLog.ActorID = user.ID.String()
	}

	if loginError != nil {
		auditLog.Fields["error"] = loginError
	}

	return auditLog
}

func TestAuditLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	a := authenticator{
		db: mockDB,
	}
	testCtx, loginRequest := setupRequest(testyUser)
	expectedAuditLog := buildAuditLog(testCtx, testyUser, loginRequest, model.AuditLogStatusSuccess, nil)

	mockDB.EXPECT().CreateAuditLog(testCtx, expectedAuditLog)
	a.auditLogin(testCtx, commitId, testyUser, loginRequest, model.AuditLogStatusSuccess, nil)
}

func TestAuditLogin_UserNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	a := authenticator{
		db: mockDB,
	}
	testCtx, loginRequest := setupRequest(model.User{})
	expectedAuditLog := buildAuditLog(testCtx, model.User{}, loginRequest, model.AuditLogStatusFailure, ErrInvalidAuth)

	mockDB.EXPECT().CreateAuditLog(testCtx, expectedAuditLog)
	a.auditLogin(testCtx, commitId, model.User{}, loginRequest, model.AuditLogStatusFailure, ErrInvalidAuth)
}

func TestValidateRequestSignature(t *testing.T) {
	NewTestAuthenticator := func(ctrl *gomock.Controller) authenticator {
		cfg := config.Configuration{
			WorkDir: os.TempDir(),
		}
		os.Mkdir(cfg.TempDirectory(), 0755)
		return authenticator{
			cfg:             cfg,
			db:              dbMocks.NewMockDatabase(ctrl),
			ctxInitializer:  dbMocks.NewMockAuthContextInitializer(ctrl),
			secretDigester:  cryptoMocks.NewMockSecretDigester(ctrl),
			concurrencyLock: make(chan struct{}, 1),
		}
	}

	t.Run("should return 400 error on missing request date header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

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
		authenticator := NewTestAuthenticator(ctrl)

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
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "no signature header")
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should return  400 error on malformed signature header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		req.Header.Add(headers.Signature.String(), "I'm a bad signature")

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "malformed signature header")
		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("should return 500 error on failure to retrieve auth token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := NewRequestSignature(sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, fmt.Errorf("all your base are belong to us"))

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "internal error")
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("should return 500 error on failure to initialize user auth context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := NewRequestSignature(sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, fmt.Errorf("somebody set up us the bomb"))

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "internal error")
		require.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("should return 403 when user is disabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := NewRequestSignature(sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{
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
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		badRequestDate := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		req.Header.Add(headers.RequestDate.String(), badRequestDate)
		signature, err := NewRequestSignature(sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{}, nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "signature too far behind")
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("should return 200 and stream request body to disk on payloads that exceed Threshold50MiB", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		payload := make([]byte, ThresholdLargePayload+1)
		req, err := http.NewRequest(http.MethodPost, "http://teapotsrus.dev", bytes.NewBuffer(payload))
		require.NoError(t, err)

		req.ContentLength = int64(len(payload))
		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := NewRequestSignature(sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{Key: "token"}, nil)
		db.EXPECT().UpdateAuthToken(gomock.Any(), gomock.Any()).Return(nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)

		tmpFiles, err := os.ReadDir(authenticator.cfg.TempDirectory())
		assert.NoError(t, err)
		assert.Len(t, slicesext.Filter[fs.DirEntry](tmpFiles, func(file fs.DirEntry) bool {
			return strings.HasPrefix(file.Name(), "bh-request-")
		}), 1)

		// Closing the body should remove the tmp file
		req.Body.Close()
		tmpFiles, err = os.ReadDir(os.TempDir())
		assert.NoError(t, err)
		assert.Len(t, slicesext.Filter[fs.DirEntry](tmpFiles, func(file fs.DirEntry) bool {
			return strings.HasPrefix(file.Name(), "bh-request-")
		}), 0)
	})

	t.Run("test handling of 'small' payloads within Threshold50MiB", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		payload := make([]byte, ThresholdLargePayload)
		req, err := http.NewRequest(http.MethodPost, "http://teapotsrus.dev", bytes.NewBuffer(payload))
		require.NoError(t, err)
		defer req.Body.Close()

		req.ContentLength = int64(len(payload) - 1)
		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))
		signature, err := NewRequestSignature(sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{Key: "token"}, nil)
		db.EXPECT().UpdateAuthToken(gomock.Any(), gomock.Any()).Return(nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		// "small" payloads should not create a tmp file
		tmpFiles, err := os.ReadDir(os.TempDir())
		assert.NoError(t, err)
		assert.Len(t, slicesext.Filter[fs.DirEntry](tmpFiles, func(file fs.DirEntry) bool {
			return strings.HasPrefix(file.Name(), "bh-request-")
		}), 0)
	})

	t.Run("test signature digest mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		datetime := time.Now().Format(time.RFC3339)
		req.Header.Add(headers.RequestDate.String(), datetime)
		signature, err := NewRequestSignature(sha256.New, "badtoken", datetime, http.MethodGet, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
			Key: "token",
		}, nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.Error(t, err)
		require.ErrorContains(t, err, "signature digest mismatch")
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("test successful signature validation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		authenticator := NewTestAuthenticator(ctrl)

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		datetime := time.Now().Format(time.RFC3339)
		req.Header.Add(headers.RequestDate.String(), datetime)
		signature, err := NewRequestSignature(sha256.New, "token", datetime, req.Method, req.RequestURI, nil)
		require.NoError(t, err)
		req.Header.Add(headers.Signature.String(), base64.StdEncoding.EncodeToString(signature))

		db := authenticator.db.(*dbMocks.MockDatabase)
		db.EXPECT().GetAuthToken(gomock.Any(), gomock.Any()).Return(model.AuthToken{
			Key: "token",
		}, nil)
		db.EXPECT().UpdateAuthToken(gomock.Any(), gomock.Any()).Return(nil)

		ctxInit := authenticator.ctxInitializer.(*dbMocks.MockAuthContextInitializer)
		ctxInit.EXPECT().InitContextFromToken(gomock.Any(), gomock.Any()).Return(auth.Context{}, nil)

		_, status, err := authenticator.ValidateRequestSignature(uuid.UUID{}, req, time.Now())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)
	})
}
