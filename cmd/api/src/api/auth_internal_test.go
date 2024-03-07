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
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/ctx"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
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

func buildAuditLog(testCtx context.Context, user model.User, loginRequest LoginRequest, status string, loginError error) model.AuditLog {
	bhCtx := ctx.Get(testCtx)

	auditLog := model.AuditLog{
		Action:          "LoginAttempt",
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
	expectedAuditLog := buildAuditLog(testCtx, testyUser, loginRequest, string(model.AuditStatusSuccess), nil)

	mockDB.EXPECT().CreateAuditLog(testCtx, expectedAuditLog)
	a.auditLogin(testCtx, commitId, testyUser, loginRequest, string(model.AuditStatusSuccess), nil)
}

func TestAuditLogin_UserNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	a := authenticator{
		db: mockDB,
	}
	testCtx, loginRequest := setupRequest(model.User{})
	expectedAuditLog := buildAuditLog(testCtx, model.User{}, loginRequest, string(model.AuditStatusFailure), ErrInvalidAuth)

	mockDB.EXPECT().CreateAuditLog(testCtx, expectedAuditLog)
	a.auditLogin(testCtx, commitId, model.User{}, loginRequest, string(model.AuditStatusFailure), ErrInvalidAuth)
}
